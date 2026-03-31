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
		Add map[string]models.Metadata
		Del map[string]struct{}
		Mod map[string]models.MetadataDiff
		Set bool
	}
	// Nested struct defined inline
	Inner struct {
		Value struct {
			Title struct {
				Value string
				Set   bool
			}
			Value struct {
				Value int
				Set   bool
			}
		}
		Set bool
	}
	Ref struct {
		Value *models.UserDiff
		Set   bool
	}
	// Map of string to slice of strings
	Categories struct {
		Add map[string][]string
		Del map[string]struct{}
		Set bool
	}
	UserCategories struct {
		Add map[string][]models.UserDiff
		Del map[string]struct{}
		Set bool
	}
}
