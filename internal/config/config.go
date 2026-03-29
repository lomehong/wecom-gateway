package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	WeCom    WeComConfig    `yaml:"wecom"`
	Auth     AuthConfig     `yaml:"auth"`
	UI       UIConfig       `yaml:"ui"`
}

// ServerConfig holds HTTP and gRPC server configuration
type ServerConfig struct {
	HTTPListen string `yaml:"http_listen"`
	GRPCListen string `yaml:"grpc_listen"`
	Mode       string `yaml:"mode"` // debug, release
	TLSCert    string `yaml:"tls_cert"`
	TLSKey     string `yaml:"tls_key"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver string `yaml:"driver"` // sqlite, postgres
	DSN    string `yaml:"dsn"`
}

// WeComConfig holds WeChat Work configuration
type WeComConfig struct {
	Corps []CorpsConfig `yaml:"corps"`
}

// CorpsConfig represents a WeChat Work corporation
type CorpsConfig struct {
	Name   string      `yaml:"name"`
	CorpID string      `yaml:"corp_id"`
	Apps   []AppConfig `yaml:"apps"`
}

// AppConfig represents a WeChat Work application
type AppConfig struct {
	Name   string `yaml:"name"`
	AgentID int64  `yaml:"agent_id"`
	Secret  string `yaml:"secret"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	AdminAPIKey  string `yaml:"admin_api_key"`
	KeyExpiryDays int   `yaml:"key_expiry_days"`
	RateLimit    int    `yaml:"rate_limit"`
}

// UIConfig holds management UI configuration
type UIConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Load YAML configuration
	cfg := &Config{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Load environment variables and override config
	if err := loadEnvOverrides(cfg); err != nil {
		return nil, fmt.Errorf("failed to load environment overrides: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithEnv loads configuration with .env file support
func LoadWithEnv(configPath string) (*Config, error) {
	// Try to load .env file if it exists
	loadEnvFile()

	return Load(configPath)
}

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() {
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return
	}

	// Parse .env file
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"`)
			value = strings.Trim(value, `'`)
			os.Setenv(key, value)
		}
	}
}

