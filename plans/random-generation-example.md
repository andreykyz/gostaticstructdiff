# Plan: Random Generation Example

## Objective
Create an example program in `examples/cmd/gen/main.go` that demonstrates using `debugging.GetTestStruct` to generate random `ComplexStruct` instances with different seeds, compute diffs between consecutive seeds, apply patches, and verify correctness.

## Requirements
- Use at least 10 generations with seed values: 0, 1, 2, 3, 10, 100, 786876, 85589098, 87878809809, 777777.
- Compute diff between each consecutive pair (seed i → seed i+1).
- Apply diff to the first struct and verify that the patched struct equals the second struct.
- Verify deterministic generation (same seed produces same struct).
- Show that different seeds produce different structs (extremely unlikely to collide).

## Design

### File Structure
```
examples/cmd/gen/main.go
```

### Program Outline

1. **Imports**:
   - `debugging` – for `GetTestStruct`
   - `examples` – for `ComplexStruct`, `ComplexStructPatch`, `ApplyComplexStructDiff`
   - `fmt`, `reflect`

2. **Main Steps**:
   - Define seed list.
   - Generate a `ComplexStruct` for each seed, store in slice.
   - Verify deterministic generation by generating each seed twice and comparing.
   - For each consecutive pair:
     - Compute diff using `examples.ComplexStructPatch`.
     - Apply diff using `examples.ApplyComplexStructDiff`.
     - Verify equality using `reflect.DeepEqual`.
     - Print success/failure and summary of changed fields.
   - Check that all seeds produce distinct structs (optional warning if collision).
   - Print final statistics.

3. **Helper Functions**:
   - `printDiffSummary` – prints which fields changed in a diff.

### Key Considerations
- The generation is deterministic across runs (same seed → same struct). This is verified.
- Diff between random structs will likely be large; we still demonstrate that patch works correctly.
- The example will output clear progress and results to the console.

## Implementation Steps

1. Create directory `examples/cmd/gen/`.
2. Write `main.go` with the outlined program.
3. Test the program by running it (ensure no compilation errors).
4. Optionally update documentation (README) to mention the new example.

## Dependencies
- The `debugging` package depends on `github.com/trailofbits/go-fuzz-utils`. This is already in `go.mod`.
- The `examples` package must have generated diff functions (`complex_diff.go`). This is already present.

## Verification
- Run the program and confirm all pairs pass verification (should pass because diff/patch is correct).
- Ensure deterministic generation passes (should pass).
- Ensure no collisions between different seeds (should pass; if collision occurs, it's extremely rare but we'll warn).

## Next Steps
After approval, switch to Code mode to implement the plan.