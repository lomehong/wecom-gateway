package todo

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func TestNewService_Todo(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestService_GetTodoList_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		GetTodoListFunc: func(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
			return &wecom.TodoListResult{
				IndexList: []wecom.TodoIndex{
					{TodoID: "todo-1", TodoStatus: 1, CreatorID: "user1"},
					{TodoID: "todo-2", TodoStatus: 0, CreatorID: "user2"},
				},
				NextCursor: "",
				HasMore:    false,
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.GetTodoList(context.Background(), authCtx, &GetTodoListRequest{Limit: 50})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.IndexList) != 2 {
		t.Errorf("expected 2 todos, got %d", len(result.IndexList))
	}
}

func TestService_GetTodoList_DefaultLimit(t *testing.T) {
	var receivedLimit int
	mockClient := &MockWeComClientForTodo{
		GetTodoListFunc: func(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
			receivedLimit = opts.Limit
			return &wecom.TodoListResult{}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetTodoList(context.Background(), authCtx, &GetTodoListRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if receivedLimit != 100 {
		t.Errorf("expected default limit 100, got %d", receivedLimit)
	}
}

func TestService_GetTodoList_Error(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		GetTodoListFunc: func(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
			return nil, errors.New("api error")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetTodoList(context.Background(), authCtx, &GetTodoListRequest{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_GetTodoDetail_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		GetTodoDetailFunc: func(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
			return []*wecom.TodoDetail{
				{
					TodoIndex: wecom.TodoIndex{TodoID: "todo-1", TodoStatus: 1, CreatorID: "user1"},
					Content:   "Test content",
					Assignees: []string{"user1"},
				},
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	details, err := svc.GetTodoDetail(context.Background(), authCtx, "todo-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(details) != 1 {
		t.Errorf("expected 1 detail, got %d", len(details))
	}
	if details[0].Content != "Test content" {
		t.Errorf("expected content 'Test content', got %s", details[0].Content)
	}
}

func TestService_GetTodoDetail_EmptyID(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetTodoDetail(context.Background(), authCtx, "")
	if err == nil {
		t.Error("expected error for empty id")
	}
}

func TestService_CreateTodo_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		CreateTodoFunc: func(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
			return "new-todo-id", nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	todoID, err := svc.CreateTodo(context.Background(), authCtx, &CreateTodoRequest{
		Content: "Test todo",
		Title:   "Test Title",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if todoID != "new-todo-id" {
		t.Errorf("expected todo id 'new-todo-id', got %s", todoID)
	}
}

func TestService_CreateTodo_EmptyContent(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.CreateTodo(context.Background(), authCtx, &CreateTodoRequest{})
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestService_UpdateTodo_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		UpdateTodoFunc: func(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	newContent := "Updated content"
	err := svc.UpdateTodo(context.Background(), authCtx, "todo-1", &UpdateTodoRequest{
		Content: &newContent,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_UpdateTodo_EmptyID(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.UpdateTodo(context.Background(), authCtx, "", &UpdateTodoRequest{})
	if err == nil {
		t.Error("expected error for empty id")
	}
}

func TestService_DeleteTodo_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		DeleteTodoFunc: func(ctx context.Context, corpName, appName string, todoID string) error {
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.DeleteTodo(context.Background(), authCtx, "todo-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_DeleteTodo_EmptyID(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.DeleteTodo(context.Background(), authCtx, "")
	if err == nil {
		t.Error("expected error for empty id")
	}
}

func TestService_DeleteTodo_Error(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		DeleteTodoFunc: func(ctx context.Context, corpName, appName string, todoID string) error {
			return errors.New("delete failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.DeleteTodo(context.Background(), authCtx, "todo-1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_ChangeUserStatus_Success(t *testing.T) {
	mockClient := &MockWeComClientForTodo{
		ChangeTodoUserStatusFunc: func(ctx context.Context, corpName, appName string, todoID string, status int) error {
			return nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.ChangeUserStatus(context.Background(), authCtx, "todo-1", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_ChangeUserStatus_InvalidStatus(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.ChangeUserStatus(context.Background(), authCtx, "todo-1", 99)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_ChangeUserStatus_EmptyID(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForTodo{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	err := svc.ChangeUserStatus(context.Background(), authCtx, "", 1)
	if err == nil {
		t.Error("expected error for empty id")
	}
}

// MockWeComClientForTodo is a mock for testing todo service
type MockWeComClientForTodo struct {
	GetTodoListFunc         func(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error)
	GetTodoDetailFunc       func(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error)
	CreateTodoFunc          func(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error)
	UpdateTodoFunc          func(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error
	DeleteTodoFunc          func(ctx context.Context, corpName, appName string, todoID string) error
	ChangeTodoUserStatusFunc func(ctx context.Context, corpName, appName string, todoID string, status int) error
}

func (m *MockWeComClientForTodo) GetTodoList(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
	if m.GetTodoListFunc != nil {
		return m.GetTodoListFunc(ctx, corpName, appName, opts)
	}
	return &wecom.TodoListResult{}, nil
}

func (m *MockWeComClientForTodo) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
	if m.GetTodoDetailFunc != nil {
		return m.GetTodoDetailFunc(ctx, corpName, appName, todoIDs)
	}
	return []*wecom.TodoDetail{}, nil
}

func (m *MockWeComClientForTodo) CreateTodo(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
	if m.CreateTodoFunc != nil {
		return m.CreateTodoFunc(ctx, corpName, appName, params)
	}
	return "mock-todo-id", nil
}

func (m *MockWeComClientForTodo) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
	if m.UpdateTodoFunc != nil {
		return m.UpdateTodoFunc(ctx, corpName, appName, todoID, params)
	}
	return nil
}

func (m *MockWeComClientForTodo) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	if m.DeleteTodoFunc != nil {
		return m.DeleteTodoFunc(ctx, corpName, appName, todoID)
	}
	return nil
}

func (m *MockWeComClientForTodo) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	if m.ChangeTodoUserStatusFunc != nil {
		return m.ChangeTodoUserStatusFunc(ctx, corpName, appName, todoID, status)
	}
	return nil
}

func (m *MockWeComClientForTodo) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	return nil
}

func (m *MockWeComClientForTodo) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}

func (m *MockWeComClientForTodo) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForTodo) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}

func (m *MockWeComClientForTodo) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}

func (m *MockWeComClientForTodo) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) CreateMeeting(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	return nil
}

func (m *MockWeComClientForTodo) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
	return nil
}

func (m *MockWeComClientForTodo) ListMeetings(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	return nil, nil
}

func (m *MockWeComClientForTodo) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForTodo) CheckAvailability(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
	return nil, nil
}
