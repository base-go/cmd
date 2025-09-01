package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gertd/go-pluralize"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var PluralizeClient *pluralize.Client

func init() {
	PluralizeClient = pluralize.NewClient()
}

func GetGoType(t string) string {
	// Handle relationship types with colon syntax
	if strings.Contains(t, ":") {
		parts := strings.Split(t, ":")
		if len(parts) >= 2 {
			relationType := parts[1]

			// Handle relationship types
			switch {
			case strings.Contains(relationType, "belongsTo") || strings.Contains(relationType, "belongs_to"):
				if len(parts) > 2 {
					return ToPascalCase(parts[2]) // Use specified model
				}
				return "uint" // Foreign key field

			case strings.Contains(relationType, "hasMany") || strings.Contains(relationType, "has_many"):
				if len(parts) > 2 {
					return "[]" + ToPascalCase(parts[2])
				}
				// Infer model from field name (plural to singular)
				return "[]" + ToPascalCase(Singularize(parts[0]))

			case strings.Contains(relationType, "hasOne") || strings.Contains(relationType, "has_one"):
				if len(parts) > 2 {
					return "*" + ToPascalCase(parts[2])
				}
				// Use field name as model name for hasOne
				return "*" + ToPascalCase(parts[0])

			case strings.Contains(relationType, "toMany") || strings.Contains(relationType, "to_many") ||
				strings.Contains(relationType, "manyToMany") || strings.Contains(relationType, "many_to_many"):
				if len(parts) > 2 {
					return "[]*" + ToPascalCase(parts[2])
				}
				return "[]*" + ToPascalCase(Singularize(parts[0]))
			}
		}
	}

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
	case "translatedField", "translation":
		return "translation.Field"

	// Convenience aliases
	case "text", "email", "url", "slug":
		return "string"
	case "datetime", "time", "date", "timestamp":
		return "time.Time"
	case "float", "decimal":
		return "float64"
	case "sort":
		return "int"
	case "image", "file", "attachment":
		return "*storage.Attachment"
	case "json", "jsonb":
		return "datatypes.JSON"

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

func TrimIdSuffix(s string) string {
	if strings.HasSuffix(s, "Id") && len(s) > 2 {
		return s[:len(s)-2]
	}
	return s
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
	// Look for the module initializer pattern: modules["pluralName"] = pluralName.Init(deps)
	moduleStr := fmt.Sprintf(`modules["%s"]`, pluralName)
	initializerStart := strings.Index(contentStr, moduleStr)
	if initializerStart != -1 {
		// Find the start of the line (including any comment above)
		lineStart := strings.LastIndex(contentStr[:initializerStart], "\n") + 1

		// Check if there's a comment line above this module
		if lineStart > 1 {
			prevLineStart := strings.LastIndex(contentStr[:lineStart-1], "\n") + 1
			prevLine := strings.TrimSpace(contentStr[prevLineStart:lineStart])
			if strings.HasPrefix(prevLine, "//") && strings.Contains(prevLine, "module") {
				lineStart = prevLineStart // Include the comment line
			}
		}

		// Find the end of the line
		initializerEnd := strings.Index(contentStr[initializerStart:], "\n") + initializerStart
		if initializerEnd != -1 {
			initializerEnd++ // Include the newline
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

func GetRequiredImports(fields []Field) map[string][]string {
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

// Unzip extracts a zip archive to a destination directory.
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Ensure that the file path is within the destination directory.
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory if it doesn't exist.
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create parent directories if necessary.
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Create the file.
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Open the file inside the zip archive.
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Copy the file content to the destination file.
		_, err = io.Copy(outFile, rc)

		// Close files.
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
