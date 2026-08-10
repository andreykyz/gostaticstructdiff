# Plan: Add `-suffix` CLI Option for Generated Identifiers and Filename

## Problem

The code generator produces fixed identifiers: `{{Name}}Diff`, `{{Name}}Patch`,
`Apply{{Name}}Diff`, plus map-alias identifiers `{{Name}}Diff` / `{{Name}}Patch` /
`Apply{{Name}}Diff`, and an output filename `<input>_diff.go`. There is no way to
namespace or distinguish multiple generated diff sets for the same structs in the
same package, which causes symbol collisions.

## Solution Overview

Add a `-suffix` flag to the CLI. When provided:

- Convert the suffix to **CamelCase** for identifier names (e.g. `my_suffix` -> `MySuffix`).
- Append the CamelCase suffix to generated **diff struct types**, **patch functions**,
  and **apply functions** (including nested references to other diff types/functions,
  and map alias types/functions).
- Keep the **original struct type name unchanged** in signatures (only the generated
  `Diff`/`Patch`/`Apply` identifiers get the suffix).
- Append the **raw (underscore) suffix** to the output **filename** (e.g.
  `complex_diff_my_suffix.go`).

### Concrete Example

Given input struct `ComplexStruct`, with `-suffix my_suffix`:

| Before                          | After (CamelCase suffix `MySuffix`)                 |
|---------------------------------|-----------------------------------------------------|
| `type ComplexStructDiff struct` | `type ComplexStructDiffMySuffix struct`            |
| `func (d ComplexStructDiff) IsEmpty() bool` | `func (d ComplexStructDiffMySuffix) IsEmpty() bool` |
| `func ComplexStructPatch(...) ComplexStructDiff` | `func ComplexStructPatchMySuffix(...) ComplexStructDiffMySuffix` |
| `func ApplyComplexStructDiff(...)` | `func ApplyComplexStructDiffMySuffix(...)`         |
| `type models.UserDiff` (nested ref) | `type models.UserDiffMySuffix` (nested ref)        |
| `models.ApplyUserDiff(...)` (nested call) | `models.ApplyUserDiffMySuffix(...)` (nested call) |
| `type MetaStringDiff struct` (map alias) | `type MetaStringDiffMySuffix struct`              |
| `func MetaStringPatch(...)` (map alias) | `func MetaStringPatchMySuffix(...)`               |
| `ApplyMetaStringDiff(...)` (map alias) | `ApplyMetaStringDiffMySuffix(...)`               |
| output file `complex_diff.go`   | output file `complex_diff_my_suffix.go` (raw suffix) |

## Detailed Changes

### 1. CLI flag (`cmd/gostaticstructdiff/main.go`)

Note: this file is currently listed in `.codeassistantignore`; the implementer must
ensure it is accessible or update the ignore rules before editing.

- Add flag: `suffix := flag.String("suffix", "", "Optional suffix to append to generated type and function names, and to the output filename")`.
- Pass the suffix (raw string) into `generator.Generate(...)`.
- When computing the default output filename, if a suffix is present and no explicit
  `-output` is given, append `_<suffix>` before `.go` on the first input's derived name,
  i.e. `<first_input>_diff_<suffix>.go`.
- Add the flag to CLI usage/help and to `README.md`.

### 2. `generator/generator.go` — thread suffix through generation

- Change `Generate(structs, packageName, imports, version, typeDefs, verbose)` to accept a
  new `suffix string` parameter (append at the end to avoid breaking existing positional call
  sites unless they are updated together).
- Compute a CamelCase version of the suffix once:
  ```go
  func toCamelCase(suffix string) string {
      // split on '_' and '-', capitalize first letter of each part, join
  }
  ```
- Add a `Suffix string` field to `StructTemplateData` and to `MapAliasTemplateData`.
- In `convertToTemplateData(s, knownStructs, typeDefs, suffix)`:
  - Pass the suffix into the returned `StructTemplateData`.
  - When constructing any **generated** identifier/type/function name that embeds another
    struct's or map alias's diff type, append the CamelCase suffix:
    - `ValueDiffType` (`{{.ValueType}}Diff` -> `{{.ValueType}}Diff{{Suffix}}`)
    - `ElementDiffType`
    - `PointerDiffType`
    - `StructDiffFunc` (`pkg.Apply<Name>Diff` -> `pkg.Apply<Name>Diff<Suffix>`)
    - `ValueDiffFunc`, `ElementDiffFunc`, `PointerDiffFunc`
- In `collectReferencedMapAliases(structs, typeDefs, knownStructs, suffix)`:
  - Set `Suffix` on each `MapAliasTemplateData`.
  - Append CamelCase suffix to `ValueDiffType` and `ValueDiffFunc`.

