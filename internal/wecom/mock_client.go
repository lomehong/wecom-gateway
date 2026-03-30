package wecom

import (
	"context"
	"time"
)

// MockClient is a mock implementation of WeCom Client for testing
type MockClient struct {
	// Add fields to track calls if needed
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) CreateSchedule(ctx context.Context, corpName, appName string, params *ScheduleParams) (*Schedule, error) {
	return &Schedule{
		ScheduleID: "test-schedule-id",
	}, nil
}

func (m *MockClient) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*Schedule, error) {
	return []*Schedule{
		{ScheduleID: "schedule-1"},
		{ScheduleID: "schedule-2"},
	}, nil
}

func (m *MockClient) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *ScheduleParams) error {
	return nil
}

func (m *MockClient) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}

func (m *MockClient) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *RoomQueryOptions) ([]*MeetingRoom, string, error) {
	rooms := []*MeetingRoom{
		{MBookingID: "room-1", Name: "Room 101", Capacity: 10},
		{MBookingID: "room-2", Name: "Room 102", Capacity: 20},
	}
	return rooms, "", nil
}

func (m *MockClient) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*TimeSlot, error) {
	return []*TimeSlot{
		{StartTime: start, EndTime: end},
	}, nil
}

func (m *MockClient) BookMeetingRoom(ctx context.Context, corpName, appName string, params *BookingParams) (*BookingResult, error) {
	return &BookingResult{
		BookingID: "booking-123",
	}, nil
}

func (m *MockClient) SendText(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error) {
	return &SendResult{
		InvalidUserIDs:  []string{},
		InvalidPartyIDs: []string{},
		InvalidTagIDs:   []string{},
	}, nil
}

func (m *MockClient) SendMarkdown(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error) {
	return &SendResult{
		InvalidUserIDs:  []string{},
		InvalidPartyIDs: []string{},
		InvalidTagIDs:   []string{},
	}, nil
}

func (m *MockClient) SendImage(ctx context.Context, corpName, appName string, params *ImageMessageParams) (*SendResult, error) {
	return &SendResult{
		InvalidUserIDs:  []string{},
		InvalidPartyIDs: []string{},
		InvalidTagIDs:   []string{},
	}, nil
}

func (m *MockClient) SendFile(ctx context.Context, corpName, appName string, params *FileMessageParams) (*SendResult, error) {
	return &SendResult{
		InvalidUserIDs:  []string{},
		InvalidPartyIDs: []string{},
		InvalidTagIDs:   []string{},
	}, nil
}

func (m *MockClient) SendCard(ctx context.Context, corpName, appName string, params *CardMessageParams) (*SendResult, error) {
	return &SendResult{
		InvalidUserIDs:  []string{},
		InvalidPartyIDs: []string{},
		InvalidTagIDs:   []string{},
	}, nil
}

func (m *MockClient) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "media-id-123", nil
}

func (m *MockClient) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return "mock-access-token", nil
}

// Contact operations

func (m *MockClient) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*ContactUser, error) {
	return []*ContactUser{
		{UserID: "user1", Name: "张三"},
		{UserID: "user2", Name: "李四"},
	}, nil
}

func (m *MockClient) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*ContactUser, error) {
	return []*ContactUser{
		{UserID: "user1", Name: "张三"},
	}, nil
}

// Todo operations

func (m *MockClient) GetTodoList(ctx context.Context, corpName, appName string, opts *TodoListOptions) (*TodoListResult, error) {
	return &TodoListResult{
		IndexList: []TodoIndex{
			{TodoID: "todo-1", CreatorID: "user1"},
			{TodoID: "todo-2", CreatorID: "user2"},
		},
		HasMore: false,
	}, nil
}

func (m *MockClient) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*TodoDetail, error) {
	details := make([]*TodoDetail, len(todoIDs))
	for i, id := range todoIDs {
		details[i] = &TodoDetail{TodoIndex: TodoIndex{TodoID: id}, Content: "Mock todo content"}
	}
	return details, nil
}

func (m *MockClient) CreateTodo(ctx context.Context, corpName, appName string, params *CreateTodoParams) (string, error) {
	return "todo-new-1", nil
}

func (m *MockClient) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *UpdateTodoParams) error {
	return nil
}

func (m *MockClient) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	return nil
}

func (m *MockClient) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	return nil
}

// Meeting appointment operations (Phase 1.3)

func (m *MockClient) CreateMeeting(ctx context.Context, corpName, appName string, params *CreateMeetingParams) (*MeetingInfo, error) {
	return &MeetingInfo{
		MeetingID:     "meeting-123",
		Title:         params.Title,
		Status:        0,
		StartDateTime: params.StartDateTime,
		Duration:      params.Duration,
		Creator:       "creator-user",
	}, nil
}

func (m *MockClient) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	return nil
}

func (m *MockClient) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *MeetingInvitees) error {
	return nil
}

func (m *MockClient) ListMeetings(ctx context.Context, corpName, appName string, opts *MeetingListOptions) (*MeetingListResult, error) {
	return &MeetingListResult{
		Meetings: []MeetingInfo{
			{MeetingID: "meeting-1", Title: "Test Meeting 1"},
			{MeetingID: "meeting-2", Title: "Test Meeting 2"},
		},
		NextCursor: "",
	}, nil
}

func (m *MockClient) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*MeetingInfo, error) {
	return &MeetingInfo{
		MeetingID:   meetingID,
		Title:       "Test Meeting",
		Status:      0,
		Creator:     "creator-user",
		MeetingLink: "https://meeting.qq.com/dm/r/test",
	}, nil
}

// Message pull operations (Phase 3.1)

func (m *MockClient) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*ChatListResult, error) {
	return &ChatListResult{
		ChatList: []ChatInfo{
			{ChatID: "chat-1", ChatType: 1, Name: "Group A"},
			{ChatID: "chat-2", ChatType: 1, Name: "Group B"},
		},
	}, nil
}

func (m *MockClient) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*ChatMessagesResult, error) {
	return &ChatMessagesResult{
		MsgList: []ChatMessage{
			{MsgID: "msg-1", MsgType: "text", From: "user1"},
			{MsgID: "msg-2", MsgType: "text", From: "user2"},
		},
	}, nil
}

func (m *MockClient) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	return []byte("mock-media-content"), "mock-file.txt", nil
}

// Schedule availability operations (Phase 3.2)

func (m *MockClient) CheckAvailability(ctx context.Context, corpName, appName string, opts *AvailabilityOptions) ([]*UserAvailability, error) {
	result := make([]*UserAvailability, len(opts.UserIDs))
	for i, uid := range opts.UserIDs {
		result[i] = &UserAvailability{
			UserID: uid,
			Slots:  []AvailabilitySlot{},
		}
	}
	return result, nil
}
