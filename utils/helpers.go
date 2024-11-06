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

	newInitializer := fmt.Sprintf(`	"%s": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return %s.New%sModule(db, router) },`,
		packageName, packageName, structName)

	updatedContent := contentStr[:markerIndex] + newInitializer + "\n        " + contentStr[markerIndex:]

	return []byte(updatedContent), true
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
