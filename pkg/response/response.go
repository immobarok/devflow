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

// successResponse is the standard envelope for successful API responses.
type successResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// paginatedResponse extends the success envelope with pagination metadata.
type paginatedResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Meta   Meta        `json:"meta"`
}

// errorResponse is the standard envelope for error API responses.
type errorResponse struct {
	Status string      `json:"status"`
	Error  errorDetail `json:"error"`
}

// errorDetail holds the error code, message, and optional field-level details.
type errorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// Success sends a standard success response.
//
//	response.Success(c, http.StatusOK, gin.H{"user": user})
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, successResponse{
		Status: "success",
		Data:   data,
	})
}

// SuccessWithMeta sends a success response with pagination metadata.
//
//	response.SuccessWithMeta(c, http.StatusOK, tasks, response.Meta{Page: 1, PerPage: 20, Total: 100, TotalPages: 5})
func SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta Meta) {
	c.JSON(statusCode, paginatedResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	})
}

// Error sends a standard error response.
//
//	response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, errorResponse{
		Status: "error",
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError sends a 422 response with field-level validation errors.
//
//	response.ValidationError(c, []response.FieldError{{Field: "title", Message: "must not be empty"}})
func ValidationError(c *gin.Context, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, errorResponse{
		Status: "error",
		Error: errorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "One or more fields failed validation",
			Details: details,
		},
	})
}
