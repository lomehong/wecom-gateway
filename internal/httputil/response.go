package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the standard API response format
type Response struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"ok"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// Error codes
const (
	ErrCodeSuccess          = 0
	ErrCodeInvalidParams    = 40001
	ErrCodeUnauthorized     = 40101
	ErrCodeForbidden        = 40301
	ErrCodeNotFound         = 40401
	ErrCodeConflict         = 40901
	ErrCodeRateLimitExceed  = 42901
	ErrCodeInternalError    = 50001
	ErrCodeBadGateway       = 50201
	ErrCodeServiceUnavail   = 50301
)

// Error messages
const (
	MsgSuccess           = "ok"
	MsgInvalidParams     = "invalid parameters"
	MsgUnauthorized      = "unauthorized"
	MsgForbidden         = "forbidden"
	MsgNotFound          = "resource not found"
	MsgConflict          = "resource conflict"
	MsgRateLimitExceed   = "rate limit exceeded"
	MsgInternalError     = "internal server error"
	MsgBadGateway        = "bad gateway"
	MsgServiceUnavail    = "service unavailable"
	MsgInvalidAPIKey     = "invalid or missing API key"
	MsgAPIKeyDisabled    = "API key is disabled"
	MsgAPIKeyExpired     = "API key has expired"
	MsgPermissionDenied  = "permission denied"
	MsgInvalidCredential = "invalid WeChat Work credentials"
	MsgTokenExpired      = "access token expired"
)

// HTTP status code mapping
var errorStatusCodes = map[int]int{
	ErrCodeSuccess:         http.StatusOK,
	ErrCodeInvalidParams:   http.StatusBadRequest,
	ErrCodeUnauthorized:    http.StatusUnauthorized,
	ErrCodeForbidden:       http.StatusForbidden,
	ErrCodeNotFound:        http.StatusNotFound,
	ErrCodeConflict:        http.StatusConflict,
	ErrCodeRateLimitExceed: http.StatusTooManyRequests,
	ErrCodeInternalError:   http.StatusInternalServerError,
	ErrCodeBadGateway:      http.StatusBadGateway,
	ErrCodeServiceUnavail:  http.StatusServiceUnavailable,
}

// GetStatusCode returns the HTTP status code for an error code
func GetStatusCode(errorCode int) int {
	if statusCode, ok := errorStatusCodes[errorCode]; ok {
		return statusCode
	}
	return http.StatusInternalServerError
}

// GetDefaultMessage returns the default message for an error code
func GetDefaultMessage(errorCode int) string {
	switch errorCode {
	case ErrCodeSuccess:
		return MsgSuccess
	case ErrCodeInvalidParams:
		return MsgInvalidParams
	case ErrCodeUnauthorized:
		return MsgUnauthorized
	case ErrCodeForbidden:
		return MsgForbidden
	case ErrCodeNotFound:
		return MsgNotFound
	case ErrCodeConflict:
		return MsgConflict
	case ErrCodeRateLimitExceed:
		return MsgRateLimitExceed
	case ErrCodeInternalError:
		return MsgInternalError
	case ErrCodeBadGateway:
		return MsgBadGateway
	case ErrCodeServiceUnavail:
		return MsgServiceUnavail
	default:
		return MsgInternalError
	}
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(GetStatusCode(ErrCodeSuccess), Response{
		Code:    ErrCodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    ErrCodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	})
}

// Error sends an error response with default message
func Error(c *gin.Context, errorCode int) {
	c.JSON(GetStatusCode(errorCode), Response{
		Code:    errorCode,
		Message: GetDefaultMessage(errorCode),
	})
}

// ErrorWithMessage sends an error response with custom message
func ErrorWithMessage(c *gin.Context, errorCode int, message string) {
	c.JSON(GetStatusCode(errorCode), Response{
		Code:    errorCode,
		Message: message,
	})
}

// ErrorWithData sends an error response with data
func ErrorWithData(c *gin.Context, errorCode int, message string, data interface{}) {
	c.JSON(GetStatusCode(errorCode), Response{
		Code:    errorCode,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a 400 bad request error
func BadRequest(c *gin.Context, details ...string) {
	resp := Response{
		Code:    ErrCodeInvalidParams,
		Message: MsgInvalidParams,
	}
	if len(details) > 0 {
		resp.Data = map[string]interface{}{"details": details}
	}
	c.JSON(http.StatusBadRequest, resp)
}

// Unauthorized sends a 401 unauthorized error
func Unauthorized(c *gin.Context, message ...string) {
	msg := MsgUnauthorized
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code:    ErrCodeUnauthorized,
		Message: msg,
	})
}

// Forbidden sends a 403 forbidden error
func Forbidden(c *gin.Context, message ...string) {
	msg := MsgForbidden
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusForbidden, Response{
		Code:    ErrCodeForbidden,
		Message: msg,
	})
}

// NotFound sends a 404 not found error
func NotFound(c *gin.Context, resource ...string) {
	msg := MsgNotFound
	if len(resource) > 0 && resource[0] != "" {
		msg = resource[0] + " not found"
	}
	c.JSON(http.StatusNotFound, Response{
		Code:    ErrCodeNotFound,
		Message: msg,
	})
}

// Conflict sends a 409 conflict error
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Code:    ErrCodeConflict,
		Message: message,
	})
}

// RateLimitExceeded sends a 429 rate limit exceeded error
func RateLimitExceeded(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    ErrCodeRateLimitExceed,
		Message: MsgRateLimitExceed,
	})
}

// InternalError sends a 500 internal server error
func InternalError(c *gin.Context, message ...string) {
	msg := MsgInternalError
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    ErrCodeInternalError,
		Message: msg,
	})
}

// BadGateway sends a 502 bad gateway error
func BadGateway(c *gin.Context, message string) {
	c.JSON(http.StatusBadGateway, Response{
		Code:    ErrCodeBadGateway,
		Message: message,
	})
}

// ServiceUnavailable sends a 503 service unavailable error
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Code:    ErrCodeServiceUnavail,
		Message: message,
	})
}
