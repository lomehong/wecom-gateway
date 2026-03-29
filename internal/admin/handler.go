package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/apikey"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/httputil"
	"wecom-gateway/internal/store"
)

// Handler handles HTTP requests for admin operations
type Handler struct {
	service    *Service
	apiKeySvc  *apikey.Service
	auditQuer  *audit.Querier
	authMgr    *AuthManager
}

// NewHandler creates a new admin handler
func NewHandler(service *Service, apiKeySvc *apikey.Service, auditQuer *audit.Querier, authMgr *AuthManager) *Handler {
	return &Handler{
		service:   service,
		apiKeySvc: apiKeySvc,
		auditQuer: auditQuer,
		authMgr:   authMgr,
	}
}

// CreateAPIKey handles POST /v1/admin/api-keys
// @Summary Create API key
// @Description Create a new API key
// @Tags admin
// @Accept json
// @Produce json
// @Param request body apikey.CreateKeyRequest true "API key parameters"
// @Success 201 {object} apikey.APIKeyResponse
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/api-keys [post]
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req apikey.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	key, err := h.apiKeySvc.CreateKey(c.Request.Context(), &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, key)
}

// ListAPIKeys handles GET /v1/admin/api-keys
// @Summary List API keys
// @Description List all API keys
// @Tags admin
// @Accept json
// @Produce json
// @Param disabled query bool false "Filter by disabled status"
// @Param limit query int false "Limit (default: 50)"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/api-keys [get]
func (h *Handler) ListAPIKeys(c *gin.Context) {
	var disabled *bool
	if val := c.Query("disabled"); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			disabled = &parsed
		}
	}

	limit := 50
	if val := c.Query("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	keys, cursor, err := h.apiKeySvc.ListKeys(c.Request.Context(), disabled, limit)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"keys":       keys,
		"count":      len(keys),
		"next_cursor": cursor,
	})
}

// DeleteAPIKey handles DELETE /v1/admin/api-keys/:id
// @Summary Delete API key
// @Description Delete an API key
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "API Key ID"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/api-keys/{id} [delete]
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "api key id is required")
		return
	}

	err := h.apiKeySvc.DeleteKey(c.Request.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			httputil.NotFound(c, "api key")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message": "api key deleted successfully",
		"id":      id,
	})
}

// QueryAuditLogs handles GET /v1/admin/audit-logs
// @Summary Query audit logs
// @Description Query audit logs with filtering
// @Tags admin
// @Accept json
// @Produce json
// @Param api_key_name query string false "Filter by API key name"
// @Param method query string false "Filter by HTTP method"
// @Param path query string false "Filter by request path"
// @Param status_code query int false "Filter by status code"
// @Param start_time query string false "Filter by start time (RFC3339)"
// @Param end_time query string false "Filter by end time (RFC3339)"
// @Param limit query int false "Limit (default: 50)"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/audit-logs [get]
func (h *Handler) QueryAuditLogs(c *gin.Context) {
	opts := &audit.QueryOptions{
		Limit: 50,
	}

	if val := c.Query("api_key_name"); val != "" {
		opts.APIKeyName = &val
	}
	if val := c.Query("method"); val != "" {
		opts.Method = &val
	}
	if val := c.Query("path"); val != "" {
		opts.Path = &val
	}
	if val := c.Query("status_code"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			opts.StatusCode = &parsed
		}
	}
	if val := c.Query("limit"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 && parsed <= 100 {
			opts.Limit = parsed
		}
	}
	if val := c.Query("start_time"); val != "" {
		if parsed, err := time.Parse(time.RFC3339, val); err == nil {
			opts.StartTime = &parsed
		}
	}
	if val := c.Query("end_time"); val != "" {
		if parsed, err := time.Parse(time.RFC3339, val); err == nil {
			opts.EndTime = &parsed
		}
	}

	logs, cursor, err := h.auditQuer.Query(c.Request.Context(), opts)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"logs":        logs,
		"count":       len(logs),
		"next_cursor": cursor,
	})
}

