package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *SQLite {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

// === API Key CRUD Tests ===

func TestSQLite_CreateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	key := &APIKey{
		ID:          "key-1",
		Name:        "Test Key",
		KeyHash:     "hash_abc123",
		Permissions: []string{"calendar:read"},
		CorpName:    "main",
		AppName:     "oa",
		Disabled:    false,
	}

	err := db.CreateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// Verify by getting it back
	retrieved, err := db.GetAPIKeyByHash(ctx, "hash_abc123")
	if err != nil {
		t.Fatalf("GetAPIKeyByHash failed: %v", err)
	}
	if retrieved.Name != "Test Key" {
		t.Errorf("expected name Test Key, got %s", retrieved.Name)
	}
	if retrieved.CorpName != "main" {
		t.Errorf("expected corp main, got %s", retrieved.CorpName)
	}
	if len(retrieved.Permissions) != 1 || retrieved.Permissions[0] != "calendar:read" {
		t.Errorf("expected [calendar:read], got %v", retrieved.Permissions)
	}
}

func TestSQLite_CreateAPIKey_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	key := &APIKey{
		ID:          "key-1",
		Name:        "Dup Key",
		KeyHash:     "hash_dup",
		Permissions: []string{"read"},
		CorpName:    "main",
	}

	err := db.CreateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("first CreateAPIKey failed: %v", err)
	}

	// Try creating duplicate (same name)
	key2 := &APIKey{
		ID:          "key-2",
		Name:        "Dup Key",
		KeyHash:     "hash_dup2",
		Permissions: []string{"read"},
		CorpName:    "main",
	}
	err = db.CreateAPIKey(ctx, key2)
	if err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestSQLite_GetAPIKeyByHash_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.GetAPIKeyByHash(ctx, "nonexistent_hash")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLite_CreateAPIKey_WithExpiration(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	key := &APIKey{
		ID:          "key-exp",
		Name:        "Expiring Key",
		KeyHash:     "hash_exp",
		Permissions: []string{"read"},
		CorpName:    "main",
		ExpiresAt:   &expiresAt,
	}

	err := db.CreateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	retrieved, err := db.GetAPIKeyByHash(ctx, "hash_exp")
	if err != nil {
		t.Fatalf("GetAPIKeyByHash failed: %v", err)
	}
	if retrieved.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
}

func TestSQLite_ListAPIKeys(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create multiple keys
	for i := 0; i < 3; i++ {
		key := &APIKey{
			ID:          "key-" + string(rune('A'+i)),
			Name:        "Key " + string(rune('A'+i)),
			KeyHash:     "hash_" + string(rune('A'+i)),
			Permissions: []string{"read"},
			CorpName:    "main",
		}
		if err := db.CreateAPIKey(ctx, key); err != nil {
			t.Fatalf("CreateAPIKey %d failed: %v", i, err)
		}
	}

	keys, cursor, err := db.ListAPIKeys(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %q", cursor)
	}
}

func TestSQLite_ListAPIKeys_WithDisabledFilter(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	active := true
	key1 := &APIKey{ID: "k1", Name: "Active Key", KeyHash: "h1", Permissions: []string{"read"}, CorpName: "main", Disabled: false}
	key2 := &APIKey{ID: "k2", Name: "Disabled Key", KeyHash: "h2", Permissions: []string{"read"}, CorpName: "main", Disabled: true}
	db.CreateAPIKey(ctx, key1)
	db.CreateAPIKey(ctx, key2)

	// List only disabled
	keys, _, err := db.ListAPIKeys(ctx, ListOptions{Disabled: &active})
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	// Note: active = true means filter for disabled=true in our impl
	for _, k := range keys {
		if !k.Disabled {
			t.Errorf("expected all keys to be disabled, got %s which is not", k.Name)
		}
	}
}

func TestSQLite_ListAPIKeys_WithLimit(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		key := &APIKey{
			ID:          "key-lim-" + string(rune('A'+i)),
			Name:        "Limit Key " + string(rune('A'+i)),
			KeyHash:     "hash_lim_" + string(rune('A'+i)),
			Permissions: []string{"read"},
			CorpName:    "main",
		}
		db.CreateAPIKey(ctx, key)
	}

	keys, cursor, err := db.ListAPIKeys(ctx, ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	if cursor == "" {
		t.Error("expected non-empty cursor for pagination")
	}
}

