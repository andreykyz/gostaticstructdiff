package types

import (
	"go/ast"
	"strings"
)

// Category represents the classification of a Go type.
type Category int

const (
	CategoryBasic Category = iota
	CategoryPointer
	CategorySlice
	CategoryMap
	CategoryStruct
	CategoryUnknown
)

// String returns a human-readable representation of the category.
func (c Category) String() string {
	switch c {
	case CategoryBasic:
		return "basic"
	case CategoryPointer:
		return "pointer"
	case CategorySlice:
		return "slice"
	case CategoryMap:
		return "map"
	case CategoryStruct:
		return "struct"
	default:
		return "unknown"
	}
}

// TypeInfo holds information about a parsed field type.
type TypeInfo struct {
	Category Category
	// Underlying type as string (e.g., "int", "string", "*float64")
	TypeString string
	// For pointer: points to TypeInfo of the element
	// For slice: element TypeInfo
	// For map: key and value TypeInfo
	// For struct: fields (optional)
	Element *TypeInfo
	Key     *TypeInfo
	Value   *TypeInfo
}

// Classify analyzes an AST expression and returns a TypeInfo.
// knownStructs is a set of type names that are known to be structs (including package qualifiers).
// If nil, no types are assumed to be structs.
func Classify(expr ast.Expr, knownStructs map[string]bool, typeDefs map[string]ast.Expr) *TypeInfo {
	switch t := expr.(type) {
	case *ast.Ident:
		// Basic type or named type
		name := t.Name
		if isBasicType(name) {
			return &TypeInfo{
				Category:   CategoryBasic,
				TypeString: name,
			}
		}
		// Check if it's a known struct
		if knownStructs != nil && knownStructs[name] {
			return &TypeInfo{
				Category:   CategoryStruct,
				TypeString: name,
			}
		}
		// Look up in type definitions
		if typeDefs != nil {
			if underlying, ok := typeDefs[name]; ok {
				// Recursively classify the underlying type
				return Classify(underlying, knownStructs, typeDefs)
			}
		}
		// Assume it's a basic (comparable) type (e.g., type alias)
		return &TypeInfo{
			Category:   CategoryBasic,
			TypeString: name,
		}
	case *ast.StarExpr:
		// Pointer type
		elem := Classify(t.X, knownStructs, typeDefs)
		return &TypeInfo{
			Category:   CategoryPointer,
			TypeString: "*" + elem.TypeString,
			Element:    elem,
		}
	case *ast.ArrayType:
		// Slice (if Len is nil) or array (if Len present)
		// For simplicity, treat both as slice
		elem := Classify(t.Elt, knownStructs, typeDefs)
		return &TypeInfo{
			Category:   CategorySlice,
			TypeString: "[]" + elem.TypeString,
			Element:    elem,
		}
	case *ast.MapType:
		key := Classify(t.Key, knownStructs, typeDefs)
		value := Classify(t.Value, knownStructs, typeDefs)
		return &TypeInfo{
			Category:   CategoryMap,
			TypeString: "map[" + key.TypeString + "]" + value.TypeString,
			Key:        key,
			Value:      value,
		}
	case *ast.SelectorExpr:
		// Qualified identifier (e.g., "models.User")
		typeStr := exprToString(expr)
		// Check if it's a known struct
		if knownStructs != nil && knownStructs[typeStr] {
			return &TypeInfo{
				Category:   CategoryStruct,
				TypeString: typeStr,
			}
		}
		// Cannot determine underlying type; treat as unknown (use reflect.DeepEqual)
		return &TypeInfo{
			Category:   CategoryUnknown,
			TypeString: typeStr,
		}
	case *ast.StructType:
		// Anonymous struct
		return &TypeInfo{
			Category:   CategoryStruct,
			TypeString: exprToString(expr),
		}
	default:
		// Fallback
		return &TypeInfo{
			Category:   CategoryUnknown,
			TypeString: exprToString(expr),
		}
	}
}

// isBasicType returns true if the given name is a Go basic type.
func isBasicType(name string) bool {
	basicTypes := []string{
		"bool", "string",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"complex64", "complex128",
		"byte", "rune",
	}
	for _, t := range basicTypes {
		if name == t {
			return true
		}
	}
	return false
}

// exprToString converts an ast.Expr to a string representation.
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
		// Array with length
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StructType:
		return structTypeToString(t)
	case *ast.BasicLit:
		return t.Value
	default:
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

// DiffStrategy defines the structure of a diff field for a given type.
type DiffStrategy struct {
	// Name of the template to use (e.g., "basic", "pointer", "map")
	TemplateName string
	// Additional data for the template
	Data map[string]interface{}
}

// DetermineDiffStrategy returns the appropriate diff strategy for a TypeInfo.
func DetermineDiffStrategy(typeInfo *TypeInfo) DiffStrategy {
	switch typeInfo.Category {
	case CategoryBasic:
		return DiffStrategy{
			TemplateName: "basic",
			Data: map[string]interface{}{
				"Type": typeInfo.TypeString,
			},
		}
	case CategoryPointer:
		return DiffStrategy{
			TemplateName: "pointer",
			Data: map[string]interface{}{
				"Type": typeInfo.TypeString,
				"Elem": typeInfo.Element,
			},
		}
	case CategorySlice:
		return DiffStrategy{
			TemplateName: "slice",
			Data: map[string]interface{}{
				"Type": typeInfo.TypeString,
				"Elem": typeInfo.Element,
			},
		}
	case CategoryMap:
		return DiffStrategy{
			TemplateName: "map",
			Data: map[string]interface{}{
				"KeyType":   typeInfo.Key.TypeString,
				"ValueType": typeInfo.Value.TypeString,
			},
		}
	case CategoryStruct:
		// For structs, we need to decide: if it's a named struct from another package,
		// we should generate a diff type for that struct (recursive).
		// For simplicity, we'll treat as nested diff.
		return DiffStrategy{
			TemplateName: "struct",
			Data: map[string]interface{}{
				"Type": typeInfo.TypeString,
			},
		}
	default:
		// Unknown type - fallback to basic
		return DiffStrategy{
			TemplateName: "basic",
			Data: map[string]interface{}{
				"Type": typeInfo.TypeString,
			},
		}
	}
}

// IsSupportedType returns true if the type can be handled by the generator.
func IsSupportedType(typeInfo *TypeInfo) bool {
	// For now, support all categories except unknown
	return typeInfo.Category != CategoryUnknown
}
