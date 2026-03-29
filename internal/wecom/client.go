package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
)

// Client defines the interface for WeChat Work API client
type Client interface {
	// Schedule operations
	CreateSchedule(ctx context.Context, corpName, appName string, params *ScheduleParams) (*Schedule, error)
	GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*Schedule, error)
	UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *ScheduleParams) error
	DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error

	// Meeting room operations
	ListMeetingRooms(ctx context.Context, corpName, appName string, opts *RoomQueryOptions) ([]*MeetingRoom, string, error)
	GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*TimeSlot, error)
	BookMeetingRoom(ctx context.Context, corpName, appName string, params *BookingParams) (*BookingResult, error)

	// Message operations
	SendText(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error)
	SendMarkdown(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error)
	SendImage(ctx context.Context, corpName, appName string, params *ImageMessageParams) (*SendResult, error)
	SendFile(ctx context.Context, corpName, appName string, params *FileMessageParams) (*SendResult, error)
	SendCard(ctx context.Context, corpName, appName string, params *CardMessageParams) (*SendResult, error)

	// Media operations
	UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error)
}

// HTTPClient represents the HTTP client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// impl implements the Client interface
type impl struct {
	config   *config.Config
	httpCli  HTTPClient
	tokens   map[string]*tokenInfo
	tokensMu sync.RWMutex
	encKey   []byte // Encryption key for storing secrets
}

type tokenInfo struct {
	token      string
	expiresAt  time.Time
	corpID     string
	agentID    int64
	secret     string // Encrypted secret
	nonce      string // Nonce for AES-GCM
}

// NewClient creates a new WeChat Work API client
func NewClient(cfg *config.Config, encKey []byte) Client {
	return &impl{
		config:  cfg,
		httpCli: &http.Client{Timeout: 30 * time.Second},
		tokens:  make(map[string]*tokenInfo),
		encKey:  encKey,
	}
}

