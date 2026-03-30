package contact

import (
	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for contact operations
type Handler struct {
	service *Service
}

// NewHandler creates a new contact handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetUserList handles GET /v1/contacts/users
// @Summary Get contact user list
// @Description Get the list of contact users for a department
// @Tags contacts
// @Accept json
// @Produce json
// @Param department_id query int false "Department ID (default: 1)"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/contacts/users [get]
func (h *Handler) GetUserList(c *gin.Context) {
	var req GetUserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.GetUserList(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// SearchUser handles GET /v1/contacts/users/search
// @Summary Search contact users
// @Description Search for contact users by name, alias, or user ID
// @Tags contacts
// @Accept json
// @Produce json
// @Param query query string true "Search query (name, alias, or user ID)"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/contacts/users/search [get]
func (h *Handler) SearchUser(c *gin.Context) {
	var req SearchUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SearchUser(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}
