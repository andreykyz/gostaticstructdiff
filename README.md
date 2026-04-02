# gostaticstructdiff

A Go code generation tool for creating type-safe diff structures and patch functions from your Go structs.

## Overview

`gostaticstructdiff` analyzes Go structs annotated with `structtomap` tags and generates:
- **Diff structures** (`StructNameDiff`) with field-level change tracking
- **Patch functions** for computing and applying diffs
- **Type-safe operations** for structural diffing and patching

Perfect for applications that need to track changes between versions of data structures, implement partial updates, or maintain audit trails.

## Features

- ✅ **Type-safe diff structures** for all Go basic types
- ✅ **Nested struct support** with recursive diff generation
- ✅ **Map and slice handling** with efficient diff operations
- ✅ **Zero-dependency generated code** (pure Go)
- ✅ **`go generate` integration** for automated code generation
- ✅ **Comprehensive type support** (pointers, embedded structs, etc.)
- ✅ **Performance optimized** with minimal allocations

## Quick Start

### Installation

```bash
go install github.com/andreykyz/gostaticstructdiff@latest
```

### Basic Usage

1. **Annotate your struct**:

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
original := models.User{ID: 1, Username: "alice", Email: "alice@example.com", Active: true}
updated := models.User{ID: 1, Username: "alice", Email: "alice@work.com", Active: true}

// Compute diff between two structs
diff := models.UserPatch(original, updated)

// Apply diff to original
patched := models.UserPatch(original, diff)
```

## Examples

Check out the [examples directory](./examples/) for complete examples:

- [Basic User struct](./examples/models/user.go) with generated diff
- [Complex nested structures](./examples/complex.go) with maps and slices
- [Metadata handling](./examples/models/metadata.go) with map operations

## Documentation

- [User Guide](./doc/user-guide.md) - Installation, usage, and examples
- [Developer Guide](./doc/developer-guide.md) - Architecture and development setup
- [API Reference](./doc/api-reference.md) - Complete API documentation
- [Tutorial](./doc/tutorial.md) - Step-by-step walkthroughs
- [Best Practices](./doc/best-practices.md) - Production recommendations

## Command Line Interface

```bash
gostaticstructdiff -input <input-file> -output <output-file> [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input Go source file | Required |
| `-output` | Output file for generated code | `<input>_diff.go` |
| `-struct` | Specific struct to generate | All structs |
| `-package` | Package name for generated code | Same as input |
| `-verbose` | Enable verbose logging | false |
| `-version` | Show version information | N/A |

## Integration with `go generate`

Add a `go:generate` directive to your source files:

```go
//go:generate gostaticstructdiff -input $GOFILE -output ${GOFILE%.go}_diff.go

type Product struct {
    ID    int     `structtomap:"id"`
    Name  string  `structtomap:"name"`
    Price float64 `structtomap:"price"`
}
```

Then run:

```bash
go generate ./...
```

## Generated Code Patterns

### Diff Structures

For each field in the source struct, a corresponding diff field is generated:

```go
type UserDiff struct {
    ID struct {
        Value int
        Set   bool  // true if field was changed
    }
    Username struct {
        Value string
        Set   bool
    }
    // ... more fields
}
```

### Patch Functions

Two functions are generated:

1. **Compute diff**: `UserPatch(original, new User) UserDiff`
2. **Apply diff**: `UserPatch(original User, diff UserDiff) User`

## Use Cases

- **API partial updates**: Send only changed fields to clients
- **Audit logging**: Track exactly what changed between versions
- **Conflict resolution**: Merge concurrent modifications
- **Event sourcing**: Store diffs instead of full states
- **Change notifications**: Trigger actions based on specific field changes

## Performance

- **Time complexity**: O(N + M) where N = field count, M = map/slice elements
- **Memory usage**: Diff structs are 2-3x larger than original structs
- **Zero allocations** for unchanged fields in patch operations

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [Open an issue](https://github.com/andreykyz/gostaticstructdiff/issues) for bug reports or feature requests
- Check the [FAQ](./doc/faq.md) for common questions
- Review the [troubleshooting guide](./doc/user-guide.md#troubleshooting) for help

## Acknowledgments

- Inspired by the need for type-safe diff operations in Go
- Built using Go's excellent standard library packages (`go/ast`, `go/parser`, `text/template`)
- Thanks to all contributors and users

---

**gostaticstructdiff** - Type-safe structural diffing for Go