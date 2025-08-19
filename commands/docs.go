package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	outputDir    string
	generateStatic bool
	noStatic     bool
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate OpenAPI 3.0 documentation",
	Long:  `Generate OpenAPI 3.0 documentation by scanning controller annotations and create static files (JSON, YAML, docs.go).`,
	Run:   generateDocs,
}

func init() {
	docsCmd.Flags().StringVarP(&outputDir, "output", "o", "docs", "Output directory for generated files")
	docsCmd.Flags().BoolVarP(&generateStatic, "static", "s", true, "Generate static swagger files (JSON, YAML, docs.go)")
	docsCmd.Flags().BoolVar(&noStatic, "no-static", false, "Skip generating static files, only generate auto_generated.go")
	rootCmd.AddCommand(docsCmd)
}

// SwaggerMainInfo holds the parsed swagger info from main.go
type SwaggerMainInfo struct {
	Title       string
	Version     string
	Description string
	BasePath    string
	Host        string
	Schemes     []string
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

	// If we're in cmd directory, look for base directory
	baseDir := cwd
	if strings.HasSuffix(cwd, "/cmd") {
		baseDir = filepath.Join(filepath.Dir(cwd), "base")
	} else if _, err := os.Stat("../base"); err == nil {
		baseDir = "../base"
	}

	fmt.Printf("🔍 Scanning for swagger annotations in %s...\n", baseDir)

	// Find all controller files
	controllerFiles, err := findControllerFiles(baseDir)
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

	fmt.Printf("📝 Found %d documented API endpoints\n", len(allAnnotations))

	fmt.Println("✅ OpenAPI 3.0 documentation scanned successfully!")
	fmt.Printf("📋 Found %d endpoints across %d controller files\n", len(allAnnotations), len(controllerFiles))
	
	// Generate static files by default, unless --no-static is specified
	if generateStatic && !noStatic {
		fmt.Printf("📄 Generating static files to %s...\n", outputDir)
		err = generateStaticSwaggerFiles(outputDir, allAnnotations)
		if err != nil {
			fmt.Printf("Error generating static files: %v\n", err)
			return
		}
		fmt.Printf("✅ Static files generated in %s/\n", outputDir)
		fmt.Println("   - swagger.json")
		fmt.Println("   - swagger.yaml") 
		fmt.Println("   - docs.go")
	}
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

// generateStaticSwaggerFiles creates static swagger documentation files
func generateStaticSwaggerFiles(outputDir string, annotations []SwaggerAnnotation) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate OpenAPI 3.0 spec
	swaggerDoc := generateOpenAPISpec(annotations)
	
	// Generate swagger.json
	jsonData, err := json.MarshalIndent(swaggerDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal swagger spec: %w", err)
	}
	
	jsonPath := filepath.Join(outputDir, "swagger.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write swagger.json: %w", err)
	}
	
	// Generate swagger.yaml
	yamlContent := generateSwaggerYAML(swaggerDoc)
	yamlPath := filepath.Join(outputDir, "swagger.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write swagger.yaml: %w", err)
	}
	
	// Generate docs.go
	docsGoContent := generateSwaggoDocsGo(swaggerDoc)
	docsGoPath := filepath.Join(outputDir, "docs.go")
	if err := os.WriteFile(docsGoPath, []byte(docsGoContent), 0644); err != nil {
		return fmt.Errorf("failed to write docs.go: %w", err)
	}
	
	return nil
}

