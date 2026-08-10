# Plan: Fix time.Time Type Regression in Code Generation

## Problem

Field `TS time.Time` in [`examples/complex.go:41`](examples/complex.go:41) generates
incorrect diff code that references non-existent types and functions.

Broken output in [`examples/complex_diff.go:246-249`](examples/complex_diff.go:246):

```go
TSNestedDiff := time.TimePatch(original.TS, new.TS)   // <-- time.TimePatch does not exist
diff.TS = &struct { Value time.TimeDiff }{}            // <-- time.TimeDiff does not exist
diff.TS.Value = TSNestedDiff
```

And in the Apply function at [`examples/complex_diff.go:354-355`](examples/complex_diff.go:354):

```go
if diff.TS != nil{
    result.TS = time.ApplyTimeDiff(original.TS, diff.TS.Value)   // <-- time.ApplyTimeDiff does not exist
}
```

This regression was introduced by commit `c544ae8` (plan `006-add-suffix-cli-option.md`).

## Root Cause Analysis

The bug is in the `*ast.SelectorExpr` branch of [`Classify()`](types/types.go:124).

For a qualified type like `time.Time`:

1. [`Classify()`](types/types.go:128) checks `knownStructs["time.Time"]` → not found (Time is not a user struct).
2. [`Classify()`](types/types.go:135) looks up `typeDefs["Time"]` → not found, because
   `resolveImportTypeDefs()` at [`generator/generator.go:243-247`](generator/generator.go:243)
   **skips all standard library imports** (it only processes imports containing a `.`).
   Therefore `time` is never resolved and `time.Time` is never added to `typeDefs`.
3. Falls through to the fallback at [`types/types.go:161-166`](types/types.go:161), which
   **assumes any unresolved qualified identifier is a struct**:

```go
// Assume it's a struct (most likely scenario for imported types)
// This enables nested struct recursion for cross-package structs.
return &TypeInfo{
    Category:   CategoryStruct,
    TypeString: typeStr,
}
```

Because `time.Time` is classified as `CategoryStruct`, the generator emits a "Nested struct
diff" in [`patch_func.go.tmpl:113-118`](generator/templates/patch_func.go.tmpl) producing
`time.TimePatch(...)` and `time.TimeDiff` — neither of which the tool ever generates (the
tool cannot generate diff code for the external `time` package).

### Why this is a regression vs. earlier behavior