### 3. Templates — reference the suffix on generated identifiers

All templates currently emit generated names inline (e.g. `{{.Name}}Diff`,
`Apply{{.Name}}Diff`, `{{.Type}}Diff`). Update them to include `{{.Suffix}}` after
`Diff`/`Patch`/`Apply` where the name is a **generated identifier**. The original
struct name `{{.Name}}` (as an argument type) must NOT get the suffix.

**`generator/templates/struct_diff.go.tmpl`:**
- `type {{.Name}}Diff struct` -> `type {{.Name}}Diff{{.Suffix}} struct`
- `func (d {{.Name}}Diff) IsEmpty() bool` -> `func (d {{.Name}}Diff{{.Suffix}}) IsEmpty() bool`
- Field refs already use `{{.Type}}Diff`/`{{.PointerDiffType}}`/`{{.ValueDiffType}}` etc.;
  these are populated with the suffix in `convertToTemplateData`, so no change needed here
  **except** where the type string is built inline. Audit each field template.

**`generator/templates/patch_func.go.tmpl`:**
- `func {{.Name}}Patch(original, new {{.Name}}) {{.Name}}Diff` ->
  `func {{.Name}}Patch{{.Suffix}}(original, new {{.Name}}) {{.Name}}Diff{{.Suffix}}`
- `var diff {{.Name}}Diff` -> `var diff {{.Name}}Diff{{.Suffix}}`
- `func Apply{{.Name}}Diff(original {{.Name}}, diff {{.Name}}Diff) {{.Name}}` ->
  `func Apply{{.Name}}Diff{{.Suffix}}(original {{.Name}}, diff {{.Name}}Diff{{.Suffix}}) {{.Name}}`
- Nested calls (`{{.PointerElementType}}Patch`, `{{.ValueType}}Patch`, `{{.ValueDiffFunc}}`,
  `{{.ElementDiffFunc}}`, `{{.StructDiffFunc}}`) rely on data fields that are suffixed in
  `convertToTemplateData`; verify each is covered.

**`generator/templates/map_alias_diff.go.tmpl`:**
- `type {{.Name}}Diff struct` -> `type {{.Name}}Diff{{.Suffix}} struct`
- `func (d {{.Name}}Diff) IsEmpty() bool` -> `func (d {{.Name}}Diff{{.Suffix}}) IsEmpty() bool`

**`generator/templates/map_alias_patch.go.tmpl`:**
- `func {{.Name}}Patch(original, new {{.Name}}) {{.Name}}Diff` -> suffix on `Patch` and `Diff`
- `var diff {{.Name}}Diff` -> suffixed
- `func Apply{{.Name}}Diff(original {{.Name}}, diff {{.Name}}Diff) {{.Name}}` -> suffix on generated names
- `nestedDiff := {{.ValueType}}Patch(...)` -> suffixed via data field or template addition
- `{{.ValueDiffFunc}}(...)` -> suffixed via data field

> Design decision: to keep changes localized and consistent, prefer populating suffixed
> values in the data structs (`convertToTemplateData` / `collectReferencedMapAliases`)
> rather than scattering `{{.Suffix}}` in many inline template spots. Only add `{{.Suffix}}`
> in templates where the name is built directly (diff struct type, patch/apply function
> names). Field-level nested names should be fully resolved in the data layer.

### 4. Filename handling

- When suffix is provided and `-output` is empty, derive the default as
  `<first_input_base>_diff_<suffix>.go` (raw suffix, underscores preserved).

### 5. Tests

- Add generator tests for `-suffix`:
  - A suffix is applied to diff struct type, IsEmpty receiver, Patch, and Apply names.
  - Nested struct/map references include the suffix (e.g. `models.UserDiffMySuffix`).
  - Map alias types/functions are suffixed.
  - Original struct type name is unchanged.
  - Empty suffix produces current (unsuffixed) output (backward compatibility).
- Add unit tests for `toCamelCase` (e.g. `my_suffix` -> `MySuffix`, empty -> empty,
  `foo-bar` -> `FooBar`, `Foo` -> `Foo`).
- Run `make test` and `make example` to confirm existing examples still compile and the
  new flag works end-to-end.

### 6. Documentation

- Update `README.md` to document the `-suffix` flag and show an example invocation.

## Backward Compatibility

- Default behavior (no `-suffix`) must remain byte-for-byte identical to current output.
- `Generate` signature change: update all call sites in tests (`generator/generator_test.go`)
  and any callers (e.g. `debugging/`) to pass the new suffix argument (`""` for default).