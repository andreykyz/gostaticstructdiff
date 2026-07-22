# Plan 004: Fix Missing `reflect` Import for Map Alias Types

## Problem

When generating diff code for map alias types (e.g., `type MetaString map[string]string`), the generated patch function uses `reflect.DeepEqual` to compare map values. However, the `needsReflect` check in [`generator/generator.go:98-111`](../generator/generator.go:98) only iterates over **struct fields** to determine if the `reflect` package needs to be imported. Standalone map alias types (not referenced as struct fields) are never checked, causing the generated code to reference `reflect.DeepEqual` without importing `reflect`.

### Example

Source: [`examples/models/nested/id.go:8`](../examples/models/nested/id.go:8)
```go
type MetaString map[string]string
```

Generated: [`examples/models/nested/id_diff.go:19`](../examples/models/nested/id_diff.go:19)
```go
if !ok || !reflect.DeepEqual(oldV, v) {
```

But the generated file has no `import "reflect"` statement, causing a compilation error:
```
../../models/nested/id_diff.go:19:14: undefined: reflect
```

## Root Cause

The `needsReflect` detection loop at [`generator/generator.go:98-111`](../generator/generator.go:98) only examines struct fields:

```go
needsReflect := false
for _, s := range structs {
    for _, f := range s.Fields {
        typeInfo := types.Classify(f.TypeExpr, knownStructs, typeDefs)
        if typeInfo.Category == types.CategorySlice || typeInfo.Category == types.CategoryMap || typeInfo.Category == types.CategoryMapAlias || typeInfo.Category == types.CategoryUnknown {
            needsReflect = true
            break
        }
    }
    if needsReflect {
        break
    }
}
```

Map alias types are processed later at line 167 via `collectReferencedMapAliases()`, which scans both struct fields and `typeDefs` for map alias types. But the `needsReflect` check happens before this collection, and it only checks struct fields — not the `typeDefs` map for standalone map alias types.

## Solution

Extend the `needsReflect` check to also scan `typeDefs` for map alias types (types whose underlying AST expression is `*ast.MapType`). This mirrors the logic already present in `collectReferencedMapAliases()` at lines 448-456.

### Changes

**File: [`generator/generator.go`](../generator/generator.go)**

After the existing struct field loop (lines 100-111), add a second loop that checks `typeDefs` for map alias types:

```go
// Also check typeDefs for map alias types (standalone types like type MetaString map[string]string)
if !needsReflect {
    for _, underlying := range typeDefs {
        if _, ok := underlying.(*ast.MapType); ok {
            needsReflect = true
            break
        }
    }
}
```

This ensures that any map alias type defined in the package will trigger the `reflect` import, regardless of whether it's referenced as a struct field.

## Testing

1. Build the tool: `make build`
2. Regenerate the nested example: `./gostaticstructdiff -input examples/models/nested/id.go -output examples/models/nested/id_diff.go`
3. Verify `examples/models/nested/id_diff.go` now includes `import "reflect"`
4. Run `make example` to confirm the full example pipeline works
5. Run `make test` to ensure no regressions