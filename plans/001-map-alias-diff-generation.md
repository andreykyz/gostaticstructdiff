# Plan: Generate Diff Types for Map Alias Types

## Problem

The generator does not produce diff types for **map alias types** — types defined as `type X map[K]V` (e.g., `type MetaMeta map[nested.ID]Metadata`).

### Root Cause

In [`types/types.go:111-126`](types/types.go:111-126), the `Classify` function handles `*ast.SelectorExpr` (qualified identifiers like `models.MetaMeta`) by checking `knownStructs` first. If the type is not in `knownStructs`, it **assumes it's a struct** (line 123-126). This is incorrect for map aliases.

When `MetaMeta` (a map alias) is used as a field type in [`examples/complex.go:20`](examples/complex.go:20), the classifier returns `CategoryStruct`, causing the generator to:
- Generate `MetaMeta *struct { Value models.MetaMetaDiff }` in the diff struct
- Generate `models.MetaMetaPatch(...)` and `models.ApplyMetaMetaDiff(...)` calls
- But **never generate the `MetaMetaDiff` type itself**, since the parser only generates diffs for actual struct types

### Current Generated Code (broken)

```go
// In complex_diff.go:
MetaMeta *struct {
    Value models.MetaMetaDiff  // references a type that doesn't exist
}
```

## Solution Overview

The fix requires changes in three areas:

1. **`types/types.go`** — Add a new `CategoryMapAlias` category and detect map alias types by looking up `typeDefs` for `*ast.SelectorExpr` types
2. **`generator/generator.go`** — Handle `CategoryMapAlias` in `convertToTemplateData` to populate map-related fields (KeyType, ValueType, ValueIsStruct, etc.)
3. **`generator/templates/`** — Add `mapAlias` handling in `struct_diff.go.tmpl` and `patch_func.go.tmpl` templates

## Detailed Changes

### 1. `types/types.go` — Add `CategoryMapAlias`

**New category constant:**
```go
const (
    CategoryBasic Category = iota
    CategoryPointer
    CategorySlice
    CategoryMap
    CategoryStruct
    CategoryMapAlias   // NEW: type X map[K]V
    CategoryUnknown
)
```

**Modified `Classify` for `*ast.SelectorExpr` (lines 111-126):**

When a `*ast.SelectorExpr` is encountered:
1. Check `knownStructs` — if found, return `CategoryStruct` (existing behavior)
2. Look up the type in `typeDefs` using the unqualified name (last segment after dot) — **NEW**
3. If the underlying type in `typeDefs` is a `*ast.MapType`, return `CategoryMapAlias` with the resolved key/value types
4. Otherwise, fall through to existing behavior (assume struct)

**Modified `Classify` for `*ast.Ident` (lines 57-84):**

When an `*ast.Ident` is encountered and it's not a basic type or known struct:
1. Look up in `typeDefs` — if the underlying type is a `*ast.MapType`, return `CategoryMapAlias` with resolved key/value types
2. Otherwise, existing behavior (recursively classify or assume basic)

**New helper method on `TypeInfo`:**
```go
func (ti *TypeInfo) IsMapAlias() bool {
    return ti.Category == CategoryMapAlias
}
```

### 2. `generator/generator.go` — Handle `CategoryMapAlias`

**In `convertToTemplateData` function (after line 191):**

Add handling for `CategoryMapAlias` similar to `CategoryMap`:

```go
// For map alias types (type X map[K]V)
if typeInfo.Category == types.CategoryMapAlias && typeInfo.Key != nil && typeInfo.Value != nil {
    category = "mapAlias"
    fieldData.KeyType = typeInfo.Key.TypeString
    fieldData.ValueType = typeInfo.Value.TypeString
    if typeInfo.Value.Category == types.CategoryStruct {
        fieldData.ValueIsStruct = true
        fieldData.ValueDiffType = typeInfo.Value.TypeString + "Diff"
        pkg, name := splitType(typeInfo.Value.TypeString)
        fieldData.ValueTypePackage = pkg
        fieldData.ValueTypeName = name
        if pkg != "" {
            fieldData.ValueDiffFunc = pkg + ".Apply" + name + "Diff"
        } else {
            fieldData.ValueDiffFunc = "Apply" + name + "Diff"
        }
    }
}
```

