package todo

import (
	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for todo operations
type Handler struct {
	service *Service
}

// NewHandler creates a new todo handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetTodoList handles GET /v1/todos
// @Summary Get todo list
// @Description Get the list of todos with optional time filters
// @Tags todos
// @Accept json
// @Produce json
// @Param create_begin_time query string false "Create begin time (RFC3339)"
// @Param create_end_time query string false "Create end time (RFC3339)"
// @Param remind_begin_time query string false "Remind begin time (RFC3339)"
// @Param remind_end_time query string false "Remind end time (RFC3339)"
// @Param limit query int false "Limit (default: 100)"
// @Param cursor query string false "Pagination cursor"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos [get]
func (h *Handler) GetTodoList(c *gin.Context) {
	var req GetTodoListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.GetTodoList(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// GetTodoDetail handles GET /v1/todos/:id
// @Summary Get todo detail
// @Description Get detailed information about a specific todo
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos/{id} [get]
func (h *Handler) GetTodoDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "todo id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	details, err := h.service.GetTodoDetail(c.Request.Context(), authCtx, id)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	if len(details) == 0 {
		httputil.NotFound(c, "todo")
		return
	}

	httputil.Success(c, details[0])
}

// CreateTodo handles POST /v1/todos
// @Summary Create todo
// @Description Create a new todo
// @Tags todos
// @Accept json
// @Produce json
// @Param request body CreateTodoRequest true "Todo parameters"
// @Success 201 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos [post]
func (h *Handler) CreateTodo(c *gin.Context) {
	var req CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	todoID, err := h.service.CreateTodo(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"todo_id": todoID,
	})
}

// UpdateTodo handles PUT /v1/todos/:id
// @Summary Update todo
// @Description Update an existing todo
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param request body UpdateTodoRequest true "Update parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos/{id} [put]
func (h *Handler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "todo id is required")
		return
	}

	var req UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.UpdateTodo(c.Request.Context(), authCtx, id, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":  "todo updated successfully",
		"todo_id": id,
	})
}

// DeleteTodo handles DELETE /v1/todos/:id
// @Summary Delete todo
// @Description Delete a todo
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos/{id} [delete]
func (h *Handler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "todo id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.DeleteTodo(c.Request.Context(), authCtx, id)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":  "todo deleted successfully",
		"todo_id": id,
	})
}

// ChangeUserStatus handles PUT /v1/todos/:id/status
// @Summary Change todo user status
// @Description Change the user processing status of a todo
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param request body ChangeUserStatusRequest true "Status parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/todos/{id}/status [put]
func (h *Handler) ChangeUserStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "todo id is required")
		return
	}

	var req ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	err := h.service.ChangeUserStatus(c.Request.Context(), authCtx, id, req.Status)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message":  "todo status updated successfully",
		"todo_id": id,
		"status":   req.Status,
	})
}
