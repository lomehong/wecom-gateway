package store

import (
	"context"
	"testing"
	"time"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "test_field: test message"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		field string
		msg   string
	}{
		{
			name:  "ErrNotFound",
			err:   ErrNotFound,
			field: "record",
			msg:   "record not found",
		},
		{
			name:  "ErrDuplicate",
			err:   ErrDuplicate,
			field: "record",
			msg:   "duplicate record",
		},
		{
			name:  "ErrInvalidCursor",
			err:   ErrInvalidCursor,
			field: "cursor",
			msg:   "invalid cursor",
		},
		{
			name:  "ErrInvalidDriver",
			err:   ErrInvalidDriver,
			field: "driver",
			msg:   "invalid database driver",
		},
		{
			name:  "ErrInvalidLimit",
			err:   ErrInvalidLimit,
			field: "limit",
			msg:   "limit must be between 1 and 100",
		},
		{
			name:  "ErrInvalidTimeRange",
			err:   ErrInvalidTimeRange,
			field: "time_range",
			msg:   "invalid time range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErr, ok := tt.err.(*ValidationError)
			if !ok {
				t.Errorf("expected ValidationError type")
				return
			}
			if validationErr.Field != tt.field {
				t.Errorf("expected field %s, got %s", tt.field, validationErr.Field)
			}
			if validationErr.Message != tt.msg {
				t.Errorf("expected message %s, got %s", tt.msg, validationErr.Message)
			}
		})
	}
}

func TestAPIKey_Timestamps(t *testing.T) {
	now := time.Now()
	key := &APIKey{
		ID:        "test-id",
		Name:      "test-key",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if key.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if key.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestAPIKey_ExpiresAt(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		isNil     bool
	}{
		{
			name:      "with expiration",
			expiresAt: &future,
			isNil:     false,
		},
		{
			name:      "without expiration",
			expiresAt: nil,
			isNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := &APIKey{
				ExpiresAt: tt.expiresAt,
			}

			if tt.isNil && key.ExpiresAt != nil {
				t.Error("expected nil ExpiresAt")
			}
			if !tt.isNil && key.ExpiresAt == nil {
				t.Error("expected non-nil ExpiresAt")
			}
		})
	}
}

func TestWeComCorp_Fields(t *testing.T) {
	corp := &WeComCorp{
		ID:        "test-id",
		Name:      "test-corp",
		CorpID:    "test-corp-id",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if corp.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", corp.ID)
	}
	if corp.Name != "test-corp" {
		t.Errorf("expected Name test-corp, got %s", corp.Name)
	}
	if corp.CorpID != "test-corp-id" {
		t.Errorf("expected CorpID test-corp-id, got %s", corp.CorpID)
	}
}

func TestWeComApp_Fields(t *testing.T) {
	agentID := int64(123456)
	token := "test-token"
	expiresAt := time.Now().Add(2 * time.Hour)

	app := &WeComApp{
		ID:             "test-id",
		Name:           "test-app",
		CorpName:       "test-corp",
		AgentID:        agentID,
		SecretEnc:      "encrypted-secret",
		AccessToken:    &token,
		TokenExpiresAt: &expiresAt,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if app.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", app.ID)
	}
	if app.Name != "test-app" {
		t.Errorf("expected Name test-app, got %s", app.Name)
	}
	if app.AgentID != agentID {
		t.Errorf("expected AgentID %d, got %d", agentID, app.AgentID)
	}
	if app.AccessToken == nil || *app.AccessToken != token {
		t.Error("AccessToken not set correctly")
	}
}

func TestAuditLog_Fields(t *testing.T) {
	now := time.Now()
	method := "GET"
	path := "/api/v1/test"
	statusCode := 200
	clientIP := "127.0.0.1"
	errorMsg := "internal error"

	log := &AuditLog{
		ID:         1,
		Timestamp:  now,
		Protocol:   "http",
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		ClientIP:   &clientIP,
		ErrorMsg:   &errorMsg,
	}

	if log.Method != method {
		t.Errorf("expected Method %s, got %s", method, log.Method)
	}
	if log.Path != path {
		t.Errorf("expected Path %s, got %s", path, log.Path)
	}
	if log.StatusCode != statusCode {
		t.Errorf("expected StatusCode %d, got %d", statusCode, log.StatusCode)
	}
	if log.ClientIP == nil || *log.ClientIP != clientIP {
		t.Error("ClientIP not set correctly")
	}
	if log.ErrorMsg == nil || *log.ErrorMsg != errorMsg {
		t.Error("ErrorMsg not set correctly")
	}
}

func TestListOptions_DefaultValues(t *testing.T) {
	opts := ListOptions{
		Cursor: "",
		Limit:  50,
	}

	if opts.Limit != 50 {
		t.Errorf("expected Limit 50, got %d", opts.Limit)
	}
	if opts.Cursor != "" {
		t.Errorf("expected empty Cursor, got %s", opts.Cursor)
	}
	if opts.Disabled != nil {
		t.Error("expected nil Disabled")
	}
}