Earlier, unresolved external types (including `time.Time`) were mapped to the `CategoryUnknown`
path and generated using `reflect.DeepEqual` as a safe fallback (see the "Treat unknown types
as slice" handling in [`generator.go:366-369`](generator/generator.go:366)). The aggressive
"assume struct" fallback in the `SelectorExpr` branch now overrides that safe behavior for any
external type the tool cannot introspect, producing broken code for `time.Time` and any other
external non-struct type.

## Design Decision: Known Standard-Library Types Should Be Treated as Basic Comparable

`time.Time` is a value type that supports `==`/`!=` comparison and is directly assignable. For
diff generation it should be treated like a **basic** type: store the new value and compare
with `!=`, exactly like `int`/`string`. This keeps generated code correct without requiring
the tool to understand the internals of `time.Time`.

Rather than special-casing only `time.Time`, add a small **denylist/safelist of standard-library
value types** that are safe to compare with `==`. When a `*ast.SelectorExpr` resolves to a
non-generatable external package (specifically a standard-library type the tool cannot
introspect), classify it as `CategoryBasic` instead of assuming it is a struct.

## Solution Overview

Modify the fallback behavior in `Classify()`'s `*ast.SelectorExpr` branch so that known
standard-library value types are classified as `CategoryBasic` (comparable), rather than
assuming every unresolved qualified identifier is a struct.

## Detailed Changes

### 1. `types/types.go` — classify known standard-library value types as basic

**Location**: [`Classify()`](types/types.go:124), `*ast.SelectorExpr` branch, lines 161-166.

Before the "assume struct" fallback, add a check: if the selector's package (`t.X` name) is a
standard-library package and the type is in a known set of comparable value types, return
`CategoryBasic` with the qualified `TypeString` (e.g. `time.Time`).

```go
case *ast.SelectorExpr:
    // Qualified identifier (e.g., "models.User")
    typeStr := exprToString(expr)
    // ... existing knownStructs and typeDefs checks ...

    // Standard-library value types that are safe to compare with == / !=
    // and can be stored directly. The tool cannot generate diff code for
    // external packages, so treat these as basic comparable types.
    if isStdLibComparableType(typeStr) {
        return &TypeInfo{
            Category:   CategoryBasic,
            TypeString: typeStr,
        }
    }

    // Assume it's a struct (most likely scenario for imported types)
    return &TypeInfo{
        Category:   CategoryStruct,
        TypeString: typeStr,
    }
```

Add a helper function:

```go
// isStdLibComparableType returns true for standard-library value types that
// are safe to compare with == / != and store directly (no generated diff needed).
func isStdLibComparableType(typeStr string) bool {
    switch typeStr {
    case "time.Time", "time.Duration":
        return true
    // Add other comparable standard-library value types here as needed.
    }
    return false
}
```

> Rationale for `time.Duration`: it is a named `int64` and comparable, so it also fits the
> basic strategy. Other standard-library types (e.g. `url.URL`, `net.IP`) are either
> non-comparable or semantically richer and are intentionally left out; they will continue to
> follow the existing (safe) `reflect.DeepEqual` path if desired, or can be added later.

### 2. Ensure unknown external types still fall back safely (defensive)

**Location**: [`generator.go:366-369`](generator/generator.go:366) and
[`generator.go:128`](generator/generator.go:128).

The `CategoryUnknown → slice` fallback and the `needsReflect` detection already handle
`CategoryUnknown`. No change is required for `time.Time` since it will now be `CategoryBasic`.
However, confirm the `needsReflect` check at [`generator.go:128`](generator/generator.go:128)
does **not** include `CategoryStruct` — it does not, so a `CategoryBasic` classification will
correctly avoid adding an unnecessary `reflect` import.

### 3. Tests

**`types/types_test.go`** — add:

```go
func TestClassify_TimeTime(t *testing.T) {
    expr := parseExpr(t, "time.Time")
    info := Classify(expr, nil, nil)
    if info.Category != CategoryBasic {
        t.Errorf("Category = %v, want CategoryBasic", info.Category)
    }
    if info.TypeString != "time.Time" {
        t.Errorf("TypeString = %q, want time.Time", info.TypeString)
    }
}
```

Also add a test that an unknown external package type (e.g. `pkg.UnknownThing`) still falls
back to `CategoryStruct` (preserving existing behavior for real cross-package structs).

### 4. Regenerate example

Run `make generate_example` (or the equivalent `-input examples/complex.go` invocation) and
verify [`examples/complex_diff.go`](examples/complex_diff.go) now contains correct basic
handling for `TS`:

```go
TS *struct {
    Value time.Time
}
```

with a patch body like:

```go
if original.TS != new.TS {
    diff.TS = &struct { Value time.Time }{}
    diff.TS.Value = new.TS
}
```

and apply body like:

```go
if diff.TS != nil{
    result.TS = diff.TS.Value
}
```

### 5. Verification

- `make test` — all unit tests pass.
- `make example` — examples compile and run correctly.
- `go vet ./...` and `go build ./...` — no compile errors in generated code.
- Manually inspect [`examples/complex_diff.go`](examples/complex_diff.go) to confirm no
  `time.TimeDiff`, `time.TimePatch`, or `time.ApplyTimeDiff` references remain.

## Files to Modify

| File | Change |
|------|--------|
| [`types/types.go`](types/types.go:124) | Add `isStdLibComparableType` helper and use it in the `*ast.SelectorExpr` fallback |
| [`types/types_test.go`](types/types_test.go) | Add `time.Time` classification test + external-type fallback test |
| [`examples/complex_diff.go`](examples/complex_diff.go) | Regenerate to reflect corrected `TS` handling |

## Backward Compatibility

- User structs in `knownStructs` are checked **before** the std-lib check, so behavior for
  real cross-package structs (e.g. `models.User`) is unchanged.
- Only known standard-library value types are added to the basic path; all other unresolved
  qualified identifiers retain the existing "assume struct" fallback.
- Default output (no suffix) is unaffected.