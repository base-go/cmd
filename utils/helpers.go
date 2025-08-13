package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"path/filepath"
	"regexp"

	"github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var PluralizeClient *pluralize.Client

func init() {
	PluralizeClient = pluralize.NewClient()
}

func GetGoType(t string) string {
	switch t {
	// Exact Go types - pass through as-is
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return t
	case "float32", "float64":
		return t
	case "bool":
		return t
	case "string":
		return t
	case "byte", "rune":
		return t
		
	// Base framework special types
	case "translatedField":
		return "translation.Field"
		
	// Convenience aliases
	case "text", "email", "url", "slug":
		return "string"
	case "datetime", "time", "date":
		return "types.DateTime"
	case "float", "decimal":
		return "float64"
	case "sort":
		return "int"
	case "image", "file":
		return "*storage.Attachment"
		
	// Default: assume it's already a valid Go type or custom type
	default:
		return t
	}
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

func ToTitle(s string) string {
	return cases.Title(language.Und).String(s)
}

func ToLowerPlural(s string) string {
	return strings.ToLower(PluralizeClient.Plural(s))
}

func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

func ToCamelCase(s string) string {
	s = ToPascalCase(s)
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func ToPascalCase(s string) string {
	words := splitIntoWords(s)
	for i, word := range words {
		words[i] = cases.Title(language.Und).String(word)
	}
	return strings.Join(words, "")
}

func splitIntoWords(s string) []string {
	var words []string
	var currentWord strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(s[i-1])) || unicode.IsLower(r)) {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		if r == '_' || r == ' ' || r == '-' {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else {
			currentWord.WriteRune(r)
		}
	}
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}
	return words
}

func ToPlural(s string) string {
	return PluralizeClient.Plural(s)
}

func UpdateInitFile(singularName, pluralName string) error {
	initFilePath := "app/init.go"

	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	packageName := ToSnakeCase(singularName)

	importStr := fmt.Sprintf("\"base/app/%s\"", ToSnakeCase(pluralName))
	content, importAdded := AddImport(content, importStr)

	content, initializerAdded := AddModuleInitializer(content, packageName, singularName)

	if importAdded || initializerAdded {
		return os.WriteFile(initFilePath, content, 0644)
	}

	return nil
}

func AddModuleInitializer(content []byte, packageName, singularName string) ([]byte, bool) {
	contentStr := string(content)

	// Find the return modules line to add before it
	returnIndex := strings.Index(contentStr, "return modules")
	if returnIndex == -1 {
		return content, false
	}

	if strings.Contains(contentStr[:returnIndex], fmt.Sprintf(`modules["%s"]`, packageName)) {
		return content, false
	}

	structName := ToPascalCase(singularName)

	newInitializer := fmt.Sprintf(`	modules["%s"] = %s.New%sModule(deps.DB)
`,
		packageName, packageName, structName)

	// Find the line before return modules
	lineStart := strings.LastIndex(contentStr[:returnIndex], "\n") + 1
	
	updatedContent := contentStr[:lineStart] + newInitializer + contentStr[lineStart:]

	return []byte(updatedContent), true
}