// loadEnvOverrides loads environment variables and overrides config values
func loadEnvOverrides(cfg *Config) error {
	// Server overrides
	if v := os.Getenv("WECOM_HTTP_LISTEN"); v != "" {
		cfg.Server.HTTPListen = v
	}
	if v := os.Getenv("WECOM_GRPC_LISTEN"); v != "" {
		cfg.Server.GRPCListen = v
	}
	if v := os.Getenv("WECOM_SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("WECOM_TLS_CERT"); v != "" {
		cfg.Server.TLSCert = v
	}
	if v := os.Getenv("WECOM_TLS_KEY"); v != "" {
		cfg.Server.TLSKey = v
	}

	// Database overrides
	if v := os.Getenv("WECOM_DB_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}
	if v := os.Getenv("WECOM_DB_DSN"); v != "" {
		cfg.Database.DSN = v
	}

	// WeCom overrides for specific corps
	for i := range cfg.WeCom.Corps {
		corpName := cfg.WeCom.Corps[i].Name
		if v := os.Getenv(fmt.Sprintf("WECOM_CORP_%s_ID", strings.ToUpper(corpName))); v != "" {
			cfg.WeCom.Corps[i].CorpID = v
		}

		// App overrides
		for j := range cfg.WeCom.Corps[i].Apps {
			appName := cfg.WeCom.Corps[i].Apps[j].Name
			if v := os.Getenv(fmt.Sprintf("WECOM_APP_%s_%s_SECRET", strings.ToUpper(corpName), strings.ToUpper(appName))); v != "" {
				cfg.WeCom.Corps[i].Apps[j].Secret = v
			}
		}
	}

	// Auth overrides
	if v := os.Getenv("WECOM_ADMIN_API_KEY"); v != "" {
		cfg.Auth.AdminAPIKey = v
	}
	if v := os.Getenv("WECOM_KEY_EXPIRY_DAYS"); v != "" {
		var days int
		if _, err := fmt.Sscanf(v, "%d", &days); err == nil {
			cfg.Auth.KeyExpiryDays = days
		}
	}
	if v := os.Getenv("WECOM_RATE_LIMIT"); v != "" {
		var limit int
		if _, err := fmt.Sscanf(v, "%d", &limit); err == nil {
			cfg.Auth.RateLimit = limit
		}
	}

	// UI overrides
	if v := os.Getenv("WECOM_UI_ENABLED"); v != "" {
		cfg.UI.Enabled = strings.ToLower(v) == "true" || v == "1"
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.HTTPListen == "" {
		return fmt.Errorf("server.http_listen is required")
	}
	if c.Server.GRPCListen == "" {
		return fmt.Errorf("server.grpc_listen is required")
	}
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("server.mode must be 'debug' or 'release'")
	}

	// Validate database config
	if c.Database.Driver != "sqlite" && c.Database.Driver != "postgres" {
		return fmt.Errorf("database.driver must be 'sqlite' or 'postgres'")
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}

	// For SQLite, ensure directory exists
	if c.Database.Driver == "sqlite" {
		dir := filepath.Dir(c.Database.DSN)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}
	}

	// Validate WeCom config
	if len(c.WeCom.Corps) == 0 {
		return fmt.Errorf("at least one WeChat Work corporation must be configured")
	}

	corpNames := make(map[string]bool)
	for _, corp := range c.WeCom.Corps {
		if corp.Name == "" {
			return fmt.Errorf("corporation name is required")
		}
		if corp.CorpID == "" {
			return fmt.Errorf("corp_id is required for corporation '%s'", corp.Name)
		}
		if corpNames[corp.Name] {
			return fmt.Errorf("duplicate corporation name: '%s'", corp.Name)
		}
		corpNames[corp.Name] = true

		appNames := make(map[string]bool)
		for _, app := range corp.Apps {
			if app.Name == "" {
				return fmt.Errorf("app name is required for corporation '%s'", corp.Name)
			}
			if app.AgentID == 0 {
				return fmt.Errorf("agent_id is required for app '%s' in corporation '%s'", app.Name, corp.Name)
			}
			if app.Secret == "" {
				return fmt.Errorf("secret is required for app '%s' in corporation '%s'", app.Name, corp.Name)
			}
			if appNames[app.Name] {
				return fmt.Errorf("duplicate app name: '%s' in corporation '%s'", app.Name, corp.Name)
			}
			appNames[app.Name] = true
		}

		if len(corp.Apps) == 0 {
			return fmt.Errorf("at least one app must be configured for corporation '%s'", corp.Name)
		}
	}

	// Validate auth config
	if c.Auth.RateLimit <= 0 {
		return fmt.Errorf("auth.rate_limit must be positive")
	}
	if c.Auth.KeyExpiryDays <= 0 {
		c.Auth.KeyExpiryDays = 365 // default
	}

	return nil
}

// GetCorpByName retrieves corporation configuration by name
func (c *Config) GetCorpByName(name string) (*CorpsConfig, error) {
	for i := range c.WeCom.Corps {
		if c.WeCom.Corps[i].Name == name {
			return &c.WeCom.Corps[i], nil
		}
	}
	return nil, fmt.Errorf("corporation '%s' not found", name)
}

// GetAppByName retrieves app configuration by corporation and app name
func (c *Config) GetAppByName(corpName, appName string) (*AppConfig, error) {
	corp, err := c.GetCorpByName(corpName)
	if err != nil {
		return nil, err
	}

	for i := range corp.Apps {
		if corp.Apps[i].Name == appName {
			return &corp.Apps[i], nil
		}
	}
	return nil, fmt.Errorf("app '%s' not found in corporation '%s'", appName, corpName)
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPListen: ":8080",
			GRPCListen: ":9090",
			Mode:       "release",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "data/wecom.db",
		},
		Auth: AuthConfig{
			KeyExpiryDays: 365,
			RateLimit:     100,
		},
		UI: UIConfig{
			Enabled: true,
		},
	}
}
