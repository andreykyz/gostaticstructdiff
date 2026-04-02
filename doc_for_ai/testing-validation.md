# Testing and Validation Guidelines for AI Implementation

## Overview

This document provides testing and validation guidelines for AI agents implementing the `gostaticstructdiff` tool. It defines what constitutes a successful implementation and how to verify correctness at each stage.

## Testing Philosophy

### Test-Driven Development Approach
AI agents should follow a test-driven approach:
1. Write tests for expected behavior first
2. Implement functionality to pass tests
3. Refactor while keeping tests passing
4. Add more tests for edge cases

### Validation Levels
1. **Unit Tests**: Individual components (parser, generator, etc.)
2. **Integration Tests**: Component interactions
3. **End-to-End Tests**: Full CLI workflow
4. **Golden Tests**: Compare generated output with expected output

## Test Categories

### 1. Parser Tests
**Objective**: Verify the AST parser correctly extracts struct information.

**Test Cases**:
- Parse simple struct with `structtomap` tags
- Parse struct without tags (should be ignored)
- Parse nested structs
- Parse embedded structs
- Parse pointer, slice, and map fields
- Handle package imports
- Error handling for invalid Go syntax

**Validation Criteria**:
- Correct struct names extracted
- Correct field names and types identified
- Tags properly parsed
- Nested structures handled recursively
- No panics on invalid input

### 2. Type System Tests
**Objective**: Verify type classification and diff strategy selection.

**Test Cases**:
- Classify basic types (int, string, bool, float64)
- Classify pointer types (*T)
- Classify slice types ([]T)
- Classify map types (map[K]V)
- Classify struct types
- Handle type aliases
- Handle imported types

**Validation Criteria**:
- Correct classification for all Go types
- Appropriate diff strategy selected
- Matches patterns in example files

### 3. Template Tests
**Objective**: Verify templates generate correct code.

**Test Cases**:
- Basic field template generates `struct { Value T; Set bool }`
- Pointer field template generates `struct { Value *T; Set bool }`
- Slice field template generates `struct { Value []T; Set bool }`
- Map field template generates `struct { Add map[K]V; Del map[K]struct{}; Set bool }`
- Struct field template generates recursive diff
- Full struct template includes all fields
- Patch function templates generate correct signatures

**Validation Criteria**:
- Generated code matches example patterns
- Proper Go syntax
- Correct imports
- No template rendering errors

### 4. Generator Tests
**Objective**: Verify the generator integrates components correctly.

**Test Cases**:
- Generate diff for simple User struct (match example)
- Generate diff for Metadata struct (match example)
- Generate diff for ComplexStruct (match example)
- Handle multiple structs in one file
- Generate proper package declaration
- Include necessary imports
- Format code correctly

**Validation Criteria**:
- Generated code identical to example diff files
- Code compiles without errors
- All required components present

### 5. Patch Function Tests
**Objective**: Verify patch functions work correctly.

**Test Cases**:
- `UserPatch(original, new)` computes correct diff
- `UserPatch(original, diff)` applies diff correctly
- Round-trip: `Patch(Patch(original, new)) == new`
- Handle zero values correctly
- Handle nil pointers
- Handle empty maps and slices
- Handle nested struct diffs
- Performance: no excessive allocations

**Validation Criteria**:
- Diff computation is correct
- Diff application reconstructs target
- Round-trip property holds
- No panics on edge cases

### 6. CLI Tests
**Objective**: Verify command-line interface works correctly.

**Test Cases**:
- `-help` flag shows usage
- `-version` flag shows version
- `-input` with valid file generates output
- `-output` flag changes output location
- `-struct` flag filters specific structs
- Error handling for missing input file
- Error handling for invalid Go file
- Verbose mode shows progress

**Validation Criteria**:
- CLI responds correctly to all flags
- Appropriate error messages
- Files created in correct locations
- Exit codes correct (0 for success, non-zero for errors)

### 7. Integration Tests
**Objective**: Verify the complete workflow.

**Test Cases**:
- Run tool on example files, compare with existing diff files
- Compile generated code
- Use generated code in sample program
- Test with `go generate` integration
- Test with various real-world struct patterns

**Validation Criteria**:
- End-to-end workflow succeeds
- Generated code is usable
- No regressions from examples

## Golden File Testing

### Purpose
Golden files store expected output for regression testing. When the generator changes, golden tests verify output hasn't regressed.

### Implementation
```go
func TestGoldenUserDiff(t *testing.T) {
    // Generate code
    generated := generateCode("testdata/user.go", "User")
    
    // Read golden file
    golden, err := os.ReadFile("testdata/user_diff.golden.go")
    if err != nil {
        t.Fatal(err)
    }
    
    // Compare
    if string(generated) != string(golden) {
        t.Errorf("Generated code doesn't match golden file")
        
        // Update golden file if requested
        if *updateFlag {
            os.WriteFile("testdata/user_diff.golden.go", generated, 0644)
        }
    }
}
```

