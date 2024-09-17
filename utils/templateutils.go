package utils

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
)

// Initialize the pluralize client
var PluralizeClient *pluralize.Client

func init() {
	PluralizeClient = pluralize.NewClient()
}

// Embed the templates directory

//go:embed templates/*
var TemplateFS embed.FS

// GenerateFileFromTemplate generates a file from a template
func GenerateFileFromTemplate(dir, filename, templateFile, singularName, pluralName string, fields []string) {
	tmplContent, err := TemplateFS.ReadFile(templateFile)
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templateFile, err)
		return
	}

	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
		"toTitle": strings.Title,
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateFile, err)
		return
	}

	fieldStructs := GenerateFieldStructs(fields)

	data := map[string]interface{}{
		"PackageName": pluralName,
		"StructName":  ToTitle(singularName),
		"PluralName":  ToTitle(pluralName),
		"RouteName":   pluralName,
		"Fields":      fieldStructs,
		"TableName":   pluralName,
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Printf("Error executing template for %s: %v\n", filename, err)
	}
}

// FieldStruct represents a field in the model
type FieldStruct struct {
	Name           string
	Type           string
	JSONName       string
	DBName         string
	AssociatedType string
	PluralType     string
	Relationship   string
}

// GenerateFieldStructs processes the fields and returns a slice of FieldStruct
func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct

	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) >= 2 {
			name := ToTitle(parts[0])
			fieldType := parts[1]
			jsonName := ToLower(parts[0])
			dbName := ToLower(parts[0])
			var associatedType, pluralType, relationship string

			goType := GetGoType(fieldType)

			switch fieldType {
			case "belongs_to":
				relationship = "belongs_to"
				if len(parts) > 2 {
					associatedType = ToTitle(parts[2])
					// Add ID field for belongs_to relationships
					fieldStructs = append(fieldStructs, FieldStruct{
						Name:         name + "ID",
						Type:         "uint",
						JSONName:     jsonName + "Id",
						DBName:       dbName + "_id",
						Relationship: "belongs_to_id",
					})
					goType = associatedType
				}
			case "has_one":
				relationship = "has_one"
				if len(parts) > 2 {
					associatedType = ToTitle(parts[2])
					// Add ID field for has_one relationships
					fieldStructs = append(fieldStructs, FieldStruct{
						Name:         name + "ID",
						Type:         "uint",
						JSONName:     jsonName + "Id",
						DBName:       dbName + "_id",
						Relationship: "has_one_id",
					})
					goType = associatedType
				}
			case "has_many":
				relationship = "has_many"
				if len(parts) > 2 {
					associatedType = ToTitle(parts[2])
					pluralType = PluralizeClient.Plural(ToLower(parts[2]))
					goType = "[]" + associatedType
				}
			}

			fieldStructs = append(fieldStructs, FieldStruct{
				Name:           name,
				Type:           goType,
				JSONName:       jsonName,
				DBName:         dbName,
				AssociatedType: associatedType,
				PluralType:     pluralType,
				Relationship:   relationship,
			})
		}
	}

	return fieldStructs
}

// ParseTemplate parses a template string with custom functions
func ParseTemplate(name, content string) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{
		"getInputType": GetInputType,
	}).Parse(content)
}

// GetInputType maps Go types to HTML input types
func GetInputType(goType string) string {
	switch goType {
	case "int", "int64", "uint", "uint64":
		return "number"
	case "float64":
		return "number"
	case "bool":
		return "checkbox"
	case "time.Time":
		return "datetime-local"
	default:
		return "text"
	}
}
