package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"pharmacy-backend/internal/domain"
)

// Envelope is the standard success response shape.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta carries pagination information.
type Meta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// ErrorEnvelope is the standard error response shape.
type ErrorEnvelope struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// ErrorBody describes a single error.
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// OK writes a 200 success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// Paginated writes a 200 success response with pagination meta.
func Paginated(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    data,
		Meta:    &Meta{Page: page, PageSize: pageSize, Total: total},
	})
}

// NoContent writes a 204 response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error inspects err and writes the appropriate error envelope. Domain
// AppErrors map to their declared code/status; anything else becomes a 500.
func Error(c *gin.Context, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPCode, ErrorEnvelope{
			Success: false,
			Error:   ErrorBody{Code: appErr.Code, Message: appErr.Message},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, ErrorEnvelope{
		Success: false,
		Error:   ErrorBody{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred"},
	})
}

// ValidationError writes a 400 with optional field details.
func ValidationError(c *gin.Context, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, ErrorEnvelope{
		Success: false,
		Error:   ErrorBody{Code: "VALIDATION_ERROR", Message: message, Details: details},
	})
}

// Abort writes an error envelope and aborts the middleware chain.
func Abort(c *gin.Context, appErr *domain.AppError) {
	c.AbortWithStatusJSON(appErr.HTTPCode, ErrorEnvelope{
		Success: false,
		Error:   ErrorBody{Code: appErr.Code, Message: appErr.Message},
	})
}
