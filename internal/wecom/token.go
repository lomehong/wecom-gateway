package wecom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
)

// TokenManager manages access tokens for WeChat Work API
type TokenManager struct {
	config   *config.Config
	httpCli  *http.Client
	tokens   map[string]*tokenInfo
	tokensMu sync.RWMutex
	encKey   []byte
}

// NewTokenManager creates a new token manager
func NewTokenManager(cfg *config.Config, encKey []byte) *TokenManager {
	return &TokenManager{
		config:  cfg,
		httpCli: &http.Client{Timeout: 30 * time.Second},
		tokens:  make(map[string]*tokenInfo),
		encKey:  encKey,
	}
}

// GetToken retrieves or fetches an access token for the given corp and app
func (tm *TokenManager) GetToken(ctx context.Context, corpName, appName string) (string, error) {
	key := corpName + "/" + appName

	// Check if we have a valid token
	tm.tokensMu.RLock()
	tok, exists := tm.tokens[key]
	tm.tokensMu.RUnlock()

	if exists && tok.expiresAt.After(time.Now().Add(5*time.Minute)) {
		return tok.token, nil
	}

	// Need to fetch a new token
	tm.tokensMu.Lock()
	defer tm.tokensMu.Unlock()

	// Double-check after acquiring write lock
	if tok, exists := tm.tokens[key]; exists && tok.expiresAt.After(time.Now().Add(5*time.Minute)) {
		return tok.token, nil
	}

	// Get app configuration
	app, err := tm.config.GetAppByName(corpName, appName)
	if err != nil {
		return "", fmt.Errorf("failed to get app config: %w", err)
	}

	corp, err := tm.config.GetCorpByName(corpName)
	if err != nil {
		return "", fmt.Errorf("failed to get corp config: %w", err)
	}

	// Decrypt secret
	secret, err := crypto.DecryptString(app.Secret, tm.encKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Fetch access token
	token, expiresAt, err := tm.fetchAccessToken(ctx, corp.CorpID, app.AgentID, secret)
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Store token
	tm.tokens[key] = &tokenInfo{
		token:     token,
		expiresAt: expiresAt,
		corpID:    corp.CorpID,
		agentID:   app.AgentID,
	}

	return token, nil
}

// InvalidateToken invalidates a cached token
func (tm *TokenManager) InvalidateToken(corpName, appName string) {
	key := corpName + "/" + appName
	tm.tokensMu.Lock()
	defer tm.tokensMu.Unlock()
	delete(tm.tokens, key)
}

// ClearAllTokens clears all cached tokens
func (tm *TokenManager) ClearAllTokens() {
	tm.tokensMu.Lock()
	defer tm.tokensMu.Unlock()
	tm.tokens = make(map[string]*tokenInfo)
}

// fetchAccessToken fetches a new access token from WeChat Work API
func (tm *TokenManager) fetchAccessToken(ctx context.Context, corpID string, agentID int64, secret string) (string, time.Time, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, secret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	resp, err := tm.httpCli.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", time.Time{}, err
	}

	if result.ErrCode != 0 {
		return "", time.Time{}, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	// Token expires in 7200 seconds (2 hours), subtract 5 minutes for safety
	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	return result.AccessToken, expiresAt, nil
}
