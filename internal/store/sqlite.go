package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite implements the Database interface for SQLite
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite database instance
func NewSQLite(dsn string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite doesn't support multiple writers
	db.SetMaxIdleConns(1)

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLite{db: db}, nil
}

// runMigrations creates all necessary tables
func runMigrations(db *sql.DB) error {
	migrations := []string{
		// Admin Users table
		`CREATE TABLE IF NOT EXISTS admin_users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			display_name TEXT,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username)`,

		// API Keys table
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			key_hash TEXT NOT NULL UNIQUE,
			permissions TEXT NOT NULL,
			corp_name TEXT NOT NULL,
			app_name TEXT,
			expires_at TIMESTAMP,
			disabled BOOLEAN NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_name ON api_keys(name)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_corp_app ON api_keys(corp_name, app_name)`,

		// WeCom Corps table
		`CREATE TABLE IF NOT EXISTS wecom_corps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			corp_id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_wecom_corps_name ON wecom_corps(name)`,

		// WeCom Apps table
		`CREATE TABLE IF NOT EXISTS wecom_apps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			corp_name TEXT NOT NULL,
			app_type TEXT NOT NULL DEFAULT 'app',
			agent_id INTEGER NOT NULL DEFAULT 0,
			bot_id TEXT,
			secret_enc TEXT NOT NULL,
			nonce TEXT NOT NULL,
			access_token TEXT,
			token_expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, corp_name),
			FOREIGN KEY(corp_name) REFERENCES wecom_corps(name) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_wecom_apps_corp_name ON wecom_apps(corp_name)`,

		// Audit Logs table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			protocol TEXT NOT NULL DEFAULT 'http',
			api_key_id TEXT,
			api_key_name TEXT,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			query TEXT,
			body TEXT,
			status_code INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			client_ip TEXT,
			error_msg TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_api_key_name ON audit_logs(api_key_name)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_path ON audit_logs(path)`,

		// Hourly Stats table
		`CREATE TABLE IF NOT EXISTS stats_hourly (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hour TIMESTAMP NOT NULL,
			total_count INTEGER NOT NULL DEFAULT 0,
			error_count INTEGER NOT NULL DEFAULT 0,
			key_name TEXT NOT NULL DEFAULT '',
			UNIQUE(hour, key_name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_stats_hourly_time ON stats_hourly(hour, key_name)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %s: %w", migration, err)
		}
	}

	// Alter-table migrations for Bot support (safe to run multiple times)
	alterMigrations := []string{
		"ALTER TABLE wecom_apps ADD COLUMN app_type TEXT NOT NULL DEFAULT 'app'",
		"ALTER TABLE wecom_apps ADD COLUMN bot_id TEXT",
	}
	for _, m := range alterMigrations {
		db.Exec(m) // ignore "duplicate column" errors
	}

	// Enable foreign key constraints (SQLite defaults to OFF)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}

// Ping checks if the database is accessible
func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the database connection
func (s *SQLite) Close() error {
	return s.db.Close()
}

// API Key operations

func (s *SQLite) CreateAPIKey(ctx context.Context, key *APIKey) error {
	permissionsJSON, err := json.Marshal(key.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	now := time.Now()
	query := `
		INSERT INTO api_keys (id, name, key_hash, permissions, corp_name, app_name, expires_at, disabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		key.ID, key.Name, key.KeyHash, string(permissionsJSON),
		key.CorpName, key.AppName, key.ExpiresAt, key.Disabled,
		now, now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("failed to create api key: %w", err)
	}

	return nil
}

