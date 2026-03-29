package auth

import (
	"context"
	"testing"
	"time"

	"wecom-gateway/internal/store"
)

// MockStore implements store.Interface for testing
type MockStore struct {
	apiKeys      map[string]*store.APIKey
	userExists   bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		apiKeys: make(map[string]*store.APIKey),
	}
}

func (m *MockStore) CreateAPIKey(ctx context.Context, key *store.APIKey) error {
	m.apiKeys[key.KeyHash] = key
	return nil
}

func (m *MockStore) GetAPIKeyByHash(ctx context.Context, keyHash string) (*store.APIKey, error) {
	key, exists := m.apiKeys[keyHash]
	if !exists {
		return nil, store.ErrNotFound
	}
	return key, nil
}

func (m *MockStore) ListAPIKeys(ctx context.Context, opts store.ListOptions) ([]*store.APIKey, string, error) {
	keys := make([]*store.APIKey, 0, len(m.apiKeys))
	for _, key := range m.apiKeys {
		keys = append(keys, key)
	}
	return keys, "", nil
}

func (m *MockStore) UpdateAPIKey(ctx context.Context, key *store.APIKey) error {
	return nil
}

func (m *MockStore) DeleteAPIKey(ctx context.Context, id string) error {
	delete(m.apiKeys, id)
	return nil
}

func (m *MockStore) CreateAuditLog(ctx context.Context, log *store.AuditLog) error {
	return nil
}

func (m *MockStore) CreateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStore) GetAdminUserByUsername(ctx context.Context, username string) (*store.AdminUser, error) {
	return nil, store.ErrNotFound
}

func (m *MockStore) ListAdminUsers(ctx context.Context) ([]*store.AdminUser, error) {
	return nil, nil
}

func (m *MockStore) UpdateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStore) DeleteAdminUser(ctx context.Context, username string) error {
	return nil
}

func (m *MockStore) QueryAuditLogs(ctx context.Context, opts store.AuditQueryOptions) ([]*store.AuditLog, string, error) {
	return []*store.AuditLog{}, "", nil
}

func (m *MockStore) CreateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockStore) GetWeComCorpByName(ctx context.Context, name string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStore) GetWeComCorpByID(ctx context.Context, id string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStore) ListWeComCorps(ctx context.Context) ([]*store.WeComCorp, error) {
	return []*store.WeComCorp{}, nil
}

func (m *MockStore) UpdateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	return nil
}

func (m *MockStore) DeleteWeComCorp(ctx context.Context, name string) error {
	return nil
}

func (m *MockStore) CreateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockStore) GetWeComApp(ctx context.Context, corpName, appName string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStore) GetWeComAppByID(ctx context.Context, id string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStore) ListWeComApps(ctx context.Context, corpName string) ([]*store.WeComApp, error) {
	return []*store.WeComApp{}, nil
}

func (m *MockStore) UpdateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockStore) DeleteWeComApp(ctx context.Context, id string) error {
	return nil
}

func (m *MockStore) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	return nil
}

func (m *MockStore) CreateHourlyStats(ctx context.Context, stats *store.HourlyStats) error {
	return nil
}

func (m *MockStore) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*store.HourlyStats, error) {
	return []*store.HourlyStats{}, nil
}

func (m *MockStore) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

func (m *MockStore) Ping(ctx context.Context) error {
	return nil
}

func TestHashAPIKey(t *testing.T) {
	key := "WGK_TEST_API_KEY_12345"

	hash1 := hashAPIKey(key)
	hash2 := hashAPIKey(key)

	// Same key should produce same hash
	if hash1 != hash2 {
		t.Error("Same API key should produce same hash")
	}

	// Hash should be different from original
	if hash1 == key {
		t.Error("Hash should be different from original key")
	}

	// Different key should produce different hash
	hash3 := hashAPIKey("different-key")
	if hash1 == hash3 {
		t.Error("Different API keys should produce different hashes")
	}
}

func TestNewAPIKeyAuthenticator(t *testing.T) {
	store := NewMockStore()
	auth := NewAPIKeyAuthenticator(store)

	if auth == nil {
		t.Fatal("NewAPIKeyAuthenticator returned nil")
	}

	if auth.db == nil {
		t.Error("db field should not be nil")
	}
}

func TestValidateAPIKey(t *testing.T) {
	mockStore := NewMockStore()
	auth := NewAPIKeyAuthenticator(mockStore)

	ctx := context.Background()

	// Create a test API key
	testKey := "wgk_test_valid_key"
	testHash := hashAPIKey(testKey)

	expiresAt := time.Now().Add(24 * time.Hour)

	apiKey := &store.APIKey{
		ID:          "test-id",
		Name:        "Test Key",
		KeyHash:     testHash,
		Permissions: []string{"calendar:read"},
		CorpName:    "main",
		AppName:     "oa",
		ExpiresAt:   &expiresAt,
		Disabled:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockStore.CreateAPIKey(ctx, apiKey)

	testCases := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{"valid key", testKey, false},
		{"invalid key", "wgk_invalid_key", true},
		{"empty key", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authCtx, err := auth.Authenticate(ctx, tc.apiKey)

			if (err != nil) != tc.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				if authCtx == nil {
					t.Error("Expected auth context, got nil")
				} else {
					if authCtx.KeyID != apiKey.ID {
						t.Errorf("Expected KeyID %s, got %s", apiKey.ID, authCtx.KeyID)
					}
					if authCtx.KeyName != apiKey.Name {
						t.Errorf("Expected KeyName %s, got %s", apiKey.Name, authCtx.KeyName)
					}
					if authCtx.CorpName != apiKey.CorpName {
						t.Errorf("Expected CorpName %s, got %s", apiKey.CorpName, authCtx.CorpName)
					}
					if authCtx.AppName != apiKey.AppName {
						t.Errorf("Expected AppName %s, got %s", apiKey.AppName, authCtx.AppName)
					}
				}
			}
		})
	}
}

func TestValidateAPIKeyExpired(t *testing.T) {
	mockStore := NewMockStore()
	auth := NewAPIKeyAuthenticator(mockStore)

	ctx := context.Background()

	// Create an expired API key
	testKey := "wgk_test_expired_key"
	testHash := hashAPIKey(testKey)

	expiresAt := time.Now().Add(-24 * time.Hour)

	apiKey := &store.APIKey{
		ID:          "expired-id",
		Name:        "Expired Key",
		KeyHash:     testHash,
		Permissions: []string{"calendar:read"},
		CorpName:    "main",
		AppName:     "oa",
		ExpiresAt:   &expiresAt, // Expired
		Disabled:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockStore.CreateAPIKey(ctx, apiKey)

	_, err := auth.Authenticate(ctx, testKey)
	if err == nil {
		t.Error("Expected error for expired API key")
	}
}

func TestValidateAPIKeyDisabled(t *testing.T) {
	mockStore := NewMockStore()
	auth := NewAPIKeyAuthenticator(mockStore)

	ctx := context.Background()

	// Create a disabled API key
	testKey := "wgk_test_disabled_key"
	testHash := hashAPIKey(testKey)

	expiresAt := time.Now().Add(24 * time.Hour)

	apiKey := &store.APIKey{
		ID:          "disabled-id",
		Name:        "Disabled Key",
		KeyHash:     testHash,
		Permissions: []string{"calendar:read"},
		CorpName:    "main",
		AppName:     "oa",
		ExpiresAt:   &expiresAt,
		Disabled:    true, // Disabled
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockStore.CreateAPIKey(ctx, apiKey)

	_, err := auth.Authenticate(ctx, testKey)
	if err == nil {
		t.Error("Expected error for disabled API key")
	}
}
