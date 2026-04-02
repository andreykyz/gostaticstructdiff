# Task Specifications for AI Implementation

## Project: gostaticstructdiff CLI Utility

### Overall Goal
Create a Go CLI tool that generates type-safe diff structures and patch functions from Go structs annotated with `structtomap` tags.

## Task Breakdown

### Task 1: Project Structure Setup
**Objective**: Create the basic Go module structure and directory layout.

**Specifications**:
1. Create `cmd/gostaticstructdiff/main.go` as the CLI entry point
2. Create `internal/` directory with subdirectories: `parser/`, `generator/`, `types/`, `templates/`
3. Update `go.mod` with proper module name and dependencies
4. Create `Makefile` or build scripts for development

**Expected Output**:
```
gostaticstructdiff/
├── cmd/
│   └── gostaticstructdiff/
│       └── main.go
├── internal/
│   ├── parser/
│   ├── generator/
│   ├── types/
│   └── templates/
├── go.mod
└── go.sum
```

**Validation**:
- `go build ./cmd/gostaticstructdiff` should succeed
- `go test ./...` should run (even if no tests yet)

### Task 2: CLI Interface Implementation
**Objective**: Implement command-line argument parsing and basic file I/O.

**Specifications**:
1. Use `flag` package or `cobra` for CLI argument parsing
2. Support these flags:
   - `-input string`: Input Go file (required)
   - `-output string`: Output file (default: `<input>_diff.go`)
   - `-struct string`: Specific struct to generate (default: all)
   - `-verbose bool`: Enable verbose logging
   - `-version bool`: Show version
3. Implement basic file reading/writing
4. Add help text and error handling for missing arguments

**Expected Output**:
- Command `gostaticstructdiff -help` shows usage information
- Command `gostaticstructdiff -input test.go` generates `test_diff.go`
- Missing required arguments show appropriate error messages

**Validation**:
- CLI compiles and runs
- Help text is clear and complete
- File operations work correctly

### Task 3: AST Parser Implementation
**Objective**: Parse Go files and extract struct information.

**Specifications**:
1. Create `internal/parser/parser.go` with `ParseFile` function
2. Use `go/parser` and `go/ast` packages to parse Go files
3. Extract all struct type definitions with `structtomap` tags
4. For each struct, extract:
   - Struct name
   - Field names and types
   - Tag values
   - Field comments (optional)
5. Handle nested structs recursively
6. Support pointer, slice, map, and basic types

**Expected Output**:
- Parser can read `examples/models/user.go` and extract `User` struct
- Parser correctly identifies fields with `structtomap` tags
- Parser ignores fields without tags
- Parser handles all types in example files

**Validation**:
- Unit tests for parser with sample structs
- Parser works with all example files
- Edge cases handled (empty structs, no tags, etc.)

### Task 4: Type System and Diff Strategies
**Objective**: Create type classification and diff generation strategies.

**Specifications**:
1. Create `internal/types/` package with type classification
2. Implement diff strategies for:
   - Basic types (int, string, bool, float64): `struct { Value T; Set bool }`
   - Pointer types: `struct { Value *T; Set bool }`
   - Slice types: `struct { Value []T; Set bool }`
   - Map types: `struct { Add map[K]V; Del map[K]struct{}; Set bool }`
   - Struct types: Recursive diff struct
3. Handle nested/embedded structs
4. Create template mappings for each type

**Expected Output**:
- Type classifier correctly identifies field types
- Appropriate diff strategy selected for each type
- Strategies match patterns in example diff files

**Validation**:
- Type classification tests
- Diff strategy selection tests
- Comparison with example diff patterns

### Task 5: Template System
**Objective**: Create Go templates for code generation.

**Specifications**:
1. Create template files in `internal/templates/`:
   - `struct_diff.tmpl`: Main struct template
   - `field_basic.tmpl`: Basic type fields
   - `field_pointer.tmpl`: Pointer type fields
   - `field_slice.tmpl`: Slice type fields
   - `field_map.tmpl`: Map type fields
   - `field_struct.tmpl`: Nested struct fields
   - `patch_func.tmpl`: Patch function template
2. Templates should generate code matching examples
3. Include proper imports, package declaration, and formatting
4. Add template functions for code formatting helpers

