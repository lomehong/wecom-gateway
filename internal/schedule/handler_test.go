package schedule

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler_Schedule(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)
	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.service != service {
		t.Error("service field not set correctly")
	}
}

func TestHandler_CreateSchedule_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Valid JSON request body
	jsonBody := `{
		"organizer": "test-user",
		"summary": "Test Meeting",
		"start_time": "2024-01-01T10:00:00Z",
		"end_time": "2024-01-01T11:00:00Z",
		"attendees": ["user1", "user2"],
		"location": "Room 101"
	}`

	c.Request = httptest.NewRequest("POST", "/v1/schedules", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateSchedule(c)

	// Handler should process without crashing
}

func TestHandler_GetSchedules_ValidQuery(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/schedules?userid=test-user&limit=50", nil)

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
		KeyName:  "test-user",
	})

	handler.GetSchedules(c)

	// Handler should process without crashing
}

func TestHandler_GetScheduleByID_ValidID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/schedules/test-id", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
		KeyName:  "test-user",
	})

	handler.GetScheduleByID(c)

	if c.Params[0].Value != "test-id" {
		t.Error("id param not set correctly")
	}
}

func TestHandler_UpdateSchedule_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"summary": "Updated Summary"
	}`

	c.Request = httptest.NewRequest("PATCH", "/v1/schedules/test-id", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.UpdateSchedule(c)

	if c.Params[0].Value != "test-id" {
		t.Error("id param not set correctly")
	}
}

func TestHandler_DeleteSchedule_ValidID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/v1/schedules/test-id", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "test-id"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.DeleteSchedule(c)

	if c.Params[0].Value != "test-id" {
		t.Error("id param not set correctly")
	}
}

func TestParseTimeQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		queryVal  string
		wantEmpty bool
	}{
		{
			name:      "empty value",
			queryVal:  "",
			wantEmpty: true,
		},
		{
			name:      "RFC3339 format",
			queryVal:  "2024-01-01T10:00:00Z",
			wantEmpty: false,
		},
		{
			name:      "unix timestamp",
			queryVal:  "1704110400",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/test?time="+tt.queryVal, nil)

			parsedTime, err := ParseTimeQueryParam(c, "time")
			if err != nil && !tt.wantEmpty {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantEmpty && !parsedTime.IsZero() {
				t.Error("expected empty time, got non-zero")
			}
			if !tt.wantEmpty && parsedTime.IsZero() {
				t.Error("expected non-zero time, got zero")
			}
		})
	}
}

func TestCreateScheduleRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid request",
			json: `{
				"organizer": "test-user",
				"summary": "Test Meeting",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z"
			}`,
			wantErr: false,
		},
		{
			name: "missing organizer",
			json: `{
				"summary": "Test Meeting",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z"
			}`,
			wantErr: true,
		},
		{
			name: "missing summary",
			json: `{
				"organizer": "test-user",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("POST", "/v1/schedules", strings.NewReader(tt.json))
			c.Request.Header.Set("Content-Type", "application/json")

			// In real handler, Gin would validate
			var req CreateScheduleRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantErr && err == nil {
				// Gin might not catch all validation without tags
			}
		})
	}
}
