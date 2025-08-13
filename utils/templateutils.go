package utils

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/model.tmpl
var modelTemplate string

//go:embed templates/controller.tmpl
var controllerTemplate string

//go:embed templates/service.tmpl
var serviceTemplate string

//go:embed templates/module.tmpl
var moduleTemplate string

//go:embed templates/model_test.tmpl
var modelTestTemplate string

//go:embed templates/service_test.tmpl
var serviceTestTemplate string

//go:embed templates/controller_test.tmpl
var controllerTestTemplate string

//go:embed templates/validator.tmpl
var validatorTemplate string

// FieldStruct represents a field in the model
type FieldStruct struct {
	Name               string
	Type               string
	JSONName           string
	DBName             string
	AssociatedType     string
	AssociatedTable    string
	PluralType         string
	Relationship       string
	RelatedModel       string
	IsRequired         bool
	IsRelation         bool   // Whether this field represents a relationship
	GORM               string // Add this field for GORM tags
	TestValue          string // Test value for this field
	UpdateTestValue    string // Update test value for this field
	TestValueWithIndex string // Test value with index for loops
	TestValueUnique    string // Unique test value for constraint tests
	IsUnique           bool   // Whether this field has unique constraint
	
	// Parent context variables to be accessible inside range loops
	StructName         string // Parent struct name for use in range loops
	LowerName          string // Lowercase struct name for use in range loops
}

// getTestValues generates test values for different field types
func getTestValues(fieldType, fieldName string) (testValue, updateTestValue, testValueWithIndex, testValueUnique string, isUnique bool) {
	switch fieldType {
	case "string":
		testValue = fmt.Sprintf(`"Test %s"`, ToPascalCase(fieldName))
		updateTestValue = fmt.Sprintf(`"Updated %s"`, ToPascalCase(fieldName))
		testValueWithIndex = fmt.Sprintf(`fmt.Sprintf("Test %s %%d", i)`, ToPascalCase(fieldName))
		testValueUnique = fmt.Sprintf(`"Unique %s"`, ToPascalCase(fieldName))
		isUnique = (fieldName == "email" || fieldName == "username")
	case "int", "uint":
		testValue = "123"
		updateTestValue = "456"
		testValueWithIndex = "uint(100 + i)"
		testValueUnique = "789"
	case "float64":
		testValue = "123.45"
		updateTestValue = "678.90"
		testValueWithIndex = "float64(100.5 + float64(i))"
		testValueUnique = "999.99"
	case "bool":
		testValue = "true"
		updateTestValue = "false"
		testValueWithIndex = "(i%2 == 0)"
		testValueUnique = "false"
	case "time.Time":
		testValue = "time.Now()"
		updateTestValue = "time.Now().Add(time.Hour)"
		testValueWithIndex = "time.Now().Add(time.Duration(i) * time.Minute)"
		testValueUnique = "time.Now().Add(time.Hour * 24)"
	default:
		// For relationships and other types
		testValue = "nil"
		updateTestValue = "nil"
		testValueWithIndex = "nil"
		testValueUnique = "nil"
	}
	return
}

// normalizeRelationship normalizes relationship types to standard format
func normalizeRelationship(rel string) string {
	switch rel {
	case "belongs_to", "belongsTo":
		return "belongs_to"
	case "has_many", "hasMany":
		return "has_many"
	case "has_one", "hasOne":
		return "has_one"
	default:
		return rel
	}
}

func processRelationshipField(name string, relatedModel string, relationship string) []FieldStruct {
	var fields []FieldStruct

	switch relationship {
	case "belongs_to", "belongsTo":
		// Add the relationship field only
		fields = append(fields, FieldStruct{
			Name:         ToPascalCase(name),
			Type:         "*" + relatedModel,
			JSONName:     ToSnakeCase(name),
			DBName:       ToSnakeCase(name),
			Relationship: relationship,
			RelatedModel: relatedModel,
			IsRequired:   false,
		})

	case "has_many", "hasMany":
		// Add the relationship field
		fields = append(fields, FieldStruct{
			Name:         ToPascalCase(name),
			Type:         "[]*" + relatedModel,
			JSONName:     ToSnakeCase(name),
			DBName:       ToSnakeCase(name),
			Relationship: relationship,
			RelatedModel: relatedModel,
			GORM:         fmt.Sprintf("foreignKey:%sId;references:Id", strings.TrimSuffix(ToPascalCase(name), "s")),
			IsRequired:   false,
		})
	}

	return fields
}

