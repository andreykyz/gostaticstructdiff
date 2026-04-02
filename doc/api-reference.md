# gostaticstructdiff API Reference

## Package Overview

The `gostaticstructdiff` package provides both a command-line interface and programmatic API for generating diff structures and patch functions from Go structs.

## Command Line Interface

### `gostaticstructdiff` Command

```bash
gostaticstructdiff [options]
```

#### Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-input` | string | (required) | Input Go source file path |
| `-output` | string | `<input>_diff.go` | Output file for generated code |
| `-struct` | string | (all) | Specific struct name to generate (default: all structs) |
| `-package` | string | (auto) | Package name for generated code |
| `-verbose` | bool | false | Enable verbose logging |
| `-version` | bool | false | Show version information |
| `-help` | bool | false | Show help message |

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | File I/O error |
| 4 | Code generation error |

#### Examples

```bash
# Basic usage
gostaticstructdiff -input models/user.go -output models/user_diff.go

# Generate for specific struct only
gostaticstructdiff -input models.go -output models_diff.go -struct User

# Verbose mode for debugging
gostaticstructdiff -input api/types.go -output api/types_diff.go -verbose
```

## Generated Code API

### Diff Structures

For each struct `T` with `structtomap` tags, a corresponding `TDiff` type is generated.

#### Field Patterns

**Scalar Fields** (int, string, bool, float64, etc.):

```go
FieldName struct {
    Value T
    Set   bool
}
```

**Pointer Fields** (*T):

```go
FieldName struct {
    Value *T
    Set   bool
}
```

**Slice Fields** ([]T):

```go
FieldName struct {
    Value []T
    Set   bool
}
```

**Map Fields** (map[K]V):

```go
FieldName struct {
    Add map[K]V
    Del map[K]struct{}
    Set bool
}
```

**Map Fields with Struct Values** (map[K]StructType):

```go
FieldName struct {
    Add map[K]StructType
    Del map[K]struct{}
    Mod map[K]StructTypeDiff
    Set bool
}
```

**Nested Struct Fields** (StructType):

```go
FieldName struct {
    Value StructTypeDiff
    Set   bool
}
```

**Inline Struct Fields**:

```go
FieldName struct {
    Value struct {
        NestedField1 struct {
            Value T1
            Set   bool
        }
        NestedField2 struct {
            Value T2
            Set   bool
        }
    }
    Set bool
}
```

### Patch Functions

Two patch functions are generated for each struct:

#### `TPatch(original, new T) TDiff`

Computes the difference between two struct instances.

**Parameters:**
- `original T`: The original struct instance
- `new T`: The new struct instance

**Returns:**
- `TDiff`: A diff structure describing changes from `original` to `new`

**Behavior:**
- Fields with equal values have `Set: false`
- Fields with different values have `Set: true` and `Value` set to the new value
- For maps: `Add` contains new or modified entries, `Del` contains removed entries
- For nested structs: Recursively computes diffs

**Example:**
```go
original := User{ID: 1, Name: "Alice"}
updated := User{ID: 1, Name: "Bob"}
diff := UserPatch(original, updated)
// diff.Name.Set == true, diff.Name.Value == "Bob"
// diff.ID.Set == false
```

#### `TPatch(original T, diff TDiff) T`

Applies a diff to a struct instance.

**Parameters:**
- `original T`: The original struct instance
- `diff TDiff`: The diff to apply

**Returns:**
- `T`: A new struct with the diff applied

**Behavior:**
- Fields with `Set: false` are copied from `original`
- Fields with `Set: true` take the value from `diff.Value`
- For maps: `Add` entries are added/updated, `Del` entries are removed
- Returns a new struct; does not modify `original`

**Example:**
```go
original := User{ID: 1, Name: "Alice"}
diff := UserDiff{Name: struct{Value string; Set bool}{Value: "Bob", Set: true}}
patched := UserPatch(original, diff)
// patched == User{ID: 1, Name: "Bob"}
```

### Zero Value Handling

The diff system distinguishes between:
- `Set: false` - field was not changed (keep original value)
- `Set: true` with zero `Value` - field was explicitly set to zero value

```go
// Differentiating cases:
diff1 := UserDiff{Name: struct{Value string; Set bool}{Value: "", Set: true}}
// Name was explicitly set to empty string

diff2 := UserDiff{Name: struct{Value string; Set bool}{Value: "", Set: false}}
// Name was not changed (keep whatever value original had)
```