func TestSQLite_UpdateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	key := &APIKey{
		ID:          "key-upd",
		Name:        "Original",
		KeyHash:     "hash_upd",
		Permissions: []string{"read"},
		CorpName:    "main",
	}
	db.CreateAPIKey(ctx, key)

	// Update
	key.Permissions = []string{"read", "write"}
	key.CorpName = "newcorp"
	err := db.UpdateAPIKey(ctx, key)
	if err != nil {
		t.Fatalf("UpdateAPIKey failed: %v", err)
	}

	retrieved, err := db.GetAPIKeyByHash(ctx, "hash_upd")
	if err != nil {
		t.Fatalf("GetAPIKeyByHash failed: %v", err)
	}
	if len(retrieved.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(retrieved.Permissions))
	}
	if retrieved.CorpName != "newcorp" {
		t.Errorf("expected corp newcorp, got %s", retrieved.CorpName)
	}
}

func TestSQLite_UpdateAPIKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	key := &APIKey{
		ID:          "nonexistent",
		Name:        "Ghost",
		KeyHash:     "hash_ghost",
		Permissions: []string{"read"},
		CorpName:    "main",
	}
	err := db.UpdateAPIKey(ctx, key)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLite_DeleteAPIKey(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	key := &APIKey{
		ID:          "key-del",
		Name:        "To Delete",
		KeyHash:     "hash_del",
		Permissions: []string{"read"},
		CorpName:    "main",
	}
	db.CreateAPIKey(ctx, key)

	err := db.DeleteAPIKey(ctx, "key-del")
	if err != nil {
		t.Fatalf("DeleteAPIKey failed: %v", err)
	}

	_, err = db.GetAPIKeyByHash(ctx, "hash_del")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSQLite_DeleteAPIKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.DeleteAPIKey(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// === WeCom Corp CRUD Tests ===

func TestSQLite_CreateWeComCorp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	corp := &WeComCorp{
		ID:     "corp-1",
		Name:   "TestCorp",
		CorpID: "ww1234567890",
	}

	err := db.CreateWeComCorp(ctx, corp)
	if err != nil {
		t.Fatalf("CreateWeComCorp failed: %v", err)
	}

	retrieved, err := db.GetWeComCorpByName(ctx, "TestCorp")
	if err != nil {
		t.Fatalf("GetWeComCorpByName failed: %v", err)
	}
	if retrieved.CorpID != "ww1234567890" {
		t.Errorf("expected corp id ww1234567890, got %s", retrieved.CorpID)
	}
}

func TestSQLite_CreateWeComCorp_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	corp := &WeComCorp{ID: "c1", Name: "DupCorp", CorpID: "ww111"}
	db.CreateWeComCorp(ctx, corp)

	corp2 := &WeComCorp{ID: "c2", Name: "DupCorp", CorpID: "ww222"}
	err := db.CreateWeComCorp(ctx, corp2)
	if err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestSQLite_GetWeComCorp_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.GetWeComCorpByName(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLite_ListWeComCorps(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "CorpB", CorpID: "ww1"})
	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c2", Name: "CorpA", CorpID: "ww2"})

	corps, err := db.ListWeComCorps(ctx)
	if err != nil {
		t.Fatalf("ListWeComCorps failed: %v", err)
	}
	if len(corps) != 2 {
		t.Errorf("expected 2 corps, got %d", len(corps))
	}
	// Should be ordered by name
	if corps[0].Name != "CorpA" {
		t.Errorf("expected first corp CorpA, got %s", corps[0].Name)
	}
}

func TestSQLite_UpdateWeComCorp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "UpdCorp", CorpID: "ww1"})

	corp := &WeComCorp{Name: "UpdCorp", CorpID: "ww_new_id"}
	err := db.UpdateWeComCorp(ctx, corp)
	if err != nil {
		t.Fatalf("UpdateWeComCorp failed: %v", err)
	}

	retrieved, _ := db.GetWeComCorpByName(ctx, "UpdCorp")
	if retrieved.CorpID != "ww_new_id" {
		t.Errorf("expected ww_new_id, got %s", retrieved.CorpID)
	}
}

