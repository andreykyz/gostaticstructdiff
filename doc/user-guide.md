# gostaticstructdiff User Guide

## Overview

`gostaticstructdiff` is a Go code generation tool that creates type-safe diff structures and patch functions from your Go structs. It enables efficient structural diffing and patching operations for applications that need to track changes between versions of data structures.

### Key Features

- **Type-safe diff structures**: Generates `StructNameDiff` types with field-level change tracking
- **Patch functions**: Creates `StructNamePatch` functions for applying diffs and computing differences
- **Comprehensive type support**: Handles basic types, pointers, slices, maps, and nested structs
- **Go generate integration**: Works seamlessly with `go:generate` directives
- **Zero runtime dependencies**: Generated code is pure Go with no external dependencies

## Installation

### Using go install

```bash
go install github.com/andreykyz/gostaticstructdiff@latest
```

### Building from source

```bash
git clone https://github.com/andreykyz/gostaticstructdiff
cd gostaticstructdiff
go build -o gostaticstructdiff ./cmd/gostaticstructdiff
```

### Verifying installation

```bash
gostaticstructdiff --version
```

## Quick Start

1. **Annotate your struct** with `structtomap` tags:

```go
// user.go
package models

type User struct {
    ID       int    `structtomap:"id"`
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
    Active   bool   `structtomap:"active"`
}
```

2. **Generate diff code**:

```bash
gostaticstructdiff -input user.go -output user_diff.go
```

3. **Use the generated code**:

```go
import "your-project/models"

func main() {
    original := models.User{ID: 1, Username: "alice", Email: "alice@example.com", Active: true}
    updated := models.User{ID: 1, Username: "alice", Email: "alice@work.com", Active: true}
    
    // Compute diff between two structs
    diff := models.UserPatch(original, updated)
    
    // Apply diff to original
    patched := models.UserPatch(original, diff)
}
```

## Annotating Structs

### Required Tags

Add `structtomap:"field_name"` tags to fields you want included in diff generation:

```go
type Product struct {
    ID          int       `structtomap:"id"`
    Name        string    `structtomap:"name"`
    Price       float64   `structtomap:"price"`
    Categories  []string  `structtomap:"categories"`
    Metadata    map[string]string `structtomap:"metadata"`
    CreatedAt   time.Time `structtomap:"created_at"`
}
```

### Supported Field Types

| Type | Example | Diff Structure |
|------|---------|----------------|
| Basic types | `int`, `string`, `bool`, `float64` | `struct { Value T; Set bool }` |
| Pointer types | `*string`, `*int` | `struct { Value *T; Set bool }` |
| Slices | `[]string`, `[]int` | `struct { Value []T; Set bool }` |
| Maps | `map[string]int` | `struct { Add map[K]V; Del map[K]struct{}; Set bool }` |
| Nested structs | `struct { ... }` | Recursive diff struct |
| Embedded structs | `Embedded` | Flattened fields |
| Time types | `time.Time` | `struct { Value time.Time; Set bool }` |

### Nested Structs

For nested structs, the tool generates recursive diff structures:

```go
type Address struct {
    Street string `structtomap:"street"`
    City   string `structtomap:"city"`
}

type Customer struct {
    Name    string  `structtomap:"name"`
    Address Address `structtomap:"address"`
}
```

Generates `CustomerDiff` with nested `AddressDiff`.

### Maps with Struct Values

For maps with struct values, the tool generates specialized diff structures with Add, Del, and Mod operations:

```go
type Config struct {
    Settings map[string]Setting `structtomap:"settings"`
}
```

Generates diff with `Add`, `Del`, and `Mod` fields for efficient map diffing.

## Command Line Interface

### Basic Usage

```bash
gostaticstructdiff -input <input-file> -output <output-file>
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input Go source file | Required |
| `-output` | Output file for generated code | `<input>_diff.go` |
| `-struct` | Specific struct to generate (default: all) | All structs |
| `-package` | Package name for generated code | Same as input |
| `-verbose` | Enable verbose logging | false |
| `-version` | Show version information | N/A |

### Integration with go generate

Add a `go:generate` directive to your source files:

```go
//go:generate gostaticstructdiff -input $GOFILE -output ${GOFILE%.go}_diff.go

