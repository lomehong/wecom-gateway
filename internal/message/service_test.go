package message

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// MockWeComClient for testing message service
type MockWeComClient struct {
	SendTextFunc      func(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error)
	SendMarkdownFunc  func(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error)
	SendImageFunc     func(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error)
	SendFileFunc      func(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error)
	SendCardFunc      func(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error)
	GetChatListFunc   func(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error)
	GetChatMessagesFunc func(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error)
	DownloadMediaFunc func(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error)
}

func (m *MockWeComClient) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClient) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClient) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	return nil
}

func (m *MockWeComClient) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}

func (m *MockWeComClient) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}

func (m *MockWeComClient) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}

func (m *MockWeComClient) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}

func (m *MockWeComClient) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	if m.SendTextFunc != nil {
		return m.SendTextFunc(ctx, corpName, appName, params)
	}
	return &wecom.SendResult{}, nil
}

func (m *MockWeComClient) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	if m.SendMarkdownFunc != nil {
		return m.SendMarkdownFunc(ctx, corpName, appName, params)
	}
	return &wecom.SendResult{}, nil
}

func (m *MockWeComClient) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	if m.SendImageFunc != nil {
		return m.SendImageFunc(ctx, corpName, appName, params)
	}
	return &wecom.SendResult{}, nil
}

func (m *MockWeComClient) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	if m.SendFileFunc != nil {
		return m.SendFileFunc(ctx, corpName, appName, params)
	}
	return &wecom.SendResult{}, nil
}

func (m *MockWeComClient) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	if m.SendCardFunc != nil {
		return m.SendCardFunc(ctx, corpName, appName, params)
	}
	return &wecom.SendResult{}, nil
}

func (m *MockWeComClient) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}

func TestNewService(t *testing.T) {
	mockClient := &MockWeComClient{}
	svc := NewService(mockClient)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	// Cannot access private field wecomClient directly
	// Test by calling methods instead
}

