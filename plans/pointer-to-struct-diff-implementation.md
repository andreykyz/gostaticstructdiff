# Implementation Plan: Pointer-to-Struct Nested Diffs

## Objective
Modify `gostaticstructdiff` to generate nested diffs for pointer-to-struct fields instead of storing the whole pointer. The diff structure will support both pointer changes (nil vs non-nil) and struct content changes.

## Current Behavior
- Pointer fields are categorized as `CategoryPointer`.
- Diff struct contains `Value {{.Type}}` where `.Type` is the pointer type (e.g., `*models.User`).
- Patch function compares dereferenced values and stores the new pointer if they differ.
- Apply function assigns the stored pointer directly.

## New Behavior
For a field `Ref *models.User`:
- Diff struct will have two optional fields: `NewValue` (whole pointer) for pointer changes, and `Diff` (nested diff) for struct content changes.
- At most one field is set; if both are nil, no change.
- Patch function logic:
  1. If pointers are equal (both nil or same address): no diff.
  2. If one nil, other non-nil: set `NewValue` to the new pointer.
  3. If both non-nil and dereferenced values differ: compute nested diff and set `Diff`.
  4. If both non-nil and values equal: no diff.
- Apply function logic:
  1. If `diff.Ref.NewValue` != nil: replace pointer with `NewValue`.
  2. If `diff.Ref.Diff` != nil: apply diff to the original pointer's target (if original pointer is nil, treat as zero value?).

## Implementation Steps

### 1. Update Type Classification (`types/types.go`)
- Modify `DetermineDiffStrategy` to detect pointer-to-struct:
  - If `typeInfo.Category == CategoryPointer` and `typeInfo.Element.Category == CategoryStruct`, return a new template name `"pointerStruct"` with additional data:
    - `ElemType`: the underlying struct type (e.g., `models.User`)
    - `DiffType`: the diff type (e.g., `models.UserDiff`)
    - `DiffFunc`: the apply function name (e.g., `models.ApplyUserDiff`)
- Keep existing `CategoryPointer` for pointers to non-struct types.

### 2. Extend Generator Data Structures (`generator/generator.go`)
- Add fields to `FieldTemplateData`:
  ```go
  PointerElementIsStruct bool
  PointerElementType    string // e.g., "models.User"
  PointerDiffType      string // e.g., "models.UserDiff"
  PointerDiffFunc      string // e.g., "models.ApplyUserDiff"
  ```
- Update `convertToTemplateData` to populate these fields when category is `"pointerStruct"`.

### 3. Modify Templates
- **`struct_diff.go.tmpl`**: Add a new condition for `{{else if eq .Category "pointerStruct"}}` that generates:
  ```go
  {{.Name}} *struct {
      NewValue {{.PointerElementType}}
      Diff     {{.PointerDiffType}}
  }
  ```
- **`patch_func.go.tmpl`**: Add corresponding patch logic:
  - Determine which field to set based on pointer equality and content changes.
  - Use `{{.PointerElementType}}Patch` to compute nested diff.
- **`patch_func.go.tmpl`** apply section: Add logic to handle `NewValue` and `Diff`.

### 4. Create Helper Functions
- May need to add a helper function to compare pointers and compute diff. This can be inline in the template.

### 5. Update Example Generation
- Regenerate `examples/complex_diff.go` to reflect the new structure.
- Ensure the example still compiles and works.

### 6. Update Tests (`generator/generator_test.go`)
- Add test case for pointer-to-struct.
- Update existing tests if they rely on pointer field structure.

### 7. Backward Compatibility
- This is a breaking change for any generated code that uses pointer-to-struct diffs. Since the only example is `complex_diff.go`, we can update it.
- Consider adding a flag to preserve old behavior? Not required per user request.

## Detailed Template Changes

### `struct_diff.go.tmpl` snippet:
```go
{{- else if eq .Category "pointerStruct"}}
    {{.Name}} *struct {
        NewValue {{.PointerElementType}}
        Diff     {{.PointerDiffType}}
    }
```

### `patch_func.go.tmpl` snippet (patch generation):
```go
{{- else if eq .Category "pointerStruct"}}
    // Pointer-to-struct diff
    if original.{{.Name}} == nil && new.{{.Name}} == nil {
        // both nil, no change
    } else if original.{{.Name}} == nil && new.{{.Name}} != nil {
        // added pointer
        diff.{{.Name}} = &struct {
            NewValue {{.PointerElementType}}
            Diff     {{.PointerDiffType}}
        }{}
        diff.{{.Name}}.NewValue = new.{{.Name}}
    } else if original.{{.Name}} != nil && new.{{.Name}} == nil {
        // removed pointer
        diff.{{.Name}} = &struct {
            NewValue {{.PointerElementType}}
            Diff     {{.PointerDiffType}}
        }{}
        diff.{{.Name}}.NewValue = nil
    } else {
        // both non-nil, compare content
        if !reflect.DeepEqual(*original.{{.Name}}, *new.{{.Name}}) {
            nestedDiff := {{.PointerElementType}}Patch(*original.{{.Name}}, *new.{{.Name}})
            diff.{{.Name}} = &struct {
                NewValue {{.PointerElementType}}
                Diff     {{.PointerDiffType}}
            }{}
            diff.{{.Name}}.Diff = &nestedDiff
        }
    }
```

### `patch_func.go.tmpl` snippet (apply):
```go
{{- else if eq .Category "pointerStruct"}}
    if diff.{{.Name}} != nil {
        if diff.{{.Name}}.NewValue != nil {
            result.{{.Name}} = diff.{{.Name}}.NewValue
        } else if diff.{{.Name}}.Diff != nil {
            // Apply diff to existing pointer (must be non-nil)
            if original.{{.Name}} == nil {
                // If original is nil, create zero value
                zero := {{.PointerElementType}}{}
                result.{{.Name}} = &zero
            } else {
                result.{{.Name}} = original.{{.Name}}
            }
            *result.{{.Name}} = {{.PointerDiffFunc}}(*original.{{.Name}}, *diff.{{.Name}}.Diff)
        }
    }
```

## Open Questions
1. Should `Diff` be a pointer to diff (`*models.UserDiff`) or value (`models.UserDiff`)? Storing as pointer allows nil, but we already have `NewValue` nil to indicate pointer removal. We'll store as pointer to diff for consistency with other fields (where diff is a value). However, note that `StaticUser` stores diff as value (`models.UserDiff`). We'll follow that pattern: store diff as value, but wrap in pointer? Actually the diff field is `Diff {{.PointerDiffType}}` where `PointerDiffType` is `models.UserDiff` (non-pointer). That's fine; we can store zero value diff? That would be ambiguous. Better to store as pointer to diff (`*models.UserDiff`) to allow nil. Let's decide.

2. How to handle `reflect.DeepEqual` for structs that may contain slices/maps? Already used elsewhere.

3. Should we treat pointer equality (same address) as no change even if content changed? Probably not; we still want to detect content changes. We'll compare dereferenced values.

## Timeline
- Estimated effort: 2-3 days of development and testing.
- Priority: High (user requested).

## Approval
- User has approved the design.

## Next Steps
- Switch to Code mode to implement changes.
- Start with updating `types/types.go` and `generator/generator.go`.
- Then modify templates.
- Finally, regenerate examples and run tests.