// GetDashboardStats handles GET /v1/admin/dashboard
// @Summary Get dashboard statistics
// @Description Get dashboard statistics
// @Tags admin
// @Accept json
// @Produce json
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} DashboardStats
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/dashboard [get]
func (h *Handler) GetDashboardStats(c *gin.Context) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if val := c.Query("start_time"); val != "" {
		if parsed, err := time.Parse(time.RFC3339, val); err == nil {
			startTime = parsed
		}
	}
	if val := c.Query("end_time"); val != "" {
		if parsed, err := time.Parse(time.RFC3339, val); err == nil {
			endTime = parsed
		}
	}

	stats, err := h.service.GetDashboardStats(c.Request.Context(), startTime, endTime)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, stats)
}

// InitializeSystem handles POST /v1/admin/initialize
// @Summary Initialize system
// @Description Initialize system with default data
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/initialize [post]
func (h *Handler) InitializeSystem(c *gin.Context) {
	err := h.service.InitializeSystem(c.Request.Context())
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message": "system initialized successfully",
	})
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// Login handles POST /v1/admin/login
// @Summary Admin login
// @Description Login with username and password to get JWT token
// @Tags admin
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} httputil.Response
// @Failure 401 {object} httputil.Response
// @Router /v1/admin/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	token, err := h.authMgr.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			httputil.Unauthorized(c, "Invalid username or password")
			return
		}
		if err == ErrUserDisabled {
			httputil.Forbidden(c, "User account is disabled")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	// Get user info
	user, err := h.service.DB.GetAdminUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(TokenExpiration.Seconds()),
		UserID:      user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword handles POST /v1/admin/change-password
// @Summary Change password
// @Description Change current user's password
// @Tags admin
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 401 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	// Get username from JWT claims
	claims, exists := c.Get("jwt_claims")
	if !exists {
		httputil.Unauthorized(c, "Authentication required")
		return
	}

	jwtClaims, ok := claims.(*JWTClaims)
	if !ok {
		httputil.Unauthorized(c, "Invalid authentication context")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// Validate new password
	if len(req.NewPassword) < 6 {
		httputil.BadRequest(c, "New password must be at least 6 characters long")
		return
	}

	err := h.authMgr.ChangePassword(c.Request.Context(), jwtClaims.Username, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == ErrInvalidCredentials {
			httputil.Unauthorized(c, "Current password is incorrect")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message": "Password changed successfully",
	})
}

// WeCom Corp Management

// ListWeComCorps handles GET /v1/admin/corps
// @Summary List WeCom corps
// @Description List all WeChat Work corporations
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps [get]
func (h *Handler) ListWeComCorps(c *gin.Context) {
	corps, err := h.service.DB.ListWeComCorps(c.Request.Context())
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	// Enrich corps with app counts
	result := make([]gin.H, 0, len(corps))
	for _, corp := range corps {
		apps, err := h.service.DB.ListWeComApps(c.Request.Context(), corp.Name)
		appCount := 0
		if err == nil {
			appCount = len(apps)
		}

		result = append(result, gin.H{
			"id":         corp.ID,
			"name":       corp.Name,
			"corp_id":    corp.CorpID,
			"app_count":  appCount,
			"created_at": corp.CreatedAt,
			"updated_at": corp.UpdatedAt,
		})
	}

	httputil.Success(c, gin.H{
		"corps": result,
		"count": len(result),
	})
}

// CreateWeComCorpRequest represents create corp request
type CreateWeComCorpRequest struct {
	Name  string `json:"name" binding:"required"`
	CorpID string `json:"corp_id" binding:"required"`
}

// CreateWeComCorp handles POST /v1/admin/corps
// @Summary Create WeCom corp
// @Description Create a new WeChat Work corporation
// @Tags admin
// @Accept json
// @Produce json
// @Param request body CreateWeComCorpRequest true "Corp parameters"
// @Success 201 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps [post]
func (h *Handler) CreateWeComCorp(c *gin.Context) {
	var req CreateWeComCorpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	corpID := generateID("corp_")
	corp := &store.WeComCorp{
		ID:     corpID,
		Name:   req.Name,
		CorpID: req.CorpID,
	}

	if err := h.service.DB.CreateWeComCorp(c.Request.Context(), corp); err != nil {
		if err == store.ErrDuplicate {
			httputil.Conflict(c, "Corp with this name already exists")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"id":      corp.ID,
		"name":    corp.Name,
		"corp_id": corp.CorpID,
	})
}

// UpdateWeComCorpRequest represents update corp request
type UpdateWeComCorpRequest struct {
	CorpID string `json:"corp_id" binding:"required"`
}

// UpdateWeComCorp handles PUT /v1/admin/corps/:id
// @Summary Update WeCom corp
// @Description Update a WeChat Work corporation
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Corp ID"
// @Param request body UpdateWeComCorpRequest true "Corp parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{id} [put]
func (h *Handler) UpdateWeComCorp(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "corp id is required")
		return
	}

	var req UpdateWeComCorpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// Get existing corp
	corp, err := h.service.DB.GetWeComCorpByID(c.Request.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			httputil.NotFound(c, "corp")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	// Update fields
	corp.CorpID = req.CorpID

	if err := h.service.DB.UpdateWeComCorp(c.Request.Context(), corp); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"id":      corp.ID,
		"name":    corp.Name,
		"corp_id": corp.CorpID,
	})
}

