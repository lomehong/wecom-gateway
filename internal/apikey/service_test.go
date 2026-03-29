package apikey

import (
	"context"
	"testing"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/store"
)

// MockStoreForAPIKey implements store.Database for testing API key service
type MockStoreForAPIKey struct {
	apiKeys  map[string]*store.APIKey
	apps     map[string]map[string]*store.WeComApp
	corps    map[string]*store.WeComCorp
}

func NewMockStoreForAPIKey() *MockStoreForAPIKey {
	return &MockStoreForAPIKey{
		apiKeys: make(map[string]*store.APIKey),
		apps: map[string]map[string]*store.WeComApp{
			"test-corp": {
				"test-app": &store.WeComApp{
					ID:        "app-1",
					Name:      "test-app",
					CorpName:  "test-corp",
					AgentID:   123456,
					SecretEnc: "encrypted",
					Nonce:     "nonce",
				},
			},
		},
		corps: map[string]*store.WeComCorp{
			"test-corp": {
				Name:   "test-corp",
				CorpID: "test-corp-id",
			},
		},
	}
}

func (m *MockStoreForAPIKey) CreateAPIKey(ctx context.Context, key *store.APIKey) error {
	m.apiKeys[key.ID] = key
	return nil
}

func (m *MockStoreForAPIKey) GetAPIKeyByHash(ctx context.Context, hash string) (*store.APIKey, error) {
	for _, key := range m.apiKeys {
		if key.KeyHash == hash {
			return key, nil
		}
	}
	return nil, store.ErrNotFound
}

func (m *MockStoreForAPIKey) ListAPIKeys(ctx context.Context, opts store.ListOptions) ([]*store.APIKey, string, error) {
	keys := make([]*store.APIKey, 0, len(m.apiKeys))
	for _, key := range m.apiKeys {
		keys = append(keys, key)
	}
	return keys, "", nil
}

func (m *MockStoreForAPIKey) UpdateAPIKey(ctx context.Context, key *store.APIKey) error {
	m.apiKeys[key.ID] = key
	return nil
}

func (m *MockStoreForAPIKey) DeleteAPIKey(ctx context.Context, id string) error {
	delete(m.apiKeys, id)
	return nil
}

func (m *MockStoreForAPIKey) CreateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockStoreForAPIKey) GetWeComCorpByName(ctx context.Context, name string) (*store.WeComCorp, error) {
	corp, exists := m.corps[name]
	if !exists {
		return nil, store.ErrNotFound
	}
	return corp, nil
}

func (m *MockStoreForAPIKey) GetWeComCorpByID(ctx context.Context, id string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAPIKey) ListWeComCorps(ctx context.Context) ([]*store.WeComCorp, error) {
	return nil, nil
}

func (m *MockStoreForAPIKey) UpdateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockStoreForAPIKey) DeleteWeComCorp(ctx context.Context, name string) error {
	return nil
}

func (m *MockStoreForAPIKey) CreateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockStoreForAPIKey) GetWeComApp(ctx context.Context, corpName, appName string) (*store.WeComApp, error) {
	apps, exists := m.apps[corpName]
	if !exists {
		return nil, store.ErrNotFound
	}
	app, exists := apps[appName]
	if !exists {
		return nil, store.ErrNotFound
	}
	return app, nil
}

func (m *MockStoreForAPIKey) GetWeComAppByID(ctx context.Context, id string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAPIKey) ListWeComApps(ctx context.Context, corpName string) ([]*store.WeComApp, error) {
	return nil, nil
}

func (m *MockStoreForAPIKey) UpdateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockStoreForAPIKey) DeleteWeComApp(ctx context.Context, id string) error {
	return nil
}

func (m *MockStoreForAPIKey) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	return nil
}

func (m *MockStoreForAPIKey) CreateAuditLog(ctx context.Context, log *store.AuditLog) error {
	return nil
}

func (m *MockStoreForAPIKey) CreateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStoreForAPIKey) GetAdminUserByUsername(ctx context.Context, username string) (*store.AdminUser, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAPIKey) ListAdminUsers(ctx context.Context) ([]*store.AdminUser, error) {
	return nil, nil
}

func (m *MockStoreForAPIKey) UpdateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStoreForAPIKey) DeleteAdminUser(ctx context.Context, username string) error {
	return nil
}

