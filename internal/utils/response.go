package utils

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Response represents the standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *Error      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ResponseHelper provides utility methods for sending standardized responses
type ResponseHelper struct {
	logger *zap.Logger
}

func NewResponseHelper(logger *zap.Logger) *ResponseHelper {
	return &ResponseHelper{logger: logger}
}

func (r *ResponseHelper) buildBaseResponse(c echo.Context) Response {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = "unknown"
	}

	return Response{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
	}
}

// Success responses
func (r *ResponseHelper) Success(c echo.Context, data interface{}) error {
	response := r.buildBaseResponse(c)
	response.Success = true
	response.Data = data

	r.logger.Debug("API success response",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
		zap.Int("status", http.StatusOK),
	)

	return c.JSON(http.StatusOK, response)
}

func (r *ResponseHelper) SuccessWithMessage(c echo.Context, data interface{}, message string) error {
	response := r.buildBaseResponse(c)
	response.Success = true
	response.Message = message
	response.Data = data

	return c.JSON(http.StatusOK, response)
}

func (r *ResponseHelper) Created(c echo.Context, data interface{}) error {
	response := r.buildBaseResponse(c)
	response.Success = true
	response.Data = data

	r.logger.Info("Resource created",
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
	)

	return c.JSON(http.StatusCreated, response)
}

// Error responses with logging
func (r *ResponseHelper) Error(c echo.Context, status int, errorCode, message string, err error) error {
	response := r.buildBaseResponse(c)
	response.Success = false
	response.Error = &Error{
		Code:    errorCode,
		Message: message,
	}

	// Log appropriately based on status code
	logFields := []zap.Field{
		zap.String("path", c.Path()),
		zap.String("method", c.Request().Method),
		zap.Int("status", status),
		zap.String("error_code", errorCode),
		zap.String("client_message", message),
	}

	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}

	if status >= 500 {
		r.logger.Error("Server error", logFields...)
	} else {
		r.logger.Warn("Client error", logFields...)
	}

	return c.JSON(status, response)
}

// Convenience methods
func (r *ResponseHelper) BadRequest(c echo.Context, message string, err error) error {
	return r.Error(c, http.StatusBadRequest, "BAD_REQUEST", message, err)
}

func (r *ResponseHelper) Unauthorized(c echo.Context, message string, err error) error {
	return r.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, err)
}

func (r *ResponseHelper) Forbidden(c echo.Context, message string, err error) error {
	return r.Error(c, http.StatusForbidden, "FORBIDDEN", message, err)
}

func (r *ResponseHelper) NotFound(c echo.Context, message string, err error) error {
	return r.Error(c, http.StatusNotFound, "NOT_FOUND", message, err)
}

func (r *ResponseHelper) Conflict(c echo.Context, message string, err error) error {
	return r.Error(c, http.StatusConflict, "CONFLICT", message, err)
}

func (r *ResponseHelper) ValidationError(c echo.Context, details interface{}, err error) error {
	response := r.buildBaseResponse(c)
	response.Success = false
	response.Error = &Error{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: details,
	}

	r.logger.Warn("Validation error",
		zap.String("path", c.Path()),
		zap.Any("details", details),
		zap.Error(err),
	)

	return c.JSON(http.StatusBadRequest, response)
}

func (r *ResponseHelper) InternalServerError(c echo.Context, err error) error {
	return r.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", err)
}
