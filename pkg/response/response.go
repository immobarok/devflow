package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Meta holds pagination metadata for list responses.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// FieldError represents a single validation error on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse is the standard envelope for successful API responses.
type SuccessResponse struct {
	Status string `json:"status" example:"success"`
	Data   any    `json:"data"`
}

// PaginatedResponse extends the success envelope with pagination metadata.
type PaginatedResponse struct {
	Status string `json:"status" example:"success"`
	Data   any    `json:"data"`
	Meta   Meta   `json:"meta"`
}

// ErrorResponse is the standard envelope for error API responses.
type ErrorResponse struct {
	Status string      `json:"status" example:"error"`
	Error  ErrorDetail `json:"error"`
}

// ErrorDetail holds the error code, message, and optional field-level details.
type ErrorDetail struct {
	Code    string       `json:"code" example:"BAD_REQUEST"`
	Message string       `json:"message" example:"Invalid request body"`
	Details []FieldError `json:"details,omitempty"`
}

// Success sends a standard success response.
//
//	response.Success(c, http.StatusOK, gin.H{"user": user})
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

// SuccessWithMeta sends a success response with pagination metadata.
//
//	response.SuccessWithMeta(c, http.StatusOK, tasks, response.Meta{Page: 1, PerPage: 20, Total: 100, TotalPages: 5})
func SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta Meta) {
	c.JSON(statusCode, PaginatedResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	})
}

// Error sends a standard error response.
//
//	response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, ErrorResponse{
		Status: "error",
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError sends a 422 response with field-level validation errors.
//
//	response.ValidationError(c, []response.FieldError{{Field: "title", Message: "must not be empty"}})
func ValidationError(c *gin.Context, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
		Status: "error",
		Error: ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "One or more fields failed validation",
			Details: details,
		},
	})
}
