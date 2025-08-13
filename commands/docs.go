package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate swagger documentation",
	Long:  `Generate swagger documentation by scanning controller annotations.`,
	Run:   generateDocs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

type SwaggerAnnotation struct {
	Summary     string
	Description string
	Tags        string
	Method      string
	Route       string
	Security    []string
	Parameters  []ParamAnnotation
	Responses   []ResponseAnnotation
}

type ParamAnnotation struct {
	Name        string
	In          string // query, path, header, body
	Description string
	Required    bool
	Type        string
	Example     string
}

type ResponseAnnotation struct {
	Code        string
	Description string
	Schema      string
	Example     string
}

func generateDocs(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return
	}

	fmt.Println("🔍 Scanning for swagger annotations...")

	// Find all controller files
	controllerFiles, err := findControllerFiles(cwd)
	if err != nil {
		fmt.Printf("Error finding controller files: %v\n", err)
		return
	}

	fmt.Printf("Found %d controller files\n", len(controllerFiles))

	// Parse annotations from each file
	var allAnnotations []SwaggerAnnotation
	for _, file := range controllerFiles {
		annotations, err := parseSwaggerAnnotations(file)
		if err != nil {
			fmt.Printf("Warning: Error parsing %s: %v\n", file, err)
			continue
		}
		allAnnotations = append(allAnnotations, annotations...)
	}

	fmt.Printf("📝 Found %d swagger-documented endpoints\n", len(allAnnotations))

	// Generate registration code
	err = generateSwaggerRegistration(cwd, allAnnotations)
	if err != nil {
		fmt.Printf("Error generating swagger registration: %v\n", err)
		return
	}

	fmt.Println("✅ Swagger documentation generated successfully!")
	fmt.Println("   Generated: base/core/swagger/auto_generated.go")
}

func findControllerFiles(rootDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "controller.go") || strings.Contains(path, "/controllers/") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func parseSwaggerAnnotations(filename string) ([]SwaggerAnnotation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var annotations []SwaggerAnnotation
	var currentAnnotation SwaggerAnnotation
	var inAnnotationBlock bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "// @Summary") {
			currentAnnotation = SwaggerAnnotation{} // Reset
			currentAnnotation.Summary = strings.TrimPrefix(line, "// @Summary ")
			inAnnotationBlock = true
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Description") {
			currentAnnotation.Description = strings.TrimPrefix(line, "// @Description ")
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Tags") {
			currentAnnotation.Tags = strings.TrimPrefix(line, "// @Tags ")
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Router") {
			// Parse @Router /path [method]
			routerRegex := regexp.MustCompile(`// @Router\s+(.+?)\s+\[(.+?)\]`)
			matches := routerRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				currentAnnotation.Route = matches[1]
				currentAnnotation.Method = strings.ToUpper(matches[2])
			}
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Param") {
			// Parse @Param name in type required description
			param := parseParamAnnotation(line)
			if param != nil {
				currentAnnotation.Parameters = append(currentAnnotation.Parameters, *param)
			}
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Success") {
			// Parse @Success code description
			resp := parseResponseAnnotation(line, true)
			if resp != nil {
				currentAnnotation.Responses = append(currentAnnotation.Responses, *resp)
			}
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Failure") {
			// Parse @Failure code description
			resp := parseResponseAnnotation(line, false)
			if resp != nil {
				currentAnnotation.Responses = append(currentAnnotation.Responses, *resp)
			}
		} else if inAnnotationBlock && strings.HasPrefix(line, "// @Security") {
			// Parse @Security scheme
			security := strings.TrimPrefix(line, "// @Security ")
			currentAnnotation.Security = append(currentAnnotation.Security, security)
		} else if inAnnotationBlock && strings.HasPrefix(line, "func ") {
			// End of annotation block - we've reached the function
			if currentAnnotation.Summary != "" && currentAnnotation.Route != "" {
				annotations = append(annotations, currentAnnotation)
			}
			inAnnotationBlock = false
		} else if !strings.HasPrefix(line, "//") && line != "" {
			// Non-comment line - end annotation block
			inAnnotationBlock = false
		}
	}

	return annotations, scanner.Err()
}

