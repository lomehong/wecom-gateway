package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Handler handles MCP HTTP requests
type Handler struct {
	wecomClient   wecom.Client
	authenticator auth.Authenticator
	tools         []Tool
}

// NewHandler creates a new MCP handler
func NewHandler(wecomClient wecom.Client, authenticator auth.Authenticator) *Handler {
	return &Handler{
		wecomClient:   wecomClient,
		authenticator: authenticator,
		tools:         AllTools(),
	}
}

// HandleRPC handles POST /mcp — JSON-RPC 2.0 endpoint
func (h *Handler) HandleRPC(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ParseError, Message: "failed to read request body"},
		})
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ParseError, Message: "invalid JSON"},
		})
		return
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: InvalidRequest, Message: "invalid jsonrpc version, expected 2.0"},
		})
		return
	}

	// Authenticate request
	authCtx, err := h.authenticate(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: InvalidRequest, Message: "authentication failed: " + err.Error()},
		})
		return
	}
	_ = authCtx // auth context available for future use

	// Route to method handler
	switch req.Method {
	case "initialize":
		h.handleInitialize(c, &req)
	case "tools/list":
		h.handleToolsList(c, &req)
	case "tools/call":
		h.handleToolsCall(c, &req)
	default:
		c.JSON(http.StatusOK, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: MethodNotFound, Message: "method not found: " + req.Method},
		})
	}
}

// HandleSSE handles GET /mcp — SSE endpoint (simplified: returns connection info)
func (h *Handler) HandleSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Send a single event with server info
	data, _ := json.Marshal(map[string]interface{}{
		"name":    "wecom-gateway",
		"version": "1.0.0",
		"note":    "Use POST /mcp for JSON-RPC 2.0 requests",
	})
	c.SSEvent("server_info", string(data))
	c.SSEvent("end", "")
}

func (h *Handler) handleInitialize(c *gin.Context, req *JSONRPCRequest) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "wecom-gateway",
			Version: "1.0.0",
		},
	}

	c.JSON(http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	})
}

func (h *Handler) handleToolsList(c *gin.Context, req *JSONRPCRequest) {
	result := ListToolsResult{
		Tools: h.tools,
	}

	c.JSON(http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	})
}

func (h *Handler) handleToolsCall(c *gin.Context, req *JSONRPCRequest) {
	// Parse params
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			c.JSON(http.StatusOK, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &RPCError{Code: InvalidParams, Message: "invalid params: " + err.Error()},
			})
			return
		}
	}

	if params.Name == "" {
		c.JSON(http.StatusOK, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: InvalidParams, Message: "tool name is required"},
		})
		return
	}

	if params.Arguments == nil {
		params.Arguments = make(map[string]interface{})
	}

	result, err := CallTool(h.wecomClient, params.Name, params.Arguments)
	if err != nil {
		if rpcErr, ok := err.(*RPCError); ok {
			c.JSON(http.StatusOK, JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   rpcErr,
			})
			return
		}
		c.JSON(http.StatusOK, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: InternalError, Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	})
}

// authenticate validates the API key from the request
func (h *Handler) authenticate(c *gin.Context) (*auth.AuthContext, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	// Parse Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	apiKey := authHeader[7:]
	return h.authenticator.Authenticate(context.Background(), apiKey)
}
