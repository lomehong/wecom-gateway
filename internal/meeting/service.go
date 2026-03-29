package meeting

import (
	"context"
	"fmt"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Service handles meeting room business logic
type Service struct {
	wecomClient wecom.Client
}

// NewService creates a new meeting service
func NewService(wecomClient wecom.Client) *Service {
	return &Service{wecomClient: wecomClient}
}

// ListMeetingRooms lists available meeting rooms
func (s *Service) ListMeetingRooms(ctx context.Context, authCtx *auth.AuthContext, city, building, floor string, capacity int, equipment []string, limit int) ([]*wecom.MeetingRoom, string, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	opts := &wecom.RoomQueryOptions{
		City:      city,
		Building:  building,
		Floor:     floor,
		Capacity:  capacity,
		Equipment: equipment,
		Limit:     limit,
	}

	rooms, cursor, err := s.wecomClient.ListMeetingRooms(ctx, corpName, appName, opts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list meeting rooms: %w", err)
	}

	return rooms, cursor, nil
}

// GetRoomAvailability gets available time slots for a meeting room
func (s *Service) GetRoomAvailability(ctx context.Context, authCtx *auth.AuthContext, roomID string, startTime, endTime time.Time) ([]*wecom.TimeSlot, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	slots, err := s.wecomClient.GetRoomAvailability(ctx, corpName, appName, roomID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get room availability: %w", err)
	}

	return slots, nil
}

// BookMeetingRoom books a meeting room
func (s *Service) BookMeetingRoom(ctx context.Context, authCtx *auth.AuthContext, roomID, subject string, startTime, endTime time.Time, booker string, attendees []string) (*wecom.BookingResult, error) {
	// Validate booking time
	if startTime.Before(time.Now()) {
		return nil, fmt.Errorf("start_time cannot be in the past")
	}

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end_time must be after start_time")
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.BookingParams{
		MeetingRoomID: roomID,
		Subject:       subject,
		StartTime:     startTime,
		EndTime:       endTime,
		Booker:        booker,
		Attendees:     attendees,
	}

	result, err := s.wecomClient.BookMeetingRoom(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to book meeting room: %w", err)
	}

	return result, nil
}
