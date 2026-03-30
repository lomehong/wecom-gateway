package admin

import (
	"context"
	"time"

	"wecom-gateway/internal/apikey"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/store"
)

// Service handles admin operations
type Service struct {
	DB          store.Database
	config      *config.Config
	apiKeySvc   *apikey.Service
	auditLogger *audit.Logger
	encKey      []byte
}

// NewService creates a new admin service
func NewService(db store.Database, cfg *config.Config, apiKeySvc *apikey.Service, auditLogger *audit.Logger, encKey []byte) *Service {
	return &Service{
		DB:          db,
		config:      cfg,
		apiKeySvc:   apiKeySvc,
		auditLogger: auditLogger,
		encKey:      encKey,
	}
}

// InitializeSystem initializes the system with default data
func (s *Service) InitializeSystem(ctx context.Context) error {
	// Check if already initialized
	_, err := s.DB.GetWeComCorpByName(ctx, "main")
	if err == nil {
		return nil // Already initialized
	}

	// Create default corp from config
	if len(s.config.WeCom.Corps) == 0 {
		return nil
	}

	// Initialize corps and apps from config
	for _, corpCfg := range s.config.WeCom.Corps {
		// Create corp
		corpID := generateID("corp_")
		corp := &store.WeComCorp{
			ID:        corpID,
			Name:      corpCfg.Name,
			CorpID:    corpCfg.CorpID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.DB.CreateWeComCorp(ctx, corp); err != nil {
			if err != store.ErrDuplicate {
				return err
			}
		}

		// Create apps
		for _, appCfg := range corpCfg.Apps {
			// Encrypt secret
			nonce, err := crypto.GenerateRandomBytes(12)
			if err != nil {
				return err
			}

			secretEnc, err := crypto.EncryptString(appCfg.Secret, s.encKey)
			if err != nil {
				return err
			}

			appID := generateID("app_")
			app := &store.WeComApp{
				ID:        appID,
				Name:      appCfg.Name,
				CorpName:  corpCfg.Name,
				AppType:   "app",
				AgentID:   appCfg.AgentID,
				SecretEnc: secretEnc,
				Nonce:     string(nonce),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := s.DB.CreateWeComApp(ctx, app); err != nil {
				if err != store.ErrDuplicate {
					return err
				}
			}
		}
	}

	// Create admin API key if configured
	if s.config.Auth.AdminAPIKey != "" {
		// Parse the admin key
		if len(s.config.WeCom.Corps) > 0 {
			mainCorp := s.config.WeCom.Corps[0]
			_, err := s.apiKeySvc.CreateKey(ctx, &apikey.CreateKeyRequest{
				Name:        "admin",
				Permissions: []string{"*"},
				CorpName:    mainCorp.Name,
				AppName:     "", // Admin key
				ExpiresDays: 0, // No expiration
			})
			if err != nil {
				// Key might already exist
				return nil
			}
		}
	}

	return nil
}

// GetDashboardStats retrieves dashboard statistics
func (s *Service) GetDashboardStats(ctx context.Context, startTime, endTime time.Time) (*DashboardStats, error) {
	stats := &DashboardStats{
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Get total API key count
	keys, _, err := s.DB.ListAPIKeys(ctx, store.ListOptions{Limit: 1000})
	if err == nil {
		stats.TotalAPIKeys = len(keys)
		activeKeys := 0
		for _, key := range keys {
			if !key.Disabled {
				activeKeys++
			}
		}
		stats.ActiveAPIKeys = activeKeys
	}

	// Get corp and app counts
	corps, err := s.DB.ListWeComCorps(ctx)
	if err == nil {
		stats.TotalCorps = len(corps)
	}

	totalApps := 0
	for _, corp := range corps {
		apps, err := s.DB.ListWeComApps(ctx, corp.Name)
		if err == nil {
			totalApps += len(apps)
		}
	}
	stats.TotalApps = totalApps

	// Get audit log statistics for last 24 hours
	if !endTime.IsZero() && !startTime.IsZero() {
		logs, _, err := s.DB.QueryAuditLogs(ctx, store.AuditQueryOptions{
			StartTime: &startTime,
			EndTime:   &endTime,
			Limit:     10000,
		})

		if err == nil {
			stats.TotalRequests = len(logs)
			errorCount := 0
			for _, log := range logs {
				if log.StatusCode >= 400 {
					errorCount++
				}
			}
			stats.ErrorRequests = errorCount
		}
	}

	return stats, nil
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalAPIKeys  int       `json:"total_api_keys"`
	ActiveAPIKeys int       `json:"active_api_keys"`
	TotalCorps    int       `json:"total_corps"`
	TotalApps     int       `json:"total_apps"`
	TotalRequests int       `json:"total_requests"`
	ErrorRequests int       `json:"error_requests"`
}

func generateID(prefix string) string {
	return prefix + time.Now().Format("20060102150405") + "0000"
}