func UpdateInitFileForDestroy(singularName string) error {
	// Convert names consistently with generate function
	dirName := ToSnakeCase(singularName)
	pluralName := ToPlural(singularName)
	packageName := ToSnakeCase(pluralName)
	
	// For module key lookup, try both singular and plural forms
	// Prioritize plural since that's the new convention
	moduleKeys := []string{packageName, singularName, dirName}

	// Paths - use the same paths as generateModule
	initFilePath := "app/init.go"
	moduleDir := filepath.Join("app", packageName)
	modelFile := filepath.Join("app", "models", fmt.Sprintf("%s.go", dirName))

	// Read init.go first
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return fmt.Errorf("error reading init.go: %v", err)
	}

	contentStr := string(content)

	// Check if module exists in init.go - try all possible module keys
	var foundModuleKey string
	for _, key := range moduleKeys {
		simplePattern := fmt.Sprintf(`modules["%s"]`, key)
		complexPattern := fmt.Sprintf(`"%s":\s*func\(db \*gorm\.DB,`, key)
		
		if strings.Contains(contentStr, simplePattern) || regexp.MustCompile(complexPattern).MatchString(contentStr) {
			foundModuleKey = key
			break
		}
	}
	
	if foundModuleKey == "" {
		return fmt.Errorf("module '%s' not found in init.go (tried keys: %v)", singularName, moduleKeys)
	}

	// Remove import while preserving formatting - try both directory patterns
	// Prioritize plural form since that's the new convention
	importPaths := []string{
		fmt.Sprintf(`"base/app/%s"`, packageName),
		fmt.Sprintf(`"base/app/%s"`, foundModuleKey),
		fmt.Sprintf(`"base/app/%s"`, dirName),
	}
	importMarker := "// MODULE_IMPORT_MARKER"

	// Remove any matching import
	for _, importPath := range importPaths {
		markerIndex := strings.Index(contentStr, importMarker)
		if markerIndex != -1 {
			markerEnd := strings.Index(contentStr[markerIndex:], "\n") + markerIndex
			if markerEnd != -1 {
				// Look for the import line after the marker
				remainingContent := contentStr[markerEnd:]
				importPattern := regexp.MustCompile(`\n\s*` + regexp.QuoteMeta(importPath) + `\s*\n`)
				if importPattern.MatchString(remainingContent) {
					contentStr = contentStr[:markerEnd+1] + importPattern.ReplaceAllString(remainingContent, "\n")
					break
				}
			}
		} else {
			// Fallback: remove import anywhere in the import block
			importPattern := regexp.MustCompile(`\s*` + regexp.QuoteMeta(importPath) + `\s*\n`)
			if importPattern.MatchString(contentStr) {
				contentStr = importPattern.ReplaceAllString(contentStr, "")
				break
			}
		}
	}

	// Remove module initializer - handle both simple and complex patterns using the found key
	// Simple pattern: modules["products"] = products.NewProductModule(deps.DB)
	simplePattern := fmt.Sprintf(`\s*modules\["%s"\]\s*=\s*%s\.New[^(]+\(deps\.DB\)\s*\n`, foundModuleKey, foundModuleKey)
	// Complex pattern: "products": func(db *gorm.DB, ...) { return products.NewProductModule(...) },
	complexPattern := fmt.Sprintf(`\s*"%s":\s*func\(db \*gorm\.DB,\s*router \*gin\.RouterGroup,\s*log logger\.Logger,\s*emitter \*emitter\.Emitter,\s*activeStorage \*storage\.ActiveStorage\) module\.Module \{[^}]+\},\s*\n`, foundModuleKey)
	
	simpleRe := regexp.MustCompile(simplePattern)
	complexRe := regexp.MustCompile(complexPattern)

	// Try to remove simple pattern first
	if simpleRe.MatchString(contentStr) {
		contentStr = simpleRe.ReplaceAllString(contentStr, "")
	} else if complexRe.MatchString(contentStr) {
		contentStr = complexRe.ReplaceAllString(contentStr, "")
	}

	// Ensure proper formatting
	contentStr = regexp.MustCompile(`\n{3,}`).ReplaceAllString(contentStr, "\n\n")
	contentStr = regexp.MustCompile(`\{\n\n(\s*)//`).ReplaceAllString(contentStr, "{\n$1//")
	contentStr = regexp.MustCompile(`\n\n(\s*)\}`).ReplaceAllString(contentStr, "\n$1}")

	// Write updated init.go
	if err := os.WriteFile(initFilePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("error writing to init.go: %v", err)
	}

	// Try to remove module directory
	if _, err := os.Stat(moduleDir); err == nil {
		if err := os.RemoveAll(moduleDir); err != nil {
			fmt.Printf("Warning: could not remove module directory %s: %v\n", moduleDir, err)
		} else {
			fmt.Printf("Removed module directory: %s\n", moduleDir)
		}
	}

	// Try to remove model file
	if _, err := os.Stat(modelFile); err == nil {
		if err := os.Remove(modelFile); err != nil {
			fmt.Printf("Warning: could not remove model file %s: %v\n", modelFile, err)
		} else {
			fmt.Printf("Removed model file: %s\n", modelFile)
		}
	}

	fmt.Printf("Successfully removed module '%s'\n", packageName)
	return nil
}

func RemoveImport(content []byte, importStr string) []byte {
	contentStr := string(content)
	importIndex := strings.Index(contentStr, importStr)
	if importIndex != -1 {
		// Find the start of the line
		lineStart := strings.LastIndex(contentStr[:importIndex], "\n") + 1
		// Find the end of the line
		lineEnd := strings.Index(contentStr[importIndex:], "\n") + importIndex
		if lineEnd == -1 {
			lineEnd = len(contentStr)
		}
		// Remove the line
		contentStr = contentStr[:lineStart] + contentStr[lineEnd:]
	}
	return []byte(contentStr)
}

func RemoveModuleInitializer(content []byte, pluralName string) []byte {
	contentStr := string(content)
	// Look for the module initializer
	moduleStr := fmt.Sprintf(`"%s":`, pluralName)
	initializerStart := strings.Index(contentStr, moduleStr)
	if initializerStart != -1 {
		// Find the start of the line
		lineStart := strings.LastIndex(contentStr[:initializerStart], "\n") + 1
		// Find the end of the module block (closing brace followed by comma)
		initializerEnd := strings.Index(contentStr[initializerStart:], "},") + initializerStart + 2
		if initializerEnd > initializerStart {
			// Remove any trailing newline
			if len(contentStr) > initializerEnd && contentStr[initializerEnd] == '\n' {
				initializerEnd++
			}
			contentStr = contentStr[:lineStart] + contentStr[initializerEnd:]
		}
	}
	return []byte(contentStr)
}

func AddImport(content []byte, importStr string) ([]byte, bool) {
	if bytes.Contains(content, []byte(importStr)) {
		return content, false
	}

	importPos := bytes.Index(content, []byte("import ("))
	if importPos == -1 {
		return content, false
	}

	insertPos := importPos + len("import (") + 1

	newImportLine := []byte("\t" + importStr + "\n")

	updatedContent := append(content[:insertPos], append(newImportLine, content[insertPos:]...)...)

	return updatedContent, true
}

func GetRequiredImports(fields []FieldStruct) map[string][]string {
	modelImports := []string{
		"time",
		"github.com/google/uuid",
	}
	serviceImports := []string{
		"context",
		"errors",
		"fmt",
	}

	// Add conditional imports based on field types
	if HasFieldType(fields, "*storage.File") {
		modelImports = append(modelImports,
			"gorm.io/gorm",
			"base/core/storage",
		)
		serviceImports = append(serviceImports,
			"mime/multipart",
		)
	}

	return map[string][]string{
		"model":   modelImports,
		"service": serviceImports,
	}
}