func (s *SQLite) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error) {
	query := `
		SELECT id, name, key_hash, permissions, corp_name, app_name, expires_at, disabled, created_at, updated_at
		FROM api_keys WHERE key_hash = ?
	`

	row := s.db.QueryRowContext(ctx, query, hash)
	var key APIKey
	var permissionsJSON string
	var appName sql.NullString

	err := row.Scan(
		&key.ID, &key.Name, &key.KeyHash, &permissionsJSON,
		&key.CorpName, &appName, &key.ExpiresAt, &key.Disabled,
		&key.CreatedAt, &key.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}

	if err := json.Unmarshal([]byte(permissionsJSON), &key.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if appName.Valid {
		key.AppName = appName.String
	}

	return &key, nil
}

func (s *SQLite) ListAPIKeys(ctx context.Context, opts ListOptions) ([]*APIKey, string, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}

	query := `
		SELECT id, name, key_hash, permissions, corp_name, app_name, expires_at, disabled, created_at, updated_at
		FROM api_keys
		WHERE 1=1
	`
	args := []interface{}{}

	if opts.Disabled != nil {
		query += " AND disabled = ?"
		args = append(args, *opts.Disabled)
	}

	if opts.Cursor != "" {
		query += " AND name > ?"
		args = append(args, opts.Cursor)
	}

	query += " ORDER BY name ASC LIMIT ?"
	args = append(args, opts.Limit+1) // Fetch one extra to check for more results

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var key APIKey
		var permissionsJSON string
		var appName sql.NullString

		err := rows.Scan(
			&key.ID, &key.Name, &key.KeyHash, &permissionsJSON,
			&key.CorpName, &appName, &key.ExpiresAt, &key.Disabled,
			&key.CreatedAt, &key.UpdatedAt,
		)

		if err != nil {
			return nil, "", fmt.Errorf("failed to scan api key: %w", err)
		}

		if err := json.Unmarshal([]byte(permissionsJSON), &key.Permissions); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		if appName.Valid {
			key.AppName = appName.String
		}

		keys = append(keys, &key)
	}

	var nextCursor string
	if len(keys) > opts.Limit {
		keys = keys[:opts.Limit]
		nextCursor = keys[len(keys)-1].Name
	}

	return keys, nextCursor, nil
}

