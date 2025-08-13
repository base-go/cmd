package utils

import (
	"strings"
	"unicode"
)

// NamingConvention holds all naming variations derived from a single model name
type NamingConvention struct {
	// Original input (e.g., "ProductCategory")
	Original string

	// Model naming
	Model      string // ProductCategory (PascalCase)
	ModelLower string // productCategory (camelCase)
	ModelSnake string // product_category (snake_case)
	ModelKebab string // product-category (kebab-case)

	// Plural forms
	Plural      string // ProductCategories (PascalCase)
	PluralLower string // productCategories (camelCase)
	PluralSnake string // product_categories (snake_case)
	PluralKebab string // product-categories (kebab-case)

	// Package and directory naming
	PackageName string // product_categories (snake_case plural for package)
	DirName     string // product_categories (snake_case plural for directory)

	// Route naming
	RoutePath  string // /product-categories (kebab-case plural)
	RouteGroup string // product-categories (kebab-case plural)

	// Controller and Service naming
	Controller string // ProductCategoriesController
	Service    string // ProductCategoriesService

	// Database naming
	TableName string // product_categories (snake_case plural)

	// Variable naming
	VarSingle string // productCategory (camelCase)
	VarPlural string // productCategories (camelCase)
	VarId     string // productCategoryId (camelCase + Id)
}

// NewNamingConvention creates all naming variations from a single model name
func NewNamingConvention(modelName string) *NamingConvention {
	// Ensure PascalCase for the model name
	model := ToPascalCase(modelName)
	plural := PluralizeClient.Plural(model)

	nc := &NamingConvention{
		Original: modelName,

		// Model naming
		Model:      model,
		ModelLower: ToCamelCase(model),
		ModelSnake: ToSnakeCase(model),
		ModelKebab: ToKebabCase(model),

		// Plural forms
		Plural:      plural,
		PluralLower: ToCamelCase(plural),
		PluralSnake: ToSnakeCase(plural),
		PluralKebab: ToKebabCase(plural),

		// Package and directory naming
		PackageName: ToSnakeCase(plural),
		DirName:     ToSnakeCase(plural),

		// Route naming (kebab-case is URL-friendly)
		RoutePath:  "/" + ToKebabCase(plural),
		RouteGroup: ToKebabCase(plural),

		// Controller and Service naming
		Controller: plural + "Controller",
		Service:    plural + "Service",

		// Database naming
		TableName: ToSnakeCase(plural),

		// Variable naming
		VarSingle: ToCamelCase(model),
		VarPlural: ToCamelCase(plural),
		VarId:     ToCamelCase(model) + "Id",
	}

	return nc
}

// ToCapitalCase converts snake_case or kebab-case to Capital Case
func ToCapitalCase(s string) string {
	// Replace underscores and hyphens with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split by spaces and capitalize each word
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// SplitCamelCase splits a CamelCase or PascalCase string into words
func SplitCamelCase(s string) []string {
	var words []string
	var currentWord []rune

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Check if next character is lowercase (indicates new word)
			if i+1 < len(s) {
				nextRune := rune(s[i+1])
				if unicode.IsLower(nextRune) || (len(currentWord) > 1 && unicode.IsUpper(currentWord[len(currentWord)-1])) {
					if len(currentWord) > 0 {
						words = append(words, string(currentWord))
						currentWord = []rune{}
					}
				}
			} else if len(currentWord) > 0 {
				// Last character
				words = append(words, string(currentWord))
				currentWord = []rune{}
			}
		}
		currentWord = append(currentWord, r)
	}

	if len(currentWord) > 0 {
		words = append(words, string(currentWord))
	}

	return words
}
