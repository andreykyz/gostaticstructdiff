# Plan: Fix Alias Type Detection for Wrapped Primitives

## Problem Statement

When a struct field uses a type alias from another package that wraps a basic type (e.g., `type GGID string` in package `nested`), the generator incorrectly treats it as a struct type and generates code that references non-existent `Diff` types and `Patch` functions.

### Current (Broken) Generated Output

In [`examples/complex_diff.go`](examples/complex_diff.go:39-41):
```go
GGID *struct {
    Value nested.GGIDDiff  // <-- WRONG: tries to use nested.GGIDDiff which doesn't exist
}
```

Generated patch code at [`examples/complex_diff.go`](examples/complex_diff.go:125-127):
```go
GGIDNestedDiff := nested.GGIDPatch(original.GGID, new.GGID)  // <-- WRONG: calls non-existent nested.GGIDPatch
diff.GGID = &struct { Value nested.GGIDDiff }{}
diff.GGID.Value = GGIDNestedDiff
```

### Expected Output

```go
GGID *struct {
    Value nested.GGID  // <-- CORRECT: uses the original type directly
}
```

Generated patch code should be:
```go
if original.GGID != new.GGID {
    diff.GGID = &struct { Value nested.GGID }{}
    diff.GGID.Value = new.GGID
}
```

## Root Cause Analysis

The bug is in [`types/types.go`](types/types.go:124-157), specifically the `Classify()` function's handling of `*ast.SelectorExpr` (qualified identifiers like `nested.GGID`).

### Current Flow for `nested.GGID`:

1. `Classify()` receives `*ast.SelectorExpr{X: "nested", Sel: "GGID"}`
2. Line 128: Checks `knownStructs["nested.GGID"]` → not found (GGID is not a struct)
3. Line 135-150: Looks up `typeDefs["GGID"]`:
   - **If `GGID` is in `typeDefs`** (as `*ast.Ident{Name: "string"}`):
     - Recursive `Classify()` returns `{Category: CategoryBasic, TypeString: "string"}`
     - Since `inner.Category != CategoryMap`, falls to line 149: `return inner`
     - **BUG**: Returns `TypeString: "string"` instead of `"nested.GGID"`
   - **If `GGID` is NOT in `typeDefs`** (e.g., before the main.go fix):
     - Falls through to line 152-157: returns `{Category: CategoryStruct, TypeString: "nested.GGID"}`
     - **BUG**: Treats it as a struct, generating `nested.GGIDDiff` references

4. In [`generator/generator.go`](generator/generator.go:278-287), `convertToTemplateData()` sees `CategoryStruct` and sets `StructDiffFunc = "nested.ApplyGGIDDiff"`, which doesn't exist.

### Two Interdependent Bugs:

**Bug A** ([`types/types.go:149`](types/types.go:149)): When a `SelectorExpr` resolves through `typeDefs` to a basic type, the `inner` result is returned directly, losing the qualified type string (`"nested.GGID"`). The `TypeString` becomes `"string"` instead of `"nested.GGID"`.

**Bug B** ([`types/types.go:152-157`](types/types.go:152-157)): When a `SelectorExpr` does NOT resolve through `typeDefs` (because the typeDef wasn't imported), it falls back to assuming it's a struct, which is wrong for wrapped primitives.

## Fix Strategy

### Fix 1: [`types/types.go`](types/types.go) - `Classify()` SelectorExpr branch

**Location**: Lines 134-150

**Change**: When the inner type resolves to `CategoryBasic` through `typeDefs`, return a `CategoryBasic` result but with the **qualified** `TypeString` (e.g., `"nested.GGID"`) instead of the inner `TypeString` (e.g., `"string"`).

```go
// Current code (lines 134-150):
if typeDefs != nil {
    if underlying, ok := typeDefs[t.Sel.Name]; ok {
        inner := Classify(underlying, knownStructs, typeDefs)
        if inner.Category == CategoryMap {
            return &TypeInfo{
                Category:   CategoryMapAlias,
                TypeString: typeStr,
                Key:        inner.Key,
                Value:      inner.Value,
            }
        }
        // For other type aliases (e.g., type alias to struct), return as-is
        return inner  // <-- BUG: loses qualified type string for basic types
    }
}
```

**Fixed code**:
```go
if typeDefs != nil {
    if underlying, ok := typeDefs[t.Sel.Name]; ok {
        inner := Classify(underlying, knownStructs, typeDefs)
        if inner.Category == CategoryMap {
            return &TypeInfo{
                Category:   CategoryMapAlias,
                TypeString: typeStr,
                Key:        inner.Key,
                Value:      inner.Value,
            }
        }
        // For wrapped primitives (type alias to basic type like 'type GGID string'),
        // preserve the qualified type string (e.g., "nested.GGID") for correct code generation.
        if inner.Category == CategoryBasic {
            return &TypeInfo{
                Category:   CategoryBasic,
                TypeString: typeStr,
            }
        }
        // For other type aliases (e.g., type alias to struct), return as-is
        return inner
    }
}
```

### Fix 2: [`generator/generator.go`](generator/generator.go) - `convertToTemplateData()` struct/mapAlias diff function assignment

**Location**: Lines 277-287

**Change**: The current code assigns `StructDiffFunc` for both `CategoryStruct` and `CategoryMapAlias`. With Fix 1, wrapped primitives from external packages will correctly be `CategoryBasic`, so this code won't be reached for them. However, we should ensure the condition is robust.

**No change needed** - Fix 1 ensures wrapped primitives are classified as `CategoryBasic`, so they'll take the basic template path.

### Fix 3: [`generator/generator.go`](generator/generator.go) - `needsReflect` detection

**Location**: Lines 82-94

**Change**: The current code checks if any field uses `CategorySlice`, `CategoryMap`, `CategoryMapAlias`, or `CategoryUnknown` to determine if `reflect` import is needed. With Fix 1, wrapped primitives are `CategoryBasic`, so they won't trigger reflect import. This is correct since basic types use `!=` comparison.

**No change needed**.

### Fix 4: Regenerate `examples/complex_diff.go`

After applying Fix 1, regenerate the example diff file to verify the output is correct.

## Verification

1. **Unit tests**: Run `go test ./types/...` and `go test ./generator/...` to ensure no regressions
2. **Regenerate example**: Run `make generate_example` to regenerate `examples/complex_diff.go`
3. **Manual inspection**: Verify `examples/complex_diff.go` shows:
   ```go
   GGID *struct {
       Value nested.GGID
   }
   ```
   And the patch function uses `!=` comparison instead of calling `nested.GGIDPatch`

## Files to Modify

| File | Change |
|------|--------|
| [`types/types.go`](types/types.go:134-150) | Fix `Classify()` SelectorExpr branch to preserve qualified type string for wrapped primitives |
| [`examples/complex_diff.go`](examples/complex_diff.go) | Regenerate to verify fix |

## Files NOT to Modify (already correct)

- [`cmd/gostaticstructdiff/main.go`](cmd/gostaticstructdiff/main.go) - User's fix to restrict import typeDefs to wrapped primitives only is already applied
- [`generator/templates/struct_diff.go.tmpl`](generator/templates/struct_diff.go.tmpl) - Templates are correct; they just need the right category
- [`generator/templates/patch_func.go.tmpl`](generator/templates/patch_func.go.tmpl) - Templates are correct; they just need the right category