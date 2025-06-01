package client

import "fmt"

// APIError represents an error from the Mem0 API
type APIError struct {
	Message    string
	StatusCode int
	Body       string
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API request failed (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API request failed: %s", e.Message)
}

// NewAPIError creates a new APIError
func NewAPIError(message string, statusCode int, body string) *APIError {
	return &APIError{
		Message:    message,
		StatusCode: statusCode,
		Body:       body,
	}
}

// ValidationError represents a client-side validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
