package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func TestNewService_Contact(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestService_GetUserList_Success(t *testing.T) {
	mockClient := &MockWeComClientForContact{
		GetUserListFunc: func(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
			return []*wecom.ContactUser{
				{UserID: "user1", Name: "张三", Department: []int{1, 2}, Position: "工程师"},
				{UserID: "user2", Name: "李四", Department: []int{1}, Position: "经理"},
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.GetUserList(context.Background(), authCtx, &GetUserListRequest{DepartmentID: 1})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Count != 2 {
		t.Errorf("expected 2 users, got %d", result.Count)
	}
	if result.Users[0].Name != "张三" {
		t.Errorf("expected first user name 张三, got %s", result.Users[0].Name)
	}
}

func TestService_GetUserList_DefaultDepartmentID(t *testing.T) {
	var receivedDeptID int
	mockClient := &MockWeComClientForContact{
		GetUserListFunc: func(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
			receivedDeptID = departmentID
			return []*wecom.ContactUser{}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	// Request with no department ID — should default to 1
	_, err := svc.GetUserList(context.Background(), authCtx, &GetUserListRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if receivedDeptID != 1 {
		t.Errorf("expected default department ID 1, got %d", receivedDeptID)
	}
}

func TestService_GetUserList_Error(t *testing.T) {
	mockClient := &MockWeComClientForContact{
		GetUserListFunc: func(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
			return nil, errors.New("api error")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.GetUserList(context.Background(), authCtx, &GetUserListRequest{DepartmentID: 1})
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_SearchUser_Success(t *testing.T) {
	mockClient := &MockWeComClientForContact{
		SearchUserFunc: func(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
			return []*wecom.ContactUser{
				{UserID: "user1", Name: "张三", Department: []int{1}},
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	result, err := svc.SearchUser(context.Background(), authCtx, &SearchUserRequest{Query: "张三"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Count != 1 {
		t.Errorf("expected 1 user, got %d", result.Count)
	}
}

func TestService_SearchUser_EmptyQuery(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForContact{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.SearchUser(context.Background(), authCtx, &SearchUserRequest{Query: ""})
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestService_SearchUser_Error(t *testing.T) {
	mockClient := &MockWeComClientForContact{
		SearchUserFunc: func(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
			return nil, errors.New("search failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	_, err := svc.SearchUser(context.Background(), authCtx, &SearchUserRequest{Query: "test"})
	if err == nil {
		t.Error("expected error")
	}
}

// MockWeComClientForContact is a mock for testing contact service
type MockWeComClientForContact struct {
	GetUserListFunc func(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error)
	SearchUserFunc  func(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error)
}

func (m *MockWeComClientForContact) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*wecom.ContactUser, error) {
	if m.GetUserListFunc != nil {
		return m.GetUserListFunc(ctx, corpName, appName, departmentID)
	}
	return []*wecom.ContactUser{}, nil
}

func (m *MockWeComClientForContact) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*wecom.ContactUser, error) {
	if m.SearchUserFunc != nil {
		return m.SearchUserFunc(ctx, corpName, appName, query)
	}
	return []*wecom.ContactUser{}, nil
}

func (m *MockWeComClientForContact) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	return nil
}

func (m *MockWeComClientForContact) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}

func (m *MockWeComClientForContact) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForContact) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}

func (m *MockWeComClientForContact) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}

func (m *MockWeComClientForContact) GetTodoList(ctx context.Context, corpName, appName string, opts *wecom.TodoListOptions) (*wecom.TodoListResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*wecom.TodoDetail, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) CreateTodo(ctx context.Context, corpName, appName string, params *wecom.CreateTodoParams) (string, error) {
	return "", nil
}

func (m *MockWeComClientForContact) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *wecom.UpdateTodoParams) error {
	return nil
}

func (m *MockWeComClientForContact) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	return nil
}

func (m *MockWeComClientForContact) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	return nil
}

func (m *MockWeComClientForContact) CreateMeeting(ctx context.Context, corpName, appName string, params *wecom.CreateMeetingParams) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	return nil
}

func (m *MockWeComClientForContact) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *wecom.MeetingInvitees) error {
	return nil
}

func (m *MockWeComClientForContact) ListMeetings(ctx context.Context, corpName, appName string, opts *wecom.MeetingListOptions) (*wecom.MeetingListResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*wecom.MeetingInfo, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	return nil, nil
}

func (m *MockWeComClientForContact) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	return nil, "", nil
}

func (m *MockWeComClientForContact) CheckAvailability(ctx context.Context, corpName, appName string, opts *wecom.AvailabilityOptions) ([]*wecom.UserAvailability, error) {
	return nil, nil
}