func TestService_SendText(t *testing.T) {
	mockClient := &MockWeComClient{
		SendTextFunc: func(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
			if params.Content != "test message" {
				t.Errorf("expected content 'test message', got %s", params.Content)
			}
			if params.Safe != true {
				t.Error("expected safe to be true")
			}
			return &wecom.SendResult{}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.SendText(context.Background(), authCtx, "user", []string{"user1"}, "test message", true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestService_SendText_Error(t *testing.T) {
	mockClient := &MockWeComClient{
		SendTextFunc: func(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
			return nil, errors.New("send failed")
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.SendText(context.Background(), authCtx, "user", []string{"user1"}, "test message", false)
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_SendMarkdown(t *testing.T) {
	mockClient := &MockWeComClient{
		SendMarkdownFunc: func(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
			if params.Content != "**test**" {
				t.Errorf("expected content '**test**', got %s", params.Content)
			}
			return &wecom.SendResult{}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.SendMarkdown(context.Background(), authCtx, "user", []string{"user1"}, "**test**")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestService_SendImage(t *testing.T) {
	mockClient := &MockWeComClient{
		SendImageFunc: func(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
			if params.MediaID != "test-media-id" {
				t.Errorf("expected media_id 'test-media-id', got %s", params.MediaID)
			}
			return &wecom.SendResult{}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.SendImage(context.Background(), authCtx, "user", []string{"user1"}, "test-media-id", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestService_SendFile(t *testing.T) {
	mockClient := &MockWeComClient{
		SendFileFunc: func(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
			if params.MediaID != "test-file-id" {
				t.Errorf("expected media_id 'test-file-id', got %s", params.MediaID)
			}
			return &wecom.SendResult{}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.SendFile(context.Background(), authCtx, "user", []string{"user1"}, "test-file-id", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestService_SendCard(t *testing.T) {
	mockClient := &MockWeComClient{
		SendCardFunc: func(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
			if params.CardContent == nil {
				t.Error("expected card content")
			}
			return &wecom.SendResult{}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	cardContent := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
	}

	result, err := svc.SendCard(context.Background(), authCtx, "user", []string{"user1"}, cardContent)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result, got nil")
	}
}

// --- Message Pull Service Tests (Phase 3.1) ---

func TestService_GetChatList_Success(t *testing.T) {
	mockClient := &MockWeComClient{
		GetChatListFunc: func(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
			if beginTime != 1704067200 {
				t.Errorf("expected begin_time 1704067200, got %d", beginTime)
			}
			return &wecom.ChatListResult{
				ChatList: []wecom.ChatInfo{
					{ChatID: "chat-1", ChatType: 1, Name: "Group A"},
					{ChatID: "chat-2", ChatType: 1, Name: "Group B"},
				},
			}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.GetChatList(context.Background(), authCtx, 1704067200, 1704153600)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result.ChatList) != 2 {
		t.Errorf("expected 2 chats, got %d", len(result.ChatList))
	}
}

func TestService_GetChatList_Error(t *testing.T) {
	mockClient := &MockWeComClient{
		GetChatListFunc: func(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
			return nil, errors.New("get chat list failed")
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetChatList(context.Background(), authCtx, 1704067200, 1704153600)
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_GetChatMessages_Success(t *testing.T) {
	mockClient := &MockWeComClient{
		GetChatMessagesFunc: func(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
			if chatID != "chat-1" {
				t.Errorf("expected chat ID chat-1, got %s", chatID)
			}
			return &wecom.ChatMessagesResult{
				MsgList: []wecom.ChatMessage{
					{MsgID: "msg-1", MsgType: "text", From: "user1"},
					{MsgID: "msg-2", MsgType: "text", From: "user2"},
				},
			}, nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.GetChatMessages(context.Background(), authCtx, 1, "chat-1", 1704067200, 1704153600)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result.MsgList) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result.MsgList))
	}
}

func TestService_GetChatMessages_Error(t *testing.T) {
	mockClient := &MockWeComClient{
		GetChatMessagesFunc: func(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
			return nil, errors.New("get messages failed")
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetChatMessages(context.Background(), authCtx, 1, "chat-1", 1704067200, 1704153600)
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_DownloadMedia_Success(t *testing.T) {
	mockClient := &MockWeComClient{
		DownloadMediaFunc: func(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
			if mediaID != "media-123" {
				t.Errorf("expected media ID media-123, got %s", mediaID)
			}
			return []byte("media-content"), "file.txt", nil
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	data, filename, err := svc.DownloadMedia(context.Background(), authCtx, "media-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(data) != "media-content" {
		t.Errorf("expected content 'media-content', got %s", string(data))
	}
	if filename != "file.txt" {
		t.Errorf("expected filename 'file.txt', got %s", filename)
	}
}

func TestService_DownloadMedia_Error(t *testing.T) {
	mockClient := &MockWeComClient{
		DownloadMediaFunc: func(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
			return nil, "", errors.New("download failed")
		},
	}
	svc := NewService(mockClient)

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, _, err := svc.DownloadMedia(context.Background(), authCtx, "media-123")
	if err == nil {
		t.Error("expected error")
	}
}

func (m *MockWeComClient) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}

func (m *MockWeComClient) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
	return nil, nil
}

func (m *MockWeComClient) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
	return nil, nil
}

func (m *MockWeComClient) GetTodoList(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
	return nil, nil
}

func (m *MockWeComClient) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
	return nil, nil
}

func (m *MockWeComClient) CreateTodo(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
	return "", nil
}

func (m *MockWeComClient) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
	return nil
}

func (m *MockWeComClient) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	return nil
}

func (m *MockWeComClient) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	return nil
}

func (m *MockWeComClient) CreateMeeting(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClient) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	return nil
}

func (m *MockWeComClient) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
	return nil
}

func (m *MockWeComClient) ListMeetings(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
	return nil, nil
}

func (m *MockWeComClient) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClient) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	if m.GetChatListFunc != nil {
		return m.GetChatListFunc(ctx, corpName, appName, beginTime, endTime)
	}
	return &wecom.ChatListResult{
		ChatList: []wecom.ChatInfo{
			{ChatID: "chat-1", ChatType: 1, Name: "Group A"},
		},
	}, nil
}

func (m *MockWeComClient) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	if m.GetChatMessagesFunc != nil {
		return m.GetChatMessagesFunc(ctx, corpName, appName, chatType, chatID, beginTime, endTime)
	}
	return &wecom.ChatMessagesResult{
		MsgList: []wecom.ChatMessage{
			{MsgID: "msg-1", MsgType: "text", From: "user1"},
		},
	}, nil
}

func (m *MockWeComClient) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	if m.DownloadMediaFunc != nil {
		return m.DownloadMediaFunc(ctx, corpName, appName, mediaID)
	}
	return []byte("mock-media-content"), "mock-file.txt", nil
}

func (m *MockWeComClient) CheckAvailability(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
	return nil, nil
}
