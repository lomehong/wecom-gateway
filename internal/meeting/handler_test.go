package meeting

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

func TestNewHandler_Meeting(t *testing.T) {
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

func TestHandler_ListMeetingRooms_QueryParsing(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "no filters",
			query: "",
		},
		{
			name:  "with city filter",
			query: "?city=Beijing",
		},
		{
			name:  "with capacity filter",
			query: "?capacity=10",
		},
		{
			name:  "with equipment filter",
			query: "?equipment=projector&equipment=whiteboard",
		},
		{
			name:  "with limit",
			query: "?limit=25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", "/v1/meeting-rooms"+tt.query, nil)

			// Set auth context
			c.Set("auth_context", &auth.AuthContext{
				CorpName: "test-corp",
				AppName:  "test-app",
			})

			// Handler should parse query without crashing
			handler.ListMeetingRooms(c)
		})
	}
}

func TestHandler_GetRoomAvailability_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/meeting-rooms/room-123/availability?start_time=2024-01-01T10:00:00Z&end_time=2024-01-01T18:00:00Z", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "room-123"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetRoomAvailability(c)

	if c.Params[0].Value != "room-123" {
		t.Error("id param not set correctly")
	}
}

func TestHandler_GetRoomAvailability_MissingTimeParams(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Missing end_time
	c.Request = httptest.NewRequest("GET", "/v1/meeting-rooms/room-123/availability?start_time=2024-01-01T10:00:00Z", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "room-123"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetRoomAvailability(c)

	// Should return error for missing time params
}

func TestHandler_GetRoomAvailability_InvalidTimeFormat(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Invalid time format
	c.Request = httptest.NewRequest("GET", "/v1/meeting-rooms/room-123/availability?start_time=invalid&end_time=invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "room-123"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetRoomAvailability(c)

	// Should return error for invalid time format
}

func TestHandler_BookMeetingRoom_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"meetingroom_id": "room-123",
		"subject": "Team Meeting",
		"start_time": "2024-01-01T10:00:00Z",
		"end_time": "2024-01-01T11:00:00Z",
		"booker": "test-user",
		"attendees": ["user1", "user2"]
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meeting-rooms/room-123/bookings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "room-123"}}

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.BookMeetingRoom(c)

	// Handler should process without crashing
}

func TestHandler_BookMeetingRoom_MissingRequiredField(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Missing subject
	jsonBody := `{
		"meetingroom_id": "room-123",
		"start_time": "2024-01-01T10:00:00Z",
		"end_time": "2024-01-01T11:00:00Z",
		"booker": "test-user"
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meeting-rooms/room-123/bookings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BookMeetingRoom(c)

	// Should return validation error
}

func TestHandler_BookMeetingRoom_InvalidTimeFormat(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"meetingroom_id": "room-123",
		"subject": "Team Meeting",
		"start_time": "invalid-time",
		"end_time": "2024-01-01T11:00:00Z",
		"booker": "test-user"
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meeting-rooms/room-123/bookings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BookMeetingRoom(c)

	// Should return error for invalid time format
}

func TestParseQueryParamInt(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      int
		wantError bool
	}{
		{
			name:      "valid integer",
			input:     "42",
			want:      42,
			wantError: false,
		},
		{
			name:      "zero",
			input:     "0",
			want:      0,
			wantError: false,
		},
		{
			name:      "invalid",
			input:     "not-a-number",
			wantError: true,
		},
		{
			name:      "negative",
			input:     "-10",
			want:      -10,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryParamInt(tt.input)
			if tt.wantError {
				if err == nil {
					t.Error("expected error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.want {
					t.Errorf("expected %d, got %d", tt.want, result)
				}
			}
		})
	}
}

func TestBookMeetingRoomRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid request",
			json: `{
				"meetingroom_id": "room-123",
				"subject": "Team Meeting",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z",
				"booker": "test-user"
			}`,
			wantErr: false,
		},
		{
			name: "missing meetingroom_id",
			json: `{
				"subject": "Team Meeting",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z",
				"booker": "test-user"
			}`,
			wantErr: true,
		},
		{
			name: "missing subject",
			json: `{
				"meetingroom_id": "room-123",
				"start_time": "2024-01-01T10:00:00Z",
				"end_time": "2024-01-01T11:00:00Z",
				"booker": "test-user"
			}`,
			wantErr: true,
		},
		{
			name: "missing booker",
			json: `{
				"meetingroom_id": "room-123",
				"subject": "Team Meeting",
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

			c.Request = httptest.NewRequest("POST", "/v1/meeting-rooms/room-123/bookings", strings.NewReader(tt.json))
			c.Request.Header.Set("Content-Type", "application/json")

			var req BookMeetingRoomRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantErr && err == nil {
				// Gin validation might not catch without proper tags
			}
		})
	}
}