// generateOpenAPISpec creates an OpenAPI 3.0 specification from annotations
func generateOpenAPISpec(annotations []SwaggerAnnotation) map[string]any {
	// Extract swagger info from main.go
	swaggerInfo, err := extractSwaggerInfoFromMainGo()
	if err != nil {
		// Use defaults if extraction fails
		swaggerInfo = SwaggerMainInfo{
			Title:       "Base Framework API",
			Version:     "1.0.0",
			Description: "API documentation generated by Base CLI",
			BasePath:    "/api",
			Host:        "localhost:8080",
			Schemes:     []string{"http", "https"},
		}
	}
	
	// Build paths from annotations
	paths := make(map[string]any)
	
	for _, ann := range annotations {
		route := ann.Route
		if !strings.HasPrefix(route, "/api") {
			route = "/api" + route
		}
		
		if paths[route] == nil {
			paths[route] = make(map[string]any)
		}
		
		// Build operation
		operation := map[string]any{
			"summary":     ann.Summary,
			"description": ann.Description,
			"tags":        []string{ann.Tags},
			"operationId": generateOperationID(ann.Method, route),
		}
		
		// Add parameters
		if len(ann.Parameters) > 0 {
			var parameters []any
			for _, param := range ann.Parameters {
				parameters = append(parameters, map[string]any{
					"name":        param.Name,
					"in":          param.In,
					"description": param.Description,
					"required":    param.Required,
					"schema": map[string]any{
						"type": param.Type,
					},
				})
			}
			operation["parameters"] = parameters
		}
		
		// Add responses
		responses := make(map[string]any)
		if len(ann.Responses) > 0 {
			for _, resp := range ann.Responses {
				responses[resp.Code] = map[string]any{
					"description": resp.Description,
				}
			}
		} else {
			responses["200"] = map[string]any{
				"description": "Success",
			}
		}
		operation["responses"] = responses
		
		// Add security
		if len(ann.Security) > 0 {
			var security []any
			for _, sec := range ann.Security {
				security = append(security, map[string]any{
					sec: []any{},
				})
			}
			operation["security"] = security
		}
		
		paths[route].(map[string]any)[strings.ToLower(ann.Method)] = operation
	}
	
	// Build servers from extracted info
	servers := []any{}
	for _, scheme := range swaggerInfo.Schemes {
		servers = append(servers, map[string]any{
			"url":         fmt.Sprintf("%s://%s", scheme, swaggerInfo.Host),
			"description": fmt.Sprintf("Base Framework API Server (%s)", scheme),
		})
	}
	
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       swaggerInfo.Title,
			"description": swaggerInfo.Description,
			"version":     swaggerInfo.Version,
			"contact": map[string]any{
				"name":  "Base Team",
				"url":   "https://github.com/BaseTechStack",
				"email": "info@base.al",
			},
			"license": map[string]any{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": servers,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"ApiKeyAuth": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "X-Api-Key",
				},
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "Enter the token with the `Bearer: ` prefix, e.g. \"Bearer abcde12345\"",
				},
			},
			"schemas": map[string]any{
				"Error": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"error": map[string]any{
							"type":        "string",
							"description": "Error message",
						},
					},
					"required": []string{"error"},
				},
				"Success": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"message": map[string]any{
							"type":        "string",
							"description": "Success message",
						},
					},
					"required": []string{"message"},
				},
			},
		},
		"paths": paths,
	}
}

// generateOperationID creates a unique operation ID from method and route
func generateOperationID(method, route string) string {
	// Convert /api/users/{id} -> getUsersById
	parts := strings.Split(strings.Trim(route, "/"), "/")
	var opID strings.Builder
	
	opID.WriteString(strings.ToLower(method))
	
	for _, part := range parts {
		if part == "api" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// Parameter like {id} -> ById
			param := strings.Trim(part, "{}")
			opID.WriteString("By")
			opID.WriteString(strings.Title(param))
		} else {
			opID.WriteString(strings.Title(part))
		}
	}
	
	return opID.String()
}

// generateSwaggerYAML creates a YAML version of the OpenAPI spec
func generateSwaggerYAML(doc map[string]any) string {
	yamlBuilder := strings.Builder{}
	yamlBuilder.WriteString("openapi: \"3.0.3\"\n")
	
	if info, ok := doc["info"].(map[string]any); ok {
		yamlBuilder.WriteString("info:\n")
		if title, ok := info["title"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  title: \"%s\"\n", title))
		}
		if description, ok := info["description"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  description: \"%s\"\n", description))
		}
		if version, ok := info["version"].(string); ok {
			yamlBuilder.WriteString(fmt.Sprintf("  version: \"%s\"\n", version))
		}
	}
	
	if servers, ok := doc["servers"].([]any); ok && len(servers) > 0 {
		yamlBuilder.WriteString("servers:\n")
		for _, server := range servers {
			if serverMap, ok := server.(map[string]any); ok {
				if url, ok := serverMap["url"].(string); ok {
					yamlBuilder.WriteString(fmt.Sprintf("  - url: \"%s\"\n", url))
				}
				if desc, ok := serverMap["description"].(string); ok {
					yamlBuilder.WriteString(fmt.Sprintf("    description: \"%s\"\n", desc))
				}
			}
		}
	}
	
	yamlBuilder.WriteString("# Full OpenAPI spec available in swagger.json\n")
	return yamlBuilder.String()
}

