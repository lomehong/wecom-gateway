package meeting

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for meeting room operations
type Handler struct {
	service *Service
}

// NewHandler creates a new meeting handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListMeetingRooms handles GET /v1/meeting-rooms
// @Summary List meeting rooms
// @Description List available meeting rooms
// @Tags meeting-rooms
// @Accept json
// @Produce json
// @Param city query string false "Filter by city"
// @Param building query string false "Filter by building"
// @Param floor query string false "Filter by floor"
// @Param capacity query int false "Minimum capacity"
// @Param equipment query []string false "Required equipment"
// @Param limit query int false "Limit (default: 50)"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meeting-rooms [get]
func (h *Handler) ListMeetingRooms(c *gin.Context) {
	var city, building, floor string
	var capacity, limit int

	if val := c.Query("city"); val != "" {
		city = val
	}
	if val := c.Query("building"); val != "" {
		building = val
	}
	if val := c.Query("floor"); val != "" {
		floor = val
	}
	if val := c.Query("capacity"); val != "" {
		if parsed, err := parseQueryParamInt(val); err == nil {
			capacity = parsed
		}
	}
	if val := c.Query("limit"); val != "" {
		if parsed, err := parseQueryParamInt(val); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if limit == 0 {
		limit = 50
	}

	var equipment []string
	if val := c.QueryArray("equipment"); len(val) > 0 {
		equipment = val
	}

	authCtx, _ := auth.GetAuthContext(c)

	rooms, cursor, err := h.service.ListMeetingRooms(c.Request.Context(), authCtx, city, building, floor, capacity, equipment, limit)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"rooms":      rooms,
		"count":      len(rooms),
		"next_cursor": cursor,
	})
}

// GetRoomAvailability handles GET /v1/meeting-rooms/:id/availability
// @Summary Get room availability
// @Description Get available time slots for a meeting room
// @Tags meeting-rooms
// @Accept json
// @Produce json
// @Param id path string true "Meeting Room ID"
// @Param start_time query string true "Start time (RFC3339)"
// @Param end_time query string true "End time (RFC3339)"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meeting-rooms/{id}/availability [get]
func (h *Handler) GetRoomAvailability(c *gin.Context) {
	roomID := c.Param("id")
	if roomID == "" {
		httputil.BadRequest(c, "meeting room id is required")
		return
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		httputil.BadRequest(c, "start_time and end_time are required")
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		httputil.BadRequest(c, "invalid start_time format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		httputil.BadRequest(c, "invalid end_time format")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	slots, err := h.service.GetRoomAvailability(c.Request.Context(), authCtx, roomID, startTime, endTime)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"room_id": roomID,
		"slots":   slots,
		"count":   len(slots),
	})
}

// BookMeetingRoomRequest represents a booking request
type BookMeetingRoomRequest struct {
	MeetingRoomID string   `json:"meetingroom_id" binding:"required"`
	Subject       string   `json:"subject" binding:"required"`
	StartTime     string   `json:"start_time" binding:"required"`
	EndTime       string   `json:"end_time" binding:"required"`
	Booker        string   `json:"booker" binding:"required"`
	Attendees     []string `json:"attendees,omitempty"`
}

// BookMeetingRoom handles POST /v1/meeting-rooms/:id/bookings
// @Summary Book meeting room
// @Description Book a meeting room
// @Tags meeting-rooms
// @Accept json
// @Produce json
// @Param id path string true "Meeting Room ID"
// @Param request body BookMeetingRoomRequest true "Booking parameters"
// @Success 201 {object} wecom.BookingResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 409 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/meeting-rooms/{id}/bookings [post]
func (h *Handler) BookMeetingRoom(c *gin.Context) {
	var req BookMeetingRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		httputil.BadRequest(c, "invalid start_time format")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		httputil.BadRequest(c, "invalid end_time format")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.BookMeetingRoom(c.Request.Context(), authCtx, req.MeetingRoomID, req.Subject, startTime, endTime, req.Booker, req.Attendees)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, result)
}

func parseQueryParamInt(val string) (int, error) {
	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return 0, err
	}
	return result, nil
}
