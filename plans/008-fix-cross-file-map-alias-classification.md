# Plan: Fix misclassification of same-package cross-file named types (e.g. map aliases)

## Problem

Generated code treats the `VMLocks` field of `Cluster` as a **basic** type instead of a
**map alias**.

Broken output in `fenix-core/internal/model/state/cluster_diff_cluster.go:104-107`:

```go
if original.VMLocks != new.VMLocks {
    diff.VMLocks = &struct{ Value VMLocks }{}
    diff.VMLocks.Value = new.VMLocks
}
```

and in `ApplyClusterDiffCluster` at `cluster_diff_cluster.go:141-143`:

```go
if diff.VMLocks != nil {
    result.VMLocks = diff.VMLocks.Value
}
```

Expected output (map alias pattern, like `PlatformExtra` in
`node_diff_cluster.go:243-245`):

```go
VMLocksNestedDiff := VMLocksPatchCluster(original.VMLocks, new.VMLocks)
diff.VMLocks = &struct{ Value VMLocksDiffCluster }{}
diff.VMLocks.Value = VMLocksNestedDiff
```

and in Apply:

```go
if diff.VMLocks != nil {
    result.VMLocks = ApplyVMLocksDiffCluster(original.VMLocks, diff.VMLocks.Value)
}
```

The required symbols (`VMLocksDiffCluster`, `VMLocksPatchCluster`,
`ApplyVMLocksDiffCluster`) already exist in `vm_locks_diff_cluster.go` — they are just never
referenced by `cluster_diff_cluster.go`.

## Root Cause

`gostaticstructdiff` is invoked **per file** (see `fenix-core/makefile:367-369`):
`cluster.go` and `vm_locks.go` are separate inputs that produce separate outputs.

The parser builds `typeDefs` only from the **single input file**
([`parser/parser.go:91`](../parser/parser.go)):

```go
typeDefs[typeSpec.Name.Name] = typeSpec.Type
```

- `VMLocks map[vmstate.ID]time.Time` is defined in `vm_locks.go` (a sibling file), so when
  generating from `cluster.go` it is **not** in `typeDefs`.
- [`types.Classify`](../types/types.go:76) looks up the `*ast.Ident` `VMLocks` in `typeDefs`,
  does not find it, and falls back to `CategoryBasic`
  ([`types/types.go:94`](../types/types.go:94)), producing the direct `!=` comparison.

Contrast: `PlatformExtra map[string]string` is defined **in the same file** as `Node`
(`node.go:98`), so it is in `typeDefs` and classified as `CategoryMapAlias`, which is why
`node_diff_cluster.go` is correct.

The standalone map-alias diff types for `VMLocks` are generated separately during the
`vm_locks.go` invocation (`collectReferencedMapAliases` second loop over local `typeDefs`).
The problem is purely that the **classifier** does not see the sibling-file type.

## Solution Overview

Resolve same-package type definitions from **sibling `.go` files** and use them for
**classification only** (so a field like `VMLocks` is recognized as a map alias). Keep
**standalone map-alias diff generation bound to the input file's local types** so the same map
alias diff is not re-emitted into every referencing file (which would cause redeclaration
errors with the existing per-file workflow).

This also fixes the general case: any unqualified reference to a struct / wrapped primitive /
map alias defined in a sibling file of the same package.

## Detailed Changes

### 1. `parser/parser.go` — add `ParsePackageTypeDefs`

```go
// ParsePackageTypeDefs parses sibling .go files in the same directory as inputFile and
// returns their type definitions (keyed by type name), for use in classification.
// Excludes *_test.go, *_diff*.go, and files whose package clause differs from the
// package of inputFile.
func ParsePackageTypeDefs(inputFile string, verbose bool) (map[string]ast.Expr, error)
```

Implementation notes:
- Determine `dir := filepath.Dir(inputFile)` and read `inputFile`'s package name from its AST.
- `os.ReadDir(dir)`, keep names ending `.go`.
- Skip `_test.go`, `_diff.go`, and `_diff_*.go` (generated diff artifacts).
- Parse each candidate with `go/parser` and skip files whose `node.Name.Name != pkgName`.
- Collect every `*ast.TypeSpec` into the returned map (mirroring `ParseFileWithOptions`).
- On any single-file parse error, log when `verbose` and continue (do not fail the whole run).

