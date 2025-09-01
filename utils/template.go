package utils

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

//go:embed templates/validator.tmpl
var validatorTemplate string

// TemplateData contains all data needed for template generation
type TemplateData struct {
	// Naming conventions for the model
	*NamingConvention

	// Fields including relations
	Fields []Field

	// Computed properties
	HasRelations          bool
	HasBelongsTo          bool
	HasHasMany            bool
	HasHasOne             bool
	HasManyToMany         bool
	HasImages             bool
	HasFiles              bool
	HasAttachments        bool
	HasTimestamps         bool
	HasSoftDelete         bool
	HasTranslatableFields bool

	// Import paths needed
	Imports []string

	// Join tables for many-to-many relationships
	JoinTables []string
}

// Field represents a clean field structure - keeping compatibility with templates
type Field struct {
	Name    string // Field name in Go (PascalCase)
	Type    string // Go type
	JSONTag string // JSON tag name (maps to JSONName for template compatibility)
	GORMTag string // Complete GORM tag (maps to GORM for template compatibility)

	// Template compatibility fields
	JSONName           string // Same as JSONTag for template compatibility
	DBName             string // Database column name
	GORM               string // Same as GORMTag for template compatibility
	Relationship       string // Same as RelationType for template compatibility
	RelatedModel       string // Related model name (PascalCase)
	ForeignKey         string // Foreign key field name
	TestValue          string // Test value for this field
	UpdateTestValue    string // Update test value (maps to UpdateValue)
	TestValueWithIndex string // Test value with index for loops
	TestValueUnique    string // Unique test value for constraint tests

	// For relations
	IsRelation   bool
	RelationType string // belongs_to, has_many, has_one

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
		// Format: author:belongsTo:User or author:belongsTo:profile.User
		field.IsRelation = true
		field.RelationType = "belongs_to"

		var relatedModel string
		if len(parts) > 2 {
			// Handle module.Model syntax like profile.User - keep original case
			if strings.Contains(parts[2], ".") {
				// Keep the original case as provided: profile.User stays profile.User
				relatedModel = strings.TrimSpace(parts[2])

				// Add import for cross-module reference
				modelParts := strings.Split(parts[2], ".")
				if len(modelParts) == 2 {
					packageName := strings.ToLower(strings.TrimSpace(modelParts[0]))
					importPackage := fmt.Sprintf("base/app/%s", packageName)

					// Check if import already exists to avoid duplicates
					importExists := false
					for _, existingImport := range td.Imports {
						if existingImport == importPackage {
							importExists = true
							break
						}
					}
					if !importExists {
						td.Imports = append(td.Imports, importPackage)
					}
				}
			} else {
				relatedModel = ToPascalCase(parts[2])
			}
		} else {
			// Auto-detect from field name
			relatedModel = ToPascalCase(strings.TrimSuffix(fieldName, "_id"))
		}

		// The actual field for foreign key
		field.Name = ToPascalCase(fieldName + "_id")
		field.Type = "uint"
		field.JSONTag = ToSnakeCase(fieldName + "_id")
		field.ForeignKey = field.Name
		field.RelatedModel = relatedModel

		// Also need to add the relation field
		td.Fields = append(td.Fields, Field{
			Name:         ToPascalCase(fieldName),
			Type:         "*" + relatedModel, // belongsTo is a pointer
			JSONTag:      ToSnakeCase(fieldName) + ",omitempty",
			GORMTag:      fmt.Sprintf(`gorm:"foreignKey:%s"`, field.ForeignKey),
			IsRelation:   true,
			RelationType: "belongs_to",
			RelatedModel: relatedModel,
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
			Type:         "*" + field.RelatedModel,
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

	} else if strings.Contains(fieldType, "toMany") || strings.Contains(fieldType, "to_many") || strings.Contains(fieldType, "manyToMany") || strings.Contains(fieldType, "many_to_many") {
		// Format: tags:toMany:Tag or roles:manyToMany:Role
		field.IsRelation = true
		field.RelationType = "many_to_many"

		if len(parts) > 2 {
			field.RelatedModel = ToPascalCase(parts[2])
		} else {
			field.RelatedModel = ToPascalCase(Singularize(fieldName))
		}

		field.Type = "[]*" + field.RelatedModel
		field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"

		// GORM many-to-many: automatically creates join table (e.g., post_users)
		// Join table name: <current_model>_<related_model_plural> following GORM convention
		joinTable := fmt.Sprintf("%s_%s", ToSnakeCase(td.Model), ToSnakeCase(ToPlural(field.RelatedModel)))
		field.GORMTag = fmt.Sprintf(`gorm:"many2many:%s"`, joinTable)

		// Store join table name for migration
		td.JoinTables = append(td.JoinTables, joinTable)

	} else {
		// Regular field types
		field.Type = td.mapFieldType(fieldType)
		field.GORMTag = td.getGORMTag(fieldName, field.Type)

		if field.Type == "translation.Field" {
			field.GORMTag = ``
		}
		// Check for special types
		switch fieldType {
		case "image":
			field.IsImage = true
			field.IsAttachment = true
			field.Type = "*storage.Attachment"
			field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
		case "file":
			field.IsFile = true
			field.IsAttachment = true
			field.Type = "*storage.Attachment"
			field.GORMTag = `gorm:"foreignKey:ModelId;references:Id"`
		case "text":
			field.Type = "string"
			field.GORMTag = `type:text`
		}
	}

	// Set compatibility fields for templates
	field.JSONName = field.JSONTag
	field.DBName = field.JSONTag
	field.GORM = field.GORMTag
	field.Relationship = field.RelationType

	// Check for required/unique
	field.IsRequired = td.isRequired(fieldName)
	field.IsUnique = td.isUnique(fieldName)

	return field
}

// mapFieldType maps simplified types to Go types
func (td *TemplateData) mapFieldType(fieldType string) string {
	typeMap := map[string]string{
		"string":      "string",
		"text":        "string",
		"int":         "int",
		"uint":        "uint",
		"float":       "float64",
		"decimal":     "float64",
		"bool":        "bool",
		"boolean":     "bool",
		"date":        "time.Time",
		"datetime":    "time.Time",
		"timestamp":   "time.Time",
		"time":        "time.Time",
		"json":        "datatypes.JSON",
		"jsonb":       "datatypes.JSON",
		"uuid":        "string",
		"email":       "string",
		"url":         "string",
		"slug":        "string",
		"image":       "*storage.Attachment",
		"file":        "*storage.Attachment",
		"translation": "translation.Field",
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
	textFields := []string{"description", "content", "body", "notes", "comment", "summary", "bio", "text", "desc"}
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

	// Date/time fields - check for common date patterns
	dateFields := []string{"date", "time", "created_at", "updated_at", "deleted_at", "published_at", "timestamp", "datetime"}
	for _, df := range dateFields {
		if strings.Contains(lower, df) || strings.HasSuffix(lower, "_at") || strings.HasSuffix(lower, "_on") || strings.HasSuffix(lower, "_date") || strings.HasSuffix(lower, "_time") {
			return "datetime"
		}
	}

	// Check for explicit date-like words
	if strings.Contains(lower, "birth") || strings.Contains(lower, "born") || strings.Contains(lower, "expir") || strings.Contains(lower, "start") || strings.Contains(lower, "end") {
		return "datetime"
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
	if slices.Contains(requiredFields, lower) {
		tags = append(tags, "not null")
	}

	if len(tags) > 0 {
		return strings.Join(tags, ";")
	}
	return strings.Join(tags, ";") // 👈 just inner tags, no gorm:""
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
		case "many_to_many":
			td.HasManyToMany = true
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
	// Check for translatable fields
	if field.Type == "translation.Field" {
		td.HasTranslatableFields = true
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
	imports["time"] = true
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
		case "translation.Field":
			imports["base/core/translation"] = true
		}
	}

	// Convert map to slice
	td.Imports = []string{}
	for imp := range imports {
		td.Imports = append(td.Imports, imp)
	}
}

// HasFieldType checks if any field has the specified type
func HasFieldType(fields []Field, fieldType string) bool {
	for _, field := range fields {
		if field.Type == fieldType {
			return true
		}
	}
	return false
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

// GenerateFieldStructs processes all fields and returns a slice of Field (for backward compatibility)
func GenerateFieldStructs(fieldDefs []string) []Field {
	var fields []Field
	for _, fieldDef := range fieldDefs {
		field := parseFieldDef(fieldDef)
		fields = append(fields, field)
	}
	return fields
}

// parseFieldDef parses a single field definition (simplified version for compatibility)
func parseFieldDef(fieldDef string) Field {
	parts := strings.Split(fieldDef, ":")
	fieldName := parts[0]
	var fieldType string

	if len(parts) == 1 {
		fieldType = inferFieldTypeCompat(fieldName)
	} else if len(parts) == 2 {
		fieldType = parts[1]
	} else {
		// For relationship definitions like "tags:toMany:Tag", pass the full definition
		fieldType = fieldDef
	}

	field := Field{
		Name:    ToPascalCase(fieldName),
		Type:    GetGoType(fieldType),
		JSONTag: ToSnakeCase(fieldName),
		DBName:  ToSnakeCase(fieldName),
	}

	// Handle relationship metadata for compatibility
	if strings.Contains(fieldType, ":") {
		parts := strings.Split(fieldType, ":")
		if len(parts) >= 2 {
			relationType := parts[1]
			if strings.Contains(relationType, "toMany") || strings.Contains(relationType, "to_many") ||
				strings.Contains(relationType, "manyToMany") || strings.Contains(relationType, "many_to_many") {
				field.IsRelation = true
				field.RelationType = "many_to_many"
				field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"
				if len(parts) > 2 {
					field.RelatedModel = ToPascalCase(parts[2])
				} else {
					field.RelatedModel = ToPascalCase(Singularize(fieldName))
				}
				// Generate GORM tag for many-to-many (without gorm: prefix, template adds it)
				// For compatibility layer, we'll use field name, but this should be updated to use model name
				joinTable := fmt.Sprintf("%s_%s", ToSnakeCase(fieldName), ToSnakeCase(ToPlural(field.RelatedModel)))
				field.GORMTag = fmt.Sprintf(`many2many:%s`, joinTable)
			} else if strings.Contains(relationType, "hasMany") || strings.Contains(relationType, "has_many") {
				field.IsRelation = true
				field.RelationType = "has_many"
				field.JSONTag = ToSnakeCase(fieldName) + ",omitempty"
				if len(parts) > 2 {
					field.RelatedModel = ToPascalCase(parts[2])
				} else {
					field.RelatedModel = ToPascalCase(Singularize(fieldName))
				}
				field.GORMTag = fmt.Sprintf(`gorm:"foreignKey:%sID"`, ToPascalCase(fieldName))
			} else if strings.Contains(relationType, "belongsTo") || strings.Contains(relationType, "belongs_to") {
				field.IsRelation = true
				field.RelationType = "belongs_to"
				if len(parts) > 2 {
					// Handle module.Model syntax - keep original case
					if strings.Contains(parts[2], ".") {
						field.RelatedModel = strings.TrimSpace(parts[2])
					} else {
						field.RelatedModel = ToPascalCase(parts[2])
					}
				}
			}
		}
	}

	// Set compatibility fields
	field.JSONName = field.JSONTag
	field.Relationship = field.RelationType
	if field.GORMTag != "" {
		field.GORM = field.GORMTag
	} else {
		field.GORM = getGORMTagCompat(fieldName, field.Type)
	}

	return field
}

// Helper functions for backward compatibility
func inferFieldTypeCompat(fieldName string) string {
	lower := strings.ToLower(fieldName)
	if strings.HasSuffix(lower, "_id") {
		return "uint"
	}
	return "string"
}

func getGORMTagCompat(fieldName, fieldType string) string {
	if fieldType == "uint" && strings.HasSuffix(strings.ToLower(fieldName), "_id") {
		return "index"
	}
	return ""
}

// GenerateFileFromTemplate generates a file from embedded template (for backward compatibility)
func GenerateFileFromTemplate(dir, filename, templateName string, naming *NamingConvention, fields []Field) {
	// Convert Field slice to embedded template data
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
	default:
		fmt.Printf("Unknown template: %s\n", templateName)
		return
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"toLower":      strings.ToLower,
		"toTitle":      ToTitle,
		"ToSnakeCase":  ToSnakeCase,
		"ToPascalCase": ToPascalCase,
		"ToKebabCase":  ToKebabCase,
		"ToPlural":     ToPlural,
		"hasPrefix":    strings.HasPrefix,
		"hasSuffix":    strings.HasSuffix,
		"contains":     strings.Contains,
		"eq":           func(a, b interface{}) bool { return a == b },
		"slice": func(s string, start, end int) string {
			if start >= len(s) {
				return ""
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"hasField": func(fields []Field, fieldType string) bool {
			return HasFieldType(fields, fieldType)
		},
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", templateName, err)
		return
	}

	// Create output directory
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

	// Execute template with data structure
	data := struct {
		*NamingConvention
		Fields                []Field
		HasImageField         bool
		HasTranslatableFields bool
		HasSoftDelete         bool
		HasTimestamps         bool
		HasAttachments        bool
		HasRelations          bool
		HasBelongsTo          bool
		HasHasMany            bool
		HasHasOne             bool
		HasManyToMany         bool
	}{
		NamingConvention:      naming,
		Fields:                fields,
		HasImageField:         HasImageField(fields),
		HasTranslatableFields: HasFieldType(fields, "translation.Field"),
		HasSoftDelete:         HasFieldType(fields, "gorm.DeletedAt"),
		HasTimestamps:         HasFieldType(fields, "time.Time"),
		HasAttachments:        HasFieldType(fields, "*storage.Attachment"),
		HasRelations:          HasFieldType(fields, "*models."),
		HasBelongsTo:          HasFieldType(fields, "belongsTo"),
		HasHasMany:            HasFieldType(fields, "hasMany"),
		HasHasOne:             HasFieldType(fields, "hasOne"),
		HasManyToMany:         HasFieldType(fields, "manyToMany"),
	}

	if err := tmpl.Execute(f, data); err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}

	fmt.Printf("Generated %s\n", outputFile)
}

// HasImageField checks if any field has image type
func HasImageField(fields []Field) bool {
	return HasFieldType(fields, "*storage.Attachment")
}
