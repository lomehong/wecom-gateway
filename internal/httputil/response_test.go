package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		errCode    int
		wantStatus int
	}{
		{ErrCodeSuccess, http.StatusOK},
		{ErrCodeInvalidParams, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeConflict, http.StatusConflict},
		{ErrCodeRateLimitExceed, http.StatusTooManyRequests},
		{ErrCodeInternalError, http.StatusInternalServerError},
		{ErrCodeBadGateway, http.StatusBadGateway},
		{ErrCodeServiceUnavail, http.StatusServiceUnavailable},
		{99999, http.StatusInternalServerError}, // unknown code defaults to 500
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := GetStatusCode(tc.errCode)
			if got != tc.wantStatus {
				t.Errorf("GetStatusCode(%d) = %d, want %d", tc.errCode, got, tc.wantStatus)
			}
		})
	}
}

func TestGetDefaultMessage(t *testing.T) {
	tests := []struct {
		errCode int
		want    string
	}{
		{ErrCodeSuccess, MsgSuccess},
		{ErrCodeInvalidParams, MsgInvalidParams},
		{ErrCodeUnauthorized, MsgUnauthorized},
		{ErrCodeForbidden, MsgForbidden},
		{ErrCodeNotFound, MsgNotFound},
		{ErrCodeConflict, MsgConflict},
		{ErrCodeRateLimitExceed, MsgRateLimitExceed},
		{ErrCodeInternalError, MsgInternalError},
		{ErrCodeBadGateway, MsgBadGateway},
		{ErrCodeServiceUnavail, MsgServiceUnavail},
		{99999, MsgInternalError}, // unknown code defaults to internal error
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			got := GetDefaultMessage(tc.errCode)
			if got != tc.want {
				t.Errorf("GetDefaultMessage(%d) = %q, want %q", tc.errCode, got, tc.want)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeSuccess {
		t.Errorf("expected code %d, got %d", ErrCodeSuccess, resp.Code)
	}
	if resp.Message != MsgSuccess {
		t.Errorf("expected message %q, got %q", MsgSuccess, resp.Message)
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Created(c, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeSuccess {
		t.Errorf("expected code %d, got %d", ErrCodeSuccess, resp.Code)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, ErrCodeInvalidParams)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeInvalidParams {
		t.Errorf("expected code %d, got %d", ErrCodeInvalidParams, resp.Code)
	}
}

func TestErrorWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorWithMessage(c, ErrCodeNotFound, "custom not found msg")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "custom not found msg" {
		t.Errorf("expected custom message, got %q", resp.Message)
	}
}

func TestErrorWithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"field": "name"}
	ErrorWithData(c, ErrCodeInvalidParams, "validation failed", data)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "validation failed" {
		t.Errorf("expected 'validation failed', got %q", resp.Message)
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "invalid field")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeInvalidParams {
		t.Errorf("expected code %d, got %d", ErrCodeInvalidParams, resp.Code)
	}
}

func TestBadRequest_NoDetails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %d, got %d", ErrCodeUnauthorized, resp.Code)
	}
}

func TestUnauthorized_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "token expired")

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "token expired" {
		t.Errorf("expected 'token expired', got %q", resp.Message)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestForbidden_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "no access")

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "no access" {
		t.Errorf("expected 'no access', got %q", resp.Message)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "api key")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "api key not found" {
		t.Errorf("expected 'api key not found', got %q", resp.Message)
	}
}

func TestNotFound_NoResource(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != MsgNotFound {
		t.Errorf("expected %q, got %q", MsgNotFound, resp.Message)
	}
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Conflict(c, "duplicate entry")

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeConflict {
		t.Errorf("expected code %d, got %d", ErrCodeConflict, resp.Code)
	}
}

func TestRateLimitExceeded(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RateLimitExceeded(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeRateLimitExceed {
		t.Errorf("expected code %d, got %d", ErrCodeRateLimitExceed, resp.Code)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalError(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeInternalError {
		t.Errorf("expected code %d, got %d", ErrCodeInternalError, resp.Code)
	}
}

func TestInternalError_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	InternalError(c, "db connection failed")

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "db connection failed" {
		t.Errorf("expected 'db connection failed', got %q", resp.Message)
	}
}

func TestBadGateway(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadGateway(c, "upstream error")

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeBadGateway {
		t.Errorf("expected code %d, got %d", ErrCodeBadGateway, resp.Code)
	}
}

func TestServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ServiceUnavailable(c, "maintenance mode")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != ErrCodeServiceUnavail {
		t.Errorf("expected code %d, got %d", ErrCodeServiceUnavail, resp.Code)
	}
}

func TestSuccess_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Success(c, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var raw map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &raw)
	if _, exists := raw["data"]; exists {
		t.Error("expected no 'data' field when nil")
	}
}

func TestErrorConstants(t *testing.T) {
	constants := map[string]int{
		"ErrCodeSuccess":          ErrCodeSuccess,
		"ErrCodeInvalidParams":    ErrCodeInvalidParams,
		"ErrCodeUnauthorized":     ErrCodeUnauthorized,
		"ErrCodeForbidden":        ErrCodeForbidden,
		"ErrCodeNotFound":         ErrCodeNotFound,
		"ErrCodeConflict":         ErrCodeConflict,
		"ErrCodeRateLimitExceed":  ErrCodeRateLimitExceed,
		"ErrCodeInternalError":    ErrCodeInternalError,
		"ErrCodeBadGateway":       ErrCodeBadGateway,
		"ErrCodeServiceUnavail":   ErrCodeServiceUnavail,
	}

	if len(constants) != 10 {
		t.Errorf("expected 10 error codes, got %d", len(constants))
	}
}
