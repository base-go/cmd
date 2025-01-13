package utils

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Initialize the pluralize client
var PluralizeClient *pluralize.Client

func init() {
	PluralizeClient = pluralize.NewClient()
}

//go:embed templates/*
var TemplateFS embed.FS

// FieldStruct represents a field in the model
type FieldStruct struct {
	Name            string
	Type            string
	JSONName        string
	DBName          string
	AssociatedType  string
	AssociatedTable string
	PluralType      string
	Relationship    string
}

// GenerateFileFromTemplate generates a file from a template
func GenerateFileFromTemplate(dir, filename, templateName, singularName, pluralName, packageName string, fields []FieldStruct) {
	// Read template from embedded filesystem
	tmplContent, err := TemplateFS.ReadFile(filepath.Join("templates", templateName))
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templateName, err)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":     strings.ToLower,
		"toTitle":     cases.Title(language.Und).String,
		"ToSnakeCase": ToSnakeCase,
		"hasField":    func(fields []FieldStruct, fieldType string) bool {
			return HasFieldType(fields, fieldType)
		},
		"ToPlural":    PluralizeClient.Plural,
		"ToPascalCase": ToPascalCase,
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	// Create output directory if it doesn't exist
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Create output file
	outputFile := filepath.Join(dir, filename)
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer f.Close()

	// Execute template
	data := struct {
		StructName    string
		PluralName   string
		Package      string
		Fields       []FieldStruct
		HasImageField bool
	}{
		StructName:    singularName,
		PluralName:   pluralName,
		Package:      packageName,
		Fields:       fields,
		HasImageField: HasFieldType(fields, "attachment"),
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	fmt.Printf("Generated %s\n", outputFile)
}

// GenerateFieldStructs processes the fields and returns a slice of FieldStruct
func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct

	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) < 2 {
			continue
		}

		fieldStruct := FieldStruct{
			Name:     ToPascalCase(parts[0]),
			JSONName: ToSnakeCase(parts[0]),
			DBName:   ToSnakeCase(parts[0]),
		}

		// Handle relationships and field types
		switch parts[1] {
		case "string", "text":
			fieldStruct.Type = "string"
		case "int":
			fieldStruct.Type = "int"
		case "float":
			fieldStruct.Type = "float64"
		case "bool":
			fieldStruct.Type = "bool"
		case "time":
			fieldStruct.Type = "time.Time"
		case "attachment":
			fieldStruct.Type = "*storage.Attachment"
		default:
			// Check for relationships
			if strings.Contains(parts[1], "belongs_to:") {
				fieldStruct.Type = "uint"
				fieldStruct.Relationship = "belongs_to"
				fieldStruct.AssociatedType = strings.TrimPrefix(parts[1], "belongs_to:")
				fieldStruct.AssociatedTable = ToSnakeCase(PluralizeClient.Plural(fieldStruct.AssociatedType))
			} else if strings.Contains(parts[1], "has_many:") {
				fieldStruct.Relationship = "has_many"
				fieldStruct.AssociatedType = strings.TrimPrefix(parts[1], "has_many:")
				fieldStruct.AssociatedTable = ToSnakeCase(PluralizeClient.Plural(fieldStruct.AssociatedType))
				fieldStruct.PluralType = PluralizeClient.Plural(fieldStruct.AssociatedType)
				fieldStruct.Type = fmt.Sprintf("[]*%s", fieldStruct.AssociatedType)
			} else if strings.Contains(parts[1], "has_one:") {
				fieldStruct.Relationship = "has_one"
				fieldStruct.AssociatedType = strings.TrimPrefix(parts[1], "has_one:")
				fieldStruct.AssociatedTable = ToSnakeCase(PluralizeClient.Plural(fieldStruct.AssociatedType))
				fieldStruct.Type = fmt.Sprintf("*%s", fieldStruct.AssociatedType)
			}
		}

		fieldStructs = append(fieldStructs, fieldStruct)
	}

	return fieldStructs
}

// HasFieldType checks if any field has the specified type
func HasFieldType(fields []FieldStruct, fieldType string) bool {
	for _, field := range fields {
		if field.Type == fieldType {
			return true
		}
	}
	return false
}
