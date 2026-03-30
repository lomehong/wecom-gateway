package meeting

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
	"wecom-gateway/internal/wecom"
)

// CreateMeetingRequest represents a request to create a meeting
type CreateMeetingRequest struct {
	Title         string                  `json:"title" binding:"required"`
	StartDateTime string                  `json:"meeting_start_datetime" binding:"required"`
	Duration      int                     `json:"meeting_duration" binding:"required,min=60"`
	Invitees      *CreateMeetingInvitees  `json:"invitees,omitempty"`
	MeetingType   int                     `json:"meeting_type,omitempty"`
	Settings      *CreateMeetingSettings  `json:"settings,omitempty"`
}

// CreateMeetingInvitees represents invitees in create meeting request
type CreateMeetingInvitees struct {
	UserIDs []string `json:"userid,omitempty"`
	DeptIDs []string `json:"department,omitempty"`
}

// CreateMeetingSettings represents settings in create meeting request
type CreateMeetingSettings struct {
	MuteUponEntry   bool `json:"mute_upon_entry,omitempty"`
	WaitingRoom     bool `json:"waiting_room,omitempty"`
	EnableRecording bool `json:"enable_recording,omitempty"`
}

// UpdateInviteesRequest represents a request to update meeting invitees
type UpdateInviteesRequest struct {
	UserIDs []string `json:"userid,omitempty"`
	DeptIDs []string `json:"department,omitempty"`
}

// ListMeetingsRequest represents query parameters for listing meetings
type ListMeetingsRequest struct {
	BeginDatetime string `form:"begin_datetime" binding:"required"`
	EndDatetime   string `form:"end_datetime" binding:"required"`
	Limit         int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Cursor        string `form:"cursor"`
}

// --- Appointment Handler Methods ---

// CreateMeeting handles POST /v1/meetings
// @Summary Create meeting
// @Description Create a new meeting appointment
// @Tags meetings
// @Accept json
// @Produce json
// @Param request body CreateMeetingRequest true "Meeting parameters"
// @Success 201 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meetings [post]
func (h *Handler) CreateMeeting(c *gin.Context) {
	var req CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartDateTime)
	if err != nil {
		httputil.BadRequest(c, "invalid meeting_start_datetime format, use RFC3339")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.CreateMeeting(c.Request.Context(), authCtx, req.Title, startTime, req.Duration, req.Invitees, req.MeetingType, req.Settings)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, result)
}

// CancelMeeting handles DELETE /v1/meetings/:id
// @Summary Cancel meeting
// @Description Cancel a meeting
// @Tags meetings
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meetings/{id} [delete]
func (h *Handler) CancelMeeting(c *gin.Context) {
	meetingID := c.Param("id")
	if meetingID == "" {
		httputil.BadRequest(c, "meeting id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.CancelMeeting(c.Request.Context(), authCtx, meetingID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":    "meeting cancelled successfully",
		"meeting_id": meetingID,
	})
}

// UpdateInvitees handles PUT /v1/meetings/:id/invitees
// @Summary Update meeting invitees
// @Description Update the invitees of a meeting
// @Tags meetings
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Param request body UpdateInviteesRequest true "Invitees"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meetings/{id}/invitees [put]
func (h *Handler) UpdateInvitees(c *gin.Context) {
	meetingID := c.Param("id")
	if meetingID == "" {
		httputil.BadRequest(c, "meeting id is required")
		return
	}

	var req UpdateInviteesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.UpdateInvitees(c.Request.Context(), authCtx, meetingID, req.UserIDs, req.DeptIDs)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":    "invitees updated successfully",
		"meeting_id": meetingID,
	})
}

