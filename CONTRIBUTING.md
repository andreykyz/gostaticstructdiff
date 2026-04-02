# Contributing to gostaticstructdiff

Thank you for your interest in contributing to `gostaticstructdiff`! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when participating in this project.

## How to Contribute

### Reporting Issues

If you encounter a bug or have a feature request, please open an issue on GitHub. Include as much detail as possible:

- Description of the problem or feature
- Steps to reproduce (for bugs)
- Expected behavior
- Actual behavior
- Environment (Go version, OS, etc.)

### Submitting Pull Requests

1. **Fork the repository** and create a new branch from `main`.
2. **Make your changes** following the coding standards (see below).
3. **Write tests** for new functionality or bug fixes.
4. **Ensure all tests pass** (`go test ./...`).
5. **Update documentation** if needed.
6. **Submit a pull request** with a clear description of the changes.

### Development Setup

1. Clone your fork:
   ```bash
   git clone https://github.com/your-username/gostaticstructdiff
   cd gostaticstructdiff
   ```

2. Build the tool:
   ```bash
   go build ./cmd/gostaticstructdiff
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

### Coding Standards

- Follow Go conventions (effective Go).
- Use `gofmt` for formatting.
- Run `go vet` and `staticcheck` to catch issues.
- Write meaningful commit messages.

### Testing

- Unit tests should cover new functionality.
- Integration tests should verify the tool works with example files.
- Golden tests are used for regression testing of generated code.

### Documentation

- Update README.md if the tool's behavior changes.
- Add comments to public functions and types.
- Keep the examples up to date.

## Project Structure

- `cmd/gostaticstructdiff/` – CLI entry point
- `internal/parser/` – AST parsing logic
- `internal/types/` – Type classification and diff strategies
- `internal/templates/` – Go templates for code generation
- `internal/generator/` – Code generation orchestration
- `examples/` – Example structs and generated code

## Release Process

Releases are managed by the maintainers. Version numbers follow semantic versioning.

## Questions?

Feel free to reach out by opening a discussion or contacting the maintainers.

Thank you for contributing!