// DeleteWeComCorp handles DELETE /v1/admin/corps/:id
// @Summary Delete WeCom corp
// @Description Delete a WeChat Work corporation
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Corp ID"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{id} [delete]
func (h *Handler) DeleteWeComCorp(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.BadRequest(c, "corp id is required")
		return
	}

	// Get existing corp to get the name
	corp, err := h.service.DB.GetWeComCorpByID(c.Request.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			httputil.NotFound(c, "corp")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	if err := h.service.DB.DeleteWeComCorp(c.Request.Context(), corp.Name); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message": "corp deleted successfully",
		"id":      id,
	})
}

// WeCom App Management

// ListWeComApps handles GET /v1/admin/corps/:corp_name/apps
// @Summary List WeCom apps
// @Description List all WeChat Work applications for a corporation
// @Tags admin
// @Accept json
// @Produce json
// @Param corp_name path string true "Corporation name"
// @Success 200 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{corp_name}/apps [get]
func (h *Handler) ListWeComApps(c *gin.Context) {
	corpName := c.Query("corp_name")
	if corpName == "" {
		httputil.BadRequest(c, "corp name is required")
		return
	}

	apps, err := h.service.DB.ListWeComApps(c.Request.Context(), corpName)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	result := make([]gin.H, 0, len(apps))
	for _, app := range apps {
		result = append(result, gin.H{
			"id":          app.ID,
			"name":        app.Name,
			"corp_name":   app.CorpName,
			"agent_id":    app.AgentID,
			"secret_enc":  "***", // Don't expose the actual secret
			"created_at":  app.CreatedAt,
			"updated_at":  app.UpdatedAt,
		})
	}

	httputil.Success(c, gin.H{
		"apps":  result,
		"count": len(result),
	})
}

// CreateWeComAppRequest represents create app request
type CreateWeComAppRequest struct {
	Name    string `json:"name" binding:"required"`
	AgentID int64  `json:"agent_id" binding:"required"`
	Secret  string `json:"secret" binding:"required"`
}

