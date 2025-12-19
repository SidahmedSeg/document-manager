package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/SidahmedSeg/document-manager/backend/pkg/errors"
)

// Validator wraps go-playground/validator with custom validation rules
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance with custom rules
func New() *Validator {
	v := validator.New()

	// Register custom validators
	_ = v.RegisterValidation("uuid", validateUUID)
	_ = v.RegisterValidation("file_type", validateFileType)
	_ = v.RegisterValidation("alpha_space", validateAlphaSpace)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return v.formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors converts validator errors to AppError
func (v *Validator) formatValidationErrors(err error) error {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Validationf("validation failed: %v", err)
	}

	appErr := errors.ErrValidation

	for _, fieldErr := range validationErrs {
		field := fieldErr.Field()
		tag := fieldErr.Tag()
		param := fieldErr.Param()

		message := formatFieldError(field, tag, param)
		appErr = appErr.WithField(field, message)
	}

	return appErr
}

// formatFieldError creates a user-friendly error message
func formatFieldError(field, tag, param string) string {
	field = camelToSnake(field)

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "file_type":
		return fmt.Sprintf("%s must be one of the following file types: %s", field, param)
	case "alpha_space":
		return fmt.Sprintf("%s can only contain letters and spaces", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alphanum":
		return fmt.Sprintf("%s can only contain letters and numbers", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// Custom validation functions

// validateUUID validates that a string is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty, use 'required' tag for non-empty
	}
	_, err := uuid.Parse(value)
	return err == nil
}

// validateFileType validates file extensions
func validateFileType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Get allowed types from param
	allowedTypes := strings.Split(fl.Param(), "|")

	// Get file extension
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}

	ext := strings.ToLower(parts[len(parts)-1])

	// Check if extension is allowed
	for _, allowedType := range allowedTypes {
		if ext == strings.ToLower(allowedType) {
			return true
		}
	}

	return false
}

// validateAlphaSpace validates that a string contains only letters and spaces
func validateAlphaSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	// Allow letters (any language) and spaces
	matched, _ := regexp.MatchString(`^[\p{L}\s]+$`, value)
	return matched
}

// Helper functions

// camelToSnake converts camelCase to snake_case
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ValidateUUID validates a UUID string
func ValidateUUID(uuidStr string) error {
	if uuidStr == "" {
		return errors.Validationf("UUID cannot be empty")
	}
	if _, err := uuid.Parse(uuidStr); err != nil {
		return errors.Validationf("invalid UUID format: %s", uuidStr)
	}
	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.Validationf("invalid email format: %s", email)
	}
	return nil
}

// ValidateFileSize validates file size in bytes
func ValidateFileSize(size int64, maxSize int64) error {
	if size <= 0 {
		return errors.Validationf("file size must be greater than 0")
	}
	if size > maxSize {
		return errors.Validationf("file size %d bytes exceeds maximum %d bytes", size, maxSize)
	}
	return nil
}

// ValidateFileExtension validates file extension against allowed list
func ValidateFileExtension(filename string, allowedExtensions []string) error {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return errors.Validationf("file must have an extension")
	}

	ext := strings.ToLower(parts[len(parts)-1])

	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return nil
		}
	}

	return errors.Validationf("file extension .%s is not allowed (allowed: %s)",
		ext, strings.Join(allowedExtensions, ", "))
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) error {
	if page < 1 {
		return errors.Validationf("page must be at least 1")
	}
	if limit < 1 {
		return errors.Validationf("limit must be at least 1")
	}
	if limit > 100 {
		return errors.Validationf("limit cannot exceed 100")
	}
	return nil
}

// ValidateEnum validates that a value is in a list of allowed values
func ValidateEnum(value string, allowedValues []string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return errors.Validationf("value '%s' is not allowed (allowed: %s)",
		value, strings.Join(allowedValues, ", "))
}

// Global validator instance
var global = New()

// Validate validates a struct using the global validator
func Validate(i interface{}) error {
	return global.Validate(i)
}

// ValidateVar validates a single variable using the global validator
func ValidateVar(field interface{}, tag string) error {
	return global.ValidateVar(field, tag)
}
