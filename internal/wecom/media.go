package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"wecom-gateway/internal/config"
)

// MediaType represents the type of media file
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVoice MediaType = "voice"
	MediaTypeVideo MediaType = "video"
	MediaTypeFile  MediaType = "file"
)

// MediaUploadResult represents the result of uploading media
type MediaUploadResult struct {
	MediaID   string `json:"media_id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
}

// MediaUploader handles media upload operations
type MediaUploader struct {
	config   *config.Config
	tokenMgr *TokenManager
	httpCli  *http.Client
}

// NewMediaUploader creates a new media uploader
func NewMediaUploader(cfg *config.Config, tokenMgr *TokenManager) *MediaUploader {
	return &MediaUploader{
		config:   cfg,
		tokenMgr: tokenMgr,
		httpCli:  &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadMedia uploads a media file to WeChat Work
func (mu *MediaUploader) UploadMedia(ctx context.Context, corpName, appName string, mediaType MediaType, data []byte, filename string) (*MediaUploadResult, error) {
	token, err := mu.tokenMgr.GetToken(ctx, corpName, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s", token, mediaType)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := mu.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var result struct {
		ErrCode int                `json:"errcode"`
		ErrMsg  string             `json:"errmsg"`
		MediaID string             `json:"media_id"`
		Type    string             `json:"type"`
		CreatedAt string           `json:"created_at"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.ErrCode != 0 {
		// If token expired, invalidate it and retry
		apiErr := &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
		if apiErr.IsAccessTokenExpired() {
			mu.tokenMgr.InvalidateToken(corpName, appName)
			return mu.UploadMedia(ctx, corpName, appName, mediaType, data, filename)
		}
		return nil, apiErr
	}

	return &MediaUploadResult{
		MediaID:   result.MediaID,
		Type:      result.Type,
		CreatedAt: result.CreatedAt,
	}, nil
}

// UploadImage uploads an image file
func (mu *MediaUploader) UploadImage(ctx context.Context, corpName, appName string, data []byte, filename string) (string, error) {
	result, err := mu.UploadMedia(ctx, corpName, appName, MediaTypeImage, data, filename)
	if err != nil {
		return "", err
	}
	return result.MediaID, nil
}

// UploadFile uploads a file
func (mu *MediaUploader) UploadFile(ctx context.Context, corpName, appName string, data []byte, filename string) (string, error) {
	result, err := mu.UploadMedia(ctx, corpName, appName, MediaTypeFile, data, filename)
	if err != nil {
		return "", err
	}
	return result.MediaID, nil
}

// GetMediaURL returns the URL for downloading media
func (mu *MediaUploader) GetMediaURL(ctx context.Context, corpName, appName, mediaID string) (string, error) {
	token, err := mu.tokenMgr.GetToken(ctx, corpName, appName)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s", token, mediaID)
	return url, nil
}
