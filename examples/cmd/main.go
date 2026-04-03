package main

import (
	"fmt"

	"github.com/andreykyz/gostaticstructdiff/examples"
	"github.com/andreykyz/gostaticstructdiff/examples/models"
)

func main() {
	fmt.Println("=== gostaticstructdiff example: diff/patch demonstration ===")

	// Create original ComplexStruct
	original := examples.ComplexStruct{
		Name:   "Original",
		Count:  10,
		Active: true,
		Tags:   []string{"tag1", "tag2"},
		Users: []models.User{
			{ID: 1, Username: "user1", Email: "user1@example.com", Active: true},
			{ID: 2, Username: "user2", Email: "user2@example.com", Active: false},
		},
		Metadata: map[string]models.Metadata{
			"meta1": {
				Label: "First metadata",
				Values: map[string]string{
					"key1": "value1",
				},
			},
		},
		Inner: struct {
			Title string `structtomap:"title"`
			Value int    `structtomap:"value"`
		}{
			Title: "Inner Title",
			Value: 100,
		},
		Ref: &models.User{ID: 99, Username: "refuser", Email: "ref@example.com", Active: true},
		Categories: map[string][]string{
			"cat1": {"item1", "item2"},
			"cat2": {"item3"},
		},
	}

	// Create modified version
	modified := original
	modified.Name = "Modified"
	modified.Count = 20
	modified.Tags = append(modified.Tags, "tag3")
	modified.Users[0].Username = "updated_user1"
	modified.Metadata["meta2"] = models.Metadata{
		Label:  "Second metadata",
		Values: map[string]string{"key2": "value2"},
	}
	delete(modified.Categories, "cat2")
	modified.Categories["cat1"][0] = "updated_item1"

	fmt.Println("\n1. Original ComplexStruct:")
	printComplexStruct(&original)

	fmt.Println("\n2. Modified ComplexStruct:")
	printComplexStruct(&modified)

	// Compute diff
	fmt.Println("\n3. Computing diff using ComplexStructPatch...")
	diff := examples.ComplexStructPatch(original, modified)
	fmt.Printf("Diff computed: %+v\n", summarizeDiff(diff))

	// Apply diff to original to get patched version
	fmt.Println("\n4. Applying diff to original using ApplyComplexStructDiff...")
	patched := examples.ApplyComplexStructDiff(original, diff)
	fmt.Println("Patched ComplexStruct:")
	printComplexStruct(&patched)

	// Verify that patched equals modified
	fmt.Println("\n5. Verification:")
	if deepEqualComplexStruct(patched, modified) {
		fmt.Println("✓ SUCCESS: Patched struct equals modified struct")
	} else {
		fmt.Println("✗ FAILURE: Patched struct differs from modified struct")
	}

	fmt.Println("\n=== Example completed ===")
}

// printComplexStruct prints a simplified view of ComplexStruct
func printComplexStruct(cs *examples.ComplexStruct) {
	fmt.Printf("  Name: %s\n", cs.Name)
	fmt.Printf("  Count: %d\n", cs.Count)
	fmt.Printf("  Active: %v\n", cs.Active)
	fmt.Printf("  Tags: %v\n", cs.Tags)
	fmt.Printf("  Users: %d users\n", len(cs.Users))
	if len(cs.Users) > 0 {
		fmt.Printf("    First user: %s\n", cs.Users[0].Username)
	}
	fmt.Printf("  Metadata: %d entries\n", len(cs.Metadata))
	fmt.Printf("  Inner.Title: %s\n", cs.Inner.Title)
	fmt.Printf("  Inner.Value: %d\n", cs.Inner.Value)
	if cs.Ref != nil {
		fmt.Printf("  Ref: %s\n", cs.Ref.Username)
	} else {
		fmt.Printf("  Ref: nil\n")
	}
	fmt.Printf("  Categories: %d categories\n", len(cs.Categories))
	for k, v := range cs.Categories {
		fmt.Printf("    %s: %v\n", k, v)
	}
}

// summarizeDiff provides a human-readable summary of the diff
func summarizeDiff(diff examples.ComplexStructDiff) string {
	changes := []string{}
	if diff.Name.Set {
		changes = append(changes, "Name")
	}
	if diff.Count.Set {
		changes = append(changes, "Count")
	}
	if diff.Active.Set {
		changes = append(changes, "Active")
	}
	if diff.Tags.Set {
		changes = append(changes, "Tags")
	}
	if diff.Users.Set {
		changes = append(changes, "Users")
	}
	if diff.Metadata.Set {
		changes = append(changes, fmt.Sprintf("Metadata(%d added, %d deleted)", len(diff.Metadata.Add), len(diff.Metadata.Del)))
	}
	if diff.Inner.Set {
		changes = append(changes, "Inner")
	}
	if diff.Ref.Set {
		changes = append(changes, "Ref")
	}
	if diff.Categories.Set {
		changes = append(changes, fmt.Sprintf("Categories(%d added, %d deleted)", len(diff.Categories.Add), len(diff.Categories.Del)))
	}
	if len(changes) == 0 {
		return "No changes"
	}
	return fmt.Sprintf("Changed fields: %v", changes)
}

// deepEqualComplexStruct performs a simple equality check (simplified for demonstration)
func deepEqualComplexStruct(a, b examples.ComplexStruct) bool {
	if a.Name != b.Name || a.Count != b.Count || a.Active != b.Active {
		return false
	}
	if len(a.Tags) != len(b.Tags) {
		return false
	}
	for i := range a.Tags {
		if a.Tags[i] != b.Tags[i] {
			return false
		}
	}
	if len(a.Users) != len(b.Users) {
		return false
	}
	for i := range a.Users {
		if a.Users[i] != b.Users[i] {
			return false
		}
	}
	if len(a.Metadata) != len(b.Metadata) {
		return false
	}
	// Note: For simplicity, we skip deep comparison of Metadata and Categories
	// In a real scenario, you'd want proper equality checks
	return true
}
