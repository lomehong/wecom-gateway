package wecom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wecom-gateway/internal/crypto"
)

func TestNewMediaUploader(t *testing.T) {
	cfg := MockConfig()
	tm := NewTokenManager(cfg, crypto.GenerateKeyFromPassphrase("test-passphrase"))

	uploader := NewMediaUploader(cfg, tm)
	if uploader == nil {
		t.Fatal("NewMediaUploader returned nil")
	}
	if uploader.config == nil {
		t.Error("config should not be nil")
	}
	if uploader.tokenMgr == nil {
		t.Error("tokenMgr should not be nil")
	}
	if uploader.httpCli == nil {
		t.Error("httpCli should not be nil")
	}
}

func TestMediaType_Constants(t *testing.T) {
	if MediaTypeImage != "image" {
		t.Errorf("expected MediaTypeImage to be 'image', got %s", MediaTypeImage)
	}
	if MediaTypeVoice != "voice" {
		t.Errorf("expected MediaTypeVoice to be 'voice', got %s", MediaTypeVoice)
	}
	if MediaTypeVideo != "video" {
		t.Errorf("expected MediaTypeVideo to be 'video', got %s", MediaTypeVideo)
	}
	if MediaTypeFile != "file" {
		t.Errorf("expected MediaTypeFile to be 'file', got %s", MediaTypeFile)
	}
}

// newTestMediaUploader creates a MediaUploader with HTTP client redirected to the test server
// and a pre-cached token so no real API calls are made
func newTestMediaUploader(server *httptest.Server) (*MediaUploader, *TokenManager) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)

	// Override httpCli to redirect to test server
	tm.httpCli = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &roundTripperFunc{
			roundTrip: func(req *http.Request) (*http.Response, error) {
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

	// Pre-cache a valid token so GetToken won't try to hit the real API
	tm.tokens["test-corp/test-app"] = &tokenInfo{
		token:     "test-token",
		expiresAt: time.Now().Add(2 * time.Hour),
	}

	uploader := NewMediaUploader(cfg, tm)
	// Also override the uploader's http client
	uploader.httpCli = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &roundTripperFunc{
			roundTrip: func(req *http.Request) (*http.Response, error) {
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

	return uploader, tm
}

func TestMediaUploader_UploadMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"type": "image",
			"media_id": "test-media-id",
			"created_at": "2024-01-01"
		}`))
	}))
	defer server.Close()

	uploader, _ := newTestMediaUploader(server)

	data := []byte("test image data")
	result, err := uploader.UploadMedia(context.Background(), "test-corp", "test-app", MediaTypeImage, data, "test.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MediaID != "test-media-id" {
		t.Errorf("expected media_id test-media-id, got %s", result.MediaID)
	}
	if result.Type != "image" {
		t.Errorf("expected type image, got %s", result.Type)
	}
}

func TestMediaUploader_UploadImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"type": "image",
			"media_id": "test-image-id",
			"created_at": "2024-01-01"
		}`))
	}))
	defer server.Close()

	uploader, _ := newTestMediaUploader(server)

	data := []byte("test image data")
	mediaID, err := uploader.UploadImage(context.Background(), "test-corp", "test-app", data, "test.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mediaID != "test-image-id" {
		t.Errorf("expected media_id test-image-id, got %s", mediaID)
	}
}

func TestMediaUploader_UploadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"type": "file",
			"media_id": "test-file-id",
			"created_at": "2024-01-01"
		}`))
	}))
	defer server.Close()

	uploader, _ := newTestMediaUploader(server)

	data := []byte("test file data")
	mediaID, err := uploader.UploadFile(context.Background(), "test-corp", "test-app", data, "test.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mediaID != "test-file-id" {
		t.Errorf("expected media_id test-file-id, got %s", mediaID)
	}
}

func TestMediaUploader_GetMediaURL(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	tm := NewTokenManager(cfg, encKey)
	tm.tokens["test-corp/test-app"] = &tokenInfo{
		token:     "test-token",
		expiresAt: time.Now().Add(1 * time.Hour),
	}

	uploader := NewMediaUploader(cfg, tm)

	url, err := uploader.GetMediaURL(context.Background(), "test-corp", "test-app", "test-media-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

func TestMediaUploadResult(t *testing.T) {
	result := &MediaUploadResult{
		MediaID:   "test-id",
		Type:      "image",
		CreatedAt: "2024-01-01",
	}

	if result.MediaID != "test-id" {
		t.Errorf("expected MediaID test-id, got %s", result.MediaID)
	}
	if result.Type != "image" {
		t.Errorf("expected Type image, got %s", result.Type)
	}
	if result.CreatedAt != "2024-01-01" {
		t.Errorf("expected CreatedAt 2024-01-01, got %s", result.CreatedAt)
	}
}
