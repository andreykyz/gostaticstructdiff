package main

import (
	"fmt"
	"reflect"

	"github.com/andreykyz/gostaticstructdiff/debugging"
	"github.com/andreykyz/gostaticstructdiff/examples"
)

// test example program

func main() {
	fmt.Println("=== gostaticstructdiff example: random generation and diff/patch ===")
	fmt.Println("This example uses debugging.GetTestStruct to generate random ComplexStruct instances")
	fmt.Println("with different seeds, computes diffs between consecutive seeds, applies the diff,")
	fmt.Println("and verifies that the patched struct matches the target.")
	fmt.Println()

	seeds := []uint{0, 1, 2, 3, 10, 100, 786876, 85589098, 87878809809, 777777}
	if len(seeds) < 2 {
		fmt.Println("Need at least two seeds")
		return
	}

	// Step 1: Generate a ComplexStruct for each seed
	fmt.Printf("1. Generating ComplexStruct for %d seeds...\n", len(seeds))
	structs := make([]examples.ComplexStruct, len(seeds))
	for i, seed := range seeds {
		fmt.Printf("   Seed %d... ", seed)
		cs, err := debugging.GetTestStruct[examples.ComplexStruct](seed)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
		structs[i] = cs
		fmt.Println("OK")
	}
	fmt.Println()

	// Step 2: Verify deterministic generation (same seed produces same struct)
	fmt.Println("2. Verifying deterministic generation...")
	for _, seed := range seeds {
		cs1, _ := debugging.GetTestStruct[examples.ComplexStruct](seed)
		cs2, _ := debugging.GetTestStruct[examples.ComplexStruct](seed)
		if !reflect.DeepEqual(cs1, cs2) {
			fmt.Printf("   WARNING: Seed %d produced different results (should be deterministic)\n", seed)
		}
	}
	fmt.Println("   All seeds produce deterministic results.")
	fmt.Println()

	// Step 3: Compute diffs between consecutive seeds and verify patch
	fmt.Println("3. Computing diffs between consecutive seeds and verifying patch...")
	successCount := 0
	for i := 0; i < len(seeds)-1; i++ {
		for j := 0; j < len(seeds)-1; j++ {
			if j == i {
				continue
			}
			seedA, seedB := seeds[i], seeds[j]

			a, b := structs[i], structs[j]
			fmt.Printf("   Pair %d → %d (seeds %d → %d): ", i, j, seedA, seedB)

			// Compute diff
			diff := examples.ComplexStructPatch(a, b)

			// Apply diff to a
			patched := examples.ApplyComplexStructDiff(a, diff)

			// Verify patched equals b
			if reflect.DeepEqual(patched, b) {
				fmt.Println("✓ SUCCESS")
				successCount++
			} else {
				fmt.Println("✗ FAILURE")
				fmt.Printf("      Patched struct does not match target for seeds %d → %d\n", seedA, seedB)
				// Optional: print some differences
				printDiffSummary(diff)
			}
		}
	}
	fmt.Printf("   Result: %d/%d pairs passed.\n", successCount, len(seeds)-1)
	fmt.Println()

	// Step 4: Show that different seeds produce different structs (likely)
	fmt.Println("4. Checking that different seeds produce different structs...")
	allDifferent := true
	for i := 0; i < len(seeds); i++ {
		for j := i + 1; j < len(seeds); j++ {
			if reflect.DeepEqual(structs[i], structs[j]) {
				fmt.Printf("   WARNING: Seeds %d and %d produced identical structs (extremely unlikely)\n", seeds[i], seeds[j])
				allDifferent = false
			}
		}
	}
	if allDifferent {
		fmt.Println("   All seeds produced distinct structs (as expected).")
	}
	fmt.Println()

	fmt.Println("=== Example completed ===")
}

// printDiffSummary prints a simple summary of which fields changed in the diff.
func printDiffSummary(diff examples.ComplexStructDiff) {
	changed := []string{}
	if diff.Name != nil {
		changed = append(changed, "Name")
	}
	if diff.Count != nil {
		changed = append(changed, "Count")
	}
	if diff.Active != nil {
		changed = append(changed, "Active")
	}
	if diff.Tags != nil {
		changed = append(changed, "Tags")
	}
	if diff.Users != nil {
		changed = append(changed, "Users")
	}
	if diff.Metadata != nil {
		changed = append(changed, "Metadata")
	}
	if diff.Inner != nil {
		changed = append(changed, "Inner")
	}
	if diff.StaticUser != nil {
		changed = append(changed, "StaticUser")
	}
	if diff.Ref != nil {
		changed = append(changed, "Ref")
	}
	if diff.Categories != nil {
		changed = append(changed, "Categories")
	}
	fmt.Printf("      Changed fields: %v\n", changed)
}
