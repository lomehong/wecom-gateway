package schedule

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for schedule operations
type Handler struct {
	service *Service
}

// NewHandler creates a new schedule handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateSchedule handles POST /v1/schedules
// @Summary Create schedule
// @Description Create a new WeChat Work schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param request body CreateScheduleRequest true "Schedule parameters"
// @Success 201 {object} wecom.Schedule
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules [post]
func (h *Handler) CreateSchedule(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	schedule, err := h.service.CreateSchedule(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, schedule)
}

// GetSchedules handles GET /v1/schedules
// @Summary Get schedules
// @Description Get schedules for a user
// @Tags schedules
// @Accept json
// @Produce json
// @Param userid query string true "User ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules [get]
func (h *Handler) GetSchedules(c *gin.Context) {
	var req GetSchedulesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	schedules, err := h.service.GetSchedules(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"schedules": schedules,
		"count":     len(schedules),
	})
}

// GetScheduleByID handles GET /v1/schedules/:id
// @Summary Get schedule by ID
// @Description Get a specific schedule by ID
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} wecom.Schedule
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules/{id} [get]
func (h *Handler) GetScheduleByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "schedule id is required")
		return
	}

	// For simplicity, we'll get all schedules and filter
	// In production, you'd implement a dedicated GetSchedule method
	authCtx, _ := auth.GetAuthContext(c)

	req := &GetSchedulesRequest{
		UserID:    authCtx.KeyName, // Use current user
		StartTime: time.Now().AddDate(-1, 0, 0),
		EndTime:   time.Now().AddDate(1, 0, 0),
		Limit:     100,
	}

	schedules, err := h.service.GetSchedules(c.Request.Context(), authCtx, req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	for _, schedule := range schedules {
		if schedule.ScheduleID == id {
			httputil.Success(c, schedule)
			return
		}
	}

	httputil.NotFound(c, "schedule")
}

// UpdateSchedule handles PATCH /v1/schedules/:id
// @Summary Update schedule
// @Description Update an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param request body UpdateScheduleRequest true "Update parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules/{id} [patch]
func (h *Handler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "schedule id is required")
		return
	}

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.UpdateSchedule(c.Request.Context(), authCtx, id, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":     "schedule updated successfully",
		"schedule_id": id,
	})
}

// DeleteSchedule handles DELETE /v1/schedules/:id
// @Summary Delete schedule
// @Description Delete a schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules/{id} [delete]
func (h *Handler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "schedule id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.DeleteSchedule(c.Request.Context(), authCtx, id)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":     "schedule deleted successfully",
		"schedule_id": id,
	})
}

// ParseTimeQueryParam parses a time query parameter
func ParseTimeQueryParam(c *gin.Context, key string) (time.Time, error) {
	val := c.Query(key)
	if val == "" {
		return time.Time{}, nil
	}

	// Try Unix timestamp
	if ts, err := strconv.ParseInt(val, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	// Try RFC3339
	return time.Parse(time.RFC3339, val)
}
