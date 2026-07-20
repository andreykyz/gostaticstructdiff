# gostaticstructdiff

**Type-safe diff and patch code generation for Go structs.**

`gostaticstructdiff` is a Go code generation tool that reads annotated Go structs and generates type-safe diff structures and patch functions. It enables efficient structural diffing and patching operations for applications that need to track changes between versions of data structures — with zero runtime dependencies.

## Use Cases

- **API partial updates** — send only changed fields in PATCH requests
- **Audit logging** — track exact changes between versions of data
- **Conflict resolution** — merge concurrent modifications
- **Event sourcing** — store diffs instead of full state snapshots
- **Change notifications** — trigger actions based on specific field changes

## How It Works

1. Annotate your Go structs with a configurable tag (default: `structtomap`).
2. Run `gostaticstructdiff` to generate diff types and patch functions.
3. Use the generated code to compute diffs between struct values and apply them.

### Example

**Source struct** ([`user.go`](examples/models/user.go)):

```go
package models

type User struct {
    ID       int    `structtomap:"id"`
    Username string `structtomap:"username"`
    Email    string `structtomap:"email"`
    Active   bool   `structtomap:"active"`
}
```

**Generated output** ([`user_diff.go`](examples/models/user_diff.go)):

```go
// Diff type — each field is a pointer; nil means unchanged
type UserDiff struct {
    ID       *struct{ Value int    }
    Username *struct{ Value string }
    Email    *struct{ Value string }
    Active   *struct{ Value bool   }
}

// Compute diff between two struct values
func UserPatch(original, new User) UserDiff { ... }

// Apply a diff to an original struct value
func ApplyUserDiff(original User, diff UserDiff) User { ... }
```

## Installation

```bash
go install github.com/andreykyz/gostaticstructdiff/cmd/gostaticstructdiff@latest
```

Or build from source:

```bash
git clone https://github.com/andreykyz/gostaticstructdiff.git
cd gostaticstructdiff
make build
```

## Usage

```bash
gostaticstructdiff -input <file.go> -output <output.go> [options]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input Go file(s), comma-separated (required) | — |
| `-output` | Output file | `<first_input>_diff.go` |
| `-struct` | Specific struct to generate (generates all if empty) | all |
| `-tag` | Tag key to look for | `structtomap` |
| `-all` | Include all fields regardless of tags | `false` |
| `-verbose` | Enable verbose logging | `false` |
| `-version` | Show version | `false` |

### Examples

```bash
# Generate diff for all structs in user.go
gostaticstructdiff -input user.go

# Generate diff for a specific struct with custom tag
gostaticstructdiff -input models.go -struct User -tag "diff"

# Include all fields (ignore tag filtering)
gostaticstructdiff -input models.go -all

# Process multiple files (must belong to the same package)
gostaticstructdiff -input "user.go,metadata.go" -output combined_diff.go
```

### `go:generate` Integration

Add a `go:generate` directive to your source file:

```go
//go:generate gostaticstructdiff -input $GOFILE
package models
```

Then run:

```bash
go generate ./...
```

## Type Support

| Category | Diff Strategy | Example |
|----------|--------------|---------|
| **Basic types** | Value comparison (`!=`) | `int`, `string`, `bool`, `float64` |
| **Pointers** | Pointer equality + dereferenced comparison | `*string`, `*int` |
| **Pointer-to-struct** | Nested diff with `NewValue`/`Diff` fields | `*models.User` |
| **Slices** | `reflect.DeepEqual` comparison | `[]string`, `[]models.User` |
| **Maps** | Key-wise diff (added/deleted/modified) | `map[string]int` |
| **Map with struct values** | Nested diffs for modified keys | `map[string]models.Metadata` |
| **Nested structs** | Recursive diff struct generation | inline structs, imported structs |
| **Embedded types** | Fields promoted to parent diff | embedded structs |
| **Wrapped primitives** | Resolved via type definitions | `type MyString string` |
| **Unknown types** | `reflect.DeepEqual` fallback | qualified identifiers from other packages |

## Generated Code Pattern

For each annotated struct `Foo`, the tool generates:

1. **`FooDiff`** — a struct with one field per source field, each wrapped in an anonymous pointer struct. A `nil` pointer means the field is unchanged.

2. **`FooPatch(original, new Foo) FooDiff`** — computes the diff between two struct values. Only fields that differ are set in the returned diff.

3. **`ApplyFooDiff(original Foo, diff FooDiff) Foo`** — applies a diff to an original struct, returning the patched result.

### Map Diff Details

For map fields, the diff contains `Add` and `Del` maps:

```go
Tags *struct {
    Add map[string]string
    Del map[string]struct{}
}
```

For maps with struct values, an additional `Modify` field provides nested diffs for changed entries:

```go
Metadata *struct {
    Add    map[string]models.Metadata
    Del    map[string]struct{}
    Modify map[string]models.MetadataDiff
}
```

### Pointer-to-Struct Diff Details

For pointer-to-struct fields, the diff contains both a `NewValue` (for pointer changes) and a `Diff` (for content changes):

```go
Ref *struct {
    NewValue *models.User
    Diff     *models.UserDiff
}
```

At most one of `NewValue` or `Diff` is set. If both are nil, the pointer is unchanged.

## Project Structure

```
gostaticstructdiff/
├── cmd/gostaticstructdiff/main.go   # CLI entry point
├── parser/parser.go                 # AST-based struct extraction
├── generator/generator.go           # Template-based code generation
├── generator/templates/             # Go templates for code generation
│   ├── field_basic.go.tmpl
│   ├── field_slice.go.tmpl
│   ├── field_map.go.tmpl
│   ├── field_pointer.go.tmpl
│   ├── field_struct.go.tmpl
│   ├── struct_diff.go.tmpl
│   └── patch_func.go.tmpl
├── types/types.go                   # Type classification system
├── examples/                        # Usage examples
│   ├── cmd/simple/main.go           # Basic diff/patch demo
│   ├── cmd/gen/main.go              # Random generation demo
│   ├── models/user.go               # Simple struct example
│   ├── models/metadata.go           # Metadata struct example
│   ├── complex.go                   # Complex struct with nested types
│   └── *_diff.go                    # Generated diff files
└── debugging/                       # Debug and test utilities
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the CLI binary |
| `make test` | Run all Go tests |
| `make generate_example` | Regenerate example diff files |
| `make example` | Generate diffs and run the simple example |
| `make example_gen` | Generate diffs and run the random generation example |
| `make clean` | Remove binary and generated example diffs |

## Development

### Prerequisites

- Go 1.26 or later

### Running Tests

```bash
make test
```

### Regenerating Examples

```bash
make generate_example
```

### Adding a New Field Type

1. Add a new category in [`types/types.go`](types/types.go).
2. Create a new template in [`generator/templates/`](generator/templates/).
3. Update the generator to map the category to the template.
4. Add test cases in [`generator/generator_test.go`](generator/generator_test.go).

## Design Principles

- **Zero runtime dependencies** — all generated code is self-contained
- **Type safety** — all diff operations are fully typed at compile time
- **Minimal allocations** — diff structs only allocate for changed fields
- **Round-trip correctness** — `ApplyDiff(original, Patch(original, new)) == new`

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.