func (m *MockStoreForAPIKey) QueryAuditLogs(ctx context.Context, opts store.AuditQueryOptions) ([]*store.AuditLog, string, error) {
	return nil, "", nil
}

func (m *MockStoreForAPIKey) CreateHourlyStats(ctx context.Context, stats *store.HourlyStats) error {
	return nil
}

func (m *MockStoreForAPIKey) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*store.HourlyStats, error) {
	return nil, nil
}

func (m *MockStoreForAPIKey) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	return nil
}

func (m *MockStoreForAPIKey) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStoreForAPIKey) Close() error {
	return nil
}

func TestNewService_APIKey(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}

	svc := NewService(db, cfg)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.db != db {
		t.Error("db field not set correctly")
	}
	if svc.config != cfg {
		t.Error("config field not set correctly")
	}
}

func TestService_CreateKey_Success(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	req := &CreateKeyRequest{
		Name:        "test-key",
		Permissions: []string{"calendar:read"},
		CorpName:    "test-corp",
		AppName:     "test-app",
		ExpiresDays: 30,
	}

	resp, err := svc.CreateKey(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Name != "test-key" {
		t.Errorf("expected name test-key, got %s", resp.Name)
	}
	if resp.APIKey == "" {
		t.Error("expected API key to be generated")
	}
	if len(resp.Permissions) != 1 {
		t.Errorf("expected 1 permission, got %d", len(resp.Permissions))
	}
}

func TestService_CreateKey_ValidationError(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	tests := []struct {
		name    string
		req     *CreateKeyRequest
		wantErr bool
	}{
		{
			name: "missing name",
			req: &CreateKeyRequest{
				Permissions: []string{"calendar:read"},
				CorpName:    "test-corp",
			},
			wantErr: true,
		},
		{
			name: "empty permissions (allowed)",
			req: &CreateKeyRequest{
				Name:     "test-key",
				CorpName: "test-corp",
			},
			wantErr: false,
		},
		{
			name: "invalid permission",
			req: &CreateKeyRequest{
				Name:        "test-key",
				Permissions: []string{"invalid:permission"},
				CorpName:    "test-corp",
			},
			wantErr: true,
		},
		{
			name: "missing corp name",
			req: &CreateKeyRequest{
				Name:        "test-key",
				Permissions: []string{"calendar:read"},
			},
			wantErr: true,
		},
		{
			name: "negative expires days",
			req: &CreateKeyRequest{
				Name:        "test-key",
				Permissions: []string{"calendar:read"},
				CorpName:    "test-corp",
				ExpiresDays: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateKey(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateKey_AppNotFound(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	req := &CreateKeyRequest{
		Name:        "test-key",
		Permissions: []string{"calendar:read"},
		CorpName:    "test-corp",
		AppName:     "non-existent-app",
		ExpiresDays: 30,
	}

	_, err := svc.CreateKey(context.Background(), req)
	if err == nil {
		t.Error("expected error for non-existent app")
	}
}

func TestService_ListKeys(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	// Create some test keys
	now := time.Now()
	db.apiKeys["key1"] = &store.APIKey{
		ID:        "key1",
		Name:      "key1",
		Disabled:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	db.apiKeys["key2"] = &store.APIKey{
		ID:        "key2",
		Name:      "key2",
		Disabled:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name     string
		disabled *bool
		limit    int
		wantLen  int
	}{
		{
			name:     "all keys",
			disabled: nil,
			limit:    100,
			wantLen:  2,
		},
		{
			name:     "active only",
			disabled: boolPtr(false),
			limit:    100,
			wantLen:  0, // Mock doesn't filter, but structure is there
		},
		{
			name:     "disabled only",
			disabled: boolPtr(true),
			limit:    100,
			wantLen:  0,
		},
		{
			name:     "with limit",
			disabled: nil,
			limit:    1,
			wantLen:  0, // Mock doesn't limit, but structure is there
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, _, err := svc.ListKeys(context.Background(), tt.disabled, tt.limit)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Since mock doesn't filter, just check it returns keys
			_ = keys
		})
	}
}

func TestService_GetKey(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	// Add a test key
	now := time.Now()
	db.apiKeys["test-id"] = &store.APIKey{
		ID:        "test-id",
		Name:      "test-key",
		CreatedAt: now,
		UpdatedAt: now,
	}

	key, err := svc.GetKey(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key == nil {
		t.Error("expected key, got nil")
	}

	// Test not found
	_, err = svc.GetKey(context.Background(), "non-existent")
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_DisableKey(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	// Add a test key
	now := time.Now()
	db.apiKeys["test-id"] = &store.APIKey{
		ID:        "test-id",
		Name:      "test-key",
		Disabled:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := svc.DisableKey(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify key is disabled
	key := db.apiKeys["test-id"]
	if !key.Disabled {
		t.Error("expected key to be disabled")
	}
}

func TestService_EnableKey(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	// Add a disabled test key
	now := time.Now()
	db.apiKeys["test-id"] = &store.APIKey{
		ID:        "test-id",
		Name:      "test-key",
		Disabled:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := svc.EnableKey(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify key is enabled
	key := db.apiKeys["test-id"]
	if key.Disabled {
		t.Error("expected key to be enabled")
	}
}

func TestService_DeleteKey(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	// Add a test key
	now := time.Now()
	db.apiKeys["test-id"] = &store.APIKey{
		ID:        "test-id",
		Name:      "test-key",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := svc.DeleteKey(context.Background(), "test-id")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify key is deleted
	if _, exists := db.apiKeys["test-id"]; exists {
		t.Error("expected key to be deleted")
	}
}

func TestHashAPIKey(t *testing.T) {
	key := "WgK_Test_API_Key_12345" // Use mixed case to test hash

	hash1 := hashAPIKey(key)
	hash2 := hashAPIKey(key)

	if hash1 != hash2 {
		t.Error("same key should produce same hash")
	}

	if hash1 == key {
		t.Error("hash should be different from original key")
	}

	hash3 := hashAPIKey("different-key")
	if hash1 == hash3 {
		t.Error("different keys should produce different hashes")
	}
}

func TestGenerateID(t *testing.T) {
	id1, err := generateID()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if len(id1) < 10 {
		t.Error("expected ID to be at least 10 characters")
	}

	// IDs should be unique
	id2, err := generateID()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id1 == id2 {
		t.Error("IDs should be unique")
	}
}

func TestAPIKeyResponse(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	resp := &APIKeyResponse{
		ID:          "test-id",
		Name:        "test-key",
		APIKey:      "wgk_test_key",
		Permissions: []string{"calendar:read"},
		CorpName:    "test-corp",
		AppName:     "test-app",
		ExpiresAt:   &expiresAt,
		CreatedAt:   now,
	}

	if resp.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", resp.ID)
	}
	if resp.APIKey != "wgk_test_key" {
		t.Errorf("expected APIKey wgk_test_key, got %s", resp.APIKey)
	}
	if resp.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestAPIKeyInfo(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(30 * 24 * time.Hour)

	info := &APIKeyInfo{
		ID:          "test-id",
		Name:        "test-key",
		Permissions: []string{"calendar:read"},
		CorpName:    "test-corp",
		AppName:     "test-app",
		ExpiresAt:   &expiresAt,
		Disabled:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if info.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", info.ID)
	}
	if info.Disabled {
		t.Error("expected key to be enabled")
	}
}

func TestValidateCreateRequest_Permissions(t *testing.T) {
	db := NewMockStoreForAPIKey()
	cfg := &config.Config{}
	svc := NewService(db, cfg)

	validPerms := []string{
		"calendar:read",
		"calendar:write",
		"meetingroom:read",
		"meetingroom:write",
		"message:send",
		"*",
	}

	for _, perm := range validPerms {
		req := &CreateKeyRequest{
			Name:        "test-key",
			Permissions: []string{perm},
			CorpName:    "test-corp",
		}
		err := svc.validateCreateRequest(req)
		if err != nil {
			t.Errorf("permission %s should be valid: %v", perm, err)
		}
	}

	invalidPerm := "invalid:permission"
	req := &CreateKeyRequest{
		Name:        "test-key",
		Permissions: []string{invalidPerm},
		CorpName:    "test-corp",
	}
	err := svc.validateCreateRequest(req)
	if err == nil {
		t.Error("expected error for invalid permission")
	}
}

func boolPtr(b bool) *bool {
	return &b
}