func TestAuditQueryOptions_Filters(t *testing.T) {
	apiKeyName := "test-key"
	method := "POST"
	path := "/api/v1/create"
	statusCode := 201
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	opts := AuditQueryOptions{
		APIKeyName: &apiKeyName,
		Method:     &method,
		Path:       &path,
		StatusCode: &statusCode,
		StartTime:  &startTime,
		EndTime:    &endTime,
		Limit:      100,
		Cursor:     "test-cursor",
	}

	if opts.APIKeyName == nil || *opts.APIKeyName != apiKeyName {
		t.Error("APIKeyName not set correctly")
	}
	if opts.Method == nil || *opts.Method != method {
		t.Error("Method not set correctly")
	}
	if opts.Path == nil || *opts.Path != path {
		t.Error("Path not set correctly")
	}
	if opts.StatusCode == nil || *opts.StatusCode != statusCode {
		t.Error("StatusCode not set correctly")
	}
	if opts.StartTime == nil {
		t.Error("StartTime not set")
	}
	if opts.EndTime == nil {
		t.Error("EndTime not set")
	}
	if opts.Limit != 100 {
		t.Errorf("expected Limit 100, got %d", opts.Limit)
	}
	if opts.Cursor != "test-cursor" {
		t.Errorf("expected Cursor test-cursor, got %s", opts.Cursor)
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Driver: "sqlite",
		DSN:    "test.db",
	}

	if cfg.Driver != "sqlite" {
		t.Errorf("expected Driver sqlite, got %s", cfg.Driver)
	}
	if cfg.DSN != "test.db" {
		t.Errorf("expected DSN test.db, got %s", cfg.DSN)
	}
}

func TestHourlyStats_Fields(t *testing.T) {
	hour := time.Now().Truncate(time.Hour)
	stats := &HourlyStats{
		ID:         1,
		Hour:       hour,
		TotalCount: 1000,
		ErrorCount: 50,
		KeyName:    "test-key",
	}

	if stats.Hour != hour {
		t.Error("Hour not set correctly")
	}
	if stats.TotalCount != 1000 {
		t.Errorf("expected TotalCount 1000, got %d", stats.TotalCount)
	}
	if stats.ErrorCount != 50 {
		t.Errorf("expected ErrorCount 50, got %d", stats.ErrorCount)
	}
	if stats.KeyName != "test-key" {
		t.Errorf("expected KeyName test-key, got %s", stats.KeyName)
	}
}

// Test mock type for Database interface
type MockDB struct{}

func (m *MockDB) CreateAPIKey(ctx context.Context, key *APIKey) error {
	return nil
}
func (m *MockDB) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error) {
	return nil, nil
}
func (m *MockDB) ListAPIKeys(ctx context.Context, opts ListOptions) ([]*APIKey, string, error) {
	return nil, "", nil
}
func (m *MockDB) UpdateAPIKey(ctx context.Context, key *APIKey) error {
	return nil
}
func (m *MockDB) DeleteAPIKey(ctx context.Context, id string) error {
	return nil
}
func (m *MockDB) CreateWeComCorp(ctx context.Context, corp *WeComCorp) error {
	return nil
}
func (m *MockDB) GetWeComCorpByName(ctx context.Context, name string) (*WeComCorp, error) {
	return nil, nil
}

func (m *MockDB) GetWeComCorpByID(ctx context.Context, id string) (*WeComCorp, error) {
	return nil, nil
}
func (m *MockDB) ListWeComCorps(ctx context.Context) ([]*WeComCorp, error) {
	return nil, nil
}
func (m *MockDB) UpdateWeComCorp(ctx context.Context, corp *WeComCorp) error {
	return nil
}
func (m *MockDB) DeleteWeComCorp(ctx context.Context, name string) error {
	return nil
}
func (m *MockDB) CreateWeComApp(ctx context.Context, app *WeComApp) error {
	return nil
}
func (m *MockDB) GetWeComApp(ctx context.Context, corpName, appName string) (*WeComApp, error) {
	return nil, nil
}

func (m *MockDB) GetWeComAppByID(ctx context.Context, id string) (*WeComApp, error) {
	return nil, nil
}
func (m *MockDB) ListWeComApps(ctx context.Context, corpName string) ([]*WeComApp, error) {
	return nil, nil
}
func (m *MockDB) UpdateWeComApp(ctx context.Context, app *WeComApp) error {
	return nil
}
func (m *MockDB) DeleteWeComApp(ctx context.Context, id string) error {
	return nil
}
func (m *MockDB) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	return nil
}
func (m *MockDB) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	return nil
}

func (m *MockDB) CreateAdminUser(ctx context.Context, user *AdminUser) error {
	return nil
}

func (m *MockDB) GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error) {
	return nil, ErrNotFound
}

func (m *MockDB) ListAdminUsers(ctx context.Context) ([]*AdminUser, error) {
	return nil, nil
}

func (m *MockDB) UpdateAdminUser(ctx context.Context, user *AdminUser) error {
	return nil
}

func (m *MockDB) DeleteAdminUser(ctx context.Context, username string) error {
	return nil
}
func (m *MockDB) QueryAuditLogs(ctx context.Context, opts AuditQueryOptions) ([]*AuditLog, string, error) {
	return nil, "", nil
}
func (m *MockDB) CreateHourlyStats(ctx context.Context, stats *HourlyStats) error {
	return nil
}
func (m *MockDB) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*HourlyStats, error) {
	return nil, nil
}
func (m *MockDB) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	return nil
}
func (m *MockDB) Ping(ctx context.Context) error {
	return nil
}
func (m *MockDB) Close() error {
	return nil
}

func TestDatabase_Interface(t *testing.T) {
	// Test that Database interface can be implemented
	var _ Database = &MockDB{}
}
