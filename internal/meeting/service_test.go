package meeting

import (
	"context"
	"errors"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func TestNewService_Meeting(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestService_ListMeetingRooms_Success(t *testing.T) {
	mockClient := &MockWeComClientForMeeting{
		ListMeetingRoomsFunc: func(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
			return []*wecom.MeetingRoom{
				{MBookingID: "room1", Name: "Room 101", Capacity: 10},
				{MBookingID: "room2", Name: "Room 102", Capacity: 20},
			}, "", nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	rooms, cursor, err := svc.ListMeetingRooms(context.Background(), authCtx, "Beijing", "Building A", "3", 10, []string{"projector"}, 50)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %s", cursor)
	}
}

func TestService_GetRoomAvailability_Success(t *testing.T) {
	mockClient := &MockWeComClientForMeeting{
		GetRoomAvailabilityFunc: func(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
			return []*wecom.TimeSlot{
				{StartTime: start, EndTime: start.Add(1 * time.Hour)},
				{StartTime: start.Add(2 * time.Hour), EndTime: start.Add(3 * time.Hour)},
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(1 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour)

	slots, err := svc.GetRoomAvailability(context.Background(), authCtx, "room-123", startTime, endTime)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}
}

func TestService_BookMeetingRoom_Success(t *testing.T) {
	mockClient := &MockWeComClientForMeeting{
		BookMeetingRoomFunc: func(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
			return &wecom.BookingResult{
				BookingID:   "booking-123",
				ScheduleID:  "schedule-123",
				StartTime:   params.StartTime,
				EndTime:     params.EndTime,
			}, nil
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(1 * time.Hour)
	endTime := time.Now().Add(2 * time.Hour)

	result, err := svc.BookMeetingRoom(context.Background(), authCtx, "room-123", "Team Meeting", startTime, endTime, "test-user", []string{"user1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.BookingID != "booking-123" {
		t.Errorf("expected booking ID booking-123, got %s", result.BookingID)
	}
}

func TestService_BookMeetingRoom_PastTime(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForMeeting{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(-1 * time.Hour) // Past
	endTime := time.Now().Add(1 * time.Hour)

	_, err := svc.BookMeetingRoom(context.Background(), authCtx, "room-123", "Team Meeting", startTime, endTime, "test-user", []string{})
	if err == nil {
		t.Error("expected error for past start time")
	}
}

func TestService_BookMeetingRoom_InvalidTimeRange(t *testing.T) {
	svc := &Service{wecomClient: &MockWeComClientForMeeting{}}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(2 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour) // End before start

	_, err := svc.BookMeetingRoom(context.Background(), authCtx, "room-123", "Team Meeting", startTime, endTime, "test-user", []string{})
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestService_BookMeetingRoom_APIError(t *testing.T) {
	mockClient := &MockWeComClientForMeeting{
		BookMeetingRoomFunc: func(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
			return nil, errors.New("booking failed")
		},
	}
	svc := &Service{wecomClient: mockClient}

	authCtx := &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	}

	startTime := time.Now().Add(1 * time.Hour)
	endTime := time.Now().Add(2 * time.Hour)

	_, err := svc.BookMeetingRoom(context.Background(), authCtx, "room-123", "Team Meeting", startTime, endTime, "test-user", []string{})
	if err == nil {
		t.Error("expected error from API")
	}
}

// MockWeComClientForMeeting for testing meeting service
type MockWeComClientForMeeting struct {
	ListMeetingRoomsFunc   func(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error)
	GetRoomAvailabilityFunc func(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error)
	BookMeetingRoomFunc     func(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error)
}

func (m *MockWeComClientForMeeting) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error {
	return nil
}

func (m *MockWeComClientForMeeting) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	return nil
}

func (m *MockWeComClientForMeeting) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	if m.ListMeetingRoomsFunc != nil {
		return m.ListMeetingRoomsFunc(ctx, corpName, appName, opts)
	}
	return []*wecom.MeetingRoom{}, "", nil
}

func (m *MockWeComClientForMeeting) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	if m.GetRoomAvailabilityFunc != nil {
		return m.GetRoomAvailabilityFunc(ctx, corpName, appName, roomID, start, end)
	}
	return []*wecom.TimeSlot{}, nil
}

func (m *MockWeComClientForMeeting) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	if m.BookMeetingRoomFunc != nil {
		return m.BookMeetingRoomFunc(ctx, corpName, appName, params)
	}
	return &wecom.BookingResult{}, nil
}

func (m *MockWeComClientForMeeting) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}

func (m *MockWeComClientForMeeting) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "", nil
}
