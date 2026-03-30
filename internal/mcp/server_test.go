package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	client := wecom.NewMockClient()
	authenticator := &mockAuthenticator{}
	handler := NewHandler(client, authenticator)
	router.POST("/mcp", handler.HandleRPC)
	router.GET("/mcp", handler.HandleSSE)
	return router
}

type mockAuthenticator struct{}

func (m *mockAuthenticator) Authenticate(ctx context.Context, rawKey string) (*auth.AuthContext, error) {
	return &auth.AuthContext{
		KeyID:       "test-key-id",
		KeyName:     "test-key",
		Permissions: []string{"*"},
		CorpName:    "default",
		AppName:     "default",
		IsAdmin:     true,
	}, nil
}

func TestHandleRPC_Initialize(t *testing.T) {
	router := setupTestRouter()

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	if resp.ID.(float64) != 1 {
		t.Errorf("expected id 1, got %v", resp.ID)
	}

	// Check result structure
	resultBytes, _ := json.Marshal(resp.Result)
	var result InitializeResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if result.ServerInfo.Name != "wecom-gateway" {
		t.Errorf("expected server name wecom-gateway, got %s", result.ServerInfo.Name)
	}
	if result.ServerInfo.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", result.ServerInfo.Version)
	}
	if result.Capabilities.Tools == nil {
		t.Error("expected tools capability")
	}
}

func TestHandleRPC_ToolsList(t *testing.T) {
	router := setupTestRouter()

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	// Parse tools list
	resultBytes, _ := json.Marshal(resp.Result)
	var result ListToolsResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(result.Tools) < 17 {
		t.Errorf("expected at least 17 tools, got %d", len(result.Tools))
	}

	// Verify some expected tools exist
	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"wecom_get_contacts",
		"wecom_search_contact",
		"wecom_create_schedule",
		"wecom_get_schedules",
		"wecom_send_text",
		"wecom_send_markdown",
		"wecom_create_document",
		"wecom_get_todo_list",
	}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("missing expected tool: %s", name)
		}
	}
}

func TestHandleRPC_ToolsCall_GetContacts(t *testing.T) {
	router := setupTestRouter()

	params, _ := json.Marshal(map[string]interface{}{
		"name":      "wecom_get_contacts",
		"arguments": map[string]interface{}{"department_id": 1},
	})

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  params,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	// Check result is CallToolResult
	resultBytes, _ := json.Marshal(resp.Result)
	var result CallToolResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if result.IsError {
		t.Errorf("tool call returned error: %s", result.Content[0].Text)
	}
	if len(result.Content) == 0 {
		t.Error("expected content in result")
	}
}

func TestHandleRPC_ToolsCall_CreateTodo(t *testing.T) {
	router := setupTestRouter()

	params, _ := json.Marshal(map[string]interface{}{
		"name": "wecom_create_todo",
		"arguments": map[string]interface{}{
			"content":   "Test todo item",
			"assignees": []string{"user1"},
		},
	})

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  params,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestHandleRPC_InvalidJSONRPC(t *testing.T) {
	router := setupTestRouter()

	reqBody := `{"jsonrpc": "1.0", "id": 1, "method": "initialize"}`

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for invalid jsonrpc version")
	}
	if resp.Error.Code != InvalidRequest {
		t.Errorf("expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
}

func TestHandleRPC_MethodNotFound(t *testing.T) {
	router := setupTestRouter()

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "nonexistent/method",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != MethodNotFound {
		t.Errorf("expected error code %d, got %d", MethodNotFound, resp.Error.Code)
	}
}

func TestHandleRPC_NoAuth(t *testing.T) {
	router := setupTestRouter()

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "initialize",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for missing auth")
	}
}

func TestHandleSSE(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer test-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" && ct != "text/event-stream;charset=utf-8" {
		t.Errorf("expected Content-Type text/event-stream, got %s", ct)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty SSE body")
	}
}

func TestAllTools_HasRequiredTools(t *testing.T) {
	tools := AllTools()
	if len(tools) < 17 {
		t.Errorf("expected at least 17 tools, got %d", len(tools))
	}

	// Each tool must have name, description, inputSchema
	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("tool missing name")
		}
		if tool.Description == "" {
			t.Errorf("tool %s missing description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %s missing inputSchema", tool.Name)
		}
		if tool.InputSchema["type"] != "object" {
			t.Errorf("tool %s inputSchema type should be object", tool.Name)
		}
	}
}

func TestCallTool_UnknownTool(t *testing.T) {
	client := wecom.NewMockClient()
	_, err := CallTool(client, "unknown_tool", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for unknown tool")
	}
	if rpcErr, ok := err.(*RPCError); ok {
		if rpcErr.Code != MethodNotFound {
			t.Errorf("expected MethodNotFound error, got code %d", rpcErr.Code)
		}
	}
}