func (s *SQLite) UpdateAPIKey(ctx context.Context, key *APIKey) error {
	permissionsJSON, err := json.Marshal(key.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		UPDATE api_keys
		SET permissions = ?, corp_name = ?, app_name = ?, expires_at = ?, disabled = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		string(permissionsJSON), key.CorpName, key.AppName,
		key.ExpiresAt, key.Disabled, time.Now(), key.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update api key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLite) DeleteAPIKey(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// WeCom Corp operations

func (s *SQLite) CreateWeComCorp(ctx context.Context, corp *WeComCorp) error {
	query := `
		INSERT INTO wecom_corps (id, name, corp_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		corp.ID, corp.Name, corp.CorpID, now, now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("failed to create wecom corp: %w", err)
	}

	return nil
}

func (s *SQLite) GetWeComCorpByName(ctx context.Context, name string) (*WeComCorp, error) {
	query := `SELECT id, name, corp_id, created_at, updated_at FROM wecom_corps WHERE name = ?`

	row := s.db.QueryRowContext(ctx, query, name)
	var corp WeComCorp

	err := row.Scan(&corp.ID, &corp.Name, &corp.CorpID, &corp.CreatedAt, &corp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wecom corp: %w", err)
	}

	return &corp, nil
}

func (s *SQLite) ListWeComCorps(ctx context.Context) ([]*WeComCorp, error) {
	query := `SELECT id, name, corp_id, created_at, updated_at FROM wecom_corps ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list wecom corps: %w", err)
	}
	defer rows.Close()

	var corps []*WeComCorp
	for rows.Next() {
		var corp WeComCorp
		if err := rows.Scan(&corp.ID, &corp.Name, &corp.CorpID, &corp.CreatedAt, &corp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan wecom corp: %w", err)
		}
		corps = append(corps, &corp)
	}

	return corps, nil
}

func (s *SQLite) UpdateWeComCorp(ctx context.Context, corp *WeComCorp) error {
	query := `UPDATE wecom_corps SET corp_id = ?, updated_at = ? WHERE name = ?`

	result, err := s.db.ExecContext(ctx, query, corp.CorpID, time.Now(), corp.Name)
	if err != nil {
		return fmt.Errorf("failed to update wecom corp: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLite) DeleteWeComCorp(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM wecom_corps WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete wecom corp: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// WeCom App operations

func (s *SQLite) CreateWeComApp(ctx context.Context, app *WeComApp) error {
	query := `
		INSERT INTO wecom_apps (id, name, corp_name, app_type, agent_id, bot_id, secret_enc, nonce, access_token, token_expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		app.ID, app.Name, app.CorpName, app.AppType, app.AgentID, app.BotID, app.SecretEnc, app.Nonce,
		app.AccessToken, app.TokenExpiresAt, now, now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("failed to create wecom app: %w", err)
	}

	return nil
}

func (s *SQLite) GetWeComApp(ctx context.Context, corpName, appName string) (*WeComApp, error) {
	query := `
		SELECT id, name, corp_name, app_type, agent_id, bot_id, secret_enc, nonce, access_token, token_expires_at, created_at, updated_at
		FROM wecom_apps WHERE corp_name = ? AND name = ?
	`

	row := s.db.QueryRowContext(ctx, query, corpName, appName)
	var app WeComApp

	err := row.Scan(
		&app.ID, &app.Name, &app.CorpName, &app.AppType, &app.AgentID, &app.BotID, &app.SecretEnc, &app.Nonce,
		&app.AccessToken, &app.TokenExpiresAt, &app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wecom app: %w", err)
	}

	return &app, nil
}

func (s *SQLite) ListWeComApps(ctx context.Context, corpName string) ([]*WeComApp, error) {
	query := `
		SELECT id, name, corp_name, app_type, agent_id, bot_id, secret_enc, nonce, access_token, token_expires_at, created_at, updated_at
		FROM wecom_apps WHERE corp_name = ? ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, corpName)
	if err != nil {
		return nil, fmt.Errorf("failed to list wecom apps: %w", err)
	}
	defer rows.Close()

	var apps []*WeComApp
	for rows.Next() {
		var app WeComApp
		if err := rows.Scan(
			&app.ID, &app.Name, &app.CorpName, &app.AppType, &app.AgentID, &app.BotID, &app.SecretEnc, &app.Nonce,
			&app.AccessToken, &app.TokenExpiresAt, &app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan wecom app: %w", err)
		}
		apps = append(apps, &app)
	}

	return apps, nil
}

func (s *SQLite) UpdateWeComApp(ctx context.Context, app *WeComApp) error {
	query := `
		UPDATE wecom_apps
		SET name = ?, corp_name = ?, app_type = ?, agent_id = ?, bot_id = ?, secret_enc = ?, nonce = ?, access_token = ?, token_expires_at = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		app.Name, app.CorpName, app.AppType, app.AgentID, app.BotID, app.SecretEnc, app.Nonce,
		app.AccessToken, app.TokenExpiresAt, time.Now(), app.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update wecom app: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLite) DeleteWeComApp(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM wecom_apps WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete wecom app: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLite) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	query := `
		UPDATE wecom_apps
		SET access_token = ?, token_expires_at = ?, updated_at = ?
		WHERE corp_name = ? AND name = ?
	`

	result, err := s.db.ExecContext(ctx, query, token, expiresAt, time.Now(), corpName, appName)
	if err != nil {
		return fmt.Errorf("failed to update app token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Audit log operations

func (s *SQLite) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit_logs (timestamp, protocol, api_key_id, api_key_name, method, path, query, body, status_code, duration_ms, client_ip, error_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		log.Timestamp, log.Protocol, log.APIKeyID, log.APIKeyName,
		log.Method, log.Path, log.Query, log.Body, log.StatusCode,
		log.DurationMs, log.ClientIP, log.ErrorMsg,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (s *SQLite) QueryAuditLogs(ctx context.Context, opts AuditQueryOptions) ([]*AuditLog, string, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}

	query := `
		SELECT id, timestamp, protocol, api_key_id, api_key_name, method, path, query, body, status_code, duration_ms, client_ip, error_msg
		FROM audit_logs WHERE 1=1
	`
	args := []interface{}{}

	if opts.APIKeyName != nil {
		query += " AND api_key_name = ?"
		args = append(args, *opts.APIKeyName)
	}

	if opts.Method != nil {
		query += " AND method = ?"
		args = append(args, *opts.Method)
	}

	if opts.Path != nil {
		query += " AND path LIKE ?"
		args = append(args, "%"+*opts.Path+"%")
	}

	if opts.StatusCode != nil {
		query += " AND status_code = ?"
		args = append(args, *opts.StatusCode)
	}

	if opts.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, *opts.StartTime)
	}

	if opts.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, *opts.EndTime)
	}

	if opts.Cursor != "" {
		query += " AND id < ?"
		args = append(args, opts.Cursor)
	}

	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, opts.Limit+1) // Fetch one extra to check for more results

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID, &log.Timestamp, &log.Protocol, &log.APIKeyID, &log.APIKeyName,
			&log.Method, &log.Path, &log.Query, &log.Body, &log.StatusCode,
			&log.DurationMs, &log.ClientIP, &log.ErrorMsg,
		)

		if err != nil {
			return nil, "", fmt.Errorf("failed to scan audit log: %w", err)
		}

		logs = append(logs, &log)
	}

	var nextCursor string
	if len(logs) > opts.Limit {
		logs = logs[:opts.Limit]
		nextCursor = fmt.Sprintf("%d", logs[len(logs)-1].ID)
	}

	return logs, nextCursor, nil
}

// Statistics operations

func (s *SQLite) CreateHourlyStats(ctx context.Context, stats *HourlyStats) error {
	query := `
		INSERT INTO stats_hourly (hour, total_count, error_count, key_name)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		stats.Hour, stats.TotalCount, stats.ErrorCount, stats.KeyName,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("failed to create hourly stats: %w", err)
	}

	return nil
}

func (s *SQLite) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*HourlyStats, error) {
	query := `
		SELECT id, hour, total_count, error_count, key_name
		FROM stats_hourly
		WHERE key_name = ? AND hour >= ? AND hour <= ?
		ORDER BY hour ASC
	`

	rows, err := s.db.QueryContext(ctx, query, keyName, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}
	defer rows.Close()

	var stats []*HourlyStats
	for rows.Next() {
		var stat HourlyStats
		if err := rows.Scan(&stat.ID, &stat.Hour, &stat.TotalCount, &stat.ErrorCount, &stat.KeyName); err != nil {
			return nil, fmt.Errorf("failed to scan hourly stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

func (s *SQLite) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	hour := timestamp.Truncate(time.Hour).UTC()

	query := `
		INSERT INTO stats_hourly (hour, total_count, error_count, key_name)
		VALUES (?, 1, ?, ?)
		ON CONFLICT(hour, key_name) DO UPDATE SET
			total_count = total_count + 1,
			error_count = error_count + ?
	`

	errorCount := 0
	if isError {
		errorCount = 1
	}

	_, err := s.db.ExecContext(ctx, query, hour, errorCount, keyName, errorCount)
	if err != nil {
		return fmt.Errorf("failed to increment hourly stats: %w", err)
	}

	return nil
}

// Admin User operations

func (s *SQLite) CreateAdminUser(ctx context.Context, user *AdminUser) error {
	query := `
		INSERT INTO admin_users (id, username, password, display_name, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Password, user.DisplayName, user.IsActive, now, now,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicate
		}
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

func (s *SQLite) GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error) {
	query := `SELECT id, username, password, display_name, is_active, created_at, updated_at FROM admin_users WHERE username = ?`

	row := s.db.QueryRowContext(ctx, query, username)
	var user AdminUser

	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.DisplayName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}

	return &user, nil
}

func (s *SQLite) ListAdminUsers(ctx context.Context) ([]*AdminUser, error) {
	query := `SELECT id, username, password, display_name, is_active, created_at, updated_at FROM admin_users ORDER BY username`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list admin users: %w", err)
	}
	defer rows.Close()

	var users []*AdminUser
	for rows.Next() {
		var user AdminUser
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.DisplayName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan admin user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SQLite) UpdateAdminUser(ctx context.Context, user *AdminUser) error {
	query := `UPDATE admin_users SET password = ?, display_name = ?, is_active = ?, updated_at = ? WHERE username = ?`

	result, err := s.db.ExecContext(ctx, query, user.Password, user.DisplayName, user.IsActive, time.Now(), user.Username)
	if err != nil {
		return fmt.Errorf("failed to update admin user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SQLite) DeleteAdminUser(ctx context.Context, username string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM admin_users WHERE username = ?", username)
	if err != nil {
		return fmt.Errorf("failed to delete admin user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// WeCom Corp operations - add GetWeComCorpByID

func (s *SQLite) GetWeComCorpByID(ctx context.Context, id string) (*WeComCorp, error) {
	query := `SELECT id, name, corp_id, created_at, updated_at FROM wecom_corps WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	var corp WeComCorp

	err := row.Scan(&corp.ID, &corp.Name, &corp.CorpID, &corp.CreatedAt, &corp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wecom corp: %w", err)
	}

	return &corp, nil
}

// WeCom App operations - add GetWeComAppByID and DeleteWeComApp by ID

func (s *SQLite) GetWeComAppByID(ctx context.Context, id string) (*WeComApp, error) {
	query := `
		SELECT id, name, corp_name, app_type, agent_id, bot_id, secret_enc, nonce, access_token, token_expires_at, created_at, updated_at
		FROM wecom_apps WHERE id = ?
	`

	row := s.db.QueryRowContext(ctx, query, id)
	var app WeComApp

	err := row.Scan(
		&app.ID, &app.Name, &app.CorpName, &app.AppType, &app.AgentID, &app.BotID, &app.SecretEnc, &app.Nonce,
		&app.AccessToken, &app.TokenExpiresAt, &app.CreatedAt, &app.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wecom app: %w", err)
	}

	return &app, nil
}