type User struct {
    // fields...
}
```

Then run:

```bash
go generate ./...
```

## Generated Code Patterns

### Diff Structures

For a struct `User` with fields `ID int` and `Name string`, the tool generates:

```go
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
```

### Patch Functions

Two patch functions are generated:

1. **Compute diff** between two structs:

```go
func UserPatch(original, new User) UserDiff
```

2. **Apply diff** to a struct:

```go
func UserPatch(original User, diff UserDiff) User
```

### Map Field Diffs

For map fields, the diff structure includes Add and Del operations:

```go
type ConfigDiff struct {
    Settings struct {
        Add map[string]Setting
        Del map[string]struct{}
        Mod map[string]SettingDiff
        Set bool
    }
}
```

### Slice Field Diffs

For slice fields, the entire slice is replaced:

```go
type ProductDiff struct {
    Tags struct {
        Value []string
        Set   bool
    }
}
```

## Examples

### Basic Example

See `examples/models/user.go` and `examples/models/user_diff.go` for a simple example.

### Complex Example

See `examples/complex.go` and `examples/complex_diff.go` for advanced usage with slices, maps, nested structs, and pointers.

### Real-World Use Case: API Versioning

```go
// api/v1/models.go
type UserRequest struct {
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
    Settings map[string]string `structtomap:"settings"`
}

// Generate diff
//go:generate gostaticstructdiff -input models.go -output models_diff.go

// In your API handler
func handleUserUpdate(original, update UserRequest) {
    diff := UserRequestPatch(original, update)
    
    // Only send changed fields to downstream services
    if diff.Email.Set {
        sendEmailUpdate(original.Email, diff.Email.Value)
    }
    
    // Track changes for audit logging
    auditLog(diff)
}
```

## Best Practices

### 1. Use Descriptive Field Names

Choose meaningful field names in `structtomap` tags as they appear in generated code.

### 2. Handle Zero Values

Remember that `Set: false` means the field wasn't changed, not that it was set to its zero value.

### 3. Test Generated Code

Always write tests for code that uses generated diff structures:

```go
func TestUserPatch(t *testing.T) {
    original := User{ID: 1, Name: "Alice"}
    updated := User{ID: 1, Name: "Bob"}
    
    diff := UserPatch(original, updated)
    
    if !diff.Name.Set || diff.Name.Value != "Bob" {
        t.Errorf("Expected Name diff not found")
    }
    
    patched := UserPatch(original, diff)
    if patched != updated {
        t.Errorf("Patch didn't produce expected result")
    }
}
```

### 4. Consider Performance

- For large structs, diff computation is O(n) where n is the number of fields
- Map diffs are more expensive than scalar field diffs
- Consider batching updates for performance-critical applications

### 5. Version Control

- Commit generated files to version control
- Regenerate when source structs change
- Use `go generate` in your build process

## Troubleshooting

### Common Issues

1. **"No structs found with structtomap tags"**
   - Ensure all fields have `structtomap` tags
   - Check tag syntax (backticks, correct quotes)

2. **Generated code doesn't compile**
   - Check for syntax errors in source structs
   - Ensure all referenced types are imported
   - Verify Go version compatibility

3. **Missing fields in diff struct**
   - Unexported fields are ignored
   - Fields without `structtomap` tags are ignored

4. **Type not supported**
   - Check the list of supported types
   - Consider wrapping unsupported types in a supported container

### Debugging

Use verbose mode to see what the tool is processing:

```bash
gostaticstructdiff -input myfile.go -output myfile_diff.go -verbose
```

## Next Steps

- Explore the [Developer Documentation](./developer-guide.md) for implementation details
- Check [API Reference](./api-reference.md) for package documentation
- Review [Examples](../examples/) for more complex use cases
- Read [Best Practices](./best-practices.md) for production usage