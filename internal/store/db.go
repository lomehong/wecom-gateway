package store

import (
	"context"
	"time"
)

// Database defines the interface for database operations
type Database interface {
	// Admin User operations
	CreateAdminUser(ctx context.Context, user *AdminUser) error
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
	ListAdminUsers(ctx context.Context) ([]*AdminUser, error)
	UpdateAdminUser(ctx context.Context, user *AdminUser) error
	DeleteAdminUser(ctx context.Context, username string) error

	// API Key operations
	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error)
	ListAPIKeys(ctx context.Context, opts ListOptions) ([]*APIKey, string, error)
	UpdateAPIKey(ctx context.Context, key *APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error

	// WeCom Corp operations
	CreateWeComCorp(ctx context.Context, corp *WeComCorp) error
	GetWeComCorpByName(ctx context.Context, name string) (*WeComCorp, error)
	GetWeComCorpByID(ctx context.Context, id string) (*WeComCorp, error)
	ListWeComCorps(ctx context.Context) ([]*WeComCorp, error)
	UpdateWeComCorp(ctx context.Context, corp *WeComCorp) error
	DeleteWeComCorp(ctx context.Context, name string) error

	// WeCom App operations
	CreateWeComApp(ctx context.Context, app *WeComApp) error
	GetWeComApp(ctx context.Context, corpName, appName string) (*WeComApp, error)
	GetWeComAppByID(ctx context.Context, id string) (*WeComApp, error)
	ListWeComApps(ctx context.Context, corpName string) ([]*WeComApp, error)
	UpdateWeComApp(ctx context.Context, app *WeComApp) error
	DeleteWeComApp(ctx context.Context, id string) error
	UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error

	// Audit log operations
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	QueryAuditLogs(ctx context.Context, opts AuditQueryOptions) ([]*AuditLog, string, error)

	// Statistics operations
	CreateHourlyStats(ctx context.Context, stats *HourlyStats) error
	GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*HourlyStats, error)
	IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error

	// Health check
	Ping(ctx context.Context) error
	Close() error
}

// AdminUser represents an admin user entity
type AdminUser struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"` // bcrypt hash, never expose in JSON
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIKey represents an API key entity
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	KeyHash     string    `json:"-"` // Never expose hash in JSON
	Permissions []string  `json:"permissions"`
	CorpName    string    `json:"corp_name"`
	AppName     string    `json:"app_name"` // nil for admin keys
	ExpiresAt   *time.Time `json:"expires_at"`
	Disabled    bool      `json:"disabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WeComCorp represents a WeChat Work corporation entity
type WeComCorp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CorpID    string    `json:"corp_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WeComApp represents a WeChat Work application entity
type WeComApp struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	CorpName       string     `json:"corp_name"`
	AgentID        int64      `json:"agent_id"`
	SecretEnc      string     `json:"-"` // Encrypted secret
	Nonce          string     `json:"-"` // Nonce for AES-GCM
	AccessToken    *string    `json:"access_token,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          int64      `json:"id"`
	Timestamp   time.Time  `json:"timestamp"`
	Protocol    string     `json:"protocol"` // "http" or "grpc"
	APIKeyID    *string    `json:"api_key_id,omitempty"`
	APIKeyName  *string    `json:"api_key_name,omitempty"`
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	Query       *string    `json:"query,omitempty"`
	Body        *string    `json:"body,omitempty"`
	StatusCode  int        `json:"status_code"`
	DurationMs  int        `json:"duration_ms"`
	ClientIP    *string    `json:"client_ip,omitempty"`
	ErrorMsg    *string    `json:"error_msg,omitempty"`
}

// HourlyStats represents hourly statistics for monitoring
type HourlyStats struct {
	ID        int64     `json:"id"`
	Hour      time.Time `json:"hour"`
	TotalCount int64    `json:"total_count"`
	ErrorCount int64    `json:"error_count"`
	KeyName   string    `json:"key_name"`
}

// ListOptions represents common list query options
type ListOptions struct {
	Cursor   string // Pagination cursor
	Limit    int    // Maximum items to return (default: 50, max: 100)
	Disabled *bool  // Filter by disabled status (for API keys)
}

// AuditQueryOptions represents audit log query options
type AuditQueryOptions struct {
	APIKeyName *string    // Filter by API key name
	Method     *string    // Filter by HTTP method
	Path       *string    // Filter by request path
	StatusCode *int       // Filter by status code
	StartTime  *time.Time // Filter by start time
	EndTime    *time.Time // Filter by end time
	Cursor     string     // Pagination cursor
	Limit      int        // Maximum items to return (default: 50, max: 100)
}

// Config holds database configuration
type Config struct {
	Driver string // "sqlite" or "postgres"
	DSN    string // Data source name
}

// New creates a new database instance based on the configuration
func New(cfg *Config) (Database, error) {
	switch cfg.Driver {
	case "sqlite":
		return NewSQLite(cfg.DSN)
	case "postgres":
		return NewPostgreSQL(cfg.DSN)
	default:
		return nil, ErrInvalidDriver
	}
}

// Errors
var (
	ErrNotFound         = &ValidationError{Field: "record", Message: "record not found"}
	ErrDuplicate        = &ValidationError{Field: "record", Message: "duplicate record"}
	ErrInvalidCursor    = &ValidationError{Field: "cursor", Message: "invalid cursor"}
	ErrInvalidDriver    = &ValidationError{Field: "driver", Message: "invalid database driver"}
	ErrInvalidLimit     = &ValidationError{Field: "limit", Message: "limit must be between 1 and 100"}
	ErrInvalidTimeRange = &ValidationError{Field: "time_range", Message: "invalid time range"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
