package wecom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wecom-gateway/internal/crypto"
)

// newTestTokenManager creates a TokenManager with an HTTP client that routes to the test server
func newTestTokenManager(server *httptest.Server) *TokenManager {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)

	// Create a transport that redirects all requests to our test server
	tm.httpCli = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &roundTripperFunc{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				// Rewrite the URL to point to our test server
				testURL := server.URL + req.URL.Path + "?" + req.URL.RawQuery
				newReq, err := http.NewRequest(req.Method, testURL, req.Body)
				if err != nil {
					return nil, err
				}
				for k, vv := range req.Header {
					for _, v := range vv {
						newReq.Header.Add(k, v)
					}
				}
				return http.DefaultTransport.RoundTrip(newReq)
			},
		},
	}
	return tm
}

// roundTripperFunc implements http.RoundTripper
type roundTripperFunc struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (f *roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.roundTrip(req)
}

func TestNewTokenManager(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	tm := NewTokenManager(cfg, encKey)
	if tm == nil {
		t.Fatal("NewTokenManager returned nil")
	}
	if tm.config == nil {
		t.Error("config should not be nil")
	}
	if tm.tokens == nil {
		t.Error("tokens map should not be nil")
	}
	if tm.httpCli == nil {
		t.Error("httpCli should not be nil")
	}
}

func TestTokenManager_GetToken_Success(t *testing.T) {
	tokenCallCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCallCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"access_token": "test-token-123",
			"expires_in": 7200
		}`))
	}))
	defer server.Close()

	tm := newTestTokenManager(server)

	// First call should fetch token
	token1, err := tm.GetToken(context.Background(), "test-corp", "test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token1 != "test-token-123" {
		t.Errorf("expected token test-token-123, got %s", token1)
	}

	// Second call should use cached token
	token2, err := tm.GetToken(context.Background(), "test-corp", "test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token2 != token1 {
		t.Errorf("expected cached token %s, got %s", token1, token2)
	}

	if tokenCallCount != 1 {
		t.Errorf("expected 1 API call, got %d", tokenCallCount)
	}
}

func TestTokenManager_GetToken_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 40013,
			"errmsg": "invalid corpid"
		}`))
	}))
	defer server.Close()

	tm := newTestTokenManager(server)

	_, err := tm.GetToken(context.Background(), "test-corp", "test-app")
	if err == nil {
		t.Error("expected error for invalid corpid")
	}
}

func TestTokenManager_InvalidateToken(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)

	// Add a token
	tm.tokens["test-corp/test-app"] = &tokenInfo{
		token:     "test-token",
		expiresAt: time.Now().Add(1 * time.Hour),
	}

	// Invalidate it
	tm.InvalidateToken("test-corp", "test-app")

	// Check it's gone
	if _, exists := tm.tokens["test-corp/test-app"]; exists {
		t.Error("token should be invalidated")
	}
}

func TestTokenManager_ClearAllTokens(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)

	// Add multiple tokens
	tm.tokens["corp1/app1"] = &tokenInfo{token: "token1", expiresAt: time.Now().Add(1 * time.Hour)}
	tm.tokens["corp2/app2"] = &tokenInfo{token: "token2", expiresAt: time.Now().Add(1 * time.Hour)}

	// Clear all
	tm.ClearAllTokens()

	// Check all are gone
	if len(tm.tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tm.tokens))
	}
}

func TestTokenManager_TokenExpiration(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)

	// Add an expired token
	tm.tokens["test-corp/test-app"] = &tokenInfo{
		token:     "expired-token",
		expiresAt: time.Now().Add(-1 * time.Hour),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"access_token": "new-token",
			"expires_in": 7200
		}`))
	}))
	defer server.Close()

	// Replace http client with test server transport
	tm.httpCli = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &roundTripperFunc{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				testURL := server.URL + req.URL.Path + "?" + req.URL.RawQuery
				newReq, _ := http.NewRequest(req.Method, testURL, req.Body)
				return http.DefaultTransport.RoundTrip(newReq)
			},
		},
	}

	// Should fetch new token since old one is expired
	token, err := tm.GetToken(context.Background(), "test-corp", "test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "new-token" {
		t.Errorf("expected new token, got %s", token)
	}
}

func TestTokenManager_ConcurrentAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{
			"errcode": 0,
			"errmsg": "ok",
			"access_token": "concurrent-token",
			"expires_in": 7200
		}`)))
	}))
	defer server.Close()

	tm := newTestTokenManager(server)

	// Concurrent goroutines accessing the same token
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := tm.GetToken(context.Background(), "test-corp", "test-app")
			if err != nil {
				t.Errorf("concurrent access error: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have only one token in cache
	if len(tm.tokens) != 1 {
		t.Errorf("expected 1 token in cache, got %d", len(tm.tokens))
	}
}

func TestTokenInfo(t *testing.T) {
	expiresAt := time.Now().Add(1 * time.Hour)
	info := &tokenInfo{
		token:     "test-token",
		expiresAt: expiresAt,
		corpID:    "test-corp-id",
		agentID:   123456,
	}

	if info.token != "test-token" {
		t.Errorf("expected token test-token, got %s", info.token)
	}
	if info.corpID != "test-corp-id" {
		t.Errorf("expected corpID test-corp-id, got %s", info.corpID)
	}
	if info.agentID != 123456 {
		t.Errorf("expected agentID 123456, got %d", info.agentID)
	}
}

func TestFetchAccessToken_ResponseParsing(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		expectToken string
		expectError bool
	}{
		{
			name: "valid response",
			response: `{
				"errcode": 0,
				"errmsg": "ok",
				"access_token": "test-token",
				"expires_in": 7200
			}`,
			expectToken: "test-token",
			expectError: false,
		},
		{
			name: "error response",
			response: `{
				"errcode": 40014,
				"errmsg": "invalid access token"
			}`,
			expectError: true,
		},
		{
			name:        "invalid json",
			response:    `{invalid json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				ErrCode     int    `json:"errcode"`
				ErrMsg      string `json:"errmsg"`
				AccessToken string `json:"access_token"`
				ExpiresIn   int    `json:"expires_in"`
			}

			err := json.Unmarshal([]byte(tt.response), &result)

			if tt.expectError {
				if err == nil && result.ErrCode == 0 {
					t.Error("expected error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result.AccessToken != tt.expectToken {
					t.Errorf("expected token %s, got %s", tt.expectToken, result.AccessToken)
				}
			}
		})
	}
}
