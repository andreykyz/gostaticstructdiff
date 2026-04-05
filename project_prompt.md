# gostaticstructdiff - Project Context for AI Code Agent

## Project Overview
`gostaticstructdiff` is a Go code generation tool that creates type-safe diff structures and patch functions from Go structs annotated with configurable tags (default: `structtomap`). It enables efficient structural diffing and patching operations for applications that need to track changes between versions of data structures.

## Core Functionality
- **Input**: Go source file containing structs with configurable tags (default: `structtomap`)
- **Output**: Generated Go file with `_diff` suffix containing:
  - `StructNameDiff` type definitions with field-level change tracking
  - `StructNamePatch(original, new StructName) StructNameDiff` function to compute diffs
  - `ApplyStructNameDiff(original StructName, diff StructNameDiff) StructName` function to apply diffs
- **Type Support**: Basic types, pointers, slices, maps, nested structs, and embedded types
- **Flexibility**: Custom tag selection via `-tag` flag; include all fields via `-all` flag

## Project Structure
```
gostaticstructdiff/
├── cmd/gostaticstructdiff/main.go      # CLI entry point with flag parsing
├── parser/parser.go                    # AST parsing to extract structs with structtomap tags
├── generator/generator.go              # Template-based code generation
├── types/types.go                      # Type system for field type categorization
├── generator/templates/                # Go templates for different field types
│   ├── field_basic.tmpl
│   ├── field_slice.tmpl
│   ├── field_map.tmpl
│   ├── field_pointer.tmpl
│   ├── field_struct.tmpl
│   └── patch_func.tmpl
├── examples/                           # Example usage
│   ├── models/user.go                  # Simple struct example
│   ├── models/user_diff.go             # Generated diff code
│   ├── complex.go                      # Complex struct with nested types
│   └── complex_diff.go                 # Generated complex diff
└── doc_for_ai/                         # AI development guides
```

## Key Implementation Details

### 1. Parser (`parser/parser.go`)
- Uses Go's `go/ast` package to parse source files
- Extracts structs that have at least one field with the specified tag (default: `structtomap`)
- Supports `-all` flag to include all fields regardless of tags
- Returns `StructInfo` with field names, types, and tags
- Handles imports from the source file

### 2. Generator (`generator/generator.go`)
- Uses Go's `text/template` package for code generation
- Different templates for different field types (basic, slice, map, pointer, struct)
- Generates diff structs with nested `Value` and `Set` fields
- Creates patch functions with proper type comparisons

### 3. Type System (`types/types.go`)
- Categorizes field types for appropriate diff strategy
- Handles nested struct recursion
- Manages imports for generated code

### 4. CLI Interface
```bash
gostaticstructdiff -input <file.go> -output <file_diff.go> [options]
```
Options:
- `-input`: Input Go file (required)
- `-output`: Output file (default: `<input>_diff.go`)
- `-struct`: Specific struct to generate (default: all)
- `-tag`: Tag key to look for (default: `structtomap`)
- `-all`: Include all fields regardless of tags (default: false)
- `-verbose`: Enable verbose logging
- `-version`: Show version

## Generated Code Pattern

### Source Struct:
```go
type User struct {
    ID       int    `structtomap:"id"`
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
    Active   bool   `structtomap:"active"`
}
```

### Generated Diff:
```go
type UserDiff struct {
    ID struct {
        Value int
        Set   bool
    }
    Username struct {
        Value string
        Set   bool
    }
    // ... more fields
}

func UserPatch(original, new User) UserDiff {
    var diff UserDiff
    if original.ID != new.ID {
        diff.ID.Value = new.ID
        diff.ID.Set = true
    }
    // ... more field comparisons
    return diff
}

func ApplyUserDiff(original User, diff UserDiff) User {
    result := original
    if diff.ID.Set {
        result.ID = diff.ID.Value
    }
    // ... more field applications
    return result
}
```

## Use Cases
- API partial updates (send only changed fields)
- Audit logging (track exact changes between versions)
- Conflict resolution (merge concurrent modifications)
- Event sourcing (store diffs instead of full states)
- Change notifications (trigger actions based on specific field changes)

## Development Guidelines
1. **Code Generation**: Ensure generated code is valid Go and compiles without errors
2. **Type Safety**: Maintain type safety across all generated operations
3. **Performance**: Optimize for minimal allocations in patch operations
4. **Testing**: Test with various struct types (basic, nested, slices, maps, pointers)
5. **Error Handling**: Provide clear error messages for invalid input

## Quick Start for AI Agent
1. Review the `examples/` directory to understand input/output patterns
2. Examine `doc_for_ai/` for development guidance
3. Check existing implementation in `parser/`, `generator/`, and `types/` directories
4. Use `go generate` integration for automated code generation

## Notes
- The project uses Go modules (see `go.mod`)
- All generated code should have zero runtime dependencies
- Follow Go conventions and best practices
- Maintain backward compatibility where possible

## Commit Conventions for AI Agents
When making changes to the codebase, AI agents should follow these commit message guidelines:

### Commit Message Format
```
<50‑character imperative subject line>

<72‑character wrapped body explaining "why" with issue references>
```

### Guidelines
1. **Subject Line**: Use imperative mood (e.g., "Add", "Fix", "Update", "Remove") and keep under 50 characters
2. **Blank Line**: Separate subject from body with a blank line
3. **Body**: Explain the "why" behind the change, not just "what"
   - Reference relevant issues (e.g., "Fixes #123", "Addresses #456")
   - Wrap lines at 72 characters for readability
   - Focus on the rationale and impact of the change
4. **One Logical Change Per Commit**: Each commit should represent a single logical change
   - Avoid mixing unrelated changes in one commit
   - Keep commits focused and atomic

### Example
```
Fix parser handling of embedded struct fields

The parser was incorrectly skipping embedded structs without structtomap
tags, causing missing fields in generated diff structs. This fix ensures
embedded structs are properly processed while maintaining the requirement
that at least one field has a structtomap tag.

Fixes #42
```

### Best Practices
- Test changes before committing
- Ensure code compiles and tests pass
- Update documentation when necessary
- Follow the project's existing coding style