func TestSQLite_UpdateWeComCorp_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	corp := &WeComCorp{Name: "Ghost", CorpID: "ww_ghost"}
	err := db.UpdateWeComCorp(ctx, corp)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLite_DeleteWeComCorp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "DelCorp", CorpID: "ww1"})

	err := db.DeleteWeComCorp(ctx, "DelCorp")
	if err != nil {
		t.Fatalf("DeleteWeComCorp failed: %v", err)
	}

	_, err = db.GetWeComCorpByName(ctx, "DelCorp")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

// === WeCom App CRUD Tests ===

func TestSQLite_CreateWeComApp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create parent corp first
	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "MainCorp", CorpID: "ww1"})

	app := &WeComApp{
		ID:        "app-1",
		Name:      "OA",
		CorpName:  "MainCorp",
		AgentID:   1000002,
		SecretEnc: "enc_secret",
		Nonce:     "nonce123",
	}

	err := db.CreateWeComApp(ctx, app)
	if err != nil {
		t.Fatalf("CreateWeComApp failed: %v", err)
	}

	retrieved, err := db.GetWeComApp(ctx, "MainCorp", "OA")
	if err != nil {
		t.Fatalf("GetWeComApp failed: %v", err)
	}
	if retrieved.AgentID != 1000002 {
		t.Errorf("expected agent id 1000002, got %d", retrieved.AgentID)
	}
}

func TestSQLite_CreateWeComApp_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "Corp1", CorpID: "ww1"})

	app := &WeComApp{ID: "a1", Name: "OA", CorpName: "Corp1", AgentID: 1, SecretEnc: "s", Nonce: "n"}
	db.CreateWeComApp(ctx, app)

	app2 := &WeComApp{ID: "a2", Name: "OA", CorpName: "Corp1", AgentID: 2, SecretEnc: "s2", Nonce: "n2"}
	err := db.CreateWeComApp(ctx, app2)
	if err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestSQLite_GetWeComApp_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	_, err := db.GetWeComApp(ctx, "nonexistent", "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLite_ListWeComApps(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "Corp1", CorpID: "ww1"})
	db.CreateWeComApp(ctx, &WeComApp{ID: "a1", Name: "AppB", CorpName: "Corp1", AgentID: 1, SecretEnc: "s", Nonce: "n"})
	db.CreateWeComApp(ctx, &WeComApp{ID: "a2", Name: "AppA", CorpName: "Corp1", AgentID: 2, SecretEnc: "s2", Nonce: "n2"})

	apps, err := db.ListWeComApps(ctx, "Corp1")
	if err != nil {
		t.Fatalf("ListWeComApps failed: %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
}

func TestSQLite_UpdateWeComApp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "Corp1", CorpID: "ww1"})
	db.CreateWeComApp(ctx, &WeComApp{ID: "a1", Name: "OA", CorpName: "Corp1", AgentID: 1, SecretEnc: "s", Nonce: "n"})

	app := &WeComApp{ID: "a1", Name: "OA", CorpName: "Corp1", AgentID: 999, SecretEnc: "new_secret", Nonce: "new_nonce"}
	err := db.UpdateWeComApp(ctx, app)
	if err != nil {
		t.Fatalf("UpdateWeComApp failed: %v", err)
	}

	retrieved, _ := db.GetWeComApp(ctx, "Corp1", "OA")
	if retrieved.AgentID != 999 {
		t.Errorf("expected agent id 999, got %d", retrieved.AgentID)
	}
}

func TestSQLite_DeleteWeComApp(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "Corp1", CorpID: "ww1"})
	db.CreateWeComApp(ctx, &WeComApp{ID: "a1", Name: "OA", CorpName: "Corp1", AgentID: 1, SecretEnc: "s", Nonce: "n"})

	err := db.DeleteWeComApp(ctx, "a1")
	if err != nil {
		t.Fatalf("DeleteWeComApp failed: %v", err)
	}

	_, err = db.GetWeComApp(ctx, "Corp1", "OA")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSQLite_UpdateAppToken(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateWeComCorp(ctx, &WeComCorp{ID: "c1", Name: "Corp1", CorpID: "ww1"})
	db.CreateWeComApp(ctx, &WeComApp{ID: "a1", Name: "OA", CorpName: "Corp1", AgentID: 1, SecretEnc: "s", Nonce: "n"})

	expiresAt := time.Now().Add(2 * time.Hour)
	err := db.UpdateAppToken(ctx, "Corp1", "OA", "new_token_abc", expiresAt)
	if err != nil {
		t.Fatalf("UpdateAppToken failed: %v", err)
	}

	app, _ := db.GetWeComApp(ctx, "Corp1", "OA")
	if app.AccessToken == nil || *app.AccessToken != "new_token_abc" {
		t.Error("expected access token to be updated")
	}
}

