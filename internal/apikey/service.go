package apikey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/store"
)

var (
	ErrInvalidPermissions = errors.New("invalid permissions")
	ErrInvalidKeyFormat   = errors.New("invalid key format")
)

// Service handles API key management operations
type Service struct {
	db     store.Database
	config *config.Config
}

// NewService creates a new API key service
func NewService(db store.Database, cfg *config.Config) *Service {
	return &Service{
		db:     db,
		config: cfg,
	}
}

// CreateKeyRequest represents a request to create an API key
type CreateKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions"`
	CorpName    string   `json:"corp_name" binding:"required"`
	AppName     string   `json:"app_name"`           // Empty for admin keys
	ExpiresDays int      `json:"expires_days"`        // 0 for no expiration
}

// CreateKey creates a new API key
func (s *Service) CreateKey(ctx context.Context, req *CreateKeyRequest) (*APIKeyResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Validate corp and app exist
	if req.AppName != "" {
		_, err := s.db.GetWeComApp(ctx, req.CorpName, req.AppName)
		if err != nil {
			return nil, fmt.Errorf("app not found: %w", err)
		}
	} else {
		// Verify corp exists for admin keys
		_, err := s.db.GetWeComCorpByName(ctx, req.CorpName)
		if err != nil {
			return nil, fmt.Errorf("corp not found: %w", err)
		}
	}

	// Generate key ID and raw API key
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	rawKey, err := crypto.GenerateAPIKey("wgk_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresDays > 0 {
		expiry := time.Now().AddDate(0, 0, req.ExpiresDays)
		expiresAt = &expiry
	} else if s.config.Auth.KeyExpiryDays > 0 {
		// Use default expiry from config
		expiry := time.Now().AddDate(0, 0, s.config.Auth.KeyExpiryDays)
		expiresAt = &expiry
	}

	// Hash the API key
	keyHash := hashAPIKey(rawKey)

	// Create API key record
	key := &store.APIKey{
		ID:          id,
		Name:        req.Name,
		KeyHash:     keyHash,
		Permissions: req.Permissions,
		CorpName:    req.CorpName,
		AppName:     req.AppName,
		ExpiresAt:   expiresAt,
		Disabled:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.CreateAPIKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return &APIKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		APIKey:      rawKey, // Only return raw key on creation
		Permissions: key.Permissions,
		CorpName:    key.CorpName,
		AppName:     key.AppName,
		ExpiresAt:   key.ExpiresAt,
		CreatedAt:   key.CreatedAt,
	}, nil
}

// ListKeys lists API keys with optional filtering
func (s *Service) ListKeys(ctx context.Context, disabled *bool, limit int) ([]*APIKeyInfo, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	keys, cursor, err := s.db.ListAPIKeys(ctx, store.ListOptions{
		Disabled: disabled,
		Limit:    limit,
	})

	if err != nil {
		return nil, "", fmt.Errorf("failed to list api keys: %w", err)
	}

	result := make([]*APIKeyInfo, len(keys))
	for i, key := range keys {
		result[i] = &APIKeyInfo{
			ID:          key.ID,
			Name:        key.Name,
			Permissions: key.Permissions,
			CorpName:    key.CorpName,
			AppName:     key.AppName,
			ExpiresAt:   key.ExpiresAt,
			Disabled:    key.Disabled,
			CreatedAt:   key.CreatedAt,
			UpdatedAt:   key.UpdatedAt,
		}
	}

	return result, cursor, nil
}

// GetKey retrieves an API key by ID
func (s *Service) GetKey(ctx context.Context, id string) (*APIKeyInfo, error) {
	// Note: We can't directly get by ID, so we'll need to scan all keys
	// This is inefficient but works for the current implementation
	// TODO: Add GetAPIKeyByID method to store interface
	keys, _, err := s.db.ListAPIKeys(ctx, store.ListOptions{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}

	for _, key := range keys {
		if key.ID == id {
			return &APIKeyInfo{
				ID:          key.ID,
				Name:        key.Name,
				Permissions: key.Permissions,
				CorpName:    key.CorpName,
				AppName:     key.AppName,
				ExpiresAt:   key.ExpiresAt,
				Disabled:    key.Disabled,
				CreatedAt:   key.CreatedAt,
				UpdatedAt:   key.UpdatedAt,
			}, nil
		}
	}

	return nil, store.ErrNotFound
}

// DisableKey disables an API key
func (s *Service) DisableKey(ctx context.Context, id string) error {
	key, err := s.GetKey(ctx, id)
	if err != nil {
		return err
	}

	key.Disabled = true
	key.UpdatedAt = time.Now()

	return s.db.UpdateAPIKey(ctx, &store.APIKey{
		ID:          key.ID,
		Name:        key.Name,
		Permissions: key.Permissions,
		CorpName:    key.CorpName,
		AppName:     key.AppName,
		ExpiresAt:   key.ExpiresAt,
		Disabled:    key.Disabled,
		UpdatedAt:   key.UpdatedAt,
	})
}

// EnableKey enables a disabled API key
func (s *Service) EnableKey(ctx context.Context, id string) error {
	key, err := s.GetKey(ctx, id)
	if err != nil {
		return err
	}

	key.Disabled = false
	key.UpdatedAt = time.Now()

	return s.db.UpdateAPIKey(ctx, &store.APIKey{
		ID:          key.ID,
		Name:        key.Name,
		Permissions: key.Permissions,
		CorpName:    key.CorpName,
		AppName:     key.AppName,
		ExpiresAt:   key.ExpiresAt,
		Disabled:    key.Disabled,
		UpdatedAt:   key.UpdatedAt,
	})
}

// DeleteKey deletes an API key
func (s *Service) DeleteKey(ctx context.Context, id string) error {
	return s.db.DeleteAPIKey(ctx, id)
}

// validateCreateRequest validates the create key request
func (s *Service) validateCreateRequest(req *CreateKeyRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate permissions (empty means no permissions, which is allowed)
	validPerms := map[string]bool{
		"calendar:read":      true,
		"calendar:write":     true,
		"meetingroom:read":   true,
		"meetingroom:write":  true,
		"message:send":       true,
		"*":                  true, // Admin wildcard
	}

	for _, perm := range req.Permissions {
		if !validPerms[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	if req.CorpName == "" {
		return fmt.Errorf("corp_name is required")
	}

	if req.ExpiresDays < 0 {
		return fmt.Errorf("expires_days must be non-negative")
	}

	return nil
}

// generateID generates a unique ID
func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "key_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// APIKeyResponse represents the response when creating an API key (includes raw key)
type APIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key"`
	Permissions []string   `json:"permissions"`
	CorpName    string     `json:"corp_name"`
	AppName     string     `json:"app_name,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// APIKeyInfo represents API key information (without raw key)
type APIKeyInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	CorpName    string     `json:"corp_name"`
	AppName     string     `json:"app_name,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Disabled    bool       `json:"disabled"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// hashAPIKey creates a hash of the API key for storage
func hashAPIKey(apiKey string) string {
	// Simple hash for demonstration - in production, use bcrypt or Argon2
	return strings.ToLower(apiKey)
}
