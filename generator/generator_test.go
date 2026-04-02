package generator

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/andreykyz/gostaticstructdiff/parser"
)

func TestGenerate_SimpleStruct(t *testing.T) {
	// Create a mock StructInfo
	si := parser.StructInfo{
		Name: "User",
		Fields: []parser.FieldInfo{
			{
				Name:     "ID",
				Type:     "int",
				TypeExpr: &ast.Ident{Name: "int"},
			},
			{
				Name:     "Name",
				Type:     "string",
				TypeExpr: &ast.Ident{Name: "string"},
			},
		},
	}

	code, err := Generate([]parser.StructInfo{si}, "models", nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Basic sanity checks
	if !strings.Contains(code, "package models") {
		t.Error("generated code missing package declaration")
	}
	if !strings.Contains(code, "type UserDiff struct") {
		t.Error("generated code missing UserDiff struct")
	}
	if !strings.Contains(code, "ID struct") {
		t.Error("generated code missing ID field")
	}
	if !strings.Contains(code, "Name struct") {
		t.Error("generated code missing Name field")
	}
	if !strings.Contains(code, "func UserPatch") {
		t.Error("generated code missing UserPatch function")
	}
}

func TestGenerate_EmptyStructs(t *testing.T) {
	code, err := Generate([]parser.StructInfo{}, "empty", nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if !strings.Contains(code, "package empty") {
		t.Error("generated code missing package declaration")
	}
	// Should not contain any struct definitions
	if strings.Contains(code, "type") {
		t.Error("generated code should not contain type definitions for empty input")
	}
}

func TestGenerate_WithImports(t *testing.T) {
	si := parser.StructInfo{
		Name: "Item",
		Fields: []parser.FieldInfo{
			{
				Name:     "Value",
				Type:     "string",
				TypeExpr: &ast.Ident{Name: "string"},
			},
		},
	}
	imports := []string{"fmt", "time"}
	code, err := Generate([]parser.StructInfo{si}, "pkg", imports)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if !strings.Contains(code, "import (") {
		t.Error("generated code missing import block")
	}
	for _, imp := range imports {
		if !strings.Contains(code, "\""+imp+"\"") {
			t.Errorf("generated code missing import %q", imp)
		}
	}
}

func TestConvertToTemplateData(t *testing.T) {
	si := parser.StructInfo{
		Name: "Test",
		Fields: []parser.FieldInfo{
			{
				Name: "MapField",
				Type: "map[string]int",
				TypeExpr: &ast.MapType{
					Key:   &ast.Ident{Name: "string"},
					Value: &ast.Ident{Name: "int"},
				},
			},
		},
	}
	data := convertToTemplateData(si)
	if data.Name != "Test" {
		t.Errorf("expected Name Test, got %s", data.Name)
	}
	if len(data.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(data.Fields))
	}
	f := data.Fields[0]
	if f.Name != "MapField" {
		t.Errorf("expected field name MapField, got %s", f.Name)
	}
	if f.Category != "map" {
		t.Errorf("expected category map, got %s", f.Category)
	}
	if f.KeyType != "string" || f.ValueType != "int" {
		t.Errorf("expected key type string and value type int, got %s, %s", f.KeyType, f.ValueType)
	}
}
