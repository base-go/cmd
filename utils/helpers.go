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

// GetGoType maps custom type strings to Go types.
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
		return t // Return the type as-is for custom types
	}
}

// ToLower converts a string to lowercase.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToTitle converts a string to title case.
func ToTitle(s string) string {
	return cases.Title(language.Und).String(s)
}

// ToLowerPlural converts a string to its plural form in lowercase.
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
			// Start of a new word
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

// GetInputType maps Go types to HTML input types
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

// UpdateInitFile updates the app/init.go file to register a new module.
func UpdateInitFile(singularName, pluralName string) error {
	initFilePath := "app/init.go"

	// Read the current content of init.go
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	// Use the correct package name (lowercase)
	packageName := ToSnakeCase(singularName)

	// Add import for the new module if it doesn't exist
	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
	content, importAdded := AddImport(content, importStr)

	// Add module initializer if it doesn't exist
	content, initializerAdded := AddModuleInitializer(content, packageName, singularName)

	// Write the updated content back to init.go only if changes were made
	if importAdded || initializerAdded {
		return os.WriteFile(initFilePath, content, 0644)
	}

	return nil
}

// AddImport adds an import statement to the file content if it doesn't already exist.
func AddImport(content []byte, importStr string) ([]byte, bool) {
	// Check if the import already exists
	if bytes.Contains(content, []byte(importStr)) {
		return content, false
	}

	// Find the position of "import ("
	importPos := bytes.Index(content, []byte("import ("))
	if importPos == -1 {
		// If "import (" is not found, return original content
		return content, false
	}

	// Position to insert the new import (after "import (" and newline)
	insertPos := importPos + len("import (") + 1

	// Create the new import line with proper indentation
	newImportLine := []byte("\t" + importStr + "\n")

	// Insert the new import line
	updatedContent := append(content[:insertPos], append(newImportLine, content[insertPos:]...)...)

	return updatedContent, true
}

// AddModuleInitializer adds a module initializer to the app/init.go content.
func AddModuleInitializer(content []byte, packageName, singularName string) ([]byte, bool) {
	contentStr := string(content)

	// Find the module initializer marker
	markerIndex := strings.Index(contentStr, "// MODULE_INITIALIZER_MARKER")
	if markerIndex == -1 {
		return content, false
	}

	// Check if the module already exists
	if strings.Contains(contentStr[:markerIndex], fmt.Sprintf(`"%s":`, packageName)) {
		return content, false
	}

	structName := ToPascalCase(singularName)

	// Create the new initializer
	newInitializer := fmt.Sprintf(`        "%s": func(db *gorm.DB, router *gin.RouterGroup) module.Module { return %s.New%sModule(db, router) },`,
		packageName, packageName, structName)

	// Insert the new initializer before the marker
	updatedContent := contentStr[:markerIndex] + newInitializer + "\n        " + contentStr[markerIndex:]

	return []byte(updatedContent), true
}

// UpdateNavFile updates the admin/partials/nav.html file to include the new module.
func UpdateNavFile(pluralName string) {
	navFilePath := "admin/partials/nav.html"
	content, err := os.ReadFile(navFilePath)
	if err != nil {
		fmt.Printf("Error reading nav file: %v\n", err)
		return
	}

	// Find the position to insert the new menu item
	insertPos := bytes.Index(content, []byte(`<li class="auth-only"><a href="#" data-page="dashboard">Dashboard</a></li>`))
	if insertPos == -1 {
		fmt.Println("Could not find the correct position to insert the new menu item")
		return
	}

	// Move to the end of the line
	insertPos = bytes.IndexByte(content[insertPos:], '\n') + insertPos + 1

	// Create the new menu item
	newMenuItem := fmt.Sprintf(`		<li class="auth-only"><a href="#" data-page="%s">%s</a></li>`, pluralName, ToTitle(pluralName))

	// Insert the new menu item
	updatedContent := append(content[:insertPos], append([]byte(newMenuItem+"\n"), content[insertPos:]...)...)

	// Write the updated content back to the file
	if err := os.WriteFile(navFilePath, updatedContent, 0644); err != nil {
		fmt.Printf("Error writing updated nav file: %v\n", err)
	}
}