### 2. `generator/generator.go` — classification vs. generation type sets

- Change `Generate` signature to accept the input files, e.g.:

```go
func Generate(structs []parser.StructInfo, packageName string, imports []string,
    version string, typeDefs map[string]ast.Expr, inputFiles []string,
    verbose bool, suffix string) (string, error)
```

- After the existing import resolution
  ([`generator/generator.go:116-121`](../generator/generator.go:116)), build a separate
  classification set:

```go
classificationTypeDefs := copyOf(typeDefs) // local + import-resolved
for _, dir := range uniqueDirs(inputFiles) {
    siblings, err := parser.ParsePackageTypeDefs(fileOfDir, verbose)
    if err == nil {
        for name, expr := range siblings {
            if _, ok := classificationTypeDefs[name]; !ok {
                classificationTypeDefs[name] = expr // local defs win over siblings
            }
        }
    }
}
```

- Use `classificationTypeDefs` in the two `types.Classify` calls that decide **how a field is
  rendered**: the `needsReflect` scan
  ([`generator/generator.go:127`](../generator/generator.go:127)) and
  `convertToTemplateData`
  ([`generator/generator.go:350`](../generator/generator.go:350)).
- **Leave** `collectReferencedMapAliases`
  ([`generator/generator.go:449`](../generator/generator.go:449)) on the original `typeDefs`
  (local + import-resolved, no siblings). This keeps standalone map-alias diff generation tied
  to types actually defined in the input file and prevents re-emitting `VMLocksDiffCluster`
  into `cluster_diff_cluster.go`.

### 3. `cmd/gostaticstructdiff/main.go` — pass `inputFiles` through

In `processFiles`, forward the already-known `inputFiles` slice to `generator.Generate`
(after `allTypeDefs`).

### 4. Tests

- Update every `generator.Generate(...)` call in `generator/generator_test.go` for the new
  parameter (pass `nil`).
- Add `parser` test: build a temp dir with two files in the same package, one defining a map
  alias, and assert `ParsePackageTypeDefs` returns it while excluding `_diff.go` / `_test.go`.
- Add a `generator` test: two synthetic `StructInfo`s where the struct references a map alias
  name present only in the sibling `typeDefs` (simulate by passing `inputFiles` pointing at a
  fixture dir). Assert the generated patch uses the `XxxPatch...(...)` map-alias pattern
  instead of `original.X != new.X`.

### 5. Regenerate fenix-core and verify

- Rebuild the tool and run `make gen-diff` inside `fenix-core`.
- Verify `cluster_diff_cluster.go` now contains `VMLocksPatchCluster(...)` /
  `ApplyVMLocksDiffCluster(...)` and no `original.VMLocks != new.VMLocks`.
- Confirm no duplicate `VMLocksDiffCluster` type is emitted into `cluster_diff_cluster.go`.
- `go build ./...` in `fenix-core` to confirm compilation.

### 6. Note on `reflect` import

`needsReflect` will now count `VMLocks` as a map alias and conservatively add `reflect`.
In `cluster_diff_cluster.go` `reflect` is already used (slice/map fields), so no unused-import
issue there; `goimports -w` (already run by `make gen-diff`) strips any genuinely unused
import. No change to `needsReflect` logic is required.

## Files touched

- `parser/parser.go` — new `ParsePackageTypeDefs`
- `parser/parser_test.go` — new test
- `generator/generator.go` — `Generate` signature + `classificationTypeDefs`
- `generator/generator_test.go` — update call sites + new test
- `cmd/gostaticstructdiff/main.go` — pass `inputFiles`
- `fenix-core/.../*_diff_*.go` — regenerated via `make gen-diff`

## Out of scope

- Changing the per-file invocation model in `fenix-core`'s `makefile`.
- Cross-package map-alias standalone diff generation (unchanged behavior).