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
	case "int":
		return "int"
	case "string", "text":
		return "string"
	case "datetime", "time", "date":
		return "types.DateTime"
	case "float":
		return "float64"
	case "sort":
		return "int"
	case "bool":
		return "bool"
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

	markerIndex := strings.Index(contentStr, "// MODULE_INITIALIZER_MARKER")
	if markerIndex == -1 {
		return content, false
	}

	if strings.Contains(contentStr[:markerIndex], fmt.Sprintf(`"%s":`, packageName)) {
		return content, false
	}

	structName := ToPascalCase(singularName)

	newInitializer := fmt.Sprintf(`	"%s": func(db *gorm.DB, router *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, storage *storage.ActiveStorage) module.Module { return %s.New%sModule(db, router, log, emitter, storage) },`,
		packageName, packageName, structName)

	updatedContent := contentStr[:markerIndex] + newInitializer + "\n        " + contentStr[markerIndex:]

	return []byte(updatedContent), true
}

func UpdateInitFileForDestroy(singularName string) error {
	// Convert names consistently with generate function
	dirName := ToSnakeCase(singularName)
	pluralName := ToPlural(singularName)
	packageName := ToSnakeCase(pluralName)

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

	// Check if module exists in init.go
	modulePattern := fmt.Sprintf(`"%s":\s*func\(db \*gorm\.DB,`, packageName)
	if !regexp.MustCompile(modulePattern).MatchString(contentStr) {
		return fmt.Errorf("module '%s' not found in init.go", packageName)
	}

	// Remove import while preserving formatting
	importPath := fmt.Sprintf(`"base/app/%s"`, packageName)
	importMarker := "// MODULE_IMPORT_MARKER"

	// Find the marker and remove the import
	markerIndex := strings.Index(contentStr, importMarker)
	if markerIndex != -1 {
		markerEnd := strings.Index(contentStr[markerIndex:], "\n") + markerIndex
		if markerEnd != -1 {
			// Look for the import line after the marker
			remainingContent := contentStr[markerEnd:]
			importPattern := regexp.MustCompile(`\n\s*` + regexp.QuoteMeta(importPath) + `\n`)
			contentStr = contentStr[:markerEnd+1] + importPattern.ReplaceAllString(remainingContent, "\n")
		}
	}

	// Remove module initializer
	initializerPattern := fmt.Sprintf(`\n\s*"%s":\s*func\(db \*gorm\.DB,\s*router \*gin\.RouterGroup,\s*log logger\.Logger,\s*emitter \*emitter\.Emitter,\s*activeStorage \*storage\.ActiveStorage\) module\.Module \{[^}]+\},\n`, packageName)
	re := regexp.MustCompile(initializerPattern)

	// Find and remove the module initializer
	contentStr = re.ReplaceAllString(contentStr, "\n")

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
