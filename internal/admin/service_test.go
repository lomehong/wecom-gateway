package admin

import (
	"context"
	"testing"
	"time"

	"wecom-gateway/internal/apikey"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/store"
)

// MockStoreForAdmin implements store.Database for testing admin service
type MockStoreForAdmin struct {
	corps     map[string]*store.WeComCorp
	apps      map[string][]*store.WeComApp
	apiKeys   []*store.APIKey
	auditLogs []*store.AuditLog
}

func NewMockStoreForAdmin() *MockStoreForAdmin {
	return &MockStoreForAdmin{
		corps:     make(map[string]*store.WeComCorp),
		apps:      make(map[string][]*store.WeComApp),
		apiKeys:   []*store.APIKey{},
		auditLogs: []*store.AuditLog{},
	}
}

func (m *MockStoreForAdmin) CreateAPIKey(ctx context.Context, key *store.APIKey) error {
	m.apiKeys = append(m.apiKeys, key)
	return nil
}

func (m *MockStoreForAdmin) GetAPIKeyByHash(ctx context.Context, hash string) (*store.APIKey, error) {
	for _, key := range m.apiKeys {
		if key.KeyHash == hash {
			return key, nil
		}
	}
	return nil, store.ErrNotFound
}

func (m *MockStoreForAdmin) ListAPIKeys(ctx context.Context, opts store.ListOptions) ([]*store.APIKey, string, error) {
	return m.apiKeys, "", nil
}

func (m *MockStoreForAdmin) UpdateAPIKey(ctx context.Context, key *store.APIKey) error {
	for i, k := range m.apiKeys {
		if k.ID == key.ID {
			m.apiKeys[i] = key
			return nil
		}
	}
	return store.ErrNotFound
}

