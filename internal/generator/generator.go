package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/andreykyz/gostaticstructdiff/internal/parser"
	"github.com/andreykyz/gostaticstructdiff/internal/types"
)

// FieldTemplateData holds data for a field in the template.
type FieldTemplateData struct {
	Name      string
	Type      string
	Category  string
	KeyType   string
	ValueType string
}

// StructTemplateData holds data for a struct in the template.
type StructTemplateData struct {
	Name   string
	Fields []FieldTemplateData
}

// Generate generates diff code for the given structs and writes to output.
func Generate(structs []parser.StructInfo, packageName string, imports []string) (string, error) {
	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		return "", fmt.Errorf("failed to load templates: %w", err)
	}

	// Prepare data for each struct
	var output bytes.Buffer

	// Write package declaration
	output.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Write imports if any
	if len(imports) > 0 {
		output.WriteString("import (\n")
		for _, imp := range imports {
			output.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		output.WriteString(")\n\n")
	}

	// Generate each struct diff
	for _, s := range structs {
		data := convertToTemplateData(s)
		err = tmpl.ExecuteTemplate(&output, "struct_diff.tmpl", data)
		if err != nil {
			return "", fmt.Errorf("failed to execute struct template for %s: %w", s.Name, err)
		}
		output.WriteString("\n\n")

		// Generate patch functions
		err = tmpl.ExecuteTemplate(&output, "patch_func.tmpl", data)
		if err != nil {
			return "", fmt.Errorf("failed to execute patch template for %s: %w", s.Name, err)
		}
		output.WriteString("\n\n")
	}

	return output.String(), nil
}

// convertToTemplateData converts a parser.StructInfo to StructTemplateData.
func convertToTemplateData(s parser.StructInfo) StructTemplateData {
	data := StructTemplateData{
		Name:   s.Name,
		Fields: make([]FieldTemplateData, 0, len(s.Fields)),
	}

	for _, f := range s.Fields {
		typeInfo := types.Classify(f.TypeExpr)
		category := typeInfo.Category.String()

		fieldData := FieldTemplateData{
			Name:     f.Name,
			Type:     f.Type,
			Category: category,
		}

		// For map types, extract key and value types
		if typeInfo.Category == types.CategoryMap && typeInfo.Key != nil && typeInfo.Value != nil {
			fieldData.KeyType = typeInfo.Key.TypeString
			fieldData.ValueType = typeInfo.Value.TypeString
		}

		data.Fields = append(data.Fields, fieldData)
	}

	return data
}

// loadTemplates loads all template files from the templates directory.
func loadTemplates() (*template.Template, error) {
	// List of template files to load
	templateFiles := []string{
		"internal/templates/struct_diff.tmpl",
		"internal/templates/patch_func.tmpl",
	}
	// Parse all templates; each template will be named after the base name of the file.
	tmpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	return tmpl, nil
}