// parseParamAnnotation parses @Param annotations
// Format: @Param name in type required "description" example(optional)
func parseParamAnnotation(line string) *ParamAnnotation {
	// Remove @Param prefix
	paramStr := strings.TrimPrefix(line, "// @Param ")
	
	// Split by spaces but preserve quoted strings
	parts := parseQuotedString(paramStr)
	if len(parts) < 4 {
		return nil
	}
	
	param := &ParamAnnotation{
		Name: parts[0],
		In:   parts[1],
		Type: parts[2],
	}
	
	// Parse required field
	if strings.ToLower(parts[3]) == "true" {
		param.Required = true
	}
	
	// Parse description (quoted)
	if len(parts) > 4 {
		param.Description = strings.Trim(parts[4], "\"")
	}
	
	// Parse example if present
	if len(parts) > 5 {
		param.Example = strings.Trim(parts[5], "\"")
	}
	
	return param
}

// parseResponseAnnotation parses @Success and @Failure annotations
// Format: @Success code "description" schema example(optional)
func parseResponseAnnotation(line string, isSuccess bool) *ResponseAnnotation {
	var prefix string
	if isSuccess {
		prefix = "// @Success "
	} else {
		prefix = "// @Failure "
	}
	
	respStr := strings.TrimPrefix(line, prefix)
	parts := parseQuotedString(respStr)
	
	if len(parts) < 2 {
		return nil
	}
	
	resp := &ResponseAnnotation{
		Code: parts[0],
	}
	
	// Parse description (quoted)
	if len(parts) > 1 {
		resp.Description = strings.Trim(parts[1], "\"")
	}
	
	// Parse schema if present
	if len(parts) > 2 {
		resp.Schema = parts[2]
	}
	
	// Parse example if present
	if len(parts) > 3 {
		resp.Example = strings.Trim(parts[3], "\"")
	}
	
	return resp
}

// parseQuotedString splits a string by spaces while preserving quoted strings
func parseQuotedString(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	
	for i, char := range s {
		if char == '"' {
			inQuotes = !inQuotes
			current.WriteRune(char)
		} else if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(char)
		}
		
		// Add the last part
		if i == len(s)-1 && current.Len() > 0 {
			parts = append(parts, current.String())
		}
	}
	
	return parts
}

func generateSwaggerRegistration(rootDir string, annotations []SwaggerAnnotation) error {
	outputFile := filepath.Join(rootDir, "core", "swagger", "auto_generated.go")

	content := `package swagger

// Auto-generated swagger registration code
// Generated by: base docs command

func init() {
	// This function will be called when the swagger package is imported
	// It auto-registers all the documented routes
}

// RegisterAllRoutes registers all auto-discovered routes
func RegisterAllRoutes() {
`

	for _, ann := range annotations {
		// Build parameters
		paramsStr := "nil"
		if len(ann.Parameters) > 0 {
			paramsStr = "[]Parameter{"
			for _, param := range ann.Parameters {
				paramsStr += fmt.Sprintf(`
		{Name: "%s", In: "%s", Description: "%s", Required: %t, Schema: Schema{Type: "%s"}},`,
					param.Name, param.In, param.Description, param.Required, param.Type)
			}
			paramsStr += "\n\t}"
		}
		
		// Build responses
		responsesStr := "map[string]Response{\"200\": {Description: \"Success\"}}"
		if len(ann.Responses) > 0 {
			responsesStr = "map[string]Response{"
			for _, resp := range ann.Responses {
				responsesStr += fmt.Sprintf(`
		"%s": {Description: "%s"},`, resp.Code, resp.Description)
			}
			responsesStr += "\n\t}"
		}
		
		// Build security
		securityStr := "nil"
		if len(ann.Security) > 0 {
			securityStr = "[]map[string][]string{"
			for _, sec := range ann.Security {
				securityStr += fmt.Sprintf(`{"%s": {}}, `, sec)
			}
			securityStr += "}"
		}
		
		// Add /api prefix to route if not already present
		route := ann.Route
		if !strings.HasPrefix(route, "/api") {
			route = "/api" + route
		}
		
		content += fmt.Sprintf(`	AutoRegisterRouteDetailed("%s", "%s", "%s", "%s", []string{"%s"},
		%s, %s, %s)
`,
			ann.Method, route, ann.Summary, ann.Description, ann.Tags,
			paramsStr, responsesStr, securityStr)
	}

	content += `}
`

	err := os.WriteFile(outputFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write auto_generated.go: %w", err)
	}

	return nil
}