// CreateWeComApp handles POST /v1/admin/corps/:corp_name/apps
// @Summary Create WeCom app
// @Description Create a new WeChat Work application
// @Tags admin
// @Accept json
// @Produce json
// @Param corp_name path string true "Corporation name"
// @Param request body CreateWeComAppRequest true "App parameters"
// @Success 201 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{corp_name}/apps [post]
func (h *Handler) CreateWeComApp(c *gin.Context) {
	corpName := c.Query("corp_name")
	if corpName == "" {
		httputil.BadRequest(c, "corp name is required")
		return
	}

	var req CreateWeComAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// Encrypt secret
	nonce, err := crypto.GenerateRandomBytes(12)
	if err != nil {
		httputil.InternalError(c, "Failed to generate nonce: "+err.Error())
		return
	}

	secretEnc, err := crypto.EncryptString(req.Secret, h.service.encKey)
	if err != nil {
		httputil.InternalError(c, "Failed to encrypt secret: "+err.Error())
		return
	}

	appID := generateID("app_")
	app := &store.WeComApp{
		ID:        appID,
		Name:      req.Name,
		CorpName:  corpName,
		AgentID:   req.AgentID,
		SecretEnc: secretEnc,
		Nonce:     string(nonce),
	}

	if err := h.service.DB.CreateWeComApp(c.Request.Context(), app); err != nil {
		if err == store.ErrDuplicate {
			httputil.Conflict(c, "App with this name already exists in this corp")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"id":       app.ID,
		"name":     app.Name,
		"corp_name": app.CorpName,
		"agent_id": app.AgentID,
	})
}

// UpdateWeComAppRequest represents update app request
type UpdateWeComAppRequest struct {
	Name    string `json:"name" binding:"required"`
	AgentID int64  `json:"agent_id" binding:"required"`
	Secret  string `json:"secret" binding:"required"`
}

// UpdateWeComApp handles PUT /v1/admin/corps/:corp_name/apps/:id
// @Summary Update WeCom app
// @Description Update a WeChat Work application
// @Tags admin
// @Accept json
// @Produce json
// @Param corp_name path string true "Corporation name"
// @Param id path string true "App ID"
// @Param request body UpdateWeComAppRequest true "App parameters"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{corp_name}/apps/{id} [put]
func (h *Handler) UpdateWeComApp(c *gin.Context) {
	corpName := c.Query("corp_name")
	id := c.Param("id")

	if corpName == "" || id == "" {
		httputil.BadRequest(c, "corp name and app id are required")
		return
	}

	var req UpdateWeComAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	// Get existing app
	app, err := h.service.DB.GetWeComAppByID(c.Request.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			httputil.NotFound(c, "app")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	// Update fields
	app.Name = req.Name
	app.AgentID = req.AgentID

	// Only update secret if provided
	if req.Secret != "" {
		nonce, err := crypto.GenerateRandomBytes(12)
		if err != nil {
			httputil.InternalError(c, "Failed to generate nonce: "+err.Error())
			return
		}

		secretEnc, err := crypto.EncryptString(req.Secret, h.service.encKey)
		if err != nil {
			httputil.InternalError(c, "Failed to encrypt secret: "+err.Error())
			return
		}
		app.SecretEnc = secretEnc
		app.Nonce = string(nonce)
	}

	if err := h.service.DB.UpdateWeComApp(c.Request.Context(), app); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"id":       app.ID,
		"name":     app.Name,
		"corp_name": app.CorpName,
		"agent_id": app.AgentID,
	})
}

// DeleteWeComApp handles DELETE /v1/admin/corps/:corp_name/apps/:id
// @Summary Delete WeCom app
// @Description Delete a WeChat Work application
// @Tags admin
// @Accept json
// @Produce json
// @Param corp_name path string true "Corporation name"
// @Param id path string true "App ID"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 404 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/admin/corps/{corp_name}/apps/{id} [delete]
func (h *Handler) DeleteWeComApp(c *gin.Context) {
	corpName := c.Query("corp_name")
	id := c.Param("id")

	if corpName == "" || id == "" {
		httputil.BadRequest(c, "corp name and app id are required")
		return
	}

	if err := h.service.DB.DeleteWeComApp(c.Request.Context(), id); err != nil {
		if err == store.ErrNotFound {
			httputil.NotFound(c, "app")
			return
		}
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"message": "app deleted successfully",
		"id":      id,
	})
}
