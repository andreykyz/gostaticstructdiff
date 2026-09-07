package generator

import (
	"go/ast"
	"os"
	"path/filepath"
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

	code, err := Generate([]parser.StructInfo{si}, "models", nil, "0.1.0", nil, false, "")
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
	if !strings.Contains(code, "ID *struct") {
		t.Error("generated code missing ID field")
	}
	if !strings.Contains(code, "Name *struct") {
		t.Error("generated code missing Name field")
	}
	if !strings.Contains(code, "func UserPatch") {
		t.Error("generated code missing UserPatch function")
	}
	if !strings.Contains(code, "func (d UserDiff) IsEmpty() bool") {
		t.Error("generated code missing UserDiff.IsEmpty method")
	}
	if !strings.Contains(code, "d.ID == nil &&\n\t\td.Name == nil") {
		t.Error("generated code missing IsEmpty body with field nil checks")
	}
}

func TestGenerate_EmptyStructs(t *testing.T) {
	code, err := Generate([]parser.StructInfo{}, "empty", nil, "0.1.0", nil, false, "")
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
	code, err := Generate([]parser.StructInfo{si}, "pkg", imports, "0.1.0", nil, false, "")
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
	data := convertToTemplateData(si, nil, nil, "")
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

func TestGenerate_WithSuffix(t *testing.T) {
	si := parser.StructInfo{
		Name: "User",
		Fields: []parser.FieldInfo{
			{
				Name:     "ID",
				Type:     "int",
				TypeExpr: &ast.Ident{Name: "int"},
			},
		},
	}

	code, err := Generate([]parser.StructInfo{si}, "models", nil, "0.1.0", nil, false, "my_suffix")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Diff struct type should be suffixed with CamelCase suffix.
	if !strings.Contains(code, "type UserDiffMySuffix struct") {
		t.Error("generated code missing suffixed UserDiffMySuffix struct")
	}
	// Original struct name must remain unchanged in signatures.
	if !strings.Contains(code, "func UserPatchMySuffix(original, new User) UserDiffMySuffix") {
		t.Error("generated code missing suffixed UserPatchMySuffix function")
	}
	if !strings.Contains(code, "func ApplyUserDiffMySuffix(original User, diff UserDiffMySuffix) User") {
		t.Error("generated code missing suffixed ApplyUserDiffMySuffix function")
	}
	// IsEmpty receiver should be suffixed.
	if !strings.Contains(code, "func (d UserDiffMySuffix) IsEmpty() bool") {
		t.Error("generated code missing suffixed IsEmpty receiver")
	}
	// Ensure unsuffixed names are NOT present.
	if strings.Contains(code, "type UserDiff struct") {
		t.Error("generated code should not contain unsuffixed UserDiff struct")
	}
	if strings.Contains(code, "func UserPatch(") {
		t.Error("generated code should not contain unsuffixed UserPatch function")
	}
}

func TestToCamelCase(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"my_suffix", "MySuffix"},
		{"my-suffix", "MySuffix"},
		{"Foo", "Foo"},
		{"foo_bar_baz", "FooBarBaz"},
	}
	for _, c := range cases {
		if got := toCamelCase(c.in); got != c.want {
			t.Errorf("toCamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestGenerate_CrossFileMapAlias verifies that a struct field whose type is a map alias
// defined in a sibling file of the same package is classified as a map alias and generates
// the map-alias diff pattern (calling the alias's Patch/Apply functions) rather than a
// direct `!=` comparison.
func TestGenerate_CrossFileMapAlias(t *testing.T) {
	tmpDir := t.TempDir()

	// Sibling file that defines the map alias (NOT part of the parsed input below).
	sibling := `package fixtures

type VMLocks map[string]int
`
	if err := os.WriteFile(filepath.Join(tmpDir, "types.go"), []byte(sibling), 0644); err != nil {
		t.Fatal(err)
	}

	// The parsed input file: references VMLocks, which is defined in types.go.
	input := `package fixtures

type Cluster struct {
	VMLocks VMLocks ` + "`diff:\"vm_locks\"`" + `
}
`
	inputFile := filepath.Join(tmpDir, "cluster.go")
	if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	structs, imports, typeDefs, err := parser.ParseFileWithOptions(inputFile, parser.ParseOptions{
		TagKey:     "diff",
		IncludeAll: true,
	})
	if err != nil {
		t.Fatalf("ParseFileWithOptions failed: %v", err)
	}
	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	code, err := Generate(structs, "fixtures", imports, "0.1.0", typeDefs, false, "")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// The field must be classified as a map alias: it should call VMLocksPatch and reference
	// VMLocksDiffCluster instead of comparing with !=.
	if !strings.Contains(code, "VMLocksNestedDiff := VMLocksPatch(original.VMLocks, new.VMLocks)") {
		t.Errorf("expected map-alias patch pattern referencing VMLocksPatch, got:\n%s", code)
	}
	if strings.Contains(code, "original.VMLocks != new.VMLocks") {
		t.Errorf("did not expect direct != comparison for map alias field, got:\n%s", code)
	}
	// Apply must route through the alias's Apply function.
	if !strings.Contains(code, "result.VMLocks = ApplyVMLocksDiff(original.VMLocks, diff.VMLocks.Value)") {
		t.Errorf("expected map-alias apply pattern referencing ApplyVMLocksDiff, got:\n%s", code)
	}
}
