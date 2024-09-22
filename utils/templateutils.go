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
	Name           string
	Type           string
	JSONName       string
	DBName         string
	AssociatedType string
	PluralType     string
	Relationship   string
}

// SeederFieldStruct extends FieldStruct with seeder-specific information
type SeederFieldStruct struct {
	FieldStruct
	DefaultValue string
}

// GenerateFileFromTemplate generates a file from a template
func GenerateFileFromTemplate(dir, filename, templateFile, singularName, pluralName, packageName string, fields []FieldStruct) {
	tmplContent, err := TemplateFS.ReadFile(templateFile)
	if err != nil {
		fmt.Printf("Error reading template %s: %v\n", templateFile, err)
		return
	}

	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
		"toTitle": cases.Title(language.Und).String,
	}

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateFile, err)
		return
	}

	data := map[string]interface{}{
		"PackageName":           packageName,
		"StructName":            singularName,
		"LowerStructName":       strings.ToLower(singularName[:1]) + singularName[1:],
		"PluralName":            pluralName,
		"RouteName":             ToSnakeCase(singularName),
		"Fields":                fields,
		"TableName":             ToSnakeCase(pluralName),
		"LowerPluralStructName": strings.ToLower(pluralName),
		"SeederFields":          GenerateSeederFieldStructs(fields),
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

// GenerateFieldStructs processes the fields and returns a slice of FieldStruct

func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct

	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) >= 2 {
			name := ToPascalCase(parts[0])
			fieldType := parts[1]
			jsonName := ToSnakeCase(parts[0])
			dbName := ToSnakeCase(parts[0])
			var associatedType, pluralType, relationship string

			goType := GetGoType(fieldType)

			switch strings.ToLower(fieldType) {
			case "belongsto", "belongs_to":
				relationship = "belongs_to"
				if len(parts) > 2 {
					associatedType = ToPascalCase(parts[2])
					goType = "*" + associatedType
					jsonName += "_id"
					dbName += "_id"
				}
			case "hasone", "has_one":
				relationship = "has_one"
				if len(parts) > 2 {
					associatedType = ToPascalCase(parts[2])
					goType = "*" + associatedType
					jsonName = ToSnakeCase(name)
				}
			case "hasmany", "has_many":
				relationship = "has_many"
				if len(parts) > 2 {
					associatedType = ToPascalCase(parts[2])
					pluralType = PluralizeClient.Plural(ToLower(parts[2]))
					goType = "[]*" + associatedType
					jsonName = ToSnakeCase(PluralizeClient.Plural(name))
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

// GenerateSeederFieldStructs generates SeederFieldStruct slice from FieldStruct slice
func GenerateSeederFieldStructs(fields []FieldStruct) []SeederFieldStruct {
	var seederFields []SeederFieldStruct

	for _, field := range fields {
		seederField := SeederFieldStruct{
			FieldStruct: field,
		}

		if field.Relationship == "belongs_to" {
			seederField.DefaultValue = getDefaultValue("uint") // Use uint for ID fields
		} else if field.Relationship == "" {
			seederField.DefaultValue = getDefaultValue(field.Type)
		} else {
			seederField.DefaultValue = "nil" // For has_one and has_many relationships
		}

		seederFields = append(seederFields, seederField)
	}

	return seederFields
}

// getDefaultValue returns a default value string based on the Go type
func getDefaultValue(goType string) string {
	switch goType {
	case "string":
		return `"example_value"`
	case "int", "int64", "uint", "uint64":
		return "1"
	case "float64":
		return "1.0"
	case "bool":
		return "false"
	case "time.Time":
		return "time.Now()"
	default:
		return "nil"
	}
}

// ParseTemplate parses a template string with custom functions
func ParseTemplate(name, content string) (*template.Template, error) {
	return template.New(name).Funcs(template.FuncMap{
		"getInputType": GetInputType,
	}).Parse(content)
}
