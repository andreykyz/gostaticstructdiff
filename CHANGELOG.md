# Changelog

All notable changes to `gostaticstructdiff` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of the CLI tool
- Support for basic types (int, string, bool, float64)
- Support for pointer, slice, map, and nested struct fields
- AST parser for extracting structs with `structtomap` tags
- Type classification and diff strategy selection
- Template-based code generation
- Patch functions for computing and applying diffs
- Example files demonstrating usage
- Comprehensive documentation

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- N/A

### Security
- N/A

## [0.1.0] - 2026-04-02

### Added
- First release of `gostaticstructdiff`
- Basic functionality as described above

### Notes
This is an initial alpha release. The tool is functional but may have limitations with complex nested structures. Feedback and contributions are welcome.