# Plan: Add `IsEmpty` Method to Generated Diff Structs

## Objective

Add an `IsEmpty() bool` method to every generated Diff struct (both struct diffs and map alias diffs) that returns `true` when all fields of the struct are `nil`.

## Background

The generated `*Diff` structs contain fields that are either:
- **Pointer fields** (`*struct{...}`) — for basic, pointer, slice, struct, pointer-to-struct, and anonymous struct fields
- **Map alias fields** (direct map types like `Add map[...]...`, `Del map[...]struct{}`) — for map alias diffs

An `IsEmpty()` method allows consumers to quickly check whether a diff contains any changes without manually checking each field.

## Generated Code Pattern

### For struct diffs (e.g., `UserDiff`, `ComplexStructDiff`):

```go
// UserDiff example
type UserDiff struct {
    ID       *struct{ Value int }
    Username *struct{ Value string }
    Email    *struct{ Value string }
    Active   *struct{ Value bool }
}

// IsEmpty returns true if no fields have been changed.
func (d UserDiff) IsEmpty() bool {
    return d.ID == nil &&
        d.Username == nil &&
        d.Email == nil &&
        d.Active == nil
}
```

### For map alias diffs (e.g., `MetaStringDiff`, `MetaMetaDiff`):

```go
// MetaStringDiff example
type MetaStringDiff struct {
    Add map[string]string
    Del map[string]struct{}
}

// IsEmpty returns true if no changes have been made.
func (d MetaStringDiff) IsEmpty() bool {
    return len(d.Add) == 0 && len(d.Del) == 0
}
```

## Implementation Steps

### Step 1: Update `struct_diff.go.tmpl`

Add an `IsEmpty()` method after the struct definition. The method checks all fields for `nil`:

```go
// IsEmpty returns true if no fields have been changed.
func (d {{.Name}}Diff) IsEmpty() bool {
    return {{- range $i, $f := .Fields}}{{if $i}} &&
        {{end}}d.{{.Name}} == nil{{- end}}
}
```

### Step 2: Update `map_alias_diff.go.tmpl`

Add an `IsEmpty()` method after the map alias struct definition. The method checks that all map fields have zero length:

```go
// IsEmpty returns true if no changes have been made.
func (d {{.Name}}Diff) IsEmpty() bool {
    return {{- if .ValueIsStruct}}
    len(d.Add) == 0 && len(d.Modify) == 0 && len(d.Del) == 0{{- else}}
    len(d.Add) == 0 && len(d.Del) == 0{{- end}}
}
```

### Step 3: Update `generator_test.go`

Add test assertions that verify the generated code contains `IsEmpty` methods for:
- Simple struct diffs (check `func (d UserDiff) IsEmpty() bool`)
- Map alias diffs (check `func (d MetaStringDiff) IsEmpty() bool`)
- Complex struct diffs with all field types

### Step 4: Regenerate example diff files

Run `make generate_example` to regenerate all example `_diff.go` files with the new `IsEmpty()` method.

### Step 5: Verify compilation

Run `make test` and `make example` to ensure all generated code compiles and works correctly.

## Files to Modify

| File | Change |
|------|--------|
| `generator/templates/struct_diff.go.tmpl` | Add `IsEmpty()` method template after struct definition |
| `generator/templates/map_alias_diff.go.tmpl` | Add `IsEmpty()` method template after map alias struct definition |
| `generator/generator_test.go` | Add test assertions for `IsEmpty()` in generated code |

## Files to Regenerate (via `make generate_example`)

- `examples/complex_diff.go`
- `examples/models/user_diff.go`
- `examples/models/metadata_diff.go`
- `examples/models/nested/id_diff.go`

## Edge Cases Considered

1. **Empty struct (no fields)**: `IsEmpty()` would return `true` (vacuous truth) — handled naturally by the template.
2. **Map alias diffs**: Use `len()` checks instead of `nil` checks since map fields are not pointers.
3. **Struct with no tagged fields**: No diff struct is generated, so no `IsEmpty()` is needed.
4. **Single-field struct**: Template handles single field correctly (no `&&` prefix issues).