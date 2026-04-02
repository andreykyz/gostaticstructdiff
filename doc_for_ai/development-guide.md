# AI Development Guide for gostaticstructdiff

## Overview

This document provides guidance for AI code agents to implement the `gostaticstructdiff` CLI utility. The tool generates type-safe diff structures and patch functions from Go structs annotated with `structtomap` tags.

## Project Understanding

### Core Concept

`gostaticstructdiff` is a Go code generation tool that:
1. Parses Go source files to find structs with `structtomap` tags
2. Generates corresponding `StructNameDiff` types with field-level change tracking
3. Creates `StructNamePatch` functions for computing and applying diffs

### Key Requirements from Prompts

1. **Input**: Go file with structs having `structtomap` tags
2. **Output**: Generated Go file with `_diff` suffix containing:
   - `StructNameDiff` type definitions
   - `StructNamePatch` function that accepts (original, new) → returns diff
   - `StructNamePatch` function that accepts (original, diff) → returns patched struct
3. **Pattern**: Follow examples in `examples/` folder

## Implementation Strategy

### Phase 1: Project Setup
1. Create proper Go module structure
2. Set up command-line interface skeleton
3. Establish basic parsing infrastructure

### Phase 2: Core Parser
1. Implement Go AST parsing to extract struct definitions
2. Filter structs with `structtomap` tags
3. Extract field information (name, type, tags)

### Phase 3: Type Analysis
1. Categorize field types (basic, pointer, slice, map, struct)
2. Determine appropriate diff strategy for each type
3. Handle nested/recursive struct types

### Phase 4: Code Generation
1. Implement template-based code generation
2. Generate diff struct definitions
3. Generate patch function implementations
4. Ensure proper imports and package declarations

### Phase 5: CLI Integration
1. Add command-line flags and options
2. Implement file I/O operations
3. Add error handling and user feedback

### Phase 6: Testing & Validation
1. Create unit tests for each component
2. Test with example files
3. Ensure generated code compiles and works correctly

## Technical Specifications

### File Structure
```
gostaticstructdiff/
├── cmd/
│   └── gostaticstructdiff/
│       └── main.go          # CLI entry point
├── internal/
│   ├── parser/              # AST parsing
│   ├── generator/           # Code generation
│   ├── types/               # Type system
│   └── templates/           # Go templates
├── examples/                # Example structs and generated code
└── go.mod                   # Go module definition
```

### Command-Line Interface
```bash
gostaticstructdiff -input <file.go> -output <file_diff.go> [options]
```

Options:
- `-input`: Input Go file (required)
- `-output`: Output file (default: `<input>_diff.go`)
- `-struct`: Specific struct to generate (default: all)
- `-package`: Package name for generated code
- `-verbose`: Enable verbose logging

### Generated Code Patterns

#### For basic struct:
```go
// Source
type User struct {
    ID   int    `structtomap:"id"`
    Name string `structtomap:"name"`
}

// Generated
type UserDiff struct {
    ID struct {
        Value int
        Set   bool
    }
    Name struct {
        Value string
        Set   bool
    }
}

func UserPatch(original, new User) UserDiff {
    // Compute diff between original and new
}

func UserPatch(original User, diff UserDiff) User {
    // Apply diff to original
}
```

#### For map fields:
```go
// Source
type Config struct {
    Settings map[string]string `structtomap:"settings"`
}

// Generated
type ConfigDiff struct {
    Settings struct {
        Add map[string]string
        Del map[string]struct{}
        Set bool
    }
}
```

## Implementation Details

### AST Parsing
- Use `go/parser` and `go/ast` packages
- Look for `*ast.TypeSpec` with `*ast.StructType`
- Extract `*ast.Field` with `structtomap` tags
- Handle nested structs recursively

### Type Classification
- Basic types: `int`, `string`, `bool`, `float64`, etc.
- Pointer types: `*T`
- Slice types: `[]T`
- Map types: `map[K]V`
- Struct types: nested or imported

### Template System
- Use Go's `text/template`
- Separate templates for different field types
- Template functions for code formatting

### Error Handling
- Graceful error messages for common issues
- Validation of input structs
- Recovery from parsing errors

## Testing Strategy

### Unit Tests
- Test parser with various struct definitions
- Test generator with known input/output pairs
- Test CLI with different arguments

### Integration Tests
- Generate code for example structs
- Verify generated code compiles
- Test patch functions with sample data

### Golden Tests
- Compare generated output with expected output
- Update golden files when generation changes

## Success Criteria

1. **Functional**:
   - Tool generates correct diff structs for all example files
   - Generated code compiles without errors
   - Patch functions work as expected

2. **Usability**:
   - Clear command-line interface
   - Helpful error messages
   - Integration with `go generate`

3. **Code Quality**:
   - Well-structured, maintainable code
   - Comprehensive test coverage
   - Proper documentation

## Common Pitfalls to Avoid

1. **Incorrect type handling**: Ensure all Go types are handled appropriately
2. **Import management**: Generated code needs correct imports
3. **Circular dependencies**: Avoid infinite recursion with nested structs
4. **Performance**: Large structs should not cause excessive memory usage
5. **Edge cases**: Zero values, nil pointers, empty maps/slices

## Next Steps After Implementation

1. **Documentation**: Update README and user guide
2. **Examples**: Add more example use cases
3. **Optimization**: Profile and optimize performance
4. **Extensions**: Consider additional features (custom templates, plugins)

## References

- Existing examples in `examples/` folder
- Go standard library: `go/ast`, `go/parser`, `text/template`
- Similar tools: `stringer`, `easyjson`, `go-swagger`

## Getting Help

- Review example patterns in `examples/` directory
- Check Go documentation for AST parsing and templates
- Test incrementally with small structs first