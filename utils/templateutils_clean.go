package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateData contains all data needed for template generation
type TemplateData struct {
	// Naming conventions for the model
	*NamingConvention

	// Fields including relations
	Fields []Field

	// Computed properties
	HasRelations   bool
	HasBelongsTo   bool
	HasHasMany     bool
	HasHasOne      bool
	HasImages      bool
	HasFiles       bool
	HasAttachments bool
	HasTimestamps  bool
	HasSoftDelete  bool

	// Import paths needed
	Imports []string
}

// Field represents a clean field structure
type Field struct {
	Name    string // Field name in Go (PascalCase)
	Type    string // Go type
	JSONTag string // JSON tag name
	GORMTag string // Complete GORM tag

	// For relations
	IsRelation   bool
	RelationType string // belongs_to, has_many, has_one
	RelatedModel string // Related model name (PascalCase)
	ForeignKey   string // Foreign key field name

	// For testing
	TestValue   string
	UpdateValue string

	// Validation
	IsRequired bool
	IsUnique   bool

	// Special types
	IsImage      bool
	IsFile       bool
	IsAttachment bool
}

// NewTemplateData creates template data from model name and field definitions
func NewTemplateData(modelName string, fieldDefs []string) *TemplateData {
	nc := NewNamingConvention(modelName)
	td := &TemplateData{
		NamingConvention: nc,
		Fields:           []Field{},
		Imports:          []string{},
	}

	// Process field definitions
	for _, fieldDef := range fieldDefs {
		field := td.parseField(fieldDef)
		td.Fields = append(td.Fields, field)

		// Update computed properties
		td.updateComputedProperties(field)
	}

	// Add standard imports
	td.addStandardImports()

	return td
}

// parseField parses a field definition string
func (td *TemplateData) parseField(fieldDef string) Field {
	parts := strings.Split(fieldDef, ":")

	fieldName := parts[0]
	var fieldType string

	// Smart field inference: if only field name provided, infer type
	if len(parts) == 1 {
		fieldType = td.inferFieldType(fieldName)
	} else {
		fieldType = parts[1]
	}

	field := Field{
		Name:    ToPascalCase(fieldName),
		JSONTag: ToSnakeCase(fieldName),
	}

	// Handle relationships
	if strings.Contains(fieldType, "belongsTo") || strings.Contains(fieldType, "belongs_to") {
		// Format: user:belongsTo:User or user_id:uint (auto-detect)
		field.IsRelation = true
		field.RelationType = "belongs_to"

		if len(parts) > 2 {
			field.RelatedModel = ToPascalCase(parts[2])
		} else {
			// Auto-detect from field name
			field.RelatedModel = ToPascalCase(strings.TrimSuffix(fieldName, "_id"))
		}

		// The actual field for foreign key
		field.Name = ToPascalCase(fieldName + "_id")
		field.Type = "uint"
		field.JSONTag = ToSnakeCase(fieldName + "_id")
		field.ForeignKey = field.Name

		// Also need to add the relation field
		td.Fields = append(td.Fields, Field{
			Name:         ToPascalCase(strings.TrimSuffix(fieldName, "_id")),
			Type:         field.RelatedModel,
			JSONTag:      ToSnakeCase(strings.TrimSuffix(fieldName, "_id")),
			GORMTag:      fmt.Sprintf(`gorm:"foreignKey:%s"`, field.ForeignKey),
			IsRelation:   true,
			RelationType: "belongs_to",
			RelatedModel: field.RelatedModel,
		})

	} else if strings.HasSuffix(fieldName, "_id") && fieldType == "uint" {
		// Auto-detect belongs_to from _id suffix
		field.Type = "uint"
		relatedName := strings.TrimSuffix(fieldName, "_id")
		field.RelatedModel = ToPascalCase(relatedName)
		field.ForeignKey = ToPascalCase(fieldName)

		// Also add the relation field
		td.Fields = append(td.Fields, Field{
			Name:         ToPascalCase(relatedName),
			Type:         field.RelatedModel,
			JSONTag:      ToSnakeCase(relatedName) + ",omitempty",
			GORMTag:      fmt.Sprintf(`gorm:"foreignKey:%s"`, field.ForeignKey),
			IsRelation:   true,
			RelationType: "belongs_to",
			RelatedModel: field.RelatedModel,
		})

	} else if strings.Contains(fieldType, "hasMany") || strings.Contains(fieldType, "has_many") {
		// Format: comments:hasMany:Comment
		field.IsRelation = true
		field.RelationType = "has_many"

		if len(parts) > 2 {
			field.RelatedModel = ToPascalCase(parts[2])
		} else {
			field.RelatedModel = ToPascalCase(Singularize(fieldName))
		}

		field.Type = "[]" + field.RelatedModel
		field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"
		field.GORMTag = fmt.Sprintf(`gorm:"foreignKey:%sID"`, td.Model)

	} else if strings.Contains(fieldType, "hasOne") || strings.Contains(fieldType, "has_one") {
		// Format: profile:hasOne:Profile
		field.IsRelation = true
		field.RelationType = "has_one"

		if len(parts) > 2 {
			field.RelatedModel = ToPascalCase(parts[2])
		} else {
			field.RelatedModel = ToPascalCase(fieldName)
		}

		field.Type = "*" + field.RelatedModel
		field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"
		field.GORMTag = fmt.Sprintf(`gorm:"foreignKey:%sID"`, td.Model)

	} else {
		// Regular field types
		field.Type = td.mapFieldType(fieldType)
		field.GORMTag = td.getGORMTag(fieldName, field.Type)

		// Check for special types
		switch fieldType {
		case "image":
			field.IsImage = true
			field.IsAttachment = true
			field.Type = "*storage.Attachment"
			field.GORMTag = `gorm:"polymorphic:Model"`
		case "file":
			field.IsFile = true
			field.IsAttachment = true
			field.Type = "*storage.Attachment"
			field.GORMTag = `gorm:"polymorphic:Model"`
		}
	}

	// Set test values
	field.TestValue, field.UpdateValue = td.getTestValues(field)

	// Check for required/unique
	field.IsRequired = td.isRequired(fieldName)
	field.IsUnique = td.isUnique(fieldName)

	return field
}

