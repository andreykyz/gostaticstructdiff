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

// ParseOptions configures parsing behavior.
type ParseOptions struct {
	TagKey     string // tag key to look for (e.g., "structtomap", "mapstructure")
	IncludeAll bool   // if true, include all fields regardless of tags
}

// TagValue extracts the value of a structtomap tag from a tag string.
// Returns empty string if not found.
func TagValue(tag string) string {
	return TagValueWithKey(tag, "structtomap")
}

// TagValueWithKey extracts the value of a tag with the given key from a tag string.
// Returns empty string if not found.
func TagValueWithKey(tag, tagKey string) string {
	key := tagKey + ":"
	idx := strings.Index(tag, key)
	if idx == -1 {
		return ""
	}
	rest := tag[idx+len(key):]
	// Find opening quote
	start := strings.Index(rest, `"`)
	if start == -1 {
		return ""
	}
	end := strings.Index(rest[start+1:], `"`)
	if end == -1 {
		return ""
	}
	return rest[start+1 : start+1+end]
}

// ParseFile parses a Go file and returns all structs that have at least one field
// with a `structtomap` tag, along with the file's imports.
func ParseFile(filename string) ([]StructInfo, []string, error) {
	return ParseFileWithOptions(filename, ParseOptions{TagKey: "structtomap", IncludeAll: false})
}

// ParseFileWithOptions parses a Go file and returns structs according to options.
func ParseFileWithOptions(filename string, opts ParseOptions) ([]StructInfo, []string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
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

		// Extract fields according to options
		fields := extractFieldsWithOptions(structType, opts)
		// If IncludeAll is false, we still need at least one field with the tag?
		// The original behavior skipped structs with zero fields with tag.
		// We'll keep that behavior: if IncludeAll is false and no fields have tag, skip.
		if !opts.IncludeAll && len(fields) == 0 {
			return true // No fields with required tag, skip this struct
		}

		structs = append(structs, StructInfo{
			Name:   typeSpec.Name.Name,
			Fields: fields,
		})
		return true
	})

	// Collect imports
	importSet := make(map[string]bool)
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		importSet[path] = true
	}
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}

	return structs, imports, nil
}

// extractFields extracts fields from a struct type that have structtomap tags.
func extractFields(structType *ast.StructType) []FieldInfo {
	return extractFieldsWithOptions(structType, ParseOptions{TagKey: "structtomap", IncludeAll: false})
}

// extractFieldsWithOptions extracts fields from a struct type according to options.
func extractFieldsWithOptions(structType *ast.StructType, opts ParseOptions) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		// Determine if field should be included
		var tag string
		if field.Tag != nil {
			tag = strings.Trim(field.Tag.Value, "`")
		}

		// Check if field has the required tag
		hasRequiredTag := field.Tag != nil && strings.Contains(tag, opts.TagKey+":")

		if !opts.IncludeAll && !hasRequiredTag {
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
		return structTypeToString(t)
	default:
		// Fallback
		return "unknown"
	}
}

// structTypeToString converts an ast.StructType to a string representation.
func structTypeToString(st *ast.StructType) string {
	if st.Fields == nil || len(st.Fields.List) == 0 {
		return "struct{}"
	}
	var parts []string
	for _, field := range st.Fields.List {
		// Get field type
		typ := exprToString(field.Type)
		// Get tag if present
		var tag string
		if field.Tag != nil {
			tag = " " + field.Tag.Value
		}
		// Handle field names
		if field.Names == nil {
			// Embedded field (anonymous)
			parts = append(parts, typ+tag)
		} else {
			for _, name := range field.Names {
				parts = append(parts, name.Name+" "+typ+tag)
			}
		}
	}
	return "struct { " + strings.Join(parts, "; ") + " }"
}
