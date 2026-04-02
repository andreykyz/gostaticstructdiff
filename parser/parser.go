package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// FieldInfo represents a single field in a struct.
type FieldInfo struct {
	Name     string
	Type     string
	Tag      string
	TypeExpr ast.Expr
}

// StructInfo represents a parsed struct with structtomap tags.
type StructInfo struct {
	Name   string
	Fields []FieldInfo
}

// ParseFile parses a Go file and returns all structs that have at least one field
// with a `structtomap` tag.
func ParseFile(filename string) ([]StructInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structs []StructInfo

	// Walk through the AST and collect struct definitions
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Check if any field has a structtomap tag
		fields := extractFields(structType)
		if len(fields) == 0 {
			return true // No fields with structtomap tags, skip this struct
		}

		structs = append(structs, StructInfo{
			Name:   typeSpec.Name.Name,
			Fields: fields,
		})
		return true
	})

	return structs, nil
}

// extractFields extracts fields from a struct type that have structtomap tags.
func extractFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}
		tag := strings.Trim(field.Tag.Value, "`")
		if !strings.Contains(tag, "structtomap:") {
			continue
		}

		// Get field name(s) - could be multiple if same type
		if field.Names == nil {
			// Embedded field (anonymous)
			continue
		}

		for _, name := range field.Names {
			// Convert type expression to string (simplified)
			typeStr := exprToString(field.Type)

			fields = append(fields, FieldInfo{
				Name:     name.Name,
				Type:     typeStr,
				Tag:      tag,
				TypeExpr: field.Type,
			})
		}
	}

	return fields
}

// exprToString converts an ast.Expr to a string representation.
// This is a simplified version; a more robust implementation would handle
// complex types like *ast.SelectorExpr, *ast.ArrayType, etc.
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		// Handle array with length (e.g., [2]int)
		// For simplicity, we'll just return slice
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StructType:
		// Nested anonymous struct - we'll represent as "struct{...}"
		return "struct{...}"
	default:
		// Fallback
		return "unknown"
	}
}
