package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_SimpleStruct(t *testing.T) {
	// Create a temporary Go file with a simple struct
	content := `package test

type User struct {
	ID   int    ` + "`structtomap:\"id\"`" + `
	Name string ` + "`structtomap:\"name\"`" + `
	Age  int    // no tag
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	structs, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if s.Name != "User" {
		t.Errorf("expected struct name User, got %s", s.Name)
	}
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields with structtomap tags, got %d", len(s.Fields))
	}

	// Check first field
	if s.Fields[0].Name != "ID" {
		t.Errorf("expected field name ID, got %s", s.Fields[0].Name)
	}
	if s.Fields[0].Type != "int" {
		t.Errorf("expected field type int, got %s", s.Fields[0].Type)
	}
	if s.Fields[0].Tag != `structtomap:"id"` {
		t.Errorf("expected tag structtomap:\"id\", got %s", s.Fields[0].Tag)
	}

	// Check second field
	if s.Fields[1].Name != "Name" {
		t.Errorf("expected field name Name, got %s", s.Fields[1].Name)
	}
	if s.Fields[1].Type != "string" {
		t.Errorf("expected field type string, got %s", s.Fields[1].Type)
	}
	if s.Fields[1].Tag != `structtomap:"name"` {
		t.Errorf("expected tag structtomap:\"name\", got %s", s.Fields[1].Tag)
	}
}

func TestParseFile_NoStructtomapTags(t *testing.T) {
	content := `package test

type Foo struct {
	X int
	Y string
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	structs, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(structs) != 0 {
		t.Errorf("expected 0 structs, got %d", len(structs))
	}
}

func TestParseFile_MultipleStructs(t *testing.T) {
	content := `package test

type A struct {
	Field1 string ` + "`structtomap:\"field1\"`" + `
}

type B struct {
	Field2 int ` + "`structtomap:\"field2\"`" + `
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	structs, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(structs))
	}
	// order may be as defined
	names := map[string]bool{}
	for _, s := range structs {
		names[s.Name] = true
	}
	if !names["A"] || !names["B"] {
		t.Errorf("expected structs A and B, got %v", names)
	}
}

func TestParseFile_InvalidFile(t *testing.T) {
	_, err := ParseFile("nonexistent.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