### Golden Files to Create
1. `user_diff.golden.go` - From `examples/models/user.go`
2. `metadata_diff.golden.go` - From `examples/models/metadata.go`
3. `complex_diff.golden.go` - From `examples/complex.go`

## Property-Based Testing

### Round-Trip Property
For all structs S and values a, b of type S:
```go
diff := SPatch(a, b)
c := SPatch(a, diff)
assert(c == b)
```

### Identity Property
For all structs S and values a of type S:
```go
diff := SPatch(a, a)
// All fields should have Set: false
```

### Composition Property
For all structs S and values a, b, c of type S:
```go
diff1 := SPatch(a, b)
diff2 := SPatch(b, c)
combined := mergeDiffs(diff1, diff2)
result := SPatch(a, combined)
assert(result == c)
```

## Performance Testing

### Benchmarks to Implement
```go
func BenchmarkParseFile(b *testing.B) {
    for i := 0; i < b.N; i++ {
        parseFile("testdata/large.go")
    }
}

func BenchmarkGenerateDiff(b *testing.B) {
    structInfo := parseFile("testdata/large.go")
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        generateCode(structInfo)
    }
}

func BenchmarkPatchComputation(b *testing.B) {
    a := createTestStruct()
    b := createTestStruct()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = TestStructPatch(a, b)
    }
}
```

### Performance Criteria
- Parse medium struct (10 fields): < 10ms
- Generate diff for medium struct: < 50ms
- Compute patch for medium struct: < 100µs
- Memory allocations: minimal, no leaks

## Edge Case Testing

### Input Validation
- Empty input file
- File with no structs
- File with syntax errors
- Struct with no `structtomap` tags
- Very large struct (100+ fields)
- Deeply nested structs (10+ levels)

### Type Edge Cases
- `interface{}` type (should error or skip)
- `chan T` type (should error or skip)
- `func()` type (should error or skip)
- Unsupported types from other packages
- Circular type references
- Type aliases

### Value Edge Cases
- Zero values vs unset fields
- Nil pointers vs zero values
- Empty maps vs nil maps
- Empty slices vs nil slices
- NaN float values
- Time zero value

## Validation Checklist

### Before Each Commit
- [ ] All unit tests pass (`go test ./...`)
- [ ] No linting errors (`go vet ./...`, `staticcheck`)
- [ ] Code formatted (`gofmt -d .` shows no differences)
- [ ] Golden tests pass (or updated intentionally)
- [ ] Integration tests pass
- [ ] Benchmarks show no regressions

### Before Final Submission
- [ ] All example files generate matching output
- [ ] Generated code compiles without errors
- [ ] Patch functions work correctly
- [ ] CLI interface is user-friendly
- [ ] Documentation is complete and accurate
- [ ] No panics in normal operation
- [ ] Memory usage is reasonable
- [ ] Performance meets criteria

## Test Data

### Sample Test Files to Create
1. `testdata/simple.go` - Basic struct with all field types
2. `testdata/nested.go` - Deeply nested structs
3. `testdata/large.go` - Struct with many fields
4. `testdata/edge_cases.go` - Edge case scenarios
5. `testdata/real_world.go` - Real-world use case examples

### Expected Output Files
For each test file, create a corresponding `.golden.go` file with expected output.

## Continuous Integration

### Suggested CI Pipeline
1. **Lint**: Run `gofmt`, `go vet`, `staticcheck`
2. **Unit Tests**: Run `go test ./... -short`
3. **Integration Tests**: Run `go test -tags=integration`
4. **Golden Tests**: Compare generated output
5. **Benchmarks**: Run performance tests
6. **Build Verification**: Ensure tool builds successfully

### CI Success Criteria
- All tests pass
- No linting errors
- Golden files match
- Build succeeds
- No performance regressions

## Debugging Tips

### Common Issues and Solutions
1. **Generated code doesn't compile**: Check imports, syntax errors
2. **Parser misses fields**: Verify tag parsing logic
3. **Incorrect diff structure**: Compare with example patterns
4. **Performance problems**: Profile memory allocations
5. **Edge case failures**: Add specific tests for edge cases

### Debugging Tools
- Use `-verbose` flag for detailed logging
- Add debug prints to trace execution
- Use Go's `pprof` for performance analysis
- Compare generated output with expected line-by-line

## Final Validation Report

When implementation is complete, generate a validation report:

1. **Test Coverage**: Percentage of code covered
2. **Performance Metrics**: Benchmarks results
3. **Example Verification**: All examples generate correct output
4. **Edge Case Handling**: All edge cases tested
5. **User Experience**: CLI is intuitive and helpful
6. **Code Quality**: Follows Go best practices

## Conclusion

Thorough testing and validation are critical for a successful implementation. Follow these guidelines to ensure the `gostaticstructdiff` tool is robust, correct, and production-ready.