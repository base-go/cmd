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
		"ToPascalCase": ToPascalCase,
		"ToPlural":    PluralizeClient.Plural,
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
		StructName  string
		PluralName string
		Package    string
		Fields     []FieldStruct
		HasImageField bool
	}{
		StructName:  singularName,
		PluralName: pluralName,
		Package:    packageName,
		Fields:     fields,
		HasImageField: HasFieldType(fields, "*storage.Attachment"),
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

		// Convert common types to Go types
		goType := parts[1]
		relationship := ""
		var associatedType string
		var associatedTable string
		var pluralType string

		switch parts[1] {
		case "string", "text":
			goType = "string"
		case "int":
			goType = "int"
		case "bool":
			goType = "bool"
		case "float":
			goType = "float64"
		case "time":
			goType = "time.Time"
		case "attachment":
			goType = "*storage.Attachment"
			relationship = "attachment"
		}

		// Check for relationships
		if strings.HasPrefix(parts[1], "belongs_to:") {
			relationship = "belongs_to"
			associatedType = strings.TrimPrefix(parts[1], "belongs_to:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			goType = "uint"
		} else if strings.HasPrefix(parts[1], "has_many:") {
			relationship = "has_many"
			associatedType = strings.TrimPrefix(parts[1], "has_many:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			pluralType = PluralizeClient.Plural(associatedType)
			goType = fmt.Sprintf("[]*%s", associatedType)
		} else if strings.HasPrefix(parts[1], "has_one:") {
			relationship = "has_one"
			associatedType = strings.TrimPrefix(parts[1], "has_one:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			goType = fmt.Sprintf("*%s", associatedType)
		}

		fieldStruct.Type = goType
		fieldStruct.Relationship = relationship
		fieldStruct.AssociatedType = associatedType
		fieldStruct.AssociatedTable = associatedTable
		fieldStruct.PluralType = pluralType

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
