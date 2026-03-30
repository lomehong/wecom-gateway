package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
	"wecom-gateway/internal/store"
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

	// Token operations
	GetAccessToken(ctx context.Context, corpName, appName string) (string, error)

	// Contact operations
	GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*ContactUser, error)
	SearchUser(ctx context.Context, corpName, appName string, query string) ([]*ContactUser, error)

	// Todo operations
	GetTodoList(ctx context.Context, corpName, appName string, opts *TodoListOptions) (*TodoListResult, error)
	GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*TodoDetail, error)
	CreateTodo(ctx context.Context, corpName, appName string, params *CreateTodoParams) (string, error)
	UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *UpdateTodoParams) error
	DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error
	ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error

	// Meeting appointment operations (Phase 1.3)
	CreateMeeting(ctx context.Context, corpName, appName string, params *CreateMeetingParams) (*MeetingInfo, error)
	CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error
	UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *MeetingInvitees) error
	ListMeetings(ctx context.Context, corpName, appName string, opts *MeetingListOptions) (*MeetingListResult, error)
	GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*MeetingInfo, error)

	// Message pull operations (Phase 3.1)
	GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*ChatListResult, error)
	GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*ChatMessagesResult, error)
	DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error)

	// Schedule availability operations (Phase 3.2)
	CheckAvailability(ctx context.Context, corpName, appName string, opts *AvailabilityOptions) ([]*UserAvailability, error)
}

// HTTPClient represents the HTTP client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// impl implements the Client interface
type impl struct {
	config   *config.Config
	db       store.Database // Database for reading corp/app config
	httpCli  HTTPClient
	tokens   map[string]*tokenInfo
	tokensMu sync.RWMutex
	encKey   []byte // Encryption key for storing secrets
}

type tokenInfo struct {
	token     string
	expiresAt time.Time
	corpID    string
	agentID   int64
	secret    string // Encrypted secret
	nonce     string // Nonce for AES-GCM
}

// NewClient creates a new WeChat Work API client
func NewClient(cfg *config.Config, encKey []byte) Client {
	return &impl{
		config:  cfg,
		db:      nil,
		httpCli: &http.Client{Timeout: 30 * time.Second},
		tokens:  make(map[string]*tokenInfo),
		encKey:  encKey,
	}
}

// NewClientWithDB creates a new WeChat Work API client with database support
func NewClientWithDB(cfg *config.Config, db store.Database, encKey []byte) Client {
	return &impl{
		config:  cfg,
		db:      db,
		httpCli: &http.Client{Timeout: 30 * time.Second},
		tokens:  make(map[string]*tokenInfo),
		encKey:  encKey,
	}
}

// SetDB sets the database for the client (allows post-construction injection)
func (c *impl) SetDB(db store.Database) {
	c.db = db
}

// GetAccessToken retrieves or fetches an access token for the given corp and app
func (c *impl) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return c.getToken(ctx, corpName, appName)
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

	// Try database first, then fall back to config
	var corpID string
	var agentID int64
	var secret string

	if c.db != nil {
		// Look up from database
		corp, err := c.db.GetWeComCorpByName(ctx, corpName)
		if err == nil {
			corpID = corp.CorpID
		}

		app, err := c.db.GetWeComApp(ctx, corpName, appName)
		if err == nil {
			agentID = app.AgentID
			secret = app.SecretEnc
		}
	}

	// Fall back to config if not found in database
	if corpID == "" {
		corp, err := c.config.GetCorpByName(corpName)
		if err != nil {
			return "", fmt.Errorf("corporation '%s' not found in database or config", corpName)
		}
		corpID = corp.CorpID
	}

	if agentID == 0 {
		app, err := c.config.GetAppByName(corpName, appName)
		if err != nil {
			return "", fmt.Errorf("application '%s' not found in corporation '%s'", appName, corpName)
		}
		agentID = app.AgentID
		secret = app.Secret
	}

	// Decrypt secret
	decryptedSecret, err := crypto.DecryptString(secret, c.encKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Fetch access token
	token, expiresAt, err := c.fetchAccessToken(ctx, corpID, agentID, decryptedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token: %w", err)
	}

	// Store token
	c.tokens[key] = &tokenInfo{
		token:     token,
		expiresAt: expiresAt,
		corpID:    corpID,
		agentID:   agentID,
		secret:    secret,
		nonce:     "",
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
		// Download image and upload to WeCom
		data, filename, err := c.downloadURL(ctx, params.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download image: %w", err)
		}
		mediaID, err = c.UploadMedia(ctx, corpName, appName, "image", data, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}
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
		// Download file and upload to WeCom
		data, filename, err := c.downloadURL(ctx, params.FileURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download file: %w", err)
		}
		mediaID, err = c.UploadMedia(ctx, corpName, appName, "file", data, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file: %w", err)
		}
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

	uploadURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s", token, mediaType)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
		MediaID  string `json:"media_id"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return result.MediaID, nil
}

// downloadURL downloads content from a URL and returns the data and filename
func (c *impl) downloadURL(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Extract filename from URL or Content-Disposition
	filename := "downloaded_file"
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fn, ok := params["filename"]; ok {
				filename = fn
			}
		}
	}

	return data, filename, nil
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

// Contact operations implementation

func (c *impl) GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*ContactUser, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return nil, err
	}

	// Use /user/list for detailed user info
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/list?access_token=%s&department_id=%d", token, departmentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode int            `json:"errcode"`
		ErrMsg  string         `json:"errmsg"`
		UserList []ContactUser `json:"userlist"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	users := make([]*ContactUser, len(result.UserList))
	for i := range result.UserList {
		users[i] = &result.UserList[i]
	}

	return users, nil
}

func (c *impl) SearchUser(ctx context.Context, corpName, appName string, query string) ([]*ContactUser, error) {
	// Search user: first get the full user list from root department (id=1),
	// then filter by name/alias matching query
	users, err := c.GetUserList(ctx, corpName, appName, 1)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matched []*ContactUser
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Name), queryLower) ||
			strings.Contains(strings.ToLower(user.Alias), queryLower) ||
			strings.Contains(strings.ToLower(user.UserID), queryLower) {
			matched = append(matched, user)
		}
	}

	return matched, nil
}

