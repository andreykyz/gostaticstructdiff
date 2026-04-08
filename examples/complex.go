package examples

import (
	"github.com/andreykyz/gostaticstructdiff/examples/models"
)

// ComplexStruct demonstrates a variety of field types and nesting.
type ComplexStruct struct {
	// Basic fields
	Name   string `structtomap:"name"`
	Count  int    `structtomap:"count"`
	Active bool   `structtomap:"active"`
	// Slice of basic types
	Tags []string `structtomap:"tags"`
	// Slice of structs from another package
	Users []models.User `structtomap:"users"`
	// Map of string to Metadata struct (nested)
	Metadata map[string]models.Metadata `structtomap:"metadata"`
	// Nested struct defined inline
	Inner struct {
		Title string `structtomap:"title"`
		Value int    `structtomap:"value"`
	} `structtomap:"inner"`
	StaticUser models.User `structtomap:"static_user"`
	// Pointer to a struct (may be nil)
	Ref *models.User `structtomap:"ref"`
	// Map of string to slice of strings
	Categories map[string][]string `structtomap:"categories"`
}
