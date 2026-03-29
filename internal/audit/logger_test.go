package audit

import (
	"context"
	"testing"
	"time"

	"wecom-gateway/internal/store"
)

// MockAuditStore implements store.Database for testing
type MockAuditStore struct {
	logs     []*store.AuditLog
	stats    []*store.HourlyStats
}

func NewMockAuditStore() *MockAuditStore {
	return &MockAuditStore{
		logs:  make([]*store.AuditLog, 0),
		stats: make([]*store.HourlyStats, 0),
	}
}

func (m *MockAuditStore) CreateAPIKey(ctx context.Context, key *store.APIKey) error {
	return nil
}

func (m *MockAuditStore) GetAPIKeyByHash(ctx context.Context, hash string) (*store.APIKey, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) ListAPIKeys(ctx context.Context, opts store.ListOptions) ([]*store.APIKey, string, error) {
	return []*store.APIKey{}, "", nil
}

func (m *MockAuditStore) UpdateAPIKey(ctx context.Context, key *store.APIKey) error {
	return nil
}

func (m *MockAuditStore) DeleteAPIKey(ctx context.Context, id string) error {
	return nil
}

func (m *MockAuditStore) CreateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockAuditStore) GetWeComCorpByName(ctx context.Context, name string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) GetWeComCorpByID(ctx context.Context, id string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) ListWeComCorps(ctx context.Context) ([]*store.WeComCorp, error) {
	return []*store.WeComCorp{}, nil
}

func (m *MockAuditStore) UpdateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockAuditStore) DeleteWeComCorp(ctx context.Context, name string) error {
	return nil
}

func (m *MockAuditStore) CreateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockAuditStore) GetWeComApp(ctx context.Context, corpName, appName string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) GetWeComAppByID(ctx context.Context, id string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) ListWeComApps(ctx context.Context, corpName string) ([]*store.WeComApp, error) {
	return []*store.WeComApp{}, nil
}

func (m *MockAuditStore) UpdateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockAuditStore) DeleteWeComApp(ctx context.Context, id string) error {
	return nil
}

func (m *MockAuditStore) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	return nil
}

func (m *MockAuditStore) CreateAuditLog(ctx context.Context, log *store.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockAuditStore) CreateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockAuditStore) GetAdminUserByUsername(ctx context.Context, username string) (*store.AdminUser, error) {
	return nil, store.ErrNotFound
}

func (m *MockAuditStore) ListAdminUsers(ctx context.Context) ([]*store.AdminUser, error) {
	return nil, nil
}

func (m *MockAuditStore) UpdateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockAuditStore) DeleteAdminUser(ctx context.Context, username string) error {
	return nil
}

func (m *MockAuditStore) QueryAuditLogs(ctx context.Context, opts store.AuditQueryOptions) ([]*store.AuditLog, string, error) {
	// Filter logs based on options
	filtered := make([]*store.AuditLog, 0)

	for _, log := range m.logs {
		// Filter by APIKeyName
		if opts.APIKeyName != nil {
			if log.APIKeyName == nil || *log.APIKeyName != *opts.APIKeyName {
				continue
			}
		}

		// Filter by Method
		if opts.Method != nil {
			if log.Method != *opts.Method {
				continue
			}
		}

		// Filter by Path
		if opts.Path != nil {
			if log.Path != *opts.Path {
				continue
			}
		}

		// Filter by StatusCode
		if opts.StatusCode != nil {
			if log.StatusCode != *opts.StatusCode {
				continue
			}
		}

		// Filter by StartTime
		if opts.StartTime != nil {
			if log.Timestamp.Before(*opts.StartTime) {
				continue
			}
		}

		// Filter by EndTime
		if opts.EndTime != nil {
			if log.Timestamp.After(*opts.EndTime) {
				continue
			}
		}

		filtered = append(filtered, log)
	}

	// Apply limit
	limit := opts.Limit
	if limit <= 0 {
		limit = 50 // default
	}
	if limit > len(filtered) {
		limit = len(filtered)
	}

	// Apply cursor pagination (simplified - just skip first N entries)
	start := 0
	if opts.Cursor != "" {
		// In real implementation, decode cursor to get start position
		// For mock, just skip some entries
		start = 5 // arbitrary skip for testing
	}
	if start > len(filtered) {
		start = len(filtered)
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	result := filtered[start:end]

	// Generate cursor for next page
	var cursor string
	if end < len(filtered) {
		cursor = "next_page_cursor"
	}

	return result, cursor, nil
}

func (m *MockAuditStore) CreateHourlyStats(ctx context.Context, stats *store.HourlyStats) error {
	m.stats = append(m.stats, stats)
	return nil
}

func (m *MockAuditStore) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*store.HourlyStats, error) {
	return m.stats, nil
}

