package audit

import (
	"context"
	"time"

	"wecom-gateway/internal/store"
)

// Logger records audit logs
type Logger struct {
	db store.Database
}

// NewLogger creates a new audit logger
func NewLogger(db store.Database) *Logger {
	return &Logger{db: db}
}

// Log records an audit log entry asynchronously
func (l *Logger) Log(ctx context.Context, entry *LogEntry) error {
	log := &store.AuditLog{
		Timestamp:  entry.Timestamp,
		Protocol:   entry.Protocol,
		APIKeyID:   entry.APIKeyID,
		APIKeyName: entry.APIKeyName,
		Method:     entry.Method,
		Path:       entry.Path,
		Query:      entry.Query,
		Body:       entry.Body,
		StatusCode: entry.StatusCode,
		DurationMs: entry.DurationMs,
		ClientIP:   entry.ClientIP,
		ErrorMsg:   entry.ErrorMsg,
	}

	if err := l.db.CreateAuditLog(ctx, log); err != nil {
		return err
	}

	// Update statistics
	return l.updateStats(ctx, entry)
}

// LogEntry represents an audit log entry
type LogEntry struct {
	Timestamp  time.Time
	Protocol   string // "http" or "grpc"
	APIKeyID   *string
	APIKeyName *string
	Method     string
	Path       string
	Query      *string
	Body       *string
	StatusCode int
	DurationMs int
	ClientIP   *string
	ErrorMsg   *string
}

// updateStats updates hourly statistics
func (l *Logger) updateStats(ctx context.Context, entry *LogEntry) error {
	keyName := ""
	if entry.APIKeyName != nil {
		keyName = *entry.APIKeyName
	}

	isError := entry.StatusCode >= 400

	return l.db.IncrementHourlyStats(ctx, keyName, entry.Timestamp, isError)
}

// Querier handles audit log queries
type Querier struct {
	db store.Database
}

// NewQuerier creates a new audit log querier
func NewQuerier(db store.Database) *Querier {
	return &Querier{db: db}
}

// Query retrieves audit logs with filtering
func (q *Querier) Query(ctx context.Context, opts *QueryOptions) ([]*store.AuditLog, string, error) {
	auditOpts := store.AuditQueryOptions{
		Limit: opts.Limit,
	}

	if opts.APIKeyName != nil {
		auditOpts.APIKeyName = opts.APIKeyName
	}
	if opts.Method != nil {
		auditOpts.Method = opts.Method
	}
	if opts.Path != nil {
		auditOpts.Path = opts.Path
	}
	if opts.StatusCode != nil {
		auditOpts.StatusCode = opts.StatusCode
	}
	if opts.StartTime != nil {
		auditOpts.StartTime = opts.StartTime
	}
	if opts.EndTime != nil {
		auditOpts.EndTime = opts.EndTime
	}
	if opts.Cursor != "" {
		auditOpts.Cursor = opts.Cursor
	}

	return q.db.QueryAuditLogs(ctx, auditOpts)
}

// QueryOptions represents audit log query options
type QueryOptions struct {
	APIKeyName *string
	Method     *string
	Path       *string
	StatusCode *int
	StartTime  *time.Time
	EndTime    *time.Time
	Cursor     string
	Limit      int
}
