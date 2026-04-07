// Package apierror provides standardized API error types following RFC 7807
// Problem Details for HTTP APIs. It offers constructors for common HTTP error
// responses and integrates with the Gin framework for easy error handling.
package apierror

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError represents a Problem Details object as defined in RFC 7807.
// It provides a standard way to communicate error details in HTTP API responses.
type APIError struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	Instance  string `json:"instance"`
	RequestID string `json:"request_id"`
}

// Error implements the error interface and returns a formatted error string
// combining the title and detail fields.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

// NewValidation creates a new validation error with HTTP status 422 Unprocessable Entity.
// Use this for request validation failures.
func NewValidation(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/validation",
		Title:     "Validation Error",
		Status:    http.StatusUnprocessableEntity,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// NewUnauthorized creates a new unauthorized error with HTTP status 401 Unauthorized.
// Use this when authentication is required and has failed or not been provided.
func NewUnauthorized(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/unauthorized",
		Title:     "Unauthorized",
		Status:    http.StatusUnauthorized,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// NewForbidden creates a new forbidden error with HTTP status 403 Forbidden.
// Use this when the authenticated user does not have access to the requested resource.
func NewForbidden(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/forbidden",
		Title:     "Forbidden",
		Status:    http.StatusForbidden,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// NewNotFound creates a new not found error with HTTP status 404 Not Found.
// Use this when the requested resource does not exist.
func NewNotFound(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/not_found",
		Title:     "Not Found",
		Status:    http.StatusNotFound,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// NewConflict creates a new conflict error with HTTP status 409 Conflict.
// Use this when the request conflicts with the current state of the resource.
func NewConflict(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/conflict",
		Title:     "Conflict",
		Status:    http.StatusConflict,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// NewInternal creates a new internal server error with HTTP status 500 Internal Server Error.
// Use this for unexpected server-side errors.
func NewInternal(detail, instance, requestID string) *APIError {
	return &APIError{
		Type:      "https://skeleton.app/errors/internal",
		Title:     "Internal Server Error",
		Status:    http.StatusInternalServerError,
		Detail:    detail,
		Instance:  instance,
		RequestID: requestID,
	}
}

// RespondError writes the APIError as a JSON response to the Gin context
// using the error's HTTP status code.
func RespondError(c *gin.Context, apiErr *APIError) {
	c.JSON(apiErr.Status, apiErr)
}