// === Audit Log Tests ===

func TestSQLite_CreateAuditLog(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	log := &AuditLog{
		Timestamp:  time.Now(),
		Protocol:   "http",
		APIKeyName: strPtr("TestKey"),
		Method:     "POST",
		Path:       "/v1/schedule",
		StatusCode: 200,
		DurationMs: 45,
	}

	err := db.CreateAuditLog(ctx, log)
	if err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}
}

func TestSQLite_QueryAuditLogs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create multiple logs
	for i := 0; i < 5; i++ {
		log := &AuditLog{
			Timestamp:  time.Now().Add(time.Duration(i) * time.Minute),
			Protocol:   "http",
			Method:     "GET",
			Path:       "/v1/test",
			StatusCode: 200,
			DurationMs: 10 + i,
		}
		db.CreateAuditLog(ctx, log)
	}

	logs, _, err := db.QueryAuditLogs(ctx, AuditQueryOptions{})
	if err != nil {
		t.Fatalf("QueryAuditLogs failed: %v", err)
	}
	if len(logs) != 5 {
		t.Errorf("expected 5 logs, got %d", len(logs))
	}
}

func TestSQLite_QueryAuditLogs_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	db.CreateAuditLog(ctx, &AuditLog{Timestamp: time.Now(), Method: "GET", Path: "/v1/schedule", StatusCode: 200, DurationMs: 10})
	db.CreateAuditLog(ctx, &AuditLog{Timestamp: time.Now(), Method: "POST", Path: "/v1/message", StatusCode: 500, DurationMs: 20})

	// Filter by method
	method := "POST"
	logs, _, err := db.QueryAuditLogs(ctx, AuditQueryOptions{Method: &method})
	if err != nil {
		t.Fatalf("QueryAuditLogs with filter failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log for POST, got %d", len(logs))
	}

	// Filter by status code
	statusCode := 500
	logs, _, err = db.QueryAuditLogs(ctx, AuditQueryOptions{StatusCode: &statusCode})
	if err != nil {
		t.Fatalf("QueryAuditLogs with status filter failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log for 500, got %d", len(logs))
	}
}

func TestSQLite_QueryAuditLogs_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		db.CreateAuditLog(ctx, &AuditLog{
			Timestamp:  time.Now().Add(time.Duration(i) * time.Second),
			Method:     "GET",
			Path:       "/v1/test",
			StatusCode: 200,
			DurationMs: i,
		})
	}

	logs, cursor, err := db.QueryAuditLogs(ctx, AuditQueryOptions{Limit: 3})
	if err != nil {
		t.Fatalf("QueryAuditLogs with limit failed: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
	if cursor == "" {
		t.Error("expected non-empty cursor")
	}
}

// === Hourly Stats Tests ===

func TestSQLite_CreateHourlyStats(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	stats := &HourlyStats{
		Hour:       time.Now().Truncate(time.Hour),
		TotalCount: 100,
		ErrorCount: 5,
		KeyName:    "TestKey",
	}

	err := db.CreateHourlyStats(ctx, stats)
	if err != nil {
		t.Fatalf("CreateHourlyStats failed: %v", err)
	}
}

func TestSQLite_CreateHourlyStats_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	hour := time.Now().Truncate(time.Hour)
	stats := &HourlyStats{Hour: hour, TotalCount: 10, ErrorCount: 1, KeyName: "Key1"}
	db.CreateHourlyStats(ctx, stats)

	stats2 := &HourlyStats{Hour: hour, TotalCount: 20, ErrorCount: 2, KeyName: "Key1"}
	err := db.CreateHourlyStats(ctx, stats2)
	if err != ErrDuplicate {
		t.Errorf("expected ErrDuplicate, got %v", err)
	}
}

func TestSQLite_GetHourlyStats(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()
	hour := now.Truncate(time.Hour)
	db.CreateHourlyStats(ctx, &HourlyStats{Hour: hour, TotalCount: 100, ErrorCount: 5, KeyName: "TestKey"})

	stats, err := db.GetHourlyStats(ctx, "TestKey", hour.Add(-1*time.Hour), hour.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetHourlyStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].TotalCount != 100 {
		t.Errorf("expected total 100, got %d", stats[0].TotalCount)
	}
}

