package apierror

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail"`
	Instance  string `json:"instance"`
	RequestID string `json:"request_id"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

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

func RespondError(c *gin.Context, apiErr *APIError) {
	c.JSON(apiErr.Status, apiErr)
}
