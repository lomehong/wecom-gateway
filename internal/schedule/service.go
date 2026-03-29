package schedule

import (
	"context"
	"fmt"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Service handles schedule business logic
type Service struct {
	wecomClient wecom.Client
}

// NewService creates a new schedule service
func NewService(wecomClient wecom.Client) *Service {
	return &Service{
		wecomClient: wecomClient,
	}
}

// CreateScheduleRequest represents a request to create a schedule
type CreateScheduleRequest struct {
	Organizer       string    `json:"organizer" binding:"required"`
	Summary         string    `json:"summary" binding:"required"`
	Description     string    `json:"description,omitempty"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	EndTime         time.Time `json:"end_time" binding:"required"`
	Attendees       []string  `json:"attendees,omitempty"`
	Location        string    `json:"location,omitempty"`
	RemindBeforeMin int       `json:"remind_before_minutes,omitempty"`
}

// CreateSchedule creates a new schedule
func (s *Service) CreateSchedule(ctx context.Context, authCtx *auth.AuthContext, req *CreateScheduleRequest) (*wecom.Schedule, error) {
	// Validate request
	if req.EndTime.Before(req.StartTime) {
		return nil, fmt.Errorf("end_time must be after start_time")
	}

	// Set default reminder
	if req.RemindBeforeMin == 0 {
		req.RemindBeforeMin = 15
	}

	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	// Create schedule through WeCom client
	params := &wecom.ScheduleParams{
		Organizer:       req.Organizer,
		Summary:         req.Summary,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Attendees:       req.Attendees,
		Location:        req.Location,
		RemindBeforeMin: req.RemindBeforeMin,
	}

	schedule, err := s.wecomClient.CreateSchedule(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return schedule, nil
}

// GetSchedulesRequest represents a request to get schedules
type GetSchedulesRequest struct {
	UserID    string    `form:"userid" binding:"required"`
	StartTime time.Time `form:"start_time"`
	EndTime   time.Time `form:"end_time"`
	Limit     int       `form:"limit" binding:"omitempty,min=1,max=100"`
}

// GetSchedules retrieves schedules for a user
func (s *Service) GetSchedules(ctx context.Context, authCtx *auth.AuthContext, req *GetSchedulesRequest) ([]*wecom.Schedule, error) {
	// Set default time range if not provided
	startTime := req.StartTime
	if startTime.IsZero() {
		startTime = time.Now().Truncate(24 * time.Hour)
	}

	endTime := req.EndTime
	if endTime.IsZero() {
		endTime = startTime.AddDate(0, 0, 7) // Default 7 days
	}

	// Set default limit
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	schedules, err := s.wecomClient.GetSchedules(ctx, corpName, appName, req.UserID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	return schedules, nil
}

// UpdateScheduleRequest represents a request to update a schedule
type UpdateScheduleRequest struct {
	Summary         string    `json:"summary,omitempty"`
	Description     string    `json:"description,omitempty"`
	StartTime       time.Time `json:"start_time,omitempty"`
	EndTime         time.Time `json:"end_time,omitempty"`
	Attendees       []string  `json:"attendees,omitempty"`
	Location        string    `json:"location,omitempty"`
	RemindBeforeMin int       `json:"remind_before_minutes,omitempty"`
}

// UpdateSchedule updates an existing schedule
func (s *Service) UpdateSchedule(ctx context.Context, authCtx *auth.AuthContext, scheduleID string, req *UpdateScheduleRequest) error {
	// Validate time range if provided
	if !req.StartTime.IsZero() && !req.EndTime.IsZero() && req.EndTime.Before(req.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	// Build params
	params := &wecom.ScheduleParams{
		Summary:         req.Summary,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Attendees:       req.Attendees,
		Location:        req.Location,
		RemindBeforeMin: req.RemindBeforeMin,
	}

	err := s.wecomClient.UpdateSchedule(ctx, corpName, appName, scheduleID, params)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}

// DeleteSchedule deletes a schedule
func (s *Service) DeleteSchedule(ctx context.Context, authCtx *auth.AuthContext, scheduleID string) error {
	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	err := s.wecomClient.DeleteSchedule(ctx, corpName, appName, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}