// getToken retrieves or fetches an access token for the given corp and app
func (c *impl) getToken(ctx context.Context, corpName, appName string) (string, error) {
	key := corpName + "/" + appName

	// Check if we have a valid token
	c.tokensMu.RLock()
	tok, exists := c.tokens[key]
	c.tokensMu.RUnlock()

	if exists && tok.expiresAt.After(time.Now().Add(5*time.Minute)) {
		return tok.token, nil
	}

	// Need to fetch a new token
	c.tokensMu.Lock()
	defer c.tokensMu.Unlock()

	// Double-check after acquiring write lock
	if tok, exists := c.tokens[key]; exists && tok.expiresAt.After(time.Now().Add(5*time.Minute)) {
		return tok.token, nil
	}

	// Get app configuration
	app, err := c.config.GetAppByName(corpName, appName)
	if err != nil {
		return "", fmt.Errorf("failed to get app config: %w", err)
	}

	corp, err := c.config.GetCorpByName(corpName)
	if err != nil {
		return "", fmt.Errorf("failed to get corp config: %w", err)
	}

	// Decrypt secret
	secret, err := crypto.DecryptString(app.Secret, c.encKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Fetch access token
	token, expiresAt, err := c.fetchAccessToken(ctx, corp.CorpID, app.AgentID, secret)
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Store token
	c.tokens[key] = &tokenInfo{
		token:     token,
		expiresAt: expiresAt,
		corpID:    corp.CorpID,
		agentID:   app.AgentID,
		secret:    app.Secret,
		nonce:     "", // Will be populated if needed
	}

	return token, nil
}

// fetchAccessToken fetches a new access token from WeChat Work API
func (c *impl) fetchAccessToken(ctx context.Context, corpID string, agentID int64, secret string) (string, time.Time, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, secret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	resp, err := c.httpCli.Do(req)
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

// makeAPIRequest makes an API request to WeChat Work
func (c *impl) makeAPIRequest(ctx context.Context, corpName, appName, method, path string, body interface{}) ([]byte, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return nil, err
	}

	url := "https://qyapi.weixin.qq.com/cgi-bin" + path + "?access_token=" + token

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for API errors
	var errResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.ErrCode != 0 {
		// If token expired, clear it and retry
		apiErr := &WeComAPIError{ErrCode: errResp.ErrCode, ErrMsg: errResp.ErrMsg}
		if apiErr.IsAccessTokenExpired() {
			key := corpName + "/" + appName
			c.tokensMu.Lock()
			delete(c.tokens, key)
			c.tokensMu.Unlock()
			// Retry once
			return c.makeAPIRequest(ctx, corpName, appName, method, path, body)
		}
		return nil, apiErr
	}

	return respBody, nil
}

// Schedule operations implementation

func (c *impl) CreateSchedule(ctx context.Context, corpName, appName string, params *ScheduleParams) (*Schedule, error) {
	body := map[string]interface{}{
		"organizer":         params.Organizer,
		"summary":           params.Summary,
		"start_time":        params.StartTime.Unix(),
		"end_time":          params.EndTime.Unix(),
		"remind_before_min": params.RemindBeforeMin,
	}

	if params.Description != "" {
		body["description"] = params.Description
	}
	if len(params.Attendees) > 0 {
		body["attendees"] = params.Attendees
	}
	if params.Location != "" {
		body["location"] = params.Location
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/schedule/add", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
		ScheduleID string `json:"scheduleid"`
		CalID      string `json:"cal_id"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &Schedule{
		ScheduleID:  result.ScheduleID,
		Organizer:   params.Organizer,
		Summary:     params.Summary,
		Description: params.Description,
		StartTime:   params.StartTime,
		EndTime:     params.EndTime,
		Attendees:   params.Attendees,
		Location:    params.Location,
		CalID:       result.CalID,
	}, nil
}

func (c *impl) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*Schedule, error) {
	body := map[string]interface{}{
		"userid":     userID,
		"start_time": startTime.Unix(),
		"end_time":   endTime.Unix(),
		"limit":      limit,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/schedule/get", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode   int        `json:"errcode"`
		ErrMsg    string     `json:"errmsg"`
		ScheduleList []Schedule `json:"schedulelist"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	schedules := make([]*Schedule, len(result.ScheduleList))
	for i := range result.ScheduleList {
		schedules[i] = &result.ScheduleList[i]
	}

	return schedules, nil
}

func (c *impl) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *ScheduleParams) error {
	body := map[string]interface{}{
		"scheduleid": scheduleID,
	}

	if params.Organizer != "" {
		body["organizer"] = params.Organizer
	}
	if params.Summary != "" {
		body["summary"] = params.Summary
	}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if !params.StartTime.IsZero() {
		body["start_time"] = params.StartTime.Unix()
	}
	if !params.EndTime.IsZero() {
		body["end_time"] = params.EndTime.Unix()
	}
	if len(params.Attendees) > 0 {
		body["attendees"] = params.Attendees
	}
	if params.Location != "" {
		body["location"] = params.Location
	}
	if params.RemindBeforeMin > 0 {
		body["remind_before_min"] = params.RemindBeforeMin
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/schedule/update", body)
	if err != nil {
		return err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return nil
}

func (c *impl) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error {
	body := map[string]interface{}{
		"scheduleid": scheduleID,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/schedule/del", body)
	if err != nil {
		return err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return nil
}

// Meeting room operations implementation

func (c *impl) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *RoomQueryOptions) ([]*MeetingRoom, string, error) {
	body := map[string]interface{}{
		"limit": 50,
	}

	if opts != nil {
		if opts.City != "" {
			body["city"] = opts.City
		}
		if opts.Building != "" {
			body["building"] = opts.Building
		}
		if opts.Floor != "" {
			body["floor"] = opts.Floor
		}
		if opts.Capacity > 0 {
			body["capacity"] = opts.Capacity
		}
		if len(opts.Equipment) > 0 {
			body["equipment"] = opts.Equipment
		}
		if opts.Limit > 0 {
			body["limit"] = opts.Limit
		}
		if opts.Cursor != "" {
			body["cursor"] = opts.Cursor
		}
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meetingroom/list", body)
	if err != nil {
		return nil, "", err
	}

	var result struct {
		ErrCode      int          `json:"errcode"`
		ErrMsg       string       `json:"errmsg"`
		MeetingRoomList []MeetingRoom `json:"meetingroom_list"`
		NextCursor   string       `json:"next_cursor"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}

	if result.ErrCode != 0 {
		return nil, "", &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	rooms := make([]*MeetingRoom, len(result.MeetingRoomList))
	for i := range result.MeetingRoomList {
		rooms[i] = &result.MeetingRoomList[i]
	}

	return rooms, result.NextCursor, nil
}

func (c *impl) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*TimeSlot, error) {
	body := map[string]interface{}{
		"meetingroomid": roomID,
		"start_time":    start.Unix(),
		"end_time":      end.Unix(),
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meetingroom/get_available_timeslots", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode     int        `json:"errcode"`
		ErrMsg      string     `json:"errmsg"`
		TimeSlotList []TimeSlot `json:"timeslot_list"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	slots := make([]*TimeSlot, len(result.TimeSlotList))
	for i := range result.TimeSlotList {
		slots[i] = &result.TimeSlotList[i]
	}

	return slots, nil
}

func (c *impl) BookMeetingRoom(ctx context.Context, corpName, appName string, params *BookingParams) (*BookingResult, error) {
	body := map[string]interface{}{
		"meetingroomid": params.MeetingRoomID,
		"subject":       params.Subject,
		"start_time":    params.StartTime.Unix(),
		"end_time":      params.EndTime.Unix(),
		"booker":        params.Booker,
	}

	if len(params.Attendees) > 0 {
		body["attendees"] = params.Attendees
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meetingroom/book", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
		BookingID  string `json:"booking_id"`
		ScheduleID string `json:"schedule_id"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &BookingResult{
		BookingID:  result.BookingID,
		ScheduleID: result.ScheduleID,
		StartTime:  params.StartTime,
		EndTime:    params.EndTime,
	}, nil
}

// Message operations implementation

func (c *impl) SendText(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error) {
	body := map[string]interface{}{
		"touser":  joinStrings(params.ReceiverIDs, "|"),
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": params.Content,
		},
	}

	if params.Safe {
		body["safe"] = 1
	}

	return c.sendMessage(ctx, corpName, appName, body)
}

func (c *impl) SendMarkdown(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error) {
	body := map[string]interface{}{
		"touser":  joinStrings(params.ReceiverIDs, "|"),
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": params.Content,
		},
	}

	return c.sendMessage(ctx, corpName, appName, body)
}

func (c *impl) SendImage(ctx context.Context, corpName, appName string, params *ImageMessageParams) (*SendResult, error) {
	mediaID := params.MediaID
	if mediaID == "" && params.ImageURL != "" {
		// Upload image first
		// This is a simplified version - in production, you'd download and upload
		return nil, fmt.Errorf("image URL upload not implemented, please use media_id")
	}

	body := map[string]interface{}{
		"touser":  joinStrings(params.ReceiverIDs, "|"),
		"msgtype": "image",
		"image": map[string]interface{}{
			"media_id": mediaID,
		},
	}

	return c.sendMessage(ctx, corpName, appName, body)
}

func (c *impl) SendFile(ctx context.Context, corpName, appName string, params *FileMessageParams) (*SendResult, error) {
	mediaID := params.MediaID
	if mediaID == "" && params.FileURL != "" {
		return nil, fmt.Errorf("file URL upload not implemented, please use media_id")
	}

	body := map[string]interface{}{
		"touser":  joinStrings(params.ReceiverIDs, "|"),
		"msgtype": "file",
		"file": map[string]interface{}{
			"media_id": mediaID,
		},
	}

	return c.sendMessage(ctx, corpName, appName, body)
}

func (c *impl) SendCard(ctx context.Context, corpName, appName string, params *CardMessageParams) (*SendResult, error) {
	body := map[string]interface{}{
		"touser":  joinStrings(params.ReceiverIDs, "|"),
		"msgtype": "interactive",
		"interactive": params.CardContent,
	}

	return c.sendMessage(ctx, corpName, appName, body)
}

func (c *impl) sendMessage(ctx context.Context, corpName, appName string, body map[string]interface{}) (*SendResult, error) {
	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/message/send", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode        int      `json:"errcode"`
		ErrMsg         string   `json:"errmsg"`
		InvalidUser    string   `json:"invaliduser"`
		InvalidParty   string   `json:"invalidparty"`
		InvalidTag     string   `json:"invalidtag"`
		UnquotedUser   string   `json:"unquoteduser"`
		FailedUser     string   `json:"faileduser"`
		FailedParty    string   `json:"failedparty"`
		FailedTag      string   `json:"failedtag"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &SendResult{
		InvalidUserIDs:  splitStrings(result.InvalidUser, "|"),
		InvalidPartyIDs: splitStrings(result.InvalidParty, "|"),
		InvalidTagIDs:   splitStrings(result.InvalidTag, "|"),
		UnquotedUserIDs: splitStrings(result.UnquotedUser, "|"),
		FailedUserIDs:   splitStrings(result.FailedUser, "|"),
		FailedPartyIDs:  splitStrings(result.FailedParty, "|"),
		FailedTagIDs:    splitStrings(result.FailedTag, "|"),
	}, nil
}

// Media operations implementation

func (c *impl) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return "", err
	}

	_ = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s", token, mediaType)

	// Create multipart form data
	// This is a simplified version - in production, use proper multipart writer
	return "", fmt.Errorf("media upload not fully implemented")
}

// Helper functions

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func splitStrings(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