**In `Generate` function (line 52):**

The `Generate` function currently only iterates over `structs` (from parser). For map aliases, we need to also generate diff types for them. However, map aliases are not structs and won't be returned by the parser.

**Approach:** The parser already collects all `typeDefs` (including map aliases) in [`parser/parser.go:82`](parser/parser.go:82). We need to:

1. Pass `typeDefs` through to identify which types are map aliases
2. Generate additional diff structs for map alias types that are referenced by the structs being processed

**New function to collect referenced map aliases:**
```go
func collectReferencedMapAliases(structs []parser.StructInfo, typeDefs map[string]ast.Expr, knownStructs map[string]bool) []parser.StructInfo {
    // For each struct field, check if its type resolves to a map alias via typeDefs
    // If so, create a synthetic StructInfo for the map alias
}
```

This function would:
- Walk all fields of all structs
- For each field, resolve its type through `typeDefs`
- If the underlying type is a `*ast.MapType`, create a `StructInfo` with the map's key/value as "fields" (using a special marker)
- Return the list of map alias struct infos to also generate

**Alternative simpler approach:** Instead of creating synthetic structs, generate the map alias diff types directly in the `Generate` function by iterating over `typeDefs` and finding map types.

### 3. Template Changes

#### `generator/templates/struct_diff.go.tmpl`

Add a new case for `mapAlias` category:

```go
{{- else if eq .Category "mapAlias"}}
{{- if .ValueIsStruct}}
Add map[{{.KeyType}}]{{.ValueType}}
Modify map[{{.KeyType}}]{{.ValueDiffType}}
Del map[{{.KeyType}}]struct{}
{{- else}}
Add map[{{.KeyType}}]{{.ValueType}}
Del map[{{.KeyType}}]struct{}
{{- end}}
```

#### `generator/templates/patch_func.go.tmpl`

Add a new case for `mapAlias` category in both the Patch function and ApplyDiff function, mirroring the existing `map` category handling.

### 4. Parser Changes (if needed)

The parser already collects `typeDefs` in [`parser/parser.go:82`](parser/parser.go:82) and returns them from `ParseFileWithOptions`. No changes needed to the parser itself.

However, we need to ensure that when processing `*ast.SelectorExpr` types like `models.MetaMeta`, the parser's `typeDefs` map includes the unqualified name (e.g., `MetaMeta`) so the classifier can look it up. Since `typeDefs` is built from the same file, this should work for same-package types. For cross-package types, we need a different approach.

**Cross-package map aliases:** When a field has type `models.MetaMeta` and `MetaMeta` is defined in another package, the `typeDefs` from the current file won't contain it. The classifier currently assumes it's a struct. To handle this properly, we'd need to parse the imported package's type definitions too, which is a larger change.

**Short-term fix:** For the immediate issue (same-package map aliases or when the map alias is in the same file being processed), the `typeDefs` lookup will work. For cross-package map aliases, we can fall back to the existing behavior (assume struct) and document the limitation.

### 5. Generate Map Alias Diff Types

The key architectural decision is **how to generate the diff types for map aliases**.

**Option A: Generate as part of the main struct's output**

When processing a struct that references a map alias, also generate the map alias diff type in the same output file. This is simpler but could lead to duplication if multiple structs reference the same map alias.

**Option B: Generate as separate struct entries**

Create synthetic `StructInfo` entries for map aliases and add them to the struct list. This reuses the existing generation pipeline.

**Recommended: Option B** — It's cleaner and reuses existing code. The synthetic struct would have:
- `Name`: The map alias name (e.g., `MetaMeta`)
- `Fields`: Two synthetic fields representing the map's key and value types