func (m *MockAuditStore) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	return nil
}

func (m *MockAuditStore) Close() error {
	return nil
}

func (m *MockAuditStore) Ping(ctx context.Context) error {
	return nil
}

func TestNewLogger(t *testing.T) {
	mockStore := NewMockAuditStore()
	logger := NewLogger(mockStore)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.db == nil {
		t.Error("db field should not be nil")
	}
}

func TestNewQuerier(t *testing.T) {
	mockStore := NewMockAuditStore()
	querier := NewQuerier(mockStore)

	if querier == nil {
		t.Fatal("NewQuerier returned nil")
	}

	if querier.db == nil {
		t.Error("db field should not be nil")
	}
}

func TestLogger_Log(t *testing.T) {
	mockStore := NewMockAuditStore()
	logger := NewLogger(mockStore)
	ctx := context.Background()

	entry := &LogEntry{
		Timestamp:  time.Now(),
		Protocol:   "http",
		Method:     "GET",
		Path:       "/v1/schedules",
		StatusCode: 200,
		DurationMs: 100,
	}

	err := logger.Log(ctx, entry)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Verify log was created
	if len(mockStore.logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockStore.logs))
	}

	log := mockStore.logs[0]
	if log.Protocol != "http" {
		t.Errorf("Expected protocol 'http', got '%s'", log.Protocol)
	}

	if log.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", log.Method)
	}

	if log.Path != "/v1/schedules" {
		t.Errorf("Expected path '/v1/schedules', got '%s'", log.Path)
	}

	if log.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", log.StatusCode)
	}

	if log.DurationMs != 100 {
		t.Errorf("Expected duration 100ms, got %d", log.DurationMs)
	}
}

func TestLogger_LogWithNilFields(t *testing.T) {
	mockStore := NewMockAuditStore()
	logger := NewLogger(mockStore)
	ctx := context.Background()

	entry := &LogEntry{
		Timestamp:  time.Now(),
		Protocol:   "http",
		Method:     "POST",
		Path:       "/v1/messages",
		StatusCode: 201,
		DurationMs: 150,
		// All optional fields are nil
	}

	err := logger.Log(ctx, entry)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Verify log was created
	if len(mockStore.logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockStore.logs))
	}
}

func TestLogger_LogWithAllFields(t *testing.T) {
	mockStore := NewMockAuditStore()
	logger := NewLogger(mockStore)
	ctx := context.Background()

	apiKeyID := "key-id-123"
	apiKeyName := "Test Key"
	clientIP := "192.168.1.1"
	query := "limit=10"

	entry := &LogEntry{
		Timestamp:  time.Now(),
		Protocol:   "grpc",
		APIKeyID:   &apiKeyID,
		APIKeyName: &apiKeyName,
		Method:     "CreateSchedule",
		Path:       "/grpc/create",
		Query:      &query,
		StatusCode: 200,
		DurationMs: 50,
		ClientIP:   &clientIP,
	}

	err := logger.Log(ctx, entry)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Verify log was created with all fields
	if len(mockStore.logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(mockStore.logs))
	}

	log := mockStore.logs[0]
	if log.APIKeyID == nil || *log.APIKeyID != apiKeyID {
		t.Error("APIKeyID not set correctly")
	}

	if log.APIKeyName == nil || *log.APIKeyName != apiKeyName {
		t.Error("APIKeyName not set correctly")
	}

	if log.ClientIP == nil || *log.ClientIP != clientIP {
		t.Error("ClientIP not set correctly")
	}
}

func TestQuerier_Query(t *testing.T) {
	mockStore := NewMockAuditStore()
	querier := NewQuerier(mockStore)
	ctx := context.Background()

	// Add some test logs
	now := time.Now()
	mockStore.logs = append(mockStore.logs, &store.AuditLog{
		ID:         1,
		Timestamp:  now,
		Protocol:   "http",
		Method:     "GET",
		Path:       "/v1/schedules",
		StatusCode: 200,
		DurationMs: 100,
	})

	mockStore.logs = append(mockStore.logs, &store.AuditLog{
		ID:         2,
		Timestamp:  now.Add(-1 * time.Hour),
		Protocol:   "grpc",
		Method:     "CreateSchedule",
		Path:       "/wecom.gateway.WeComGateway/CreateSchedule",
		StatusCode: 200,
		DurationMs: 50,
	})

	// Query all logs
	opts := &QueryOptions{
		Limit: 10,
	}

	logs, cursor, err := querier.Query(ctx, opts)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs))
	}

	if cursor != "" {
		t.Errorf("Expected empty cursor, got '%s'", cursor)
	}
}

func TestQuerier_QueryWithFilters(t *testing.T) {
	mockStore := NewMockAuditStore()
	querier := NewQuerier(mockStore)
	ctx := context.Background()

	now := time.Now()

	// Add test logs with different methods
	mockStore.logs = append(mockStore.logs, &store.AuditLog{
		ID:         1,
		Timestamp:  now,
		Protocol:   "http",
		Method:     "GET",
		Path:       "/v1/schedules",
		StatusCode: 200,
		DurationMs: 100,
	})

	mockStore.logs = append(mockStore.logs, &store.AuditLog{
		ID:         2,
		Timestamp:  now.Add(-1 * time.Hour),
		Protocol:   "http",
		Method:     "POST",
		Path:       "/v1/messages",
		StatusCode: 201,
		DurationMs: 150,
	})

	// Query with method filter
	apiKeyName := "Test Key"
	method := "GET"
	opts := &QueryOptions{
		APIKeyName: &apiKeyName,
		Method:     &method,
		Limit:      10,
	}

	logs, _, err := querier.Query(ctx, opts)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should only return GET requests
	for _, log := range logs {
		if log.Method != "GET" {
			t.Errorf("Expected only GET requests, got %s", log.Method)
		}
	}
}

func TestQuerier_QueryWithTimeRange(t *testing.T) {
	mockStore := NewMockAuditStore()
	querier := NewQuerier(mockStore)
	ctx := context.Background()

	now := time.Now()

	// Add test logs at different times
	oldLog := &store.AuditLog{
		ID:         1,
		Timestamp:  now.Add(-25 * time.Hour), // 25 hours ago
		Protocol:   "http",
		Method:     "GET",
		Path:       "/v1/old",
		StatusCode: 200,
		DurationMs: 100,
	}

	newLog := &store.AuditLog{
		ID:         2,
		Timestamp:  now.Add(-1 * time.Hour), // 1 hour ago
		Protocol:   "http",
		Method:     "GET",
		Path:       "/v1/new",
		StatusCode: 200,
		DurationMs: 100,
	}

	mockStore.logs = append(mockStore.logs, oldLog, newLog)

	// Query for last 24 hours
	startTime := now.Add(-24 * time.Hour)
	endTime := now
	opts := &QueryOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     10,
	}

	logs, _, err := querier.Query(ctx, opts)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should only return the new log
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry within time range, got %d", len(logs))
	}

	if len(logs) > 0 && logs[0].Path != "/v1/new" {
		t.Errorf("Expected log with path '/v1/new', got '%s'", logs[0].Path)
	}
}

func TestQuerier_QueryWithPagination(t *testing.T) {
	mockStore := NewMockAuditStore()
	querier := NewQuerier(mockStore)
	ctx := context.Background()

	// Add multiple logs
	for i := 1; i <= 25; i++ {
		mockStore.logs = append(mockStore.logs, &store.AuditLog{
			ID:         int64(i),
			Timestamp:  time.Now(),
			Protocol:   "http",
			Method:     "GET",
			Path:       "/v1/test",
			StatusCode: 200,
			DurationMs: 100,
		})
	}

	// Query with limit
	opts := &QueryOptions{
		Limit: 10,
	}

	logs, cursor, err := querier.Query(ctx, opts)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(logs) != 10 {
		t.Errorf("Expected 10 log entries, got %d", len(logs))
	}

	if cursor == "" {
		t.Error("Expected non-empty cursor for pagination")
	}
}
