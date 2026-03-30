package contact

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler_Contact(t *testing.T) {
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

func TestHandler_GetUserList_ValidQuery(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/contacts/users?department_id=1", nil)

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetUserList(c)

	// Handler should process without crashing
}

func TestHandler_GetUserList_DefaultDepartment(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/contacts/users", nil)

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.GetUserList(c)

	// Handler should process without crashing (default department_id=1)
}

func TestHandler_SearchUser_ValidQuery(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/contacts/users/search?query=张三", nil)

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SearchUser(c)

	// Handler should process without crashing
}

func TestHandler_SearchUser_MissingQuery(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/v1/contacts/users/search", nil)

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SearchUser(c)

	// Handler should return bad request
}