// generateSwaggoDocsGo creates a docs.go file with OpenAPI 3.0 spec
func generateSwaggoDocsGo(doc map[string]any) string {
	jsonBytes, _ := json.Marshal(doc)
	jsonStr := string(jsonBytes)
	
	// Extract swagger info from main.go annotations
	swaggerInfo, err := extractSwaggerInfoFromMainGo()
	if err != nil {
		// Fallback to doc info if extraction fails
		info := doc["info"].(map[string]any)
		swaggerInfo = SwaggerMainInfo{
			Title:       info["title"].(string),
			Version:     info["version"].(string),
			Description: info["description"].(string),
			BasePath:    "/api",
			Host:        "localhost:8080",
			Schemes:     []string{"http", "https"},
		}
	}
	
	// Generate a pure OpenAPI 3.0 docs.go file without swaggo dependency
	return fmt.Sprintf(`// Package docs GENERATED BY BASE CLI DOCS COMMAND. DO NOT EDIT
// This file contains OpenAPI 3.0 specification
package docs

// OpenAPISpec contains the OpenAPI 3.0 specification as a JSON string
const OpenAPISpec = %q

// SwaggerInfo holds the API documentation info (for compatibility)
type SwaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// Info provides the API documentation details
var Info = SwaggerInfo{
	Version:     %q,
	Host:        %q,
	BasePath:    %q,
	Schemes:     %s,
	Title:       %q,
	Description: %q,
}

// GetOpenAPISpec returns the OpenAPI 3.0 specification as a string
func GetOpenAPISpec() string {
	return OpenAPISpec
}
`,
		jsonStr,
		swaggerInfo.Version,
		swaggerInfo.Host,
		swaggerInfo.BasePath,
		fmt.Sprintf("%#v", swaggerInfo.Schemes),
		swaggerInfo.Title,
		swaggerInfo.Description,
	)
}

// extractSwaggerInfoFromMainGo parses swagger annotations from main.go
func extractSwaggerInfoFromMainGo() (SwaggerMainInfo, error) {
	info := SwaggerMainInfo{
		Host:     "localhost:8080", // default
		BasePath: "/api",           // default
		Schemes:  []string{"http", "https"}, // default
	}
	
	// Look for main.go in the base directory
	mainGoPath := "../base/main.go"
	if _, err := os.Stat("../main.go"); err == nil {
		mainGoPath = "../main.go"
	}
	
	file, err := os.Open(mainGoPath)
	if err != nil {
		return info, fmt.Errorf("failed to open main.go: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Stop parsing main API annotations when we hit security definitions
		// to avoid overwriting main description with security description
		if strings.HasPrefix(line, "// @securityDefinitions") {
			break
		}
		
		// Parse swagger annotations
		if strings.HasPrefix(line, "// @title ") {
			info.Title = strings.TrimPrefix(line, "// @title ")
		} else if strings.HasPrefix(line, "// @version ") {
			info.Version = strings.TrimPrefix(line, "// @version ")
		} else if strings.HasPrefix(line, "// @description ") {
			info.Description = strings.TrimPrefix(line, "// @description ")
		} else if strings.HasPrefix(line, "// @BasePath ") {
			info.BasePath = strings.TrimPrefix(line, "// @BasePath ")
		} else if strings.HasPrefix(line, "// @host ") {
			// Remove http:// or https:// prefix from host if present
			host := strings.TrimPrefix(line, "// @host ")
			host = strings.TrimPrefix(host, "http://")
			host = strings.TrimPrefix(host, "https://")
			info.Host = host
		} else if strings.HasPrefix(line, "// @schemes ") {
			schemesStr := strings.TrimPrefix(line, "// @schemes ")
			info.Schemes = strings.Fields(schemesStr)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return info, fmt.Errorf("failed to read main.go: %w", err)
	}
	
	// Ensure we have minimum required info
	if info.Title == "" {
		info.Title = "Base Framework API"
	}
	if info.Version == "" {
		info.Version = "1.0.0"
	}
	if info.Description == "" {
		info.Description = "API documentation generated by Base CLI"
	}
	
	return info, nil
}