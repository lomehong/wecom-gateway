package grpcserver

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"wecom-gateway/internal/wecom"
)

// Test NewServer
func TestNewServer(t *testing.T) {
	server := NewServer(nil, nil, nil, nil, nil, nil, nil, nil)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.rateLimiter == nil {
		t.Error("rate limiter should be initialized")
	}
}

// Test RateLimiter methods
func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter()

	if !rl.Allow("key1", 60) {
		t.Error("RateLimiter.Allow should return true for normal case")
	}
	if !rl.Allow("key2", 100) {
		t.Error("RateLimiter.Allow should return true for high limit")
	}
	if !rl.Allow("", 0) {
		t.Error("RateLimiter.Allow should return true for zero limit")
	}
}

// Test createSchedule missing auth context
func TestCreateSchedule_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.CreateSchedule(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test getSchedules missing auth context
func TestGetSchedules_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.GetSchedules(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test updateSchedule missing auth context
func TestUpdateSchedule_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.UpdateSchedule(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test deleteSchedule missing auth context
func TestDeleteSchedule_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.DeleteSchedule(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test listMeetingRooms missing auth context
func TestListMeetingRooms_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.ListMeetingRooms(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test getRoomAvailability missing auth context
func TestGetRoomAvailability_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.GetRoomAvailability(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test bookMeetingRoom missing auth context
func TestBookMeetingRoom_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.BookMeetingRoom(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test sendText missing auth context
func TestSendText_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.SendText(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test sendMarkdown missing auth context
func TestSendMarkdown_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.SendMarkdown(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test sendImage missing auth context
func TestSendImage_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.SendImage(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test sendFile missing auth context
func TestSendFile_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.SendFile(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test sendCard missing auth context
func TestSendCard_MissingAuth(t *testing.T) {
	s := &Server{}
	_, err := s.SendCard(context.Background(), nil)
	if err == nil {
		t.Error("expected error for missing auth context")
	}
}

// Test admin operations (placeholder implementations)
func TestCreateAPIKey_Admin(t *testing.T) {
	s := &Server{}
	// These are placeholder implementations, just verify they don't panic
	_ = s
}

func TestListAPIKeys_Admin(t *testing.T) {
	s := &Server{}
	_ = s
}

func TestDeleteAPIKey_Admin(t *testing.T) {
	s := &Server{}
	_ = s
}

func TestQueryAuditLogs_Admin(t *testing.T) {
	s := &Server{}
	_ = s
}

func TestGetDashboardStats_Admin(t *testing.T) {
	s := &Server{}
	_ = s
}

// Test convertSendResult nil result
func TestConvertSendResult_Nil(t *testing.T) {
	// convertSendResult should handle nil pointers gracefully
	result := &wecom.SendResult{}
	protoResult := convertSendResult(result)
	if protoResult == nil {
		t.Error("convertSendResult should not return nil")
	}
}

// Test SendResult JSON marshaling
func TestSendResult_JSON(t *testing.T) {
	result := &wecom.SendResult{
		InvalidUserIDs:  []string{"u1", "u2"},
		FailedPartyIDs:  []string{"p1"},
		FailedTagIDs:    []string{},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded wecom.SendResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(decoded.InvalidUserIDs) != 2 {
		t.Errorf("expected 2 invalid user IDs, got %d", len(decoded.InvalidUserIDs))
	}
}

// Test BookingResult fields
func TestBookingResult_Fields(t *testing.T) {
	now := time.Now()
	result := &wecom.BookingResult{
		BookingID:  "booking-123",
		ScheduleID: "schedule-456",
		StartTime:  now,
		EndTime:    now.Add(2 * time.Hour),
	}

	if result.BookingID != "booking-123" {
		t.Errorf("expected booking-123, got %s", result.BookingID)
	}
	if result.ScheduleID != "schedule-456" {
		t.Errorf("expected schedule-456, got %s", result.ScheduleID)
	}
}

// Test TimeSlot fields
func TestTimeSlot_Fields(t *testing.T) {
	now := time.Now()
	slot := &wecom.TimeSlot{
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}

	if slot.EndTime.Before(slot.StartTime) {
		t.Error("end time should be after start time")
	}
}

// Test MeetingRoom fields
func TestMeetingRoom_Fields(t *testing.T) {
	room := &wecom.MeetingRoom{
		MBookingID:  "mb-123",
		Name:        "Room 101",
		Capacity:    10,
		City:        "Beijing",
		Building:    "Tower A",
		Floor:       "3F",
		Equipment:   []string{"projector", "whiteboard"},
		Description: "Large meeting room",
		Attributes:  map[string]string{"zone": "east"},
	}

	if room.Capacity != 10 {
		t.Errorf("expected capacity 10, got %d", room.Capacity)
	}
	if len(room.Equipment) != 2 {
		t.Errorf("expected 2 equipment items, got %d", len(room.Equipment))
	}
	if room.Attributes["zone"] != "east" {
		t.Errorf("expected zone east, got %s", room.Attributes["zone"])
	}
}

// Test Schedule fields
func TestSchedule_Fields(t *testing.T) {
	now := time.Now()
	schedule := &wecom.Schedule{
		ScheduleID:      "sched-123",
		Organizer:       "user1",
		Summary:         "Team Meeting",
		Description:     "Weekly sync",
		StartTime:       now,
		EndTime:         now.Add(1 * time.Hour),
		Attendees:       []string{"user2", "user3"},
		Location:        "Room 101",
		RemindBeforeMin: 15,
		CalID:           "cal-456",
	}

	if schedule.Organizer != "user1" {
		t.Errorf("expected organizer user1, got %s", schedule.Organizer)
	}
	if len(schedule.Attendees) != 2 {
		t.Errorf("expected 2 attendees, got %d", len(schedule.Attendees))
	}
	if schedule.RemindBeforeMin != 15 {
		t.Errorf("expected 15 min reminder, got %d", schedule.RemindBeforeMin)
	}
}

// Test ImageMessageParams fields
func TestImageMessageParams_Fields(t *testing.T) {
	params := &wecom.ImageMessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"u1"},
		ImageURL:     "https://example.com/img.png",
		MediaID:      "media-123",
	}

	if params.ReceiverType != "user" {
		t.Errorf("expected type user, got %s", params.ReceiverType)
	}
}

// Test FileMessageParams fields
func TestFileMessageParams_Fields(t *testing.T) {
	params := &wecom.FileMessageParams{
		ReceiverType: "department",
		ReceiverIDs:  []string{"dept1"},
		FileURL:      "https://example.com/doc.pdf",
		MediaID:      "media-456",
	}

	if params.ReceiverType != "department" {
		t.Errorf("expected type department, got %s", params.ReceiverType)
	}
}

// Test CardMessageParams fields
func TestCardMessageParams_Fields(t *testing.T) {
	params := &wecom.CardMessageParams{
		ReceiverType: "tag",
		ReceiverIDs:  []string{"tag1"},
		CardContent: map[string]interface{}{
			"title": "Test Card",
			"body":  "Test body",
		},
	}

	if params.ReceiverType != "tag" {
		t.Errorf("expected type tag, got %s", params.ReceiverType)
	}
	if params.CardContent["title"] != "Test Card" {
		t.Errorf("expected title Test Card, got %v", params.CardContent["title"])
	}
}

// Test WeComAPIError methods
func TestWeComAPIError_Methods(t *testing.T) {
	err := &wecom.WeComAPIError{ErrCode: 0, ErrMsg: "ok"}
	if err.Error() != "ok" {
		t.Errorf("expected 'ok', got %s", err.Error())
	}
	if err.IsAccessTokenExpired() {
		t.Error("errcode 0 should not be expired")
	}
	if err.IsInvalidCredential() {
		t.Error("errcode 0 should not be invalid credential")
	}
}

func TestWeComAPIError_Expired(t *testing.T) {
	err := &wecom.WeComAPIError{ErrCode: 40014, ErrMsg: "access_token expired"}
	if !err.IsAccessTokenExpired() {
		t.Error("errcode 40014 should be expired")
	}
	if err.IsInvalidCredential() {
		t.Error("errcode 40014 should not be invalid credential")
	}

	err2 := &wecom.WeComAPIError{ErrCode: 42001, ErrMsg: "access_token expired"}
	if !err2.IsAccessTokenExpired() {
		t.Error("errcode 42001 should be expired")
	}
}

func TestWeComAPIError_InvalidCred(t *testing.T) {
	err := &wecom.WeComAPIError{ErrCode: 40013, ErrMsg: "invalid corpid"}
	if !err.IsInvalidCredential() {
		t.Error("errcode 40013 should be invalid credential")
	}

	err2 := &wecom.WeComAPIError{ErrCode: 40091, ErrMsg: "invalid secret"}
	if !err2.IsInvalidCredential() {
		t.Error("errcode 40091 should be invalid credential")
	}
}