// ListMeetings handles GET /v1/meetings
// @Summary List meetings
// @Description List meetings in a time range
// @Tags meetings
// @Accept json
// @Produce json
// @Param begin_datetime query string true "Begin datetime (RFC3339)"
// @Param end_datetime query string true "End datetime (RFC3339)"
// @Param limit query int false "Limit (max 100)"
// @Param cursor query string false "Pagination cursor"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meetings [get]
func (h *Handler) ListMeetings(c *gin.Context) {
	var req ListMeetingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.ListMeetings(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"meetings":    result.Meetings,
		"count":       len(result.Meetings),
		"next_cursor": result.NextCursor,
	})
}

// GetMeetingInfo handles GET /v1/meetings/:id
// @Summary Get meeting info
// @Description Get detailed info about a meeting
// @Tags meetings
// @Accept json
// @Produce json
// @Param id path string true "Meeting ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meetings/{id} [get]
func (h *Handler) GetMeetingInfo(c *gin.Context) {
	meetingID := c.Param("id")
	if meetingID == "" {
		httputil.BadRequest(c, "meeting id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.GetMeetingInfo(c.Request.Context(), authCtx, meetingID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// --- Appointment Service Methods ---

// CreateMeeting creates a new meeting appointment
func (s *Service) CreateMeeting(ctx context.Context, authCtx *auth.AuthContext, title string, startTime time.Time, duration int, invitees *CreateMeetingInvitees, meetingType int, settings *CreateMeetingSettings) (*wecom.MeetingInfo, error) {
	if startTime.Before(time.Now()) {
		return nil, fmt.Errorf("meeting_start_datetime cannot be in the past")
	}
	if duration < 60 {
		return nil, fmt.Errorf("meeting_duration must be at least 60 seconds")
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.CreateMeetingParams{
		Title:         title,
		StartDateTime: startTime,
		Duration:      duration,
		MeetingType:   meetingType,
	}

	if invitees != nil {
		params.Invitees = &wecom.MeetingInvitees{
			UserIDs: invitees.UserIDs,
			DeptIDs: invitees.DeptIDs,
		}
	}
	if settings != nil {
		params.Settings = &wecom.MeetingSettings{
			MuteUponEntry:   settings.MuteUponEntry,
			WaitingRoom:     settings.WaitingRoom,
			EnableRecording: settings.EnableRecording,
		}
	}

	result, err := s.wecomClient.CreateMeeting(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create meeting: %w", err)
	}

	return result, nil
}

// CancelMeeting cancels a meeting
func (s *Service) CancelMeeting(ctx context.Context, authCtx *auth.AuthContext, meetingID string) error {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	err := s.wecomClient.CancelMeeting(ctx, corpName, appName, meetingID)
	if err != nil {
		return fmt.Errorf("failed to cancel meeting: %w", err)
	}

	return nil
}

// UpdateInvitees updates meeting invitees
func (s *Service) UpdateInvitees(ctx context.Context, authCtx *auth.AuthContext, meetingID string, userIDs []string, deptIDs []string) error {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	invitees := &wecom.MeetingInvitees{
		UserIDs: userIDs,
		DeptIDs: deptIDs,
	}

	err := s.wecomClient.UpdateMeetingInvitees(ctx, corpName, appName, meetingID, invitees)
	if err != nil {
		return fmt.Errorf("failed to update invitees: %w", err)
	}

	return nil
}

// ListMeetings lists meetings in a time range
func (s *Service) ListMeetings(ctx context.Context, authCtx *auth.AuthContext, req *ListMeetingsRequest) (*wecom.MeetingListResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	opts := &wecom.MeetingListOptions{
		BeginDatetime: req.BeginDatetime,
		EndDatetime:   req.EndDatetime,
		Limit:         limit,
		Cursor:        req.Cursor,
	}

	result, err := s.wecomClient.ListMeetings(ctx, corpName, appName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list meetings: %w", err)
	}

	return result, nil
}

// GetMeetingInfo gets meeting details
func (s *Service) GetMeetingInfo(ctx context.Context, authCtx *auth.AuthContext, meetingID string) (*wecom.MeetingInfo, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	result, err := s.wecomClient.GetMeetingInfo(ctx, corpName, appName, meetingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting info: %w", err)
	}

	return result, nil
}
