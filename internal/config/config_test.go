package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	if cfg == nil {
		t.Fatal("GetDefaultConfig returned nil")
	}

	if cfg.Server.HTTPListen != ":8080" {
		t.Errorf("Expected HTTPListen ':8080', got '%s'", cfg.Server.HTTPListen)
	}

	if cfg.Server.GRPCListen != ":9090" {
		t.Errorf("Expected GRPCListen ':9090', got '%s'", cfg.Server.GRPCListen)
	}

	if cfg.Server.Mode != "release" {
		t.Errorf("Expected Mode 'release', got '%s'", cfg.Server.Mode)
	}

	if cfg.Database.Driver != "sqlite" {
		t.Errorf("Expected Driver 'sqlite', got '%s'", cfg.Database.Driver)
	}

	if cfg.Database.DSN != "data/wecom.db" {
		t.Errorf("Expected DSN 'data/wecom.db', got '%s'", cfg.Database.DSN)
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Server: ServerConfig{
					HTTPListen: ":8080",
					GRPCListen: ":9090",
					Mode:       "release",
				},
				Database: DatabaseConfig{
					Driver: "sqlite",
					DSN:    "test.db",
				},
				WeCom: WeComConfig{
					Corps: []CorpsConfig{
						{
							Name:   "main",
							CorpID: "ww1234567890abcdef",
							Apps: []AppConfig{
								{
									Name:    "oa",
									AgentID: 1000001,
									Secret:  "test-secret",
								},
							},
						},
					},
				},
				Auth: AuthConfig{
					RateLimit: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "missing http_listen",
			cfg: &Config{
				Server: ServerConfig{
					GRPCListen: ":9090",
					Mode:       "release",
				},
				Database: DatabaseConfig{
					Driver: "sqlite",
					DSN:    "test.db",
				},
				WeCom: WeComConfig{
					Corps: []CorpsConfig{
						{
							Name:   "main",
							CorpID: "ww1234567890abcdef",
							Apps: []AppConfig{
								{
									Name:    "oa",
									AgentID: 1000001,
									Secret:  "test-secret",
								},
							},
						},
					},
				},
				Auth: AuthConfig{
					RateLimit: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			cfg: &Config{
				Server: ServerConfig{
					HTTPListen: ":8080",
					GRPCListen: ":9090",
					Mode:       "invalid",
				},
				Database: DatabaseConfig{
					Driver: "sqlite",
					DSN:    "test.db",
				},
				WeCom: WeComConfig{
					Corps: []CorpsConfig{
						{
							Name:   "main",
							CorpID: "ww1234567890abcdef",
							Apps: []AppConfig{
								{
									Name:    "oa",
									AgentID: 1000001,
									Secret:  "test-secret",
								},
							},
						},
					},
				},
				Auth: AuthConfig{
					RateLimit: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database driver",
			cfg: &Config{
				Server: ServerConfig{
					HTTPListen: ":8080",
					GRPCListen: ":9090",
					Mode:       "release",
				},
				Database: DatabaseConfig{
					Driver: "mysql",
					DSN:    "test.db",
				},
				WeCom: WeComConfig{
					Corps: []CorpsConfig{
						{
							Name:   "main",
							CorpID: "ww1234567890abcdef",
							Apps: []AppConfig{
								{
									Name:    "oa",
									AgentID: 1000001,
									Secret:  "test-secret",
								},
							},
						},
					},
				},
				Auth: AuthConfig{
					RateLimit: 100,
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConfig_GetCorpByName(t *testing.T) {
	cfg := &Config{
		WeCom: WeComConfig{
			Corps: []CorpsConfig{
				{
					Name:   "main",
					CorpID: "ww1234567890abcdef",
				},
				{
					Name:   "partner",
					CorpID: "ww0987654321fedcba",
				},
			},
		},
	}

	testCases := []struct {
		name    string
		corpName string
		wantErr bool
		wantID  string
	}{
		{"existing corp", "main", false, "ww1234567890abcdef"},
		{"existing corp 2", "partner", false, "ww0987654321fedcba"},
		{"non-existent corp", "nonexistent", true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			corp, err := cfg.GetCorpByName(tc.corpName)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetCorpByName() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				if corp.CorpID != tc.wantID {
					t.Errorf("Expected CorpID %s, got %s", tc.wantID, corp.CorpID)
				}
			}
		})
	}
}

func TestConfig_GetAppByName(t *testing.T) {
	cfg := &Config{
		WeCom: WeComConfig{
			Corps: []CorpsConfig{
				{
					Name:   "main",
					CorpID: "ww1234567890abcdef",
					Apps: []AppConfig{
						{
							Name:    "oa",
							AgentID: 1000001,
							Secret:  "secret1",
						},
						{
							Name:    "hr",
							AgentID: 1000002,
							Secret:  "secret2",
						},
					},
				},
			},
		},
	}

	testCases := []struct {
		name     string
		corpName string
		appName  string
		wantErr  bool
		wantID   int64
	}{
		{"existing app", "main", "oa", false, 1000001},
		{"existing app 2", "main", "hr", false, 1000002},
		{"non-existent corp", "nonexistent", "oa", true, 0},
		{"non-existent app", "main", "nonexistent", true, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app, err := cfg.GetAppByName(tc.corpName, tc.appName)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetAppByName() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				if app.AgentID != tc.wantID {
					t.Errorf("Expected AgentID %d, got %d", tc.wantID, app.AgentID)
				}
			}
		})
	}
}

func TestConfig_LoadWithEnv(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
server:
  http_listen: ":8080"
  grpc_listen: ":9090"
  mode: "release"

database:
  driver: "sqlite"
  dsn: "data/wecom.db"

wecom:
  corps:
    - name: "main"
      corp_id: "ww1234567890abcdef"
      apps:
        - name: "oa"
          agent_id: 1000001
          secret: "default_secret"

auth:
  admin_api_key: "wgk_admin_default"
  rate_limit: 100
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables
	os.Setenv("WECOM_HTTP_LISTEN", ":8888")
	os.Setenv("WECOM_RATE_LIMIT", "200")
	defer func() {
		os.Unsetenv("WECOM_HTTP_LISTEN")
		os.Unsetenv("WECOM_RATE_LIMIT")
	}()

	// Load config
	cfg, err := LoadWithEnv(configFile)
	if err != nil {
		t.Fatalf("LoadWithEnv failed: %v", err)
	}

	// Check that environment variables override config values
	if cfg.Server.HTTPListen != ":8888" {
		t.Errorf("Expected HTTPListen ':8888', got '%s'", cfg.Server.HTTPListen)
	}

	if cfg.Auth.RateLimit != 200 {
		t.Errorf("Expected RateLimit 200, got %d", cfg.Auth.RateLimit)
	}
}

func TestConfig_SQLiteDirCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	cfg := &Config{
		Server: ServerConfig{
			HTTPListen: ":8080",
			GRPCListen: ":9090",
			Mode:       "release",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    dbPath,
		},
		WeCom: WeComConfig{
			Corps: []CorpsConfig{
				{
					Name:   "main",
					CorpID: "ww1234567890abcdef",
					Apps: []AppConfig{
						{
							Name:    "oa",
							AgentID: 1000001,
							Secret:  "test-secret",
						},
					},
				},
			},
		},
		Auth: AuthConfig{
			RateLimit: 100,
		},
	}

	// Validate should create the directory
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Database directory was not created")
	}
}

func TestConfig_KeyExpiryDaysDefault(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			HTTPListen: ":8080",
			GRPCListen: ":9090",
			Mode:       "release",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "test.db",
		},
		WeCom: WeComConfig{
			Corps: []CorpsConfig{
				{
					Name:   "main",
					CorpID: "ww1234567890abcdef",
					Apps: []AppConfig{
						{
							Name:    "oa",
							AgentID: 1000001,
							Secret:  "test-secret",
						},
					},
				},
			},
		},
		Auth: AuthConfig{
			RateLimit:     100,
			KeyExpiryDays: 0, // Invalid value
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Should default to 365
	if cfg.Auth.KeyExpiryDays != 365 {
		t.Errorf("Expected KeyExpiryDays to default to 365, got %d", cfg.Auth.KeyExpiryDays)
	}
}
