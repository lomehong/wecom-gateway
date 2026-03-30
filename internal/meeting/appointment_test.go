package meeting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockWeComClientForAppointment for testing meeting appointment service
type MockWeComClientForAppointment struct {
	CreateMeetingFunc       func(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error)
	CancelMeetingFunc       func(ctx context.Context, corpName, appName string, meetingID string) error
	UpdateMeetingInviteesFunc func(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error
	ListMeetingsFunc        func(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error)
	GetMeetingInfoFunc      func(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error)
}

func (m *MockWeComClientForAppointment) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	return nil
}
func (m *MockWeComClientForAppointment) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}
func (m *MockWeComClientForAppointment) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}
func (m *MockWeComClientForAppointment) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}
func (m *MockWeComClientForAppointment) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}
func (m *MockWeComClientForAppointment) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) GetTodoList(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) CreateTodo(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
	return "", nil
}
func (m *MockWeComClientForAppointment) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
	return nil
}
func (m *MockWeComClientForAppointment) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	return nil
}
func (m *MockWeComClientForAppointment) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	return nil
}
func (m *MockWeComClientForAppointment) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	return nil, nil
}
func (m *MockWeComClientForAppointment) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	return nil, "", nil
}
func (m *MockWeComClientForAppointment) CheckAvailability(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
	return nil, nil
}

func (m *MockWeComClientForAppointment) CreateMeeting(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
	if m.CreateMeetingFunc != nil {
		return m.CreateMeetingFunc(ctx, corpName, appName, params)
	}
	return &wecom.MeetingInfo{MeetingID: "meeting-123", Title: params.Title}, nil
}

func (m *MockWeComClientForAppointment) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	if m.CancelMeetingFunc != nil {
		return m.CancelMeetingFunc(ctx, corpName, appName, meetingID)
	}
	return nil
}

func (m *MockWeComClientForAppointment) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
	if m.UpdateMeetingInviteesFunc != nil {
		return m.UpdateMeetingInviteesFunc(ctx, corpName, appName, meetingID, invitees)
	}
	return nil
}

func (m *MockWeComClientForAppointment) ListMeetings(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
	if m.ListMeetingsFunc != nil {
		return m.ListMeetingsFunc(ctx, corpName, appName, opts)
	}
	return &wecom.MeetingListResult{
		Meetings: []wecom.MeetingInfo{
			{MeetingID: "meeting-1", Title: "Test Meeting"},
		},
	}, nil
}

func (m *MockWeComClientForAppointment) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
	if m.GetMeetingInfoFunc != nil {
		return m.GetMeetingInfoFunc(ctx, corpName, appName, meetingID)
	}
	return &wecom.MeetingInfo{MeetingID: meetingID, Title: "Test Meeting"}, nil
}

// --- Service Tests ---