// inferFieldType infers field type from field name
func inferFieldType(fieldName string) string {
	lower := strings.ToLower(fieldName)
	
	// Relation patterns
	if strings.HasSuffix(lower, "_id") {
		return "uint" // Foreign key
	}
	
	// Text fields (should be TEXT in database)
	textFields := []string{"description", "content", "body", "notes", "comment", "summary", "bio", "about"}
	for _, tf := range textFields {
		if strings.Contains(lower, tf) {
			return "text"
		}
	}
	
	// Boolean fields
	boolFields := []string{"is_", "has_", "can_", "enabled", "active", "published", "verified", "confirmed"}
	for _, bf := range boolFields {
		if strings.HasPrefix(lower, bf) || strings.Contains(lower, bf) {
			return "bool"
		}
	}
	
	// Numeric fields
	numericFields := []string{"count", "price", "amount", "quantity", "number", "rating", "score", "weight", "height", "width"}
	for _, nf := range numericFields {
		if strings.Contains(lower, nf) {
			if strings.Contains(lower, "price") || strings.Contains(lower, "amount") {
				return "decimal" // For money fields
			}
			return "int"
		}
	}
	
	// Date/time fields - check suffixes first, then contains
	if strings.HasSuffix(lower, "_at") || strings.HasSuffix(lower, "_on") || strings.HasSuffix(lower, "_date") {
		return "datetime"
	}
	dateFields := []string{"date", "time", "created_at", "updated_at", "deleted_at", "published_at", "expires_at"}
	for _, df := range dateFields {
		if strings.Contains(lower, df) {
			return "datetime"
		}
	}
	
	// Email, URL fields
	if strings.Contains(lower, "email") {
		return "email"
	}
	if strings.Contains(lower, "url") || strings.Contains(lower, "link") {
		return "url"
	}
	
	// Image/file fields
	if strings.Contains(lower, "image") || strings.Contains(lower, "photo") || strings.Contains(lower, "picture") || strings.Contains(lower, "avatar") {
		return "image"
	}
	if strings.Contains(lower, "file") || strings.Contains(lower, "document") || strings.Contains(lower, "attachment") {
		return "file"
	}
	
	// Translated fields (Base framework feature)
	if strings.Contains(lower, "translation") || strings.Contains(lower, "i18n") || strings.Contains(lower, "locale") {
		return "translatedField"
	}
	
	// Default to string for varchar(255)
	return "string"
}

func ProcessField(fieldDef string) []FieldStruct {
	parts := strings.Split(fieldDef, ":")
	
	name := parts[0]
	var fieldType string
	
	// Smart field inference: if only field name provided, infer type
	if len(parts) == 1 {
		fieldType = inferFieldType(name)
	} else {
		fieldType = parts[1]
	}
	var relationship, relatedModel string

	// Check if this is a relationship field
	if len(parts) >= 3 {
		relationship = normalizeRelationship(fieldType)
		relatedModel = parts[2]
		return processRelationshipField(name, relatedModel, relationship)
	}

	// Handle attachment fields
	if fieldType == "attachment" || fieldType == "image" || fieldType == "file" {
		return []FieldStruct{{
			Name:         ToPascalCase(name),
			Type:         "*storage.Attachment",
			JSONName:     ToSnakeCase(name),
			DBName:       ToSnakeCase(name),
			Relationship: "attachment",
			IsRelation:   false, // Attachments are not relations
			GORM:         "polymorphic:Model",
			IsRequired:   false,
		}}
	}

	// Auto-detect relationships based on field name ending with _id
	if strings.HasSuffix(name, "_id") && fieldType == "uint" {
		relationName := strings.TrimSuffix(name, "_id")
		relatedModel := ToPascalCase(relationName)
		
		// Check if this is a core model (common core models)
		coreModels := map[string]string{
			"user":     "users.User",
			"author":   "users.User", // Authors are users too
			"category": "Category",   // App model
			"tag":      "Tag",        // App model  
			"media":    "media.Media",
			"file":     "media.Media",
		}
		
		modelType := relatedModel
		if coreType, exists := coreModels[strings.ToLower(relationName)]; exists {
			modelType = coreType
		}

		// Create both the foreign key field and the relationship field
		goType := GetGoType(fieldType)
		testValue, updateTestValue, testValueWithIndex, testValueUnique, isUnique := getTestValues(goType, name)
		
		return []FieldStruct{
			{
				Name:               ToPascalCase(name),
				Type:               goType,
				JSONName:           ToSnakeCase(name),
				DBName:             ToSnakeCase(name),
				GORM:               "index", // Foreign keys should be indexed
				TestValue:          testValue,
				UpdateTestValue:    updateTestValue,
				TestValueWithIndex: testValueWithIndex,
				TestValueUnique:    testValueUnique,
				IsUnique:           isUnique,
				IsRelation:         false, // Foreign key field is not a relation itself
			},
			{
				Name:               ToPascalCase(relationName),
				Type:               modelType,
				JSONName:           ToSnakeCase(relationName),
				DBName:             ToSnakeCase(relationName),
				Relationship:       "belongs_to",
				RelatedModel:       modelType,
				IsRelation:         true, // This is the actual relation field
				GORM:               fmt.Sprintf("foreignKey:%s", ToPascalCase(name)),
				TestValue:          "nil",
				UpdateTestValue:    "nil",
				TestValueWithIndex: "nil",
				TestValueUnique:    "nil",
			},
		}
	}

	// Handle regular fields
	goType := GetGoType(fieldType)
	testValue, updateTestValue, testValueWithIndex, testValueUnique, isUnique := getTestValues(goType, name)
	
	// Add GORM size for better MySQL field types
	var gormTag string
	switch fieldType {
	case "email":
		gormTag = "size:255;index" // Email should be indexed and have proper size
	case "string":
		gormTag = "size:255" // Default string size
	case "text":
		gormTag = "type:text" // Explicit text type for longer content
	case "url":
		gormTag = "size:512" // URLs can be longer
	case "slug":
		gormTag = "size:255;uniqueIndex" // Slugs should be unique and indexed
	case "decimal":
		gormTag = "type:decimal(10,2)" // Proper decimal precision for money
	case "datetime", "time", "date":
		gormTag = "" // Let GORM handle datetime fields automatically
	default:
		gormTag = "" // No special GORM tag needed
	}
	
	return []FieldStruct{{
		Name:               ToPascalCase(name),
		Type:               goType,
		JSONName:           ToSnakeCase(name),
		DBName:             ToSnakeCase(name),
		GORM:               gormTag,
		TestValue:          testValue,
		UpdateTestValue:    updateTestValue,
		TestValueWithIndex: testValueWithIndex,
		TestValueUnique:    testValueUnique,
		IsUnique:           isUnique,
	}}
}

