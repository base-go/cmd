package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var pluralizeClient *pluralize.Client

func init() {
	pluralizeClient = pluralize.NewClient()
}

func GetGoType(t string) string {
	switch t {
	case "int":
		return "int"
	case "string", "text":
		return "string"
	case "datetime", "time":
		return "time.Time"
	case "float":
		return "float64"
	case "sort":
		return "int"
	case "bool":
		return "bool"
	case "file", "image", "attachment":
		return "*storage.Attachment"
	default:
		return t
	}
}
func getExampleValue(t string) string {
	switch t {
	case "string", "text":
		return "example string"
	case "int", "int64", "uint", "uint64":
		return "1"
	case "float", "float64":
		return "1.0"
	case "bool":
		return "true"
	case "time.Time":
		return "2024-01-01T00:00:00Z"
	case "file", "image", "attachment":
		return "null"
	default:
		return ""
	}
}
func ToLower(s string) string {
	return strings.ToLower(s)
}

func ToTitle(s string) string {
	return cases.Title(language.Und).String(s)
}

func ToLowerPlural(s string) string {
	return strings.ToLower(pluralizeClient.Plural(s))
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
	return pluralizeClient.Plural(s)
}

// HasFileField checks if any field is a file type
func HasFileField(fields []FieldStruct) bool {
	for _, field := range fields {
		if isFileField(field) {
			return true
		}
	}
	return false
}

// isFileField is an internal helper to check if a field is a file type
func isFileField(field FieldStruct) bool {
	return field.Type == "file" ||
		field.Type == "image" ||
		field.Type == "attachment" ||
		field.Type == "*storage.Attachment"
}

func isRelationshipField(field FieldStruct, relType string) bool {
	return field.Relationship == relType ||
		field.Relationship == strings.Replace(relType, "_", "", -1)
}
func UpdateInitFile(singularName string, hasFileFields bool) error {
	initFilePath := "app/init.go"

	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	packageName := ToSnakeCase(singularName)
	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
	content, importAdded := AddImport(content, importStr)
	content, initializerAdded := AddModuleInitializer(content, packageName, singularName, hasFileFields)

	if importAdded || initializerAdded {
		return os.WriteFile(initFilePath, content, 0644)
	}

	return nil
}

func AddModuleInitializer(content []byte, packageName, singularName string, hasFileFields bool) ([]byte, bool) {
	contentStr := string(content)
	marker := "// MODULE_INITIALIZER_MARKER"
	markerIndex := strings.Index(contentStr, marker)
	if markerIndex == -1 {
		return content, false
	}

	lineStart := strings.LastIndex(contentStr[:markerIndex], "\n") + 1
	indentation := contentStr[lineStart:markerIndex]
	mapKeyPackageName := ToSnakeCase(packageName)

	if strings.Contains(contentStr[:markerIndex], fmt.Sprintf(`"%s":`, mapKeyPackageName)) {
		return content, false
	}

	structName := ToPascalCase(singularName)
	importPackageName := ToSnakeCase(singularName)

	var newInitializer string
	if hasFileFields {
		newInitializer = fmt.Sprintf(`%s"%s": func(db *gorm.DB, router *gin.RouterGroup, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger *zap.Logger, eventService *event.EventService) module.Module {
            return %s.New%sModule(db, router, logger, storage, eventService)
        },`,
			indentation,
			mapKeyPackageName,
			importPackageName,
			structName)
	} else {
		newInitializer = fmt.Sprintf(`%s"%s": func(db *gorm.DB, router *gin.RouterGroup, emitter *emitter.Emitter, _ *storage.ActiveStorage, logger *zap.Logger, eventService *event.EventService) module.Module {
            return %s.New%sModule(db, router, logger, eventService)
        },`,
			indentation,
			mapKeyPackageName,
			importPackageName,
			structName)
	}

	updatedContent := contentStr[:lineStart] + newInitializer + "\n" + contentStr[lineStart:]
	return []byte(updatedContent), true
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

func UpdateInitFileForDestroy(pluralName string) error {
	initFilePath := "app/init.go"

	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	importStr := fmt.Sprintf("\"base/app/%s\"", pluralName)
	content = RemoveImport(content, importStr)

	content = RemoveModuleInitializer(content, pluralName)

	return os.WriteFile(initFilePath, content, 0644)
}

func RemoveImport(content []byte, importStr string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	var newLines [][]byte

	for _, line := range lines {
		if !bytes.Contains(line, []byte(importStr)) {
			newLines = append(newLines, line)
		}
	}

	return bytes.Join(newLines, []byte("\n"))
}

func RemoveModuleInitializer(content []byte, pluralName string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	var newLines [][]byte

	for _, line := range lines {
		if !bytes.Contains(line, []byte(fmt.Sprintf(`"%s":`, pluralName))) {
			newLines = append(newLines, line)
		}
	}

	return bytes.Join(newLines, []byte("\n"))
}
