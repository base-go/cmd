package tasks

import (
	"base/app/models"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate *validator.Validate

// init initializes the validator
func init() {
	validate = validator.New()
	// Register custom validators here if needed
	// Example: validate.RegisterValidation("custom_tag", customValidationFunc)
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// ValidationErrors stores multiple validation errors
type ValidationErrors struct {
	Errors []string
}

// Error implements the error interface for ValidationErrors
func (ve *ValidationErrors) Error() string {
	return strings.Join(ve.Errors, "; ")
}

// AddError adds a validation error to the collection
func (ve *ValidationErrors) AddError(format string, args ...interface{}) {
	ve.Errors = append(ve.Errors, fmt.Sprintf(format, args...))
}

// HasErrors checks if there are any validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// CollectStructValidationErrors collects all validation errors from struct validation
func CollectStructValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			// Format each validation error
			errors = append(errors, fmt.Sprintf(
				"Field: %s, Error: %s, Value: %v",
				e.Field(),
				e.Tag(),
				e.Value(),
			))
		}
	} else {
		errors = append(errors, err.Error())
	}
	return errors
}

// ValidateTaskCreateRequest validates the create request
func ValidateTaskCreateRequest(req *models.CreateTaskRequest) error {
	// Initialize validation errors collection
	validationErrors := &ValidationErrors{}

	if req == nil {
		validationErrors.AddError("request cannot be nil")
		return validationErrors
	}

	// Collect all struct tag validation errors
	if err := validate.Struct(req); err != nil {
		for _, errMsg := range CollectStructValidationErrors(err) {
			validationErrors.AddError("%s", errMsg)
		}
	}

	// Field-specific validations
	// String validation for Title
	// if req.Title != "" && (len(req.Title) < MinLength || len(req.Title) > MaxLength) {
	// 	validationErrors.AddError("Title must be between %d and %d characters", MinLength, MaxLength)
	// }
	// String validation for Description
	// if req.Description != "" && (len(req.Description) < MinLength || len(req.Description) > MaxLength) {
	// 	validationErrors.AddError("Description must be between %d and %d characters", MinLength, MaxLength)
	// }
	// String validation for Status
	// if req.Status != "" && (len(req.Status) < MinLength || len(req.Status) > MaxLength) {
	// 	validationErrors.AddError("Status must be between %d and %d characters", MinLength, MaxLength)
	// }

	// Return all validation errors if any
	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// ValidateTaskUpdateRequest validates the update request
func ValidateTaskUpdateRequest(req *models.UpdateTaskRequest, id uint) error {
	// Initialize validation errors collection
	validationErrors := &ValidationErrors{}

	if req == nil {
		validationErrors.AddError("request cannot be nil")
	}

	if id == 0 {
		validationErrors.AddError("invalid id: cannot be zero")
	}

	// Collect all struct tag validation errors
	if req != nil && id != 0 {
		if err := validate.Struct(req); err != nil {
			for _, errMsg := range CollectStructValidationErrors(err) {
				validationErrors.AddError("%s", errMsg)
			}
		}
	}

	// Field-specific validations for update
	// String validation - only if field is provided
	// if req.Title != "" && (len(req.Title) < MinLength || len(req.Title) > MaxLength) {
	// 	validationErrors.AddError("Title must be between %d and %d characters", MinLength, MaxLength)
	// }
	// String validation - only if field is provided
	// if req.Description != "" && (len(req.Description) < MinLength || len(req.Description) > MaxLength) {
	// 	validationErrors.AddError("Description must be between %d and %d characters", MinLength, MaxLength)
	// }
	// String validation - only if field is provided
	// if req.Status != "" && (len(req.Status) < MinLength || len(req.Status) > MaxLength) {
	// 	validationErrors.AddError("Status must be between %d and %d characters", MinLength, MaxLength)
	// }

	// Return all validation errors if any
	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// ValidateTaskDeleteRequest validates the delete request
func ValidateTaskDeleteRequest(id uint) error {
	// Initialize validation errors collection
	validationErrors := &ValidationErrors{}

	if id == 0 {
		validationErrors.AddError("invalid id: cannot be zero")
	}

	// Return all validation errors if any
	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// ValidateID validates if the ID is valid
func ValidateID(id uint) error {
	// Initialize validation errors collection
	validationErrors := &ValidationErrors{}

	if id == 0 {
		validationErrors.AddError("invalid id: cannot be zero")
	}

	// Return all validation errors if any
	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
