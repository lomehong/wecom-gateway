package schedule

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func TestNewService_Schedule(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.wecomClient != nil {
		// Mock client wasn't set, but that's ok for this test
	}
}

func TestService_CreateSchedule_Success(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		CreateScheduleFunc: func(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
			return &wecom.Schedule{
				ScheduleID: "test-schedule-id",
				Summary:    params.Summary,
				StartTime:  params.StartTime,
				EndTime:    params.EndTime,
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CreateScheduleRequest{
		Organizer:   "test-user",
		Summary:     "Test Meeting",
		StartTime:   time.Now().Add(1 * time.Hour),
		EndTime:     time.Now().Add(2 * time.Hour),
		Attendees:   []string{"user1", "user2"},
		Location:    "Room 101",
		Description: "Test description",
	}

	schedule, err := svc.CreateSchedule(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if schedule == nil {
		t.Fatal("expected schedule, got nil")
	}
	if schedule.Summary != req.Summary {
		t.Errorf("expected summary %s, got %s", req.Summary, schedule.Summary)
	}
}

func TestService_CreateSchedule_ValidationError(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForSchedule{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CreateScheduleRequest{
		Organizer: "test-user",
		Summary:   "Test Meeting",
		StartTime: time.Now().Add(2 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour), // End before start
	}

	_, err := svc.CreateSchedule(context.Background(), authCtx, req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestService_CreateSchedule_DefaultReminder(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		CreateScheduleFunc: func(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
			if params.RemindBeforeMin != 15 {
				t.Errorf("expected default reminder 15, got %d", params.RemindBeforeMin)
			}
			return &wecom.Schedule{}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CreateScheduleRequest{
		Organizer: "test-user",
		Summary:   "Test Meeting",
		StartTime: time.Now().Add(1 * time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		// RemindBeforeMin not set, should default to 15
	}

	_, err := svc.CreateSchedule(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_GetSchedules_Success(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		GetSchedulesFunc: func(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
			return []*wecom.Schedule{
				{ScheduleID: "schedule1", Summary: "Meeting 1"},
				{ScheduleID: "schedule2", Summary: "Meeting 2"},
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &GetSchedulesRequest{
		UserID:    "test-user",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
		Limit:     50,
	}

	schedules, err := svc.GetSchedules(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

func TestService_GetSchedules_DefaultTimeRange(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		GetSchedulesFunc: func(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
			// Check default time range (7 days)
			duration := endTime.Sub(startTime)
			expectedDuration := 7 * 24 * time.Hour
			if duration < expectedDuration-1*time.Hour || duration > expectedDuration+1*time.Hour {
				t.Errorf("expected time range ~7 days, got %v", duration)
			}
			return []*wecom.Schedule{}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &GetSchedulesRequest{
		UserID: "test-user",
		// StartTime and EndTime not set, should use defaults
	}

	_, err := svc.GetSchedules(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_UpdateSchedule_Success(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		UpdateScheduleFunc: func(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &UpdateScheduleRequest{
		Summary: "Updated summary",
	}

	err := svc.UpdateSchedule(context.Background(), authCtx, "test-id", req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_UpdateSchedule_ValidationError(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForSchedule{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &UpdateScheduleRequest{
		StartTime: time.Now().Add(2 * time.Hour),
		EndTime:   time.Now().Add(1 * time.Hour), // End before start
	}

	err := svc.UpdateSchedule(context.Background(), authCtx, "test-id", req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestService_DeleteSchedule_Success(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		DeleteScheduleFunc: func(ctx context.Context, corpName, appName string, scheduleID string) error {
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.DeleteSchedule(context.Background(), authCtx, "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_DeleteSchedule_Error(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		DeleteScheduleFunc: func(ctx context.Context, corpName, appName string, scheduleID string) error {
			return errors.New("delete failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.DeleteSchedule(context.Background(), authCtx, "test-id")
	if err == nil {
		t.Error("expected error")
	}
}

// --- Availability Tests (Phase 3.2) ---

func TestService_CheckAvailability_Success(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		CheckAvailabilityFunc: func(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
			result := make([]*wecom.UserAvailability, len(opts.UserIDs))
			for i, uid := range opts.UserIDs {
				result[i] = &wecom.UserAvailability{
					UserID: uid,
					Slots:  []wecom.AvailabilitySlot{},
				}
			}
			return result, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CheckAvailabilityRequest{
		UserIDs:   []string{"user1", "user2"},
		StartTime: time.Now().Add(1 * time.Hour).Unix(),
		EndTime:   time.Now().Add(2 * time.Hour).Unix(),
	}

	result, err := svc.CheckAvailability(context.Background(), authCtx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestService_CheckAvailability_InvalidTimeRange(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForSchedule{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CheckAvailabilityRequest{
		UserIDs:   []string{"user1"},
		StartTime: time.Now().Add(2 * time.Hour).Unix(),
		EndTime:   time.Now().Add(1 * time.Hour).Unix(),
	}

	_, err := svc.CheckAvailability(context.Background(), authCtx, req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestService_CheckAvailability_APIError(t *testing.T) {
	mockClient := &MockWeComClientForSchedule{
		CheckAvailabilityFunc: func(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
			return nil, errors.New("check failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	req := &CheckAvailabilityRequest{
		UserIDs:   []string{"user1"},
		StartTime: time.Now().Add(1 * time.Hour).Unix(),
		EndTime:   time.Now().Add(2 * time.Hour).Unix(),
	}

	_, err := svc.CheckAvailability(context.Background(), authCtx, req)
	if err == nil {
		t.Error("expected error from API")
	}
}

// MockWeComClientForSchedule for testing schedule service
type MockWeComClientForSchedule struct {
	CreateScheduleFunc  func(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error)
	GetSchedulesFunc    func(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error)
	UpdateScheduleFunc  func(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error
	DeleteScheduleFunc  func(ctx context.Context, corpName, appName string, scheduleID string) error
	CheckAvailabilityFunc  func(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error)
}

func (m *MockWeComClientForSchedule) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	if m.CreateScheduleFunc != nil {
		return m.CreateScheduleFunc(ctx, corpName, appName, params)
	}
	return &wecom.Schedule{}, nil
}

func (m *MockWeComClientForSchedule) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	if m.GetSchedulesFunc != nil {
		return m.GetSchedulesFunc(ctx, corpName, appName, userID, startTime, endTime, limit)
	}
	return []*wecom.Schedule{}, nil
}

func (m *MockWeComClientForSchedule) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	if m.UpdateScheduleFunc != nil {
		return m.UpdateScheduleFunc(ctx, corpName, appName, scheduleID, params)
	}
	return nil
}

func (m *MockWeComClientForSchedule) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	if m.DeleteScheduleFunc != nil {
		return m.DeleteScheduleFunc(ctx, corpName, appName, scheduleID)
	}
	return nil
}

func (m *MockWeComClientForSchedule) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForSchedule) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForSchedule) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}

func (m *MockWeComClientForSchedule) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}

func (m *MockWeComClientForSchedule) CheckAvailability(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
	if m.CheckAvailabilityFunc != nil {
		return m.CheckAvailabilityFunc(ctx, corpName, appName, opts)
	}
	return []*wecom.UserAvailability{}, nil
}
func (m *MockWeComClientForSchedule) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) GetTodoList(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) CreateTodo(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
	return "", nil
}
func (m *MockWeComClientForSchedule) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
	return nil
}
func (m *MockWeComClientForSchedule) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	return nil
}
func (m *MockWeComClientForSchedule) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	return nil
}
func (m *MockWeComClientForSchedule) CreateMeeting(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	return nil
}
func (m *MockWeComClientForSchedule) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
	return nil
}
func (m *MockWeComClientForSchedule) ListMeetings(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	return nil, nil
}
func (m *MockWeComClientForSchedule) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForSchedule) GetBotMcpConfig(ctx context.Context, corpName, appName string) (*wecom.BotMcpConfig, error) { return nil, nil }
func (m *MockWeComClientForSchedule) CallBotMcp(ctx context.Context, corpName, appName string, bizType string, request interface{}) ([]byte, error) { return nil, nil }