func (m *MockStoreForAdmin) DeleteAPIKey(ctx context.Context, id string) error {
	for i, k := range m.apiKeys {
		if k.ID == id {
			m.apiKeys = append(m.apiKeys[:i], m.apiKeys[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (m *MockStoreForAdmin) CreateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	m.corps[corp.Name] = corp
	return nil
}

func (m *MockStoreForAdmin) GetWeComCorpByName(ctx context.Context, name string) (*store.WeComCorp, error) {
	corp, exists := m.corps[name]
	if !exists {
		return nil, store.ErrNotFound
	}
	return corp, nil
}

func (m *MockStoreForAdmin) GetWeComCorpByID(ctx context.Context, id string) (*store.WeComCorp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAdmin) ListWeComCorps(ctx context.Context) ([]*store.WeComCorp, error) {
	corps := make([]*store.WeComCorp, 0, len(m.corps))
	for _, corp := range m.corps {
		corps = append(corps, corp)
	}
	return corps, nil
}

func (m *MockStoreForAdmin) UpdateWeComCorp(ctx context.Context, corp *store.WeComCorp) error {
	m.corps[corp.Name] = corp
	return nil
}

func (m *MockStoreForAdmin) DeleteWeComCorp(ctx context.Context, name string) error {
	delete(m.corps, name)
	return nil
}

func (m *MockStoreForAdmin) CreateWeComApp(ctx context.Context, app *store.WeComApp) error {
	m.apps[app.CorpName] = append(m.apps[app.CorpName], app)
	return nil
}

func (m *MockStoreForAdmin) GetWeComApp(ctx context.Context, corpName, appName string) (*store.WeComApp, error) {
	apps, exists := m.apps[corpName]
	if !exists {
		return nil, store.ErrNotFound
	}
	for _, app := range apps {
		if app.Name == appName {
			return app, nil
		}
	}
	return nil, store.ErrNotFound
}

func (m *MockStoreForAdmin) GetWeComAppByID(ctx context.Context, id string) (*store.WeComApp, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAdmin) ListWeComApps(ctx context.Context, corpName string) ([]*store.WeComApp, error) {
	apps, exists := m.apps[corpName]
	if !exists {
		return []*store.WeComApp{}, nil
	}
	return apps, nil
}

func (m *MockStoreForAdmin) UpdateWeComApp(ctx context.Context, app *store.WeComApp) error {
	return nil
}

func (m *MockStoreForAdmin) DeleteWeComApp(ctx context.Context, id string) error {
	return nil
}

func (m *MockStoreForAdmin) UpdateAppToken(ctx context.Context, corpName, appName string, token string, expiresAt time.Time) error {
	return nil
}

func (m *MockStoreForAdmin) CreateAuditLog(ctx context.Context, log *store.AuditLog) error {
	m.auditLogs = append(m.auditLogs, log)
	return nil
}

func (m *MockStoreForAdmin) CreateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStoreForAdmin) GetAdminUserByUsername(ctx context.Context, username string) (*store.AdminUser, error) {
	return nil, store.ErrNotFound
}

func (m *MockStoreForAdmin) ListAdminUsers(ctx context.Context) ([]*store.AdminUser, error) {
	return nil, nil
}

func (m *MockStoreForAdmin) UpdateAdminUser(ctx context.Context, user *store.AdminUser) error {
	return nil
}

func (m *MockStoreForAdmin) DeleteAdminUser(ctx context.Context, username string) error {
	return nil
}

func (m *MockStoreForAdmin) QueryAuditLogs(ctx context.Context, opts store.AuditQueryOptions) ([]*store.AuditLog, string, error) {
	return m.auditLogs, "", nil
}

func (m *MockStoreForAdmin) CreateHourlyStats(ctx context.Context, stats *store.HourlyStats) error {
	return nil
}

func (m *MockStoreForAdmin) GetHourlyStats(ctx context.Context, keyName string, startTime, endTime time.Time) ([]*store.HourlyStats, error) {
	return nil, nil
}

func (m *MockStoreForAdmin) IncrementHourlyStats(ctx context.Context, keyName string, timestamp time.Time, isError bool) error {
	return nil
}

func (m *MockStoreForAdmin) Ping(ctx context.Context) error {
	return nil
}

func (m *MockStoreForAdmin) Close() error {
	return nil
}

// MockAPIKeyService for testing
type MockAPIKeyService struct{}

func (m *MockAPIKeyService) CreateKey(ctx context.Context, req *apikey.CreateKeyRequest) (*apikey.APIKeyResponse, error) {
	return &apikey.APIKeyResponse{
		ID:     "test-id",
		Name:   req.Name,
		APIKey: "wgk_test_key",
	}, nil
}

func TestNewService(t *testing.T) {
	db := NewMockStoreForAdmin()
	cfg := &config.Config{}
	apiKeySvc := &apikey.Service{}
	auditLogger := &audit.Logger{}
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	svc := NewService(db, cfg, apiKeySvc, auditLogger, encKey)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.DB != db {
		t.Error("db field not set correctly")
	}
	if svc.config != cfg {
		t.Error("config field not set correctly")
	}
}

func TestService_InitializeSystem_AlreadyInitialized(t *testing.T) {
	db := NewMockStoreForAdmin()
	cfg := &config.Config{
		WeCom: config.WeComConfig{
			Corps: []config.CorpsConfig{},
		},
	}
	apiKeySvc := &apikey.Service{}
	auditLogger := &audit.Logger{}
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	// Pre-create the main corp
	db.CreateWeComCorp(context.Background(), &store.WeComCorp{
		Name:   "main",
		CorpID: "test-corp-id",
	})

	svc := NewService(db, cfg, apiKeySvc, auditLogger, encKey)
	err := svc.InitializeSystem(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_InitializeSystem_NewCorps(t *testing.T) {
	db := NewMockStoreForAdmin()
	testKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	encryptedSecret, _ := crypto.EncryptString("test-secret", testKey)
	cfg := &config.Config{
		WeCom: config.WeComConfig{
			Corps: []config.CorpsConfig{
				{
					Name:   "test-corp",
					CorpID: "test-corp-id",
					Apps: []config.AppConfig{
						{
							Name:   "test-app",
							AgentID: 123456,
							Secret: encryptedSecret,
						},
					},
				},
			},
		},
	}
	apiKeySvc := &apikey.Service{}
	auditLogger := &audit.Logger{}
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	svc := NewService(db, cfg, apiKeySvc, auditLogger, encKey)
	err := svc.InitializeSystem(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify corp was created
	corp, err := db.GetWeComCorpByName(context.Background(), "test-corp")
	if err != nil {
		t.Errorf("corp not created: %v", err)
	}
	if corp.CorpID != "test-corp-id" {
		t.Errorf("expected CorpID test-corp-id, got %s", corp.CorpID)
	}
}

func TestService_GetDashboardStats(t *testing.T) {
	db := NewMockStoreForAdmin()
	cfg := &config.Config{}
	apiKeySvc := &apikey.Service{}
	auditLogger := &audit.Logger{}
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	// Add some test data
	db.apiKeys = append(db.apiKeys, &store.APIKey{
		ID:        "key1",
		Name:      "test-key",
		Disabled:  false,
		CreatedAt: time.Now(),
	})
	db.apiKeys = append(db.apiKeys, &store.APIKey{
		ID:        "key2",
		Name:      "disabled-key",
		Disabled:  true,
		CreatedAt: time.Now(),
	})
	db.corps["main"] = &store.WeComCorp{
		Name:   "main",
		CorpID: "main-corp-id",
	}

	svc := NewService(db, cfg, apiKeySvc, auditLogger, encKey)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	stats, err := svc.GetDashboardStats(context.Background(), startTime, endTime)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if stats.TotalAPIKeys != 2 {
		t.Errorf("expected TotalAPIKeys 2, got %d", stats.TotalAPIKeys)
	}
	if stats.ActiveAPIKeys != 1 {
		t.Errorf("expected ActiveAPIKeys 1, got %d", stats.ActiveAPIKeys)
	}
	if stats.TotalCorps != 1 {
		t.Errorf("expected TotalCorps 1, got %d", stats.TotalCorps)
	}
}

func TestGenerateID(t *testing.T) {
	prefix := "test_"
	id := generateID(prefix)

	if len(id) <= len(prefix) {
		t.Error("generated ID should be longer than prefix")
	}
	if id[:len(prefix)] != prefix {
		t.Errorf("generated ID should start with %s", prefix)
	}
}

func TestDashboardStats(t *testing.T) {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	stats := &DashboardStats{
		StartTime:     startTime,
		EndTime:       endTime,
		TotalAPIKeys:  10,
		ActiveAPIKeys: 8,
		TotalCorps:    2,
		TotalApps:     5,
		TotalRequests: 1000,
		ErrorRequests: 50,
	}

	if stats.TotalAPIKeys != 10 {
		t.Errorf("expected TotalAPIKeys 10, got %d", stats.TotalAPIKeys)
	}
	if stats.ActiveAPIKeys != 8 {
		t.Errorf("expected ActiveAPIKeys 8, got %d", stats.ActiveAPIKeys)
	}
	if stats.ErrorRate() != 0.05 {
		t.Errorf("expected error rate 0.05, got %f", stats.ErrorRate())
	}
}

// ErrorRate calculates the error rate (helper method for testing)
func (ds *DashboardStats) ErrorRate() float64 {
	if ds.TotalRequests == 0 {
		return 0
	}
	return float64(ds.ErrorRequests) / float64(ds.TotalRequests)
}
