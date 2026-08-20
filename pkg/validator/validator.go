// Package validator provides a singleton go-playground/validator instance
// with structured error formatting.
package validator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns the singleton Validate instance.
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()
	})
	return instance
}

// Validate validates a struct and returns a human-readable error string,
// or nil if the struct is valid.
func Validate(s any) error {
	if err := Get().Struct(s); err != nil {
		return FormatErrors(err)
	}
	return nil
}

// FormatErrors converts validator.ValidationErrors into a single readable string.
func FormatErrors(err error) error {
	var validationErrors validator.ValidationErrors
	if !asValidationErrors(err, &validationErrors) {
		return err
	}

	messages := make([]string, 0, len(validationErrors))
	for _, fe := range validationErrors {
		messages = append(messages, formatField(fe))
	}
	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func formatField(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "e164":
		return fmt.Sprintf("%s must be a valid E.164 phone number (e.g. +1234567890)", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "hexadecimal", "mongodb":
		return fmt.Sprintf("%s must be a valid hexadecimal string", field)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, fe.Tag())
	}
}

// asValidationErrors tries to unwrap err as ValidationErrors.
func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}
