package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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
	// SourceDir is the directory of the Go file this struct was parsed from. It is used
	// to resolve same-package type definitions in sibling files for classification.
	SourceDir string
}

// ParseOptions configures parsing behavior.
type ParseOptions struct {
	TagKey     string // tag key to look for (e.g., "structtomap", "mapstructure")
	IncludeAll bool   // if true, include all fields regardless of tags
	Verbose    bool   // if true, enable verbose logging
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
	structs, imports, _, err := ParseFileWithOptions(filename, ParseOptions{TagKey: "structtomap", IncludeAll: false})
	return structs, imports, err
}

// ParseFileWithOptions parses a Go file and returns structs according to options.
// It also returns a map of all type definitions (including non-structs) in the file,
// keyed by type name (without package qualifier).
func ParseFileWithOptions(filename string, opts ParseOptions) ([]StructInfo, []string, map[string]ast.Expr, error) {
	if opts.Verbose {
		fmt.Printf("Parsing file: %s\n", filename)
		fmt.Printf("Options: tagKey=%s, includeAll=%v\n", opts.TagKey, opts.IncludeAll)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, err
	}

	var structs []StructInfo
	typeDefs := make(map[string]ast.Expr)

	// Walk through the AST and collect all type definitions
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		// Store the underlying type expression
		typeDefs[typeSpec.Name.Name] = typeSpec.Type

		// If it's a struct, also process it for struct info
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
			if opts.Verbose {
				fmt.Printf("  Skipping struct %s: no fields with tag %s\n", typeSpec.Name.Name, opts.TagKey)
			}
			return true // No fields with required tag, skip this struct
		}

		if opts.Verbose {
			fmt.Printf("  Found struct %s with %d fields\n", typeSpec.Name.Name, len(fields))
		}

		structs = append(structs, StructInfo{
			Name:      typeSpec.Name.Name,
			Fields:    fields,
			SourceDir: filepath.Dir(filename),
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

	if opts.Verbose {
		fmt.Printf("Parsed %d struct(s), %d import(s), %d type definition(s)\n", len(structs), len(imports), len(typeDefs))
	}

	return structs, imports, typeDefs, nil
}

// ParsePackageTypeDefs parses sibling .go files in the same directory as inputFile and
// returns their type definitions (keyed by type name), for use in classification.
// It excludes *_test.go, *_diff*.go (generated diff artifacts), the input file itself,
// and files whose package clause differs from the package of inputFile.
// Resolution failures for individual sibling files are non-fatal; only a failure to
// read the input file's own package clause returns an error.
func ParsePackageTypeDefs(inputFile string, verbose bool) (map[string]ast.Expr, error) {
	dir := filepath.Dir(inputFile)

	// Determine the package name of the input file.
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFile, nil, parser.PackageClauseOnly)
	if err != nil {
		return nil, fmt.Errorf("parse package clause of %s: %w", inputFile, err)
	}
	pkgName := node.Name.Name

	typeDefs := make(map[string]ast.Expr)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	base := filepath.Base(inputFile)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if name == base {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if strings.HasSuffix(name, "_diff.go") || strings.Contains(name, "_diff_") {
			continue
		}

		filePath := filepath.Join(dir, name)
		fs := token.NewFileSet()
		fileNode, perr := parser.ParseFile(fs, filePath, nil, parser.ParseComments)
		if perr != nil {
			if verbose {
				fmt.Printf("  Warning: could not parse sibling file %s: %v\n", filePath, perr)
			}
			continue
		}
		if fileNode.Name.Name != pkgName {
			continue
		}

		ast.Inspect(fileNode, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, exists := typeDefs[typeSpec.Name.Name]; !exists {
				typeDefs[typeSpec.Name.Name] = typeSpec.Type
			}
			return true
		})
	}

	return typeDefs, nil
}

// ParsePackageTypeDefsInDir parses sibling .go files in the given directory that belong to
// pkgName and returns their type definitions (keyed by type name), for use in classification.
// It excludes *_test.go and *_diff*.go (generated diff artifacts), as well as files whose
// package clause differs from pkgName. Resolution failures for individual sibling files are
// non-fatal; only a failure to read the directory returns an error.
func ParsePackageTypeDefsInDir(dir, pkgName string, verbose bool) (map[string]ast.Expr, error) {
	typeDefs := make(map[string]ast.Expr)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if strings.HasSuffix(name, "_diff.go") || strings.Contains(name, "_diff_") {
			continue
		}

		filePath := filepath.Join(dir, name)
		fs := token.NewFileSet()
		fileNode, perr := parser.ParseFile(fs, filePath, nil, parser.ParseComments)
		if perr != nil {
			if verbose {
				fmt.Printf("  Warning: could not parse sibling file %s: %v\n", filePath, perr)
			}
			continue
		}
		if fileNode.Name.Name != pkgName {
			continue
		}

		ast.Inspect(fileNode, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if _, exists := typeDefs[typeSpec.Name.Name]; !exists {
				typeDefs[typeSpec.Name.Name] = typeSpec.Type
			}
			return true
		})
	}

	return typeDefs, nil
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
