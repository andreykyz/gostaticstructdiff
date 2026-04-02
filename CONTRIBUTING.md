# Contributing to gostaticstructdiff

Thank you for your interest in contributing to `gostaticstructdiff`! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We aim to foster an inclusive and welcoming community.

## Getting Started

### Prerequisites

- Go 1.26 or later
- Git
- Basic understanding of Go code generation concepts

### Development Setup

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/your-username/gostaticstructdiff
   cd gostaticstructdiff
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/andreykyz/gostaticstructdiff
   ```
4. **Install dependencies**:
   ```bash
   go mod download
   ```
5. **Build the tool**:
   ```bash
   go build -o gostaticstructdiff ./cmd/gostaticstructdiff
   ```

## Development Workflow

### 1. Create a Branch

Create a descriptive branch for your changes:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 2. Make Your Changes

Follow the project's coding standards:

- Use `gofmt` for formatting
- Follow Go naming conventions
- Write clear, descriptive commit messages
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./tests
```

### 4. Update Examples

If you change code generation patterns, update the examples:

```bash
# Regenerate example files
./scripts/regenerate-examples.sh
```

### 5. Commit Your Changes

Use descriptive commit messages:

```bash
git commit -m "feat: add support for time.Duration type"
```

Commit message conventions:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or fixes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### 6. Keep Your Branch Updated

Regularly sync with the upstream main branch:

```bash
git fetch upstream
git rebase upstream/main
```

### 7. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Project Structure

```
gostaticstructdiff/
├── cmd/gostaticstructdiff/     # CLI entry point
├── internal/                   # Internal packages
│   ├── parser/                # AST parsing and analysis
│   ├── generator/             # Code generation logic
│   ├── types/                 # Type system and strategies
│   └── templates/             # Go templates
├── examples/                  # Example code
├── doc/                      # Documentation
├── tests/                    # Test files
└── scripts/                  # Build and utility scripts
```

## Areas for Contribution

### High Priority

1. **New type support**: Add support for additional Go types
2. **Performance improvements**: Optimize diff computation
3. **Bug fixes**: Address issues reported in the tracker
4. **Documentation improvements**: Clarify usage, add examples

### Feature Ideas

1. **Custom diff strategies**: Allow users to define custom diff logic
2. **Serialization support**: Generate JSON/protobuf marshaling for diffs
3. **Validation generation**: Generate validation functions for diffs
4. **Plugin system**: Extensible architecture for custom generators

## Testing Guidelines

### Unit Tests

Each package should have comprehensive unit tests:

```go
func TestParserExtractStructs(t *testing.T) {
    // Test setup
    parser := NewParser()
    
    // Test execution
    structs, err := parser.Extract("testdata/simple.go")
    
    // Assertions
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(structs) != 1 {
        t.Errorf("expected 1 struct, got %d", len(structs))
    }
}
```

### Integration Tests

Test the end-to-end generation pipeline:

```go
func TestGenerateUserDiff(t *testing.T) {
    generator := NewGenerator()
    output, err := generator.Generate("testdata/user.go", "User")
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify generated code compiles
    if err := compileGeneratedCode(output); err != nil {
        t.Errorf("generated code doesn't compile: %v", err)
    }
}
```

### Golden File Tests

Use golden files for generated code:

```go
func TestGoldenUserDiff(t *testing.T) {
    generator := NewGenerator()
    output, err := generator.Generate("testdata/user.go", "User")
    if err != nil {
        t.Fatal(err)
    }
    
    golden := readGoldenFile("testdata/user_diff.golden.go")
    if output != golden {
        t.Errorf("generated code doesn't match golden file")
        
        // Update golden file if flag is set
        if *updateGolden {
            writeGoldenFile("testdata/user_diff.golden.go", output)
        }
    }
}
```

## Code Review Process

### Pull Request Checklist

Before submitting a PR, ensure:

- [ ] Code follows Go conventions and is formatted with `gofmt`
- [ ] Tests pass (`go test ./...`)
- [ ] No linting errors (`go vet ./...`, `staticcheck`)
- [ ] Documentation is updated
- [ ] Examples are updated if generation patterns changed
- [ ] Commit messages are clear and descriptive
- [ ] Changes are focused and minimal

### Review Guidelines

Reviewers should:

1. **Check functionality**: Does the change work as intended?
2. **Verify tests**: Are there adequate tests for the change?
3. **Assess performance**: Does the change impact performance?
4. **Review documentation**: Is the documentation updated?
5. **Consider edge cases**: Are edge cases handled properly?
6. **Evaluate design**: Is the implementation clean and maintainable?

## Documentation

### Updating Documentation

When making changes that affect users:

1. **Update user documentation** in `doc/user-guide.md`
2. **Update API reference** in `doc/api-reference.md` if APIs change
3. **Add examples** if introducing new features
4. **Update README.md** for significant changes

### Adding New Documentation

New documentation should:

- Be placed in the `doc/` directory
- Use clear, concise language
- Include examples where helpful
- Follow the existing markdown style

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backward compatible)
- **PATCH** version for bug fixes (backward compatible)

### Release Steps

1. **Update version** in `cmd/gostaticstructdiff/version.go`
2. **Update CHANGELOG.md** with release notes
3. **Create release tag**:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
4. **Create GitHub release** with release notes
5. **Update documentation** if needed

## Getting Help

### Questions and Discussions

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code review**: Ask questions in your PR comments

### Finding Issues to Work On

Check the GitHub issue tracker for:
- Issues labeled `good-first-issue`
- Issues labeled `help-wanted`
- Bugs reported by users

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).

## Acknowledgments

Thank you for contributing to `gostaticstructdiff`! Your efforts help make this tool better for everyone in the Go community.