// GenerateFieldStructs processes all fields and returns a slice of FieldStruct
func GenerateFieldStructs(fields []string) []FieldStruct {
	var fieldStructs []FieldStruct

	for _, fieldDef := range fields {
		processedFields := ProcessField(fieldDef)
		fieldStructs = append(fieldStructs, processedFields...)
	}

	return fieldStructs
}

// GenerateFileFromTemplate generates a file from a template using naming convention
func GenerateFileFromTemplate(dir, filename, templateName string, naming *NamingConvention, fields []FieldStruct) {
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
	case "validator.tmpl":
		tmplContent = validatorTemplate

	case "model_test.tmpl":
		tmplContent = modelTestTemplate
	case "service_test.tmpl":
		tmplContent = serviceTestTemplate
	case "controller_test.tmpl":
		tmplContent = controllerTestTemplate
	default:
		fmt.Printf("Unknown template: %s\n", templateName)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":     strings.ToLower,
		"toTitle":     ToTitle,
		"ToSnakeCase": ToSnakeCase,
		"hasField": func(fields []FieldStruct, fieldType string) bool {
			return HasFieldType(fields, fieldType)
		},
		"ToPascalCase": ToPascalCase,
		"ToKebabCase":  ToKebabCase,
		"ToPlural":     ToPlural,
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

	// Inject parent context variables into each field struct
	enhancedFields := make([]FieldStruct, len(fields))
	copy(enhancedFields, fields)
	for i := range enhancedFields {
		enhancedFields[i].StructName = naming.Model
		enhancedFields[i].LowerName = naming.ModelLower
	}
	
	// Execute template with enhanced data structure
	data := struct {
		*NamingConvention
		Fields        []FieldStruct
		HasImageField bool
	}{
		NamingConvention: naming,
		Fields:           enhancedFields,
		HasImageField:    HasImageField(fields),
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

// GenerateRailsStyleTests generates comprehensive Rails-style tests for a CRUD module
func GenerateTests(naming *NamingConvention, fields []FieldStruct) error {
	// Create test directory structure
	testDir := filepath.Join("test", "app_test", fmt.Sprintf("%s_test", naming.PackageName))
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Generate model tests
	GenerateFileFromTemplate(
		testDir,
		"model_test.go",
		"model_test.tmpl",
		naming,
		fields,
	)

	// Generate service tests
	GenerateFileFromTemplate(
		testDir,
		"service_test.go",
		"service_test.tmpl",
		naming,
		fields,
	)

	// Generate controller tests
	GenerateFileFromTemplate(
		testDir,
		"controller_test.go",
		"controller_test.tmpl",
		naming,
		fields,
	)

	fmt.Printf("Generated Rails-style tests in %s\n", testDir)
	return nil
}