But this is awkward because the existing templates expect struct fields with names, not map key/value pairs.

**Recommended: Option A with deduplication** — Track which map aliases have been generated and generate them once.

## Implementation Steps

### Step 1: Add `CategoryMapAlias` to `types/types.go`

- Add the new category constant
- Modify `Classify` for `*ast.Ident` to check `typeDefs` for map types
- Modify `Classify` for `*ast.SelectorExpr` to check `typeDefs` for map types
- Update `String()` method

### Step 2: Update `generator/generator.go`

- Add `CategoryMapAlias` handling in `convertToTemplateData`
- Add logic to collect and generate diff types for referenced map aliases
- Track generated map aliases to avoid duplicates

### Step 3: Update templates

- Add `mapAlias` case to `struct_diff.go.tmpl`
- Add `mapAlias` case to `patch_func.go.tmpl` (both Patch and ApplyDiff functions)

### Step 4: Regenerate examples

- Run the generator on `examples/models/metadata.go` to produce `MetaMetaDiff`
- Run the generator on `examples/complex.go` to update `complex_diff.go`

### Step 5: Run tests

- Run `go test ./...` to verify everything compiles and passes

## Expected Generated Output

For `type MetaMeta map[nested.ID]Metadata`, the generator should produce:

```go
// MetaMetaDiff represents the diff of a MetaMeta map.
type MetaMetaDiff struct {
    Add    map[nested.ID]Metadata
    Modify map[nested.ID]MetadataDiff
    Del    map[nested.ID]struct{}
}

func MetaMetaPatch(original, new MetaMeta) MetaMetaDiff {
    var diff MetaMetaDiff
    diff.Add = make(map[nested.ID]Metadata)
    diff.Modify = make(map[nested.ID]MetadataDiff)
    diff.Del = make(map[nested.ID]struct{})
    // Added or modified keys
    for k, v := range new {
        oldV, ok := original[k]
        if !ok {
            diff.Add[k] = v
        } else if !reflect.DeepEqual(oldV, v) {
            nestedDiff := MetadataPatch(oldV, v)
            diff.Modify[k] = nestedDiff
        }
    }
    // Deleted keys
    for k := range original {
        if _, ok := new[k]; !ok {
            diff.Del[k] = struct{}{}
        }
    }
    if len(diff.Add) == 0 && len(diff.Modify) == 0 && len(diff.Del) == 0 {
        diff = MetaMetaDiff{}
    }
    return diff
}

func ApplyMetaMetaDiff(original MetaMeta, diff MetaMetaDiff) MetaMeta {
    if original == nil {
        original = make(MetaMeta)
    }
    for k, v := range diff.Add {
        original[k] = v
    }
    for k, v := range diff.Modify {
        existing, ok := original[k]
        if !ok {
            existing = Metadata{}
        }
        original[k] = ApplyMetadataDiff(existing, v)
    }
    for k := range diff.Del {
        delete(original, k)
    }
    return original
}
```

## Files to Modify

| File | Change |
|------|--------|
| `types/types.go` | Add `CategoryMapAlias`, update `Classify` for `*ast.Ident` and `*ast.SelectorExpr` |
| `generator/generator.go` | Handle `CategoryMapAlias` in `convertToTemplateData`, add map alias diff generation |
| `generator/templates/struct_diff.go.tmpl` | Add `mapAlias` case |
| `generator/templates/patch_func.go.tmpl` | Add `mapAlias` case for Patch and ApplyDiff |
| `examples/models/metadata_diff.go` | Regenerated with `MetaMetaDiff` |
| `examples/complex_diff.go` | Regenerated (may change if MetaMetaDiff is now in same package) |

## Test Plan

1. **Unit test in `types/types_test.go`**: Test that `Classify` correctly identifies map alias types
2. **Unit test in `generator/generator_test.go`**: Test that `Generate` produces correct output for structs with map alias fields
3. **Integration test**: Run `make generate_example` and verify the output compiles
4. **Run `go test ./...`**: Ensure all existing tests still pass