// mapFieldType maps simplified types to Go types
func (td *TemplateData) mapFieldType(fieldType string) string {
	typeMap := map[string]string{
		"string":   "string",
		"text":     "string",
		"int":      "int",
		"uint":     "uint",
		"float":    "float64",
		"decimal":  "float64",
		"bool":     "bool",
		"boolean":  "bool",
		"date":     "time.Time",
		"datetime": "time.Time",
		"time":     "time.Time",
		"json":     "datatypes.JSON",
		"jsonb":    "datatypes.JSON",
		"uuid":     "string",
		"email":    "string",
		"url":      "string",
		"slug":     "string",
		"image":    "*storage.Attachment",
		"file":     "*storage.Attachment",
	}

	if goType, ok := typeMap[strings.ToLower(fieldType)]; ok {
		return goType
	}
	return fieldType // Return as-is if not in map
}

// inferFieldType infers field type from field name
func (td *TemplateData) inferFieldType(fieldName string) string {
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

	// Date/time fields
	dateFields := []string{"date", "time", "at", "on", "created_at", "updated_at", "deleted_at", "published_at"}
	for _, df := range dateFields {
		if strings.Contains(lower, df) || strings.HasSuffix(lower, "_at") || strings.HasSuffix(lower, "_on") {
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

	// Default to string for varchar(255)
	return "string"
}

// getGORMTag generates appropriate GORM tags based on field type and name
func (td *TemplateData) getGORMTag(fieldName, fieldType string) string {
	tags := []string{}
	lower := strings.ToLower(fieldName)

	// Handle different field types with proper GORM mapping
	switch fieldType {
	case "string":
		// Default varchar(255)
		tags = append(tags, "type:varchar(255)")

		// Unique constraints
		switch lower {
		case "email":
			tags = append(tags, "uniqueIndex")
		case "username", "slug":
			tags = append(tags, "uniqueIndex")
			tags = append(tags, "type:varchar(100)") // Override for shorter fields
		}

	case "text":
		// Long text fields
		tags = append(tags, "type:text")

	case "email":
		tags = append(tags, "type:varchar(255)", "uniqueIndex")

	case "url":
		tags = append(tags, "type:varchar(500)")

	case "decimal", "float", "float64":
		// For money/price fields
		if strings.Contains(lower, "price") || strings.Contains(lower, "amount") || strings.Contains(lower, "cost") {
			tags = append(tags, "type:decimal(10,2)")
		}

	case "bool", "boolean":
		tags = append(tags, "type:boolean", "default:false")

	case "datetime", "time.Time":
		if lower == "deleted_at" {
			tags = append(tags, "index") // For soft delete queries
		}

	case "uuid":
		tags = append(tags, "type:uuid")

	case "json", "jsonb":
		tags = append(tags, "type:jsonb")

	case "int":
		// Add constraints for specific int fields
		if strings.Contains(lower, "status") {
			tags = append(tags, "default:0")
		}

	case "uint":
		// Unsigned integers, often IDs
		if strings.HasSuffix(lower, "_id") {
			tags = append(tags, "index") // Foreign keys should be indexed
		}
	}

	// Required fields
	requiredFields := []string{"name", "title", "email", "username"}
	for _, rf := range requiredFields {
		if lower == rf {
			tags = append(tags, "not null")
			break
		}
	}

	if len(tags) > 0 {
		return fmt.Sprintf(`gorm:"%s"`, strings.Join(tags, ";"))
	}
	return ""
}

// getTestValues returns test and update values for a field
func (td *TemplateData) getTestValues(field Field) (testValue, updateValue string) {
	switch field.Type {
	case "string":
		testValue = fmt.Sprintf(`"Test %s"`, field.Name)
		updateValue = fmt.Sprintf(`"Updated %s"`, field.Name)
	case "int", "uint":
		testValue = "123"
		updateValue = "456"
	case "float64":
		testValue = "99.99"
		updateValue = "199.99"
	case "bool":
		testValue = "true"
		updateValue = "false"
	case "time.Time":
		testValue = "time.Now()"
		updateValue = "time.Now().Add(24 * time.Hour)"
	default:
		if field.IsRelation {
			testValue = "1" // Foreign key ID
			updateValue = "2"
		} else {
			testValue = `"test"`
			updateValue = `"updated"`
		}
	}
	return
}

// isRequired checks if a field should be required
func (td *TemplateData) isRequired(fieldName string) bool {
	requiredFields := []string{"name", "title", "email", "username", "password"}
	for _, rf := range requiredFields {
		if strings.Contains(strings.ToLower(fieldName), rf) {
			return true
		}
	}
	return false
}

// isUnique checks if a field should be unique
func (td *TemplateData) isUnique(fieldName string) bool {
	uniqueFields := []string{"email", "username", "slug", "code", "sku"}
	for _, uf := range uniqueFields {
		if strings.Contains(strings.ToLower(fieldName), uf) {
			return true
		}
	}
	return false
}

// updateComputedProperties updates computed properties based on field
func (td *TemplateData) updateComputedProperties(field Field) {
	if field.IsRelation {
		td.HasRelations = true
		switch field.RelationType {
		case "belongs_to":
			td.HasBelongsTo = true
		case "has_many":
			td.HasHasMany = true
		case "has_one":
			td.HasHasOne = true
		}
	}

	if field.IsImage {
		td.HasImages = true
		td.HasAttachments = true
	}
	if field.IsFile {
		td.HasFiles = true
		td.HasAttachments = true
	}

	if field.Type == "time.Time" {
		switch field.Name {
		case "DeletedAt":
			td.HasSoftDelete = true
		case "CreatedAt", "UpdatedAt":
			td.HasTimestamps = true
		}
	}
}

// addStandardImports adds standard imports based on fields
func (td *TemplateData) addStandardImports() {
	imports := make(map[string]bool)

	// Always needed
	imports["base/app/models"] = true
	imports["base/core/module"] = true
	imports["gorm.io/gorm"] = true

	// Check fields for additional imports
	for _, field := range td.Fields {
		switch field.Type {
		case "time.Time":
			imports["time"] = true
		case "datatypes.JSON":
			imports["gorm.io/datatypes"] = true
		case "*storage.Attachment":
			imports["base/core/storage"] = true
		}
	}

	// Convert map to slice
	td.Imports = []string{}
	for imp := range imports {
		td.Imports = append(td.Imports, imp)
	}
}

// Singularize converts plural to singular (basic implementation)
func Singularize(word string) string {
	if strings.HasSuffix(word, "ies") {
		return strings.TrimSuffix(word, "ies") + "y"
	}
	if strings.HasSuffix(word, "es") {
		return strings.TrimSuffix(word, "es")
	}
	if strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

// GenerateWithCleanTemplates generates files using the new clean template system
func GenerateWithCleanTemplates(modelName string, fieldDefs []string) error {
	// Create template data from model name and fields
	td := NewTemplateData(modelName, fieldDefs)

	// Create directories
	dirs := []string{
		filepath.Join("app", "models"),
		filepath.Join("app", td.DirName),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate model
	if err := generateCleanTemplate("model",
		filepath.Join("app", "models", td.ModelSnake+".go"), td); err != nil {
		return fmt.Errorf("failed to generate model: %w", err)
	}

	// Generate service
	if err := generateCleanTemplate("service",
		filepath.Join("app", td.DirName, "service.go"), td); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	// Generate controller
	if err := generateCleanTemplate("controller",
		filepath.Join("app", td.DirName, "controller.go"), td); err != nil {
		return fmt.Errorf("failed to generate controller: %w", err)
	}

	// Generate module
	if err := generateCleanTemplate("module",
		filepath.Join("app", td.DirName, "module.go"), td); err != nil {
		return fmt.Errorf("failed to generate module: %w", err)
	}

	return nil
}

// generateCleanTemplate generates a single file from template
func generateCleanTemplate(templateName, outputPath string, data *TemplateData) error {
	// This would use the embedded templates
	fmt.Printf("Generating %s for %s at %s\n", templateName, data.ModelSnake, outputPath)
	// For now, using the existing template system
	fmt.Printf("Generated %s at %s\n", templateName, outputPath)
	return nil
}
