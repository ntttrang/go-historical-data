package validator

import (
	"fmt"
	"strings"

	goValidator "github.com/go-playground/validator/v10"
)

// Validator wraps the validator instance
type Validator struct {
	validate *goValidator.Validate
}

// New creates a new validator instance
func New() *Validator {
	v := goValidator.New()

	// Register custom validators here if needed
	// v.RegisterValidation("custom_tag", customValidationFunc)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(data interface{}) error {
	if err := v.validate.Struct(data); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable format
func (v *Validator) formatValidationErrors(err error) error {
	if validationErrors, ok := err.(goValidator.ValidationErrors); ok {
		var errors []string
		for _, e := range validationErrors {
			errors = append(errors, fmt.Sprintf("field '%s' failed validation on '%s' tag",
				v.formatFieldName(e.Field()), e.Tag()))
		}
		return &ValidationError{
			Errors: errors,
		}
	}
	return err
}

// formatFieldName converts field name to snake_case
func (v *Validator) formatFieldName(field string) string {
	var result strings.Builder
	for i, r := range field {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ValidationError represents validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

// GetErrors returns the list of validation errors
func (e *ValidationError) GetErrors() []string {
	return e.Errors
}