// Todo operations implementation

func (c *impl) GetTodoList(ctx context.Context, corpName, appName string, opts *TodoListOptions) (*TodoListResult, error) {
	if opts == nil {
		opts = &TodoListOptions{}
	}

	body := map[string]interface{}{
		"size": 100,
	}
	if opts.CreateBeginTime != nil {
		body["create_begin_time"] = opts.CreateBeginTime.Unix()
	}
	if opts.CreateEndTime != nil {
		body["create_end_time"] = opts.CreateEndTime.Unix()
	}
	if opts.RemindBeginTime != nil {
		body["remind_begin_time"] = opts.RemindBeginTime.Unix()
	}
	if opts.RemindEndTime != nil {
		body["remind_end_time"] = opts.RemindEndTime.Unix()
	}
	if opts.Cursor != "" {
		body["cursor"] = opts.Cursor
	}
	if opts.Limit > 0 {
		body["size"] = opts.Limit
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/list", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode    int         `json:"errcode"`
		ErrMsg     string      `json:"errmsg"`
		IndexList  []TodoIndex `json:"index_list"`
		NextCursor string      `json:"next_cursor"`
		HasMore    bool        `json:"has_more"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &TodoListResult{
		IndexList:  result.IndexList,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}

func (c *impl) GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*TodoDetail, error) {
	if len(todoIDs) == 0 {
		return nil, fmt.Errorf("todo_ids is required")
	}

	body := map[string]interface{}{
		"todo_id_list": todoIDs,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/get", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode   int          `json:"errcode"`
		ErrMsg    string       `json:"errmsg"`
		TodoList  []TodoDetail `json:"todo_list"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	details := make([]*TodoDetail, len(result.TodoList))
	for i := range result.TodoList {
		details[i] = &result.TodoList[i]
	}

	return details, nil
}

func (c *impl) CreateTodo(ctx context.Context, corpName, appName string, params *CreateTodoParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params is required")
	}

	body := map[string]interface{}{
		"content": params.Content,
	}

	if params.Title != "" {
		body["title"] = params.Title
	}
	if len(params.Assignees) > 0 {
		body["assigned_user_list"] = params.Assignees
	}
	if params.RemindTime != nil {
		body["remind_time"] = params.RemindTime.Unix()
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/add", body)
	if err != nil {
		return "", err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		TodoID  string `json:"todo_id"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return result.TodoID, nil
}

func (c *impl) UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *UpdateTodoParams) error {
	if params == nil {
		return fmt.Errorf("params is required")
	}

	body := map[string]interface{}{
		"todo_id": todoID,
	}

	if params.Title != nil {
		body["title"] = *params.Title
	}
	if params.Content != nil {
		body["content"] = *params.Content
	}
	if params.Status != nil {
		body["status"] = *params.Status
	}
	if params.RemindTime != nil {
		body["remind_time"] = params.RemindTime.Unix()
	}
	if len(params.Assignees) > 0 {
		body["assigned_user_list"] = params.Assignees
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/update", body)
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

func (c *impl) DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error {
	body := map[string]interface{}{
		"todo_id": todoID,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/del", body)
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

func (c *impl) ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error {
	body := map[string]interface{}{
		"todo_id": todoID,
		"status":  status,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/todo/update_user_status", body)
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

// Meeting appointment operations implementation (Phase 1.3)

func (c *impl) CreateMeeting(ctx context.Context, corpName, appName string, params *CreateMeetingParams) (*MeetingInfo, error) {
	body := map[string]interface{}{
		"title":                    params.Title,
		"meeting_start_datetime":   params.StartDateTime.Unix(),
		"meeting_duration":         params.Duration,
	}

	if params.Invitees != nil {
		body["invitees"] = params.Invitees
	}
	if params.MeetingType > 0 {
		body["meeting_type"] = params.MeetingType
	}
	if params.Settings != nil {
		body["settings"] = params.Settings
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meeting/create_meeting", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MeetingInfo
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &result.MeetingInfo, nil
}

func (c *impl) CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error {
	body := map[string]interface{}{
		"meetingid": meetingID,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meeting/cancel_meeting", body)
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

func (c *impl) UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *MeetingInvitees) error {
	body := map[string]interface{}{
		"meetingid": meetingID,
		"invitees":  invitees,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meeting/set_invite_meeting_members", body)
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

func (c *impl) ListMeetings(ctx context.Context, corpName, appName string, opts *MeetingListOptions) (*MeetingListResult, error) {
	body := map[string]interface{}{
		"begin_datetime": opts.BeginDatetime,
		"end_datetime":   opts.EndDatetime,
	}

	if opts.Limit > 0 {
		body["limit"] = opts.Limit
	}
	if opts.Cursor != "" {
		body["cursor"] = opts.Cursor
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meeting/list_user_meetings", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode    int            `json:"errcode"`
		ErrMsg     string         `json:"errmsg"`
		MeetingList []MeetingInfo `json:"meeting_list"`
		NextCursor string         `json:"next_cursor"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &MeetingListResult{
		Meetings:   result.MeetingList,
		NextCursor: result.NextCursor,
	}, nil
}

func (c *impl) GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*MeetingInfo, error) {
	body := map[string]interface{}{
		"meetingid": meetingID,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/meeting/get_meeting_info", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MeetingInfo
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &result.MeetingInfo, nil
}

// Message pull operations implementation (Phase 3.1)

func (c *impl) GetChatList(ctx context.Context, corpName, appName string, beginTime, endTime int64) (*ChatListResult, error) {
	body := map[string]interface{}{
		"begin_time": beginTime,
		"end_time":   endTime,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/message/get_msg_chat_list", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode  int        `json:"errcode"`
		ErrMsg   string     `json:"errmsg"`
		ChatList []ChatInfo `json:"chat_list"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &ChatListResult{
		ChatList: result.ChatList,
	}, nil
}

func (c *impl) GetChatMessages(ctx context.Context, corpName, appName string, chatType int, chatID string, beginTime, endTime int64) (*ChatMessagesResult, error) {
	body := map[string]interface{}{
		"chat_type":  chatType,
		"chatid":     chatID,
		"begin_time": beginTime,
		"end_time":   endTime,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/message/get_message", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode int          `json:"errcode"`
		ErrMsg  string       `json:"errmsg"`
		MsgList []ChatMessage `json:"msg_list"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return &ChatMessagesResult{
		MsgList: result.MsgList,
	}, nil
}

func (c *impl) DownloadMedia(ctx context.Context, corpName, appName string, mediaID string) ([]byte, string, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return nil, "", err
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s", token, mediaID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Check for JSON error response
	if resp.Header.Get("Content-Type") == "application/json" || len(respBody) < 200 {
		var errResp struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.ErrCode != 0 {
			return nil, "", &WeComAPIError{ErrCode: errResp.ErrCode, ErrMsg: errResp.ErrMsg}
		}
	}

	// Extract filename from Content-Disposition
	filename := "downloaded_media"
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fn, ok := params["filename"]; ok {
				filename = fn
			}
		}
	}

	return respBody, filename, nil
}

// Schedule availability operations implementation (Phase 3.2)

func (c *impl) CheckAvailability(ctx context.Context, corpName, appName string, opts *AvailabilityOptions) ([]*UserAvailability, error) {
	body := map[string]interface{}{
		"userids":    opts.UserIDs,
		"start_time": opts.StartTime,
		"end_time":   opts.EndTime,
	}

	resp, err := c.makeAPIRequest(ctx, corpName, appName, "POST", "/oa/schedule/check_availability", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ErrCode   int               `json:"errcode"`
		ErrMsg    string            `json:"errmsg"`
		AvailList []UserAvailability `json:"availability_list"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &WeComAPIError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	availabilities := make([]*UserAvailability, len(result.AvailList))
	for i := range result.AvailList {
		availabilities[i] = &result.AvailList[i]
	}

	return availabilities, nil
}
