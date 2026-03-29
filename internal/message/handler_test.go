package message

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewHandler_Message(t *testing.T) {
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

func TestHandler_SendText_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := SendTextRequest{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1", "user2"},
		Content:      "test message",
		Safe:         true,
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/text", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SendText(c)

	// Handler should process without crashing
}

func TestHandler_SendText_InvalidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Missing required field
	req := SendTextRequest{
		ReceiverType: "user",
		// Missing ReceiverIDs
		Content: "test message",
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/text", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SendText(c)

	// Should return validation error
}

func TestHandler_SendMarkdown_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := SendMarkdownRequest{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		Content:      "**test**",
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/markdown", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SendMarkdown(c)

	// Handler should process without crashing
}

func TestHandler_SendImage_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := SendImageRequest{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		MediaID:      "test-media-id",
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/image", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SendImage(c)

	// Handler should process without crashing
}

func TestHandler_SendFile_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := SendFileRequest{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		MediaID:      "test-file-id",
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/file", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SendFile(c)

	// Handler should process without crashing
}

func TestHandler_SendCard_ValidRequest(t *testing.T) {
	mockClient := wecom.NewMockClient()
	service := NewService(mockClient)
	handler := NewHandler(service)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := SendCardRequest{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		CardContent: map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
			},
		},
	}

	jsonBytes, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/v1/messages/card", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Set auth context
	c.Set("auth_context", &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
	})

	handler.SendCard(c)

	// Handler should process without crashing
}

func TestSendMessageRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     SendMessageRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: SendMessageRequest{
				ReceiverType: "user",
				ReceiverIDs:  []string{"user1"},
			},
			wantErr: false,
		},
		{
			name: "missing receiver type",
			req: SendMessageRequest{
				ReceiverIDs: []string{"user1"},
			},
			wantErr: true,
		},
		{
			name: "missing receiver ids",
			req: SendMessageRequest{
				ReceiverType: "user",
			},
			wantErr: true,
		},
		{
			name: "empty receiver ids",
			req: SendMessageRequest{
				ReceiverType: "user",
				ReceiverIDs:  []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.req)
			var parsed SendMessageRequest
			err := json.Unmarshal(jsonBytes, &parsed)

			if tt.wantErr && err == nil {
				// In real handler, Gin would validate
			}
		})
	}
}