**Expected Output**:
- Templates generate code identical to example diff files
- Generated code is properly formatted
- All required imports are included

**Validation**:
- Template tests with sample data
- Generated code compiles without errors
- Output matches golden files

### Task 6: Code Generator
**Objective**: Integrate parser, type system, and templates to generate code.

**Specifications**:
1. Create `internal/generator/generator.go` with `Generate` function
2. Input: parsed struct information
3. Process: Apply appropriate templates for each field
4. Output: Complete Go source code with diff struct and patch functions
5. Handle multiple structs in one file
6. Manage imports and package dependencies

**Expected Output**:
- Generator creates `UserDiff` from `User` struct (matches example)
- Generator creates `MetadataDiff` from `Metadata` struct (matches example)
- Generator creates `ComplexStructDiff` from `ComplexStruct` (matches example)

**Validation**:
- Generated code for examples matches existing diff files
- Generated code compiles
- Patch functions work correctly

### Task 7: Patch Function Implementation
**Objective**: Implement the actual patch function logic.

**Specifications**:
1. Two patch functions per struct:
   - `StructPatch(original, new Struct) StructDiff`: Computes diff
   - `StructPatch(original Struct, diff StructDiff) Struct`: Applies diff
2. For basic types: compare values, set `Set: true` if different
3. For maps: compute added/deleted entries
4. For slices: compare entire slices
5. For nested structs: call recursively
6. Optimize for performance (avoid unnecessary allocations)

**Expected Output**:
- Patch functions compile and work
- Diff computation is correct for all field types
- Diff application reconstructs the target struct

**Validation**:
- Unit tests for patch functions with various inputs
- Property-based tests (diff then apply should reconstruct)
- Performance benchmarks

### Task 8: Integration and Testing
**Objective**: Integrate all components and create comprehensive tests.

**Specifications**:
1. Create integration tests that:
   - Run the tool on example files
   - Compare generated output with expected output
   - Compile generated code
   - Test patch functions with sample data
2. Create golden file tests for regression testing
3. Add end-to-end tests for CLI
4. Test edge cases: empty structs, nil pointers, zero values

**Expected Output**:
- All tests pass
- Tool works correctly with all example files
- Generated code is correct and compiles

**Validation**:
- `go test ./...` passes
- Integration tests verify full functionality
- No regressions in generated code

### Task 9: Documentation and Examples
**Objective**: Create user documentation and additional examples.

**Specifications**:
1. Update README.md with usage instructions
2. Create more example files demonstrating various use cases
3. Add `go:generate` directive examples
4. Document limitations and known issues
5. Create troubleshooting guide

**Expected Output**:
- Comprehensive documentation
- Clear examples for common use cases
- Help text in CLI

**Validation**:
- Documentation is clear and accurate
- Examples work as described
- Users can understand how to use the tool

## Implementation Order

Recommended sequence for AI implementation:

1. **Task 1**: Project Structure Setup
2. **Task 2**: CLI Interface Implementation  
3. **Task 3**: AST Parser Implementation
4. **Task 5**: Template System (create basic templates)
5. **Task 4**: Type System and Diff Strategies
6. **Task 6**: Code Generator
7. **Task 7**: Patch Function Implementation
8. **Task 8**: Integration and Testing
9. **Task 9**: Documentation and Examples

## Success Metrics

1. **Functional**:
   - Tool generates correct diff structs for all example files
   - Generated code compiles without errors
   - Patch functions work correctly

2. **Usability**:
   - Clear CLI interface with helpful error messages
   - Reasonable performance (sub-second for typical structs)
   - Good documentation

3. **Code Quality**:
   - Well-structured, maintainable code
   - Comprehensive test coverage (>80%)
   - Proper error handling
   - No panics in normal operation

## Testing Checklist

For each task, verify:
- [ ] Unit tests exist and pass
- [ ] Edge cases are handled
- [ ] Code follows Go conventions
- [ ] No linting errors
- [ ] Documentation is updated

## Delivery Requirements

The final implementation should:
1. Be a working Go CLI tool
2. Generate code matching the examples
3. Have comprehensive tests
4. Include documentation
5. Follow Go best practices

## References

- Example files in `examples/` directory
- Go standard library documentation
- Existing similar tools (stringer, easyjson)