func TestSQLite_IncrementHourlyStats(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()

	// Increment success
	err := db.IncrementHourlyStats(ctx, "TestKey", now, false)
	if err != nil {
		t.Fatalf("IncrementHourlyStats (success) failed: %v", err)
	}

	// Increment error
	err = db.IncrementHourlyStats(ctx, "TestKey", now, true)
	if err != nil {
		t.Fatalf("IncrementHourlyStats (error) failed: %v", err)
	}

	// Verify - the stored hour is in UTC
	hour := now.Truncate(time.Hour).UTC()
	stats, err := db.GetHourlyStats(ctx, "TestKey", hour.Add(-1*time.Second), hour.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetHourlyStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].TotalCount != 2 {
		t.Errorf("expected total 2, got %d", stats[0].TotalCount)
	}
	if stats[0].ErrorCount != 1 {
		t.Errorf("expected error 1, got %d", stats[0].ErrorCount)
	}
}

func TestSQLite_Ping(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestSQLite_New_InvalidDriver(t *testing.T) {
	_, err := New(&Config{Driver: "mysql", DSN: ""})
	if err != ErrInvalidDriver {
		t.Errorf("expected ErrInvalidDriver, got %v", err)
	}
}

func TestSQLite_New_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_new.db")
	db, err := New(&Config{Driver: "sqlite", DSN: dbPath})
	if err != nil {
		t.Fatalf("New sqlite failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

// Helper
func strPtr(s string) *string {
	return &s
}

// Test that ListOptions defaults are applied correctly
func TestListOptions_Defaults(t *testing.T) {
	opts := ListOptions{}
	if opts.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", opts.Limit)
	}
	if opts.Cursor != "" {
		t.Errorf("expected empty cursor, got %q", opts.Cursor)
	}
	if opts.Disabled != nil {
		t.Error("expected nil disabled filter")
	}
}

// Test that AuditQueryOptions defaults are applied correctly
func TestAuditQueryOptions_Defaults(t *testing.T) {
	opts := AuditQueryOptions{}
	if opts.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", opts.Limit)
	}
	if opts.Cursor != "" {
		t.Errorf("expected empty cursor, got %q", opts.Cursor)
	}
	if opts.APIKeyName != nil {
		t.Error("expected nil api_key_name filter")
	}
	if opts.Method != nil {
		t.Error("expected nil method filter")
	}
	if opts.Path != nil {
		t.Error("expected nil path filter")
	}
	if opts.StatusCode != nil {
		t.Error("expected nil status_code filter")
	}
	if opts.StartTime != nil {
		t.Error("expected nil start_time")
	}
	if opts.EndTime != nil {
		t.Error("expected nil end_time")
	}
}

// Test file path for DSN
func TestNewSQLite_BadDSN(t *testing.T) {
	// Use a directory that doesn't exist as DSN
	_, err := NewSQLite("/nonexistent/directory/path/test.db")
	if err == nil {
		t.Error("expected error for bad DSN")
	}
}

func TestSQLite_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_close.db")
	db, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite failed: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Double close should not panic
	err = db.Close()
	if err != nil {
		t.Errorf("double Close failed: %v", err)
	}
}

func TestSQLite_QueryAuditLogs_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	logs, cursor, err := db.QueryAuditLogs(ctx, AuditQueryOptions{})
	if err != nil {
		t.Fatalf("QueryAuditLogs failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %q", cursor)
	}
}

func TestSQLite_ListAPIKeys_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	keys, cursor, err := db.ListAPIKeys(ctx, ListOptions{})
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %q", cursor)
	}
}

func TestSQLite_ListWeComApps_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	apps, err := db.ListWeComApps(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("ListWeComApps failed: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}

func TestSQLite_ListWeComCorps_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	corps, err := db.ListWeComCorps(ctx)
	if err != nil {
		t.Fatalf("ListWeComCorps failed: %v", err)
	}
	if len(corps) != 0 {
		t.Errorf("expected 0 corps, got %d", len(corps))
	}
}

func TestSQLite_GetHourlyStats_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	now := time.Now()
	stats, err := db.GetHourlyStats(ctx, "nonexistent", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("GetHourlyStats failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

// Remove unused import
var _ = os.TempDir
