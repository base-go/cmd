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
func GenerateFileFromTemplate(dir, filename, templateFile, singularName, pluralName, packageName string, fields []FieldStruct) {
	// Read template file
	tmplContent, err := os.ReadFile(templateFile)
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templateFile, err)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":     strings.ToLower,
		"toTitle":     cases.Title(language.Und).String,
		"ToSnakeCase": ToSnakeCase,
		"ToPascalCase": ToPascalCase,
		"ToPlural":    PluralizeClient.Plural,
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateFile, err)
		return
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", dir, err)
		return
	}

	// Create target file
	targetFile := filepath.Join(dir, filename)
	file, err := os.Create(targetFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", targetFile, err)
		return
	}
	defer file.Close()

	// Prepare template data
	data := struct {
		StructName string
		PluralName string
		PackageName string
		Fields []FieldStruct
		HasImageField bool
	}{
		StructName:   singularName,
		PluralName:   pluralName,
		PackageName:  packageName,
		Fields:       fields,
		HasImageField: HasFieldType(fields, "*storage.Attachment"),
	}

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Error executing template for %s: %v\n", targetFile, err)
		return
	}

	fmt.Printf("Generated %s\n", targetFile)
}

// GenerateFieldStructs processes the fields and returns a slice of FieldStruct
func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct
	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) != 2 {
			continue
		}
		fieldName := ToPascalCase(parts[0])
		fieldType := parts[1]

		// Convert common types to Go types
		goType := fieldType
		relationship := ""
		var associatedType string
		var associatedTable string
		var pluralType string

		switch fieldType {
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
		if strings.HasPrefix(fieldType, "belongs_to:") {
			relationship = "belongs_to"
			associatedType = strings.TrimPrefix(fieldType, "belongs_to:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			goType = "uint"
		} else if strings.HasPrefix(fieldType, "has_many:") {
			relationship = "has_many"
			associatedType = strings.TrimPrefix(fieldType, "has_many:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			pluralType = PluralizeClient.Plural(associatedType)
			goType = fmt.Sprintf("[]*%s", associatedType)
		} else if strings.HasPrefix(fieldType, "has_one:") {
			relationship = "has_one"
			associatedType = strings.TrimPrefix(fieldType, "has_one:")
			associatedTable = ToSnakeCase(PluralizeClient.Plural(associatedType))
			goType = fmt.Sprintf("*%s", associatedType)
		}

		fieldStructs = append(fieldStructs, FieldStruct{
			Name:            fieldName,
			Type:            goType,
			JSONName:        ToSnakeCase(fieldName),
			DBName:          ToSnakeCase(fieldName),
			AssociatedType:  associatedType,
			AssociatedTable: associatedTable,
			PluralType:      pluralType,
			Relationship:    relationship,
		})
	}
	return fieldStructs
}

func HasFieldType(fields []FieldStruct, fieldType string) bool {
	for _, field := range fields {
		if field.Type == fieldType {
			return true
		}
	}
	return false
}
