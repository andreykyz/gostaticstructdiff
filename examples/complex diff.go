package examples

import (
	"github.com/andreykyz/gostaticstructdiff/examples/models"
)

// ComplexStruct demonstrates a variety of field types and nesting.
type ComplexStructDiff struct {
	Name struct {
		Value string
		Set   bool
	}
	Count struct {
		Value int
		Set   bool
	}
	Active struct {
		Value bool
		Set   bool
	}
	Tags struct {
		Value []string
		Set   bool
	}
	Users struct {
		Value []models.UserDiff
		Set   bool
	}

	// Map of string to Metadata struct (nested)
	Metadata struct {
		Value map[string]models.MetadataDiff
		Set   bool
	}
	// Nested struct defined inline
	Inner struct {
		Value struct {
			Title struct {
				Value string
				Set   bool
			}
			Value int `structtomap:"value"`
		}
		Set bool
	} `structtomap:"inner"`
	// Pointer to a struct (may be nil)
	Ref *models.User `structtomap:"ref"`
	// Map of string to slice of strings
	Categories map[string][]string `structtomap:"categories"`
}
