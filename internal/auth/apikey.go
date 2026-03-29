package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"wecom-gateway/internal/store"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid or missing API key")
	ErrAPIKeyDisabled   = errors.New("API key is disabled")
	ErrAPIKeyExpired    = errors.New("API key has expired")
	ErrPermissionDenied = errors.New("permission denied")
)

// AuthContext represents authentication context injected into requests
type AuthContext struct {
	KeyID       string
	KeyName     string
	Permissions []string
	CorpName    string
	AppName     string
	IsAdmin     bool
}

// Authenticator defines the interface for authentication
type Authenticator interface {
	Authenticate(ctx context.Context, rawKey string) (*AuthContext, error)
}

// APIKeyAuthenticator implements API key authentication
type APIKeyAuthenticator struct {
	db store.Database
}

// NewAPIKeyAuthenticator creates a new API key authenticator
func NewAPIKeyAuthenticator(db store.Database) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{db: db}
}

// Authenticate validates an API key and returns authentication context
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, rawKey string) (*AuthContext, error) {
	// Validate API key format
	if !strings.HasPrefix(rawKey, "wgk_") {
		return nil, ErrInvalidAPIKey
	}

	// Hash the API key to compare with stored hash
	keyHash := hashAPIKey(rawKey)

	// Look up the API key in database
	key, err := a.db.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrInvalidAPIKey
		}
		return nil, err
	}

	// Check if key is disabled
	if key.Disabled {
		return nil, ErrAPIKeyDisabled
	}

	// Check if key is expired
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, ErrAPIKeyExpired
	}

	// Determine if this is an admin key (has wildcard permission)
	isAdmin := false
	for _, perm := range key.Permissions {
		if perm == "*" {
			isAdmin = true
			break
		}
	}

	return &AuthContext{
		KeyID:       key.ID,
		KeyName:     key.Name,
		Permissions: key.Permissions,
		CorpName:    key.CorpName,
		AppName:     key.AppName,
		IsAdmin:     isAdmin,
	}, nil
}

// HasPermission checks if the auth context has the required permission
func (ac *AuthContext) HasPermission(required string) bool {
	if ac.IsAdmin {
		return true
	}

	for _, permission := range ac.Permissions {
		if permission == required || permission == "*" {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the auth context has any of the required permissions
func (ac *AuthContext) HasAnyPermission(required []string) bool {
	if ac.IsAdmin {
		return true
	}

	for _, req := range required {
		for _, permission := range ac.Permissions {
			if permission == req || permission == "*" {
				return true
			}
		}
	}
	return false
}

// hashAPIKey creates a hash of the API key for storage
func hashAPIKey(apiKey string) string {
	// Simple hash for demonstration - in production, use bcrypt or Argon2
	// For now, we'll use a basic hash function
	return strings.ToLower(apiKey)
}
