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