// UpdateIndexFile updates the admin/index.html file to include the new module.
func UpdateIndexFile(pluralName string) {
	indexFilePath := "admin/index.html"
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		fmt.Printf("Error reading index file: %v\n", err)
		return
	}

	// Find the position to insert the new case
	markerComment := []byte("//LoadGeneratedPage")
	insertPos := bytes.Index(content, markerComment)
	if insertPos == -1 {
		fmt.Println("Could not find the marker comment to insert the new case")
		return
	}

	// Create the new case
	newCase := fmt.Sprintf(`
						case '%s':
						$('#main-content').load('/admin/%s/index.html');
						break;
						`, pluralName, pluralName)

	// Insert the new case just before the marker comment
	updatedContent := append(content[:insertPos], append([]byte(newCase), content[insertPos:]...)...)

	// Write the updated content back to the file
	if err := os.WriteFile(indexFilePath, updatedContent, 0644); err != nil {
		fmt.Printf("Error writing updated index file: %v\n", err)
	}
}

// UpdateInitFileForDestroy updates the app/init.go file to unregister a module.
func UpdateInitFileForDestroy(pluralName string) error {
	initFilePath := "app/init.go"

	// Read the current content of init.go
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	// Remove import for the module
	importStr := fmt.Sprintf("\"base/app/%s\"", pluralName)
	content = RemoveImport(content, importStr)

	// Remove module initializer
	content = RemoveModuleInitializer(content, pluralName)

	// Write the updated content back to init.go
	return os.WriteFile(initFilePath, content, 0644)
}

// RemoveImport removes an import statement from the file content.
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

// RemoveModuleInitializer removes a module initializer from the app/init.go content.
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

// UpdateSeedersFile updates the app/init.go file to register a new seeder.
func UpdateSeedersFile(singularName, packageName string) error {
	initFilePath := "app/init.go"

	// Read the current content of init.go
	content, err := os.ReadFile(initFilePath)
	if err != nil {
		return err
	}

	// Add import for the new seeder if it doesn't exist
	importStr := fmt.Sprintf("\"base/app/%s\"", packageName)
	content, importAdded := AddImport(content, importStr)

	// Add seeder initializer if it doesn't exist
	content, seederAdded := AddSeederInitializer(content, singularName, packageName)

	// Write the updated content back to init.go only if changes were made
	if importAdded || seederAdded {
		return os.WriteFile(initFilePath, content, 0644)
	}

	return nil
}

// AddSeederInitializer adds a seeder to the InitializeSeeders function.
func AddSeederInitializer(content []byte, structName, packageName string) ([]byte, bool) {

	// Find the InitializeSeeders function
	funcIndex := bytes.Index(content, []byte("func InitializeSeeders() []module.Seeder {"))
	if funcIndex == -1 {
		return content, false
	}

	// Find the position to insert the new seeder (before the closing bracket of the seeders slice)
	seedersIndex := bytes.Index(content[funcIndex:], []byte("}"))
	if seedersIndex == -1 {
		return content, false
	}
	insertPos := funcIndex + seedersIndex

	// Check if the seeder already exists
	seederLine := fmt.Sprintf("&%s.%sSeeder{},", packageName, structName)
	if bytes.Contains(content, []byte(seederLine)) {
		return content, false
	}

	// Insert the new seeder
	newSeeder := fmt.Sprintf("\t\t%s\n", seederLine)
	updatedContent := append(content[:insertPos], append([]byte(newSeeder), content[insertPos:]...)...)

	return updatedContent, true
}

// RemoveSeederInitializer removes a seeder from the InitializeSeeders function.
func RemoveSeederInitializer(content []byte, structName, packageName string) []byte {
	// Find the InitializeSeeders function
	funcIndex := bytes.Index(content, []byte("func InitializeSeeders() []module.Seeder {"))
	if funcIndex == -1 {
		return content
	}

	// Find the position of the seeder to remove
	seederLine := fmt.Sprintf("&%s.%sSeeder{},", packageName, structName)
	seederIndex := bytes.Index(content[funcIndex:], []byte(seederLine))
	if seederIndex == -1 {
		return content
	}

	// Find the start of the line
	startIndex := funcIndex + seederIndex
	for startIndex > 0 && content[startIndex] != '\n' {
		startIndex--
	}

	// Find the end of the line
	endIndex := funcIndex + seederIndex
	for endIndex < len(content) && content[endIndex] != '\n' {
		endIndex++
	}

	// Remove the seeder line
	updatedContent := append(content[:startIndex], content[endIndex+1:]...)

	return updatedContent
}