func TestService_CreateMeeting_Success(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		CreateMeetingFunc: func(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
			return &wecom.MeetingInfo{
				MeetingID:     "meeting-123",
				Title:         params.Title,
				Duration:      params.Duration,
				MeetingLink:   "https://meeting.qq.com/dm/r/test",
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(1 * time.Hour)
	result, err := svc.CreateMeeting(context.Background(), authCtx, "Team Standup", startTime, 1800, nil, 0, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.MeetingID != "meeting-123" {
		t.Errorf("expected meeting ID meeting-123, got %s", result.MeetingID)
	}
}

func TestService_CreateMeeting_PastTime(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForAppointment{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.CreateMeeting(context.Background(), authCtx, "Test", time.Now().Add(-1*time.Hour), 1800, nil, 0, nil)
	if err == nil {
		t.Error("expected error for past start time")
	}
}

func TestService_CreateMeeting_InsufficientDuration(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForAppointment{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.CreateMeeting(context.Background(), authCtx, "Test", time.Now().Add(1*time.Hour), 30, nil, 0, nil)
	if err == nil {
		t.Error("expected error for duration < 60")
	}
}

func TestService_CreateMeeting_WithInvitees(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		CreateMeetingFunc: func(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
			if params.Invitees == nil {
				t.Error("expected invitees to be set")
			}
			if len(params.Invitees.UserIDs) != 2 {
				t.Errorf("expected 2 user invitees, got %d", len(params.Invitees.UserIDs))
			}
			return &wecom.MeetingInfo{MeetingID: "meeting-123"}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	invitees := &CreateMeetingInvitees{
		UserIDs: []string{"user1", "user2"},
		DeptIDs: []string{"dept1"},
	}

	_, err := svc.CreateMeeting(context.Background(), authCtx, "Test", time.Now().Add(1*time.Hour), 3600, invitees, 0, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_CreateMeeting_APIError(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		CreateMeetingFunc: func(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
			return nil, errors.New("create failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.CreateMeeting(context.Background(), authCtx, "Test", time.Now().Add(1*time.Hour), 3600, nil, 0, nil)
	if err == nil {
		t.Error("expected error from API")
	}
}

func TestService_CancelMeeting_Success(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		CancelMeetingFunc: func(ctx context.Context, corpName, appName string, meetingID string) error {
			if meetingID != "meeting-123" {
				t.Errorf("expected meeting ID meeting-123, got %s", meetingID)
			}
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.CancelMeeting(context.Background(), authCtx, "meeting-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_CancelMeeting_Error(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		CancelMeetingFunc: func(ctx context.Context, corpName, appName string, meetingID string) error {
			return errors.New("cancel failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.CancelMeeting(context.Background(), authCtx, "meeting-123")
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_UpdateInvitees_Success(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		UpdateMeetingInviteesFunc: func(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
			if len(invitees.UserIDs) != 2 {
				t.Errorf("expected 2 user IDs, got %d", len(invitees.UserIDs))
			}
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.UpdateInvitees(context.Background(), authCtx, "meeting-123", []string{"user1", "user2"}, []string{"dept1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_UpdateInvitees_Error(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		UpdateMeetingInviteesFunc: func(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
			return errors.New("update failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.UpdateInvitees(context.Background(), authCtx, "meeting-123", []string{"user1"}, nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_ListMeetings_Success(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		ListMeetingsFunc: func(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
			return &wecom.MeetingListResult{
				Meetings: []wecom.MeetingInfo{
					{MeetingID: "meeting-1", Title: "Meeting 1"},
					{MeetingID: "meeting-2", Title: "Meeting 2"},
				},
				NextCursor: "next-123",
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &ListMeetingsRequest{
		BeginDatetime: "2024-01-01T00:00:00Z",
		EndDatetime:   "2024-01-31T23:59:59Z",
	}

	result, err := svc.ListMeetings(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result.Meetings) != 2 {
		t.Errorf("expected 2 meetings, got %d", len(result.Meetings))
	}
	if result.NextCursor != "next-123" {
		t.Errorf("expected next cursor next-123, got %s", result.NextCursor)
	}
}

func TestService_ListMeetings_DefaultLimit(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		ListMeetingsFunc: func(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
			if opts.Limit != 50 {
				t.Errorf("expected default limit 50, got %d", opts.Limit)
			}
			return &wecom.MeetingListResult{}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &ListMeetingsRequest{
		BeginDatetime: "2024-01-01T00:00:00Z",
		EndDatetime:   "2024-01-31T23:59:59Z",
		// Limit not set, should default to 50
	}

	_, err := svc.ListMeetings(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_GetMeetingInfo_Success(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		GetMeetingInfoFunc: func(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
			return &wecom.MeetingInfo{
				MeetingID:   meetingID,
				Title:       "Test Meeting",
				MeetingLink: "https://meeting.qq.com/dm/r/test",
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.GetMeetingInfo(context.Background(), authCtx, "meeting-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.MeetingID != "meeting-123" {
		t.Errorf("expected meeting ID meeting-123, got %s", result.MeetingID)
	}
}

func TestService_GetMeetingInfo_Error(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{
		GetMeetingInfoFunc: func(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
			return nil, errors.New("get failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetMeetingInfo(context.Background(), authCtx, "meeting-123")
	if err == nil {
		t.Error("expected error")
	}
}

// --- Handler Tests ---

func TestHandler_CreateMeeting_ValidRequest(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"title": "Team Standup",
		"meeting_start_datetime": "2025-01-01T10:00:00Z",
		"meeting_duration": 1800,
		"invitees": {
			"userid": ["user1", "user2"],
			"department": ["dept1"]
		}
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meetings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateMeeting(c)
}

func TestHandler_CreateMeeting_MissingTitle(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"meeting_start_datetime": "2025-01-01T10:00:00Z",
		"meeting_duration": 1800
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meetings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateMeeting(c)
	// Should return 400
}

func TestHandler_CreateMeeting_InvalidTimeFormat(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"title": "Test",
		"meeting_start_datetime": "invalid-time",
		"meeting_duration": 1800
	}`

	c.Request = httptest.NewRequest("POST", "/v1/meetings", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateMeeting(c)
	// Should return 400
}

func TestHandler_CancelMeeting_ValidID(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/v1/meetings/meeting-123", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "meeting-123"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CancelMeeting(c)
}

func TestHandler_CancelMeeting_MissingID(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/v1/meetings/", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CancelMeeting(c)
	// Should return 400
}

func TestHandler_UpdateInvitees_ValidRequest(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"userid": ["user1", "user2"],
		"department": ["dept1"]
	}`

	c.Request = httptest.NewRequest("PUT", "/v1/meetings/meeting-123/invitees", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "meeting-123"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.UpdateInvitees(c)
}

func TestHandler_ListMeetings_ValidQuery(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/meetings?begin_datetime=2024-01-01T00:00:00Z&end_datetime=2024-01-31T23:59:59Z", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.ListMeetings(c)
}

func TestHandler_ListMeetings_MissingParams(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/meetings?begin_datetime=2024-01-01T00:00:00Z", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.ListMeetings(c)
	// Should return 400 for missing end_datetime
}

func TestHandler_GetMeetingInfo_ValidID(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/meetings/meeting-123", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "meeting-123"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetMeetingInfo(c)
}

func TestHandler_GetMeetingInfo_MissingID(t *testing.T) {
	mockClient := &MockWeComClientForAppointment{}
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/meetings/", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetMeetingInfo(c)
	// Should return 400
}

func TestCreateMeetingRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid request",
			json: `{
				"title": "Team Standup",
				"meeting_start_datetime": "2025-01-01T10:00:00Z",
				"meeting_duration": 1800
			}`,
			wantErr: false,
		},
		{
			name: "missing title",
			json: `{
				"meeting_start_datetime": "2025-01-01T10:00:00Z",
				"meeting_duration": 1800
			}`,
			wantErr: true,
		},
		{
			name: "missing start datetime",
			json: `{
				"title": "Test",
				"meeting_duration": 1800
			}`,
			wantErr: true,
		},
		{
			name: "missing duration",
			json: `{
				"title": "Test",
				"meeting_start_datetime": "2025-01-01T10:00:00Z"
			}`,
			wantErr: true,
		},
		{
			name: "duration too short",
			json: `{
				"title": "Test",
				"meeting_start_datetime": "2025-01-01T10:00:00Z",
				"meeting_duration": 30
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("POST", "/v1/meetings", bytes.NewBufferString(tt.json))
			c.Request.Header.Set("Content-Type", "application/json")

			var req CreateMeetingRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantErr && err == nil {
				t.Error("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Ensure all JSON round-trips work correctly
func TestCreateMeetingRequest_WithSettings(t *testing.T) {
	jsonBody := `{
		"title": "Test",
		"meeting_start_datetime": "2025-01-01T10:00:00Z",
		"meeting_duration": 3600,
		"settings": {
			"mute_upon_entry": true,
			"waiting_room": true,
			"enable_recording": false
		}
	}`

	var req CreateMeetingRequest
	err := json.Unmarshal([]byte(jsonBody), &req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if req.Settings == nil {
		t.Fatal("expected settings to be parsed")
	}
	if !req.Settings.MuteUponEntry {
		t.Error("expected mute_upon_entry to be true")
	}
	if !req.Settings.WaitingRoom {
		t.Error("expected waiting_room to be true")
	}
	if req.Settings.EnableRecording {
		t.Error("expected enable_recording to be false")
	}
}
