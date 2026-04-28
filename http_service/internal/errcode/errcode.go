/*
 * Unified response and error helpers.
 * 1. Define stable business error codes.
 * 2. Provide shared success and failure JSON writers.
 * 3. Keep request metadata consistent across all handlers.
 */
package errcode

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// 1. CodeOK is the shared success code.
	CodeOK = "OK"
	// 2. CodeAuthRequired is returned when login is required.
	CodeAuthRequired = "AUTH_REQUIRED"
	// 3. CodeAuthForbidden is returned when resource access is forbidden.
	CodeAuthForbidden = "AUTH_FORBIDDEN"
	// 4. CodeValidationError is returned when request validation fails.
	CodeValidationError = "VALIDATION_ERROR"
	// 5. CodeNotFound is returned when resource cannot be found.
	CodeNotFound = "RESOURCE_NOT_FOUND"
	// 6. CodeExpired is returned when a resource is expired.
	CodeExpired = "RESOURCE_EXPIRED"
	// 7. CodeHidden is returned when a resource is hidden.
	CodeHidden = "RESOURCE_HIDDEN"
	// 8. CodeVisibilityForbidden is returned when visibility rules block access.
	CodeVisibilityForbidden = "VISIBILITY_FORBIDDEN"
	// 9. CodeRateLimited is returned when request rate is limited.
	CodeRateLimited = "RATE_LIMITED"
	// 10. CodeIdempotencyConflict is reserved for idempotent write conflicts.
	CodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"
	// 11. CodeInternalError is returned for unexpected server failures.
	CodeInternalError = "INTERNAL_ERROR"
)

// 12. FieldError defines field level validation detail.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// 13. AppError is the project-wide business error type.
type AppError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
	Cause   error        `json:"-"`
}

// 14. Error implements the error interface.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

// 15. Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

// 16. New creates an AppError without field details.
func New(code string, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// 17. NewWithFields creates an AppError with validation details.
func NewWithFields(code string, message string, fields []FieldError) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Errors:  fields,
	}
}

// 18. Success writes the shared success response.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":       CodeOK,
		"message":    "success",
		"data":       data,
		"request_id": RequestID(c),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// 19. WriteError writes the shared failure response.
func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = New(CodeInternalError, "internal server error")
	}

	c.JSON(httpStatus(appErr.Code), gin.H{
		"code":       appErr.Code,
		"message":    appErr.Message,
		"errors":     appErr.Errors,
		"request_id": RequestID(c),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// 20. RequestID extracts request id from gin context.
func RequestID(c *gin.Context) string {
	value, exists := c.Get("request_id")
	if !exists {
		return ""
	}

	requestID, _ := value.(string)
	return requestID
}

// 21. httpStatus maps business codes to HTTP status.
func httpStatus(code string) int {
	switch code {
	case CodeAuthRequired:
		return http.StatusUnauthorized
	case CodeAuthForbidden, CodeVisibilityForbidden:
		return http.StatusForbidden
	case CodeValidationError:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeExpired, CodeHidden, CodeIdempotencyConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
