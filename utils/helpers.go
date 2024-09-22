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

func UpdateInitFile(singularName, pluralName string) error {
	initFilePath := "app/init.go"

	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	packageName := ToSnakeCase(singularName)

	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
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

	newInitializer := fmt.Sprintf(`        "%s": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return %s.New%sModule(db, router) },`,
		packageName, packageName, structName)

	updatedContent := contentStr[:markerIndex] + newInitializer + "\n        " + contentStr[markerIndex:]

	return []byte(updatedContent), true
}

func UpdateNavFile(pluralName string) {
	navFilePath := "admin/partials/nav.html"
	content, err := os.ReadFile(navFilePath)
	if err != nil {
		fmt.Printf("Error reading nav file: %v\n", err)
		return
	}

	insertPos := bytes.Index(content, []byte(`<li class="auth-only"><a href="#" data-page="dashboard">Dashboard</a></li>`))
	if insertPos == -1 {
		fmt.Println("Could not find the correct position to insert the new menu item")
		return
	}

	insertPos = bytes.IndexByte(content[insertPos:], '\n') + insertPos + 1

	newMenuItem := fmt.Sprintf(`		<li class="auth-only"><a href="#" data-page="%s">%s</a></li>`, pluralName, ToTitle(pluralName))

	updatedContent := append(content[:insertPos], append([]byte(newMenuItem+"\n"), content[insertPos:]...)...)

	if err := os.WriteFile(navFilePath, updatedContent, 0644); err != nil {
		fmt.Printf("Error writing updated nav file: %v\n", err)
	}
}

func UpdateIndexFile(pluralName string) {
	indexFilePath := "admin/index.html"
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		fmt.Printf("Error reading index file: %v\n", err)
		return
	}

	markerComment := []byte("//LoadGeneratedPage")
	insertPos := bytes.Index(content, markerComment)
	if insertPos == -1 {
		fmt.Println("Could not find the marker comment to insert the new case")
		return
	}

	newCase := fmt.Sprintf(`
						case '%s':
						$('#main-content').load('/admin/%s/index.html');
						break;
						`, pluralName, pluralName)

	updatedContent := append(content[:insertPos], append([]byte(newCase), content[insertPos:]...)...)

	if err := os.WriteFile(indexFilePath, updatedContent, 0644); err != nil {
		fmt.Printf("Error writing updated index file: %v\n", err)
	}
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

func UpdateSeedersFile(structName, packageName string) error {
	seedFilePath := "app/seed.go"

	content, err := os.ReadFile(seedFilePath)
	if err != nil {
		return err
	}

	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
	content, importAdded := AddImport(content, importStr)

	content, seederAdded := AddSeederInitializer(content, structName, packageName)

	if importAdded || seederAdded {
		return os.WriteFile(seedFilePath, content, 0644)
	}

	return nil
}

func AddSeederInitializer(content []byte, structName, packageName string) ([]byte, bool) {
	funcIndex := bytes.Index(content, []byte("func InitializeSeeders() []module.Seeder {"))
	if funcIndex == -1 {
		return content, false
	}

	seedersIndex := bytes.LastIndex(content[funcIndex:], []byte("}"))
	if seedersIndex == -1 {
		return content, false
	}
	insertPos := funcIndex + seedersIndex

	seederLine := fmt.Sprintf("&%s.%sSeeder{},", packageName, structName)
	if bytes.Contains(content, []byte(seederLine)) {
		return content, false
	}

	newSeeder := fmt.Sprintf("\t\t%s\n", seederLine)
	updatedContent := append(content[:insertPos], append([]byte(newSeeder), content[insertPos:]...)...)

	return updatedContent, true
}

func RemoveSeederFromSeedFile(pluralName string) error {
	seedFilePath := "app/seed.go"

	content, err := os.ReadFile(seedFilePath)
	if err != nil {
		return err
	}

	content = RemoveImport(content, fmt.Sprintf("\"base/app/%s\"", pluralName))

	content = RemoveSeederInitializer(content, pluralName)

	return os.WriteFile(seedFilePath, content, 0644)
}

func RemoveSeederInitializer(content []byte, pluralName string) []byte {
	funcStart := bytes.Index(content, []byte("func InitializeSeeders() []module.Seeder {"))
	if funcStart == -1 {
		return content
	}

	funcEnd := bytes.Index(content[funcStart:], []byte("}"))
	if funcEnd == -1 {
		return content
	}
	funcEnd += funcStart

	functionContent := content[funcStart:funcEnd]

	lines := bytes.Split(functionContent, []byte("\n"))
	var newLines [][]byte
	for _, line := range lines {
		if !bytes.Contains(line, []byte(fmt.Sprintf("&%s.", pluralName))) {
			newLines = append(newLines, line)
		}
	}

	updatedFunction := bytes.Join(newLines, []byte("\n"))

	return append(content[:funcStart], append(updatedFunction, content[funcEnd:]...)...)
}
