package message

import (
	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for message operations
type Handler struct {
	service *Service
}

// NewHandler creates a new message handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// validReceiverTypes defines allowed receiver types
var validReceiverTypes = map[string]bool{
	"user":       true,
	"department": true,
	"tag":        true,
}

// validateReceiverType checks if the receiver_type is valid
func validateReceiverType(c *gin.Context, receiverType string) bool {
	if !validReceiverTypes[receiverType] {
		httputil.BadRequest(c, "invalid receiver_type, must be one of: user, department, tag")
		return false
	}
	return true
}

// SendMessageRequest represents a common message request
type SendMessageRequest struct {
	ReceiverType string   `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string `json:"receiver_ids" binding:"required,min=1"`
}

// SendTextRequest represents a text message request
type SendTextRequest struct {
	ReceiverType string   `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string `json:"receiver_ids" binding:"required,min=1"`
	Content      string   `json:"content" binding:"required"`
	Safe         bool     `json:"safe"`
}

// SendText handles POST /v1/messages/text
// @Summary Send text message
// @Description Send a text message to users or departments
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendTextRequest true "Message parameters"
// @Success 200 {object} wecom.SendResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/messages/text [post]
func (h *Handler) SendText(c *gin.Context) {
	var req SendTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if !validateReceiverType(c, req.ReceiverType) {
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SendText(c.Request.Context(), authCtx, req.ReceiverType, req.ReceiverIDs, req.Content, req.Safe)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// SendMarkdownRequest represents a markdown message request
type SendMarkdownRequest struct {
	ReceiverType string   `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string `json:"receiver_ids" binding:"required,min=1"`
	Content      string   `json:"content" binding:"required"`
}

// SendMarkdown handles POST /v1/messages/markdown
// @Summary Send markdown message
// @Description Send a markdown message to users or departments
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendMarkdownRequest true "Message parameters"
// @Success 200 {object} wecom.SendResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/messages/markdown [post]
func (h *Handler) SendMarkdown(c *gin.Context) {
	var req SendMarkdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if !validateReceiverType(c, req.ReceiverType) {
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SendMarkdown(c.Request.Context(), authCtx, req.ReceiverType, req.ReceiverIDs, req.Content)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// SendImageRequest represents an image message request
type SendImageRequest struct {
	ReceiverType string   `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string `json:"receiver_ids" binding:"required,min=1"`
	MediaID      string   `json:"media_id" binding:"required"`
}

// SendImage handles POST /v1/messages/image
// @Summary Send image message
// @Description Send an image message to users or departments
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendImageRequest true "Message parameters"
// @Success 200 {object} wecom.SendResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/messages/image [post]
func (h *Handler) SendImage(c *gin.Context) {
	var req SendImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if !validateReceiverType(c, req.ReceiverType) {
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SendImage(c.Request.Context(), authCtx, req.ReceiverType, req.ReceiverIDs, req.MediaID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// SendFileRequest represents a file message request
type SendFileRequest struct {
	ReceiverType string   `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string `json:"receiver_ids" binding:"required,min=1"`
	MediaID      string   `json:"media_id" binding:"required"`
}

// SendFile handles POST /v1/messages/file
// @Summary Send file message
// @Description Send a file message to users or departments
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendFileRequest true "Message parameters"
// @Success 200 {object} wecom.SendResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/messages/file [post]
func (h *Handler) SendFile(c *gin.Context) {
	var req SendFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if !validateReceiverType(c, req.ReceiverType) {
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SendFile(c.Request.Context(), authCtx, req.ReceiverType, req.ReceiverIDs, req.MediaID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}

// SendCardRequest represents a card message request
type SendCardRequest struct {
	ReceiverType string                 `json:"receiver_type" binding:"required"`
	ReceiverIDs  []string               `json:"receiver_ids" binding:"required,min=1"`
	CardContent  map[string]interface{} `json:"card_content" binding:"required"`
}

// SendCard handles POST /v1/messages/card
// @Summary Send card message
// @Description Send a card message to users or departments
// @Tags messages
// @Accept json
// @Produce json
// @Param request body SendCardRequest true "Message parameters"
// @Success 200 {object} wecom.SendResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/messages/card [post]
func (h *Handler) SendCard(c *gin.Context) {
	var req SendCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	if !validateReceiverType(c, req.ReceiverType) {
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.SendCard(c.Request.Context(), authCtx, req.ReceiverType, req.ReceiverIDs, req.CardContent)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}
