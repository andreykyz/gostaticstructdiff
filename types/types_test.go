package types

import (
	"go/ast"
	"go/parser"
	"testing"
)

func parseExpr(t *testing.T, expr string) ast.Expr {
	t.Helper()
	node, err := parser.ParseExpr(expr)
	if err != nil {
		t.Fatalf("failed to parse expression %q: %v", expr, err)
	}
	return node
}

func TestClassify_BasicTypes(t *testing.T) {
	tests := []struct {
		expr     string
		expected Category
		typeStr  string
	}{
		{"int", CategoryBasic, "int"},
		{"string", CategoryBasic, "string"},
		{"bool", CategoryBasic, "bool"},
		{"float64", CategoryBasic, "float64"},
		{"byte", CategoryBasic, "byte"},
		{"rune", CategoryBasic, "rune"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			expr := parseExpr(t, tt.expr)
			info := Classify(expr)
			if info.Category != tt.expected {
				t.Errorf("Classify(%q).Category = %v, want %v", tt.expr, info.Category, tt.expected)
			}
			if info.TypeString != tt.typeStr {
				t.Errorf("Classify(%q).TypeString = %q, want %q", tt.expr, info.TypeString, tt.typeStr)
			}
		})
	}
}

func TestClassify_Pointer(t *testing.T) {
	expr := parseExpr(t, "*int")
	info := Classify(expr)
	if info.Category != CategoryPointer {
		t.Errorf("Category = %v, want CategoryPointer", info.Category)
	}
	if info.TypeString != "*int" {
		t.Errorf("TypeString = %q, want *int", info.TypeString)
	}
	if info.Element == nil {
		t.Error("Element should not be nil")
	} else if info.Element.Category != CategoryBasic || info.Element.TypeString != "int" {
		t.Errorf("Element = %+v, want basic int", info.Element)
	}
}

func TestClassify_Slice(t *testing.T) {
	expr := parseExpr(t, "[]string")
	info := Classify(expr)
	if info.Category != CategorySlice {
		t.Errorf("Category = %v, want CategorySlice", info.Category)
	}
	if info.TypeString != "[]string" {
		t.Errorf("TypeString = %q, want []string", info.TypeString)
	}
	if info.Element == nil {
		t.Error("Element should not be nil")
	} else if info.Element.Category != CategoryBasic || info.Element.TypeString != "string" {
		t.Errorf("Element = %+v, want basic string", info.Element)
	}
}

func TestClassify_Map(t *testing.T) {
	expr := parseExpr(t, "map[string]int")
	info := Classify(expr)
	if info.Category != CategoryMap {
		t.Errorf("Category = %v, want CategoryMap", info.Category)
	}
	if info.TypeString != "map[string]int" {
		t.Errorf("TypeString = %q, want map[string]int", info.TypeString)
	}
	if info.Key == nil || info.Value == nil {
		t.Error("Key or Value should not be nil")
	} else {
		if info.Key.Category != CategoryBasic || info.Key.TypeString != "string" {
			t.Errorf("Key = %+v, want basic string", info.Key)
		}
		if info.Value.Category != CategoryBasic || info.Value.TypeString != "int" {
			t.Errorf("Value = %+v, want basic int", info.Value)
		}
	}
}

func TestClassify_Struct(t *testing.T) {
	// Named struct (identifier)
	expr := parseExpr(t, "MyStruct")
	info := Classify(expr)
	if info.Category != CategoryStruct {
		t.Errorf("Category = %v, want CategoryStruct", info.Category)
	}
	if info.TypeString != "MyStruct" {
		t.Errorf("TypeString = %q, want MyStruct", info.TypeString)
	}
}

func TestClassify_SelectorExpr(t *testing.T) {
	expr := parseExpr(t, "pkg.SubType")
	info := Classify(expr)
	if info.Category != CategoryStruct {
		t.Errorf("Category = %v, want CategoryStruct", info.Category)
	}
	if info.TypeString != "pkg.SubType" {
		t.Errorf("TypeString = %q, want pkg.SubType", info.TypeString)
	}
}

func TestDetermineDiffStrategy(t *testing.T) {
	tests := []struct {
		expr     string
		wantTmpl string
	}{
		{"int", "basic"},
		{"*float64", "pointer"},
		{"[]string", "slice"},
		{"map[string]bool", "map"},
		{"MyStruct", "struct"},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			expr := parseExpr(t, tt.expr)
			info := Classify(expr)
			strategy := DetermineDiffStrategy(info)
			if strategy.TemplateName != tt.wantTmpl {
				t.Errorf("TemplateName = %q, want %q", strategy.TemplateName, tt.wantTmpl)
			}
		})
	}
}

func TestIsSupportedType(t *testing.T) {
	tests := []struct {
		category Category
		want     bool
	}{
		{CategoryBasic, true},
		{CategoryPointer, true},
		{CategorySlice, true},
		{CategoryMap, true},
		{CategoryStruct, true},
		{CategoryUnknown, false},
	}
	for _, tt := range tests {
		info := &TypeInfo{Category: tt.category}
		got := IsSupportedType(info)
		if got != tt.want {
			t.Errorf("IsSupportedType(%v) = %v, want %v", tt.category, got, tt.want)
		}
	}
}