## Programmatic API

### Generator Interface

While primarily a CLI tool, `gostaticstructdiff` can be used programmatically:

```go
import "github.com/andreykyz/gostaticstructdiff/internal/generator"

func GenerateFromFile(inputFile, outputFile string) error {
    cfg := generator.Config{
        InputFile:  inputFile,
        OutputFile: outputFile,
        StructName: "", // empty for all structs
        Verbose:    false,
    }
    
    gen := generator.New(cfg)
    return gen.Generate()
}
```

### Configuration Types

```go
package generator

type Config struct {
    InputFile  string
    OutputFile string
    StructName string
    Package    string
    Verbose    bool
}

type Generator interface {
    Generate() error
    GenerateForStruct(structName string) (string, error)
}
```

### Error Types

```go
package gostaticstructdiff

type ErrorCode int

const (
    ErrFileNotFound ErrorCode = iota
    ErrParseFailed
    ErrNoStructsFound
    ErrGenerationFailed
    ErrWriteFailed
)

type GenerationError struct {
    Code    ErrorCode
    Message string
    File    string
    Line    int
}

func (e *GenerationError) Error() string {
    return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
}
```

## Type Support Matrix

| Go Type | Supported | Diff Pattern | Notes |
|---------|-----------|--------------|-------|
| `int`, `int8-64` | ✅ | Scalar | |
| `uint`, `uint8-64` | ✅ | Scalar | |
| `float32`, `float64` | ✅ | Scalar | |
| `string` | ✅ | Scalar | |
| `bool` | ✅ | Scalar | |
| `complex64`, `complex128` | ⚠️ | Scalar | May need custom handling |
| `time.Time` | ✅ | Scalar | Treated as struct in generated code |
| `*T` (pointer) | ✅ | Pointer | Nil handling supported |
| `[]T` (slice) | ✅ | Slice | Whole slice replacement |
| `[N]T` (array) | ⚠️ | Slice-like | Fixed size preserved |
| `map[K]V` | ✅ | Map | Add/Del operations |
| `map[K]Struct` | ✅ | Map with Mod | Recursive diff for values |
| `struct{}` | ✅ | Nested | Recursive generation |
| `interface{}` | ❌ | Not supported | Cannot generate type-safe diff |
| `chan T` | ❌ | Not supported | |
| `func()` | ❌ | Not supported | |

## Template Functions

Generated code includes helper functions:

### `isZeroValue(v interface{}) bool`

Checks if a value is the zero value for its type.

### `deepEqual(a, b interface{}) bool`

Performs deep equality check, handling nested structs and maps.

### `mergeMaps(base, add, del map[K]V) map[K]V`

Helper for applying map diffs.

## Constants and Variables

Generated files may include:

```go
const (
    _generatedBy = "gostaticstructdiff"
    _version     = "1.0.0"
)

var (
    // Type assertions for compile-time safety
    _ = func() { var _ interface{} = (*UserDiff)(nil) }
)
```

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| `TPatch(original, new)` | O(N + M) | N = field count, M = map/slice elements |
| `TPatch(original, diff)` | O(N + M) | |
| Memory allocation | O(N + M) | Proportional to diff size |

### Memory Usage

- Diff structs are approximately 2-3x larger than original structs
- Map diffs store copies of added/modified entries
- No shared references with original data

## Compatibility Notes

### Go Version Compatibility

- Requires Go 1.26+ for code generation
- Generated code compatible with Go 1.19+
- Uses standard library only, no external dependencies

### Backward Compatibility

- Adding fields to source structs: ✅ Compatible
- Removing fields from source structs: ⚠️ May break existing diffs
- Changing field types: ❌ Breaking change
- Renaming fields: ❌ Breaking change

### Serialization Formats

Generated diff structs can be serialized to:

- **JSON**: Using `encoding/json` (struct tags may be needed)
- **Protocol Buffers**: Requires additional .proto definitions
- **YAML/TOML**: Using appropriate marshallers

## Examples

See the [examples directory](../examples/) for complete generated code examples.

## See Also

- [User Guide](./user-guide.md) for usage instructions
- [Developer Guide](./developer-guide.md) for implementation details
- [Tutorial](./tutorial.md) for step-by-step examples