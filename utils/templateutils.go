package utils

import (
	_ "embed"
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

//go:embed templates/model.tmpl
var modelTemplate string

//go:embed templates/controller.tmpl
var controllerTemplate string

//go:embed templates/service.tmpl
var serviceTemplate string

//go:embed templates/module.tmpl
var moduleTemplate string

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
	RelatedModel    string
}

// GenerateFileFromTemplate generates a file from a template
func GenerateFileFromTemplate(dir, filename, templateName, structName, pluralName, packageName string, fields []FieldStruct) {
	var tmplContent string
	switch templateName {
	case "model.tmpl":
		tmplContent = modelTemplate
	case "controller.tmpl":
		tmplContent = controllerTemplate
	case "service.tmpl":
		tmplContent = serviceTemplate
	case "module.tmpl":
		tmplContent = moduleTemplate
	default:
		fmt.Printf("Unknown template: %s\n", templateName)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":     strings.ToLower,
		"toTitle":     cases.Title(language.Und).String,
		"ToSnakeCase": ToSnakeCase,
		"hasField": func(fields []FieldStruct, fieldType string) bool {
			return HasFieldType(fields, fieldType)
		},
		"ToPascalCase": ToPascalCase,
		"ToPlural":     PluralizeClient.Plural,
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateName, err)
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
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		return
	}
	defer f.Close()

	// Execute template
	data := struct {
		StructName    string
		PluralName    string
		PackageName   string
		Fields        []FieldStruct
		HasImageField bool
	}{
		StructName:    structName,
		PluralName:    pluralName,
		PackageName:   packageName,
		Fields:        fields,
		HasImageField: HasImageField(fields),
	}

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	fmt.Printf("Generated %s\n", outputFile)
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

// HasImageField checks if any field has the image type
func HasImageField(fields []FieldStruct) bool {
	return HasFieldType(fields, "*storage.Attachment")
}

// GenerateFieldStructs processes the fields and returns a slice of FieldStruct
func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct
	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		fieldType := parts[1]
		relationship := ""
		relatedModel := ""

		if len(parts) >= 3 {
			relationship = fieldType
			relatedModel = parts[2]

			// First add the ID field for belongs_to relationships
			if relationship == "belongs_to" || relationship == "belongsTo" {
				idField := FieldStruct{
					Name:         ToPascalCase(name) + "Id",
					Type:         "uint",
					JSONName:     ToSnakeCase(name) + "_id",
					DBName:       ToSnakeCase(name) + "_id",
					Relationship: "",
				}
				fieldStructs = append(fieldStructs, idField)
			}

			// Then add the relationship field with proper type
			if relationship == "belongs_to" || relationship == "has_one" || relationship == "belongsTo" || relationship == "hasOne" {
				fieldType = "*" + relatedModel
			} else if relationship == "has_many" || relationship == "hasMany" || relationship == "manyToMany" || relationship == "many2many" {
				fieldType = "[]*" + relatedModel
			}
		} else {
			// For non-relationship fields, get the Go type
			fieldType = GetGoType(fieldType)
		}

		// Check for attachment fields
		if fieldType == "attachment" || fieldType == "image" || fieldType == "file" {
			relationship = "attachment"
			fieldType = "*storage.Attachment"
			
			// Add the ID field for the attachment
			idField := FieldStruct{
				Name:         ToPascalCase(name) + "Id",
				Type:         "*uint",
				JSONName:     ToSnakeCase(name) + "_id",
				DBName:       ToSnakeCase(name) + "_id",
				Relationship: "",
			}
			fieldStructs = append(fieldStructs, idField)

			// Add the attachment field with proper GORM tags
			fieldStruct := FieldStruct{
				Name:         ToPascalCase(name),
				Type:         fieldType,
				JSONName:     ToSnakeCase(name),
				DBName:       ToSnakeCase(name),
				Relationship: relationship,
				RelatedModel: "storage.Attachment",
				AssociatedType: fieldType, // Store the original type (image/file) for validation
			}
			fieldStructs = append(fieldStructs, fieldStruct)
			continue
		}

		fieldStruct := FieldStruct{
			Name:         ToPascalCase(name),
			Type:         fieldType,
			JSONName:     ToSnakeCase(name),
			DBName:       ToSnakeCase(name),
			Relationship: relationship,
			RelatedModel: relatedModel,
		}

		if relationship != "" {
			fieldStruct.AssociatedType = relatedModel
			fieldStruct.AssociatedTable = ToSnakeCase(PluralizeClient.Plural(relatedModel))
			fieldStruct.PluralType = PluralizeClient.Plural(relatedModel)
		}

		fieldStructs = append(fieldStructs, fieldStruct)
	}
	return fieldStructs
}
