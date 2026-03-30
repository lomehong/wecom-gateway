package todo

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler_Todo(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)
	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.service != service {
		t.Error("service field not set correctly")
	}
}

func TestHandler_GetTodoList(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/todos?limit=50", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetTodoList(c)
}

func TestHandler_GetTodoDetail(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/todos/todo-1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "todo-1"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetTodoDetail(c)
}

func TestHandler_GetTodoDetail_EmptyID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/todos/", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetTodoDetail(c)
}

func TestHandler_CreateTodo(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"title": "Test Todo",
		"content": "Test content",
		"assignees": ["user1", "user2"]
	}`

	c.Request = httptest.NewRequest("POST", "/v1/todos", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateTodo(c)
}

func TestHandler_CreateTodo_InvalidJSON(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/v1/todos", strings.NewReader("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.CreateTodo(c)
}

func TestHandler_UpdateTodo(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{
		"content": "Updated content"
	}`

	c.Request = httptest.NewRequest("PUT", "/v1/todos/todo-1", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "todo-1"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.UpdateTodo(c)
}

func TestHandler_UpdateTodo_EmptyID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{"content": "test"}`

	c.Request = httptest.NewRequest("PUT", "/v1/todos/", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.UpdateTodo(c)
}

func TestHandler_DeleteTodo(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/v1/todos/todo-1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "todo-1"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.DeleteTodo(c)
}

func TestHandler_DeleteTodo_EmptyID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("DELETE", "/v1/todos/", nil)

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.DeleteTodo(c)
}

func TestHandler_ChangeUserStatus(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{"status": 0}`

	c.Request = httptest.NewRequest("PUT", "/v1/todos/todo-1/status", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "todo-1"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.ChangeUserStatus(c)
}

func TestHandler_ChangeUserStatus_EmptyID(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody := `{"status": 1}`

	c.Request = httptest.NewRequest("PUT", "/v1/todos//status", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.ChangeUserStatus(c)
}

func TestHandler_ChangeUserStatus_InvalidJSON(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("PUT", "/v1/todos/todo-1/status", strings.NewReader("invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "todo-1"}}

	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.ChangeUserStatus(c)
}
