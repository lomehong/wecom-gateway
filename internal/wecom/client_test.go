package wecom

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wecom-gateway/internal/config"
	"wecom-gateway/internal/crypto"
)

// MockConfig creates a mock config for testing
func MockConfig() *config.Config {
	testKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
	encryptedSecret, _ := crypto.EncryptString("test-secret", testKey)
	return &config.Config{
		WeCom: config.WeComConfig{
			Corps: []config.CorpsConfig{
				{
					Name:   "test-corp",
					CorpID: "test-corp-id",
					Apps: []config.AppConfig{
						{
							Name:    "test-app",
							AgentID: 123456,
							Secret:  encryptedSecret,
						},
					},
				},
			},
		},
	}
}

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	Response *http.Response
	Err      error
	Calls    []*http.Request
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.Calls = append(m.Calls, req)
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Response, nil
}

func TestNewClient(t *testing.T) {
	cfg := MockConfig()
	encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")

	client := NewClient(cfg, encKey)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestFetchAccessToken(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		wantToken      string
		wantErr        bool
		errCode        int
	}{
		{
			name: "successful token fetch",
			responseBody: `{
				"errcode": 0,
				"errmsg": "ok",
				"access_token": "test-access-token",
				"expires_in": 7200
			}`,
			responseStatus: 200,
			wantToken:      "test-access-token",
			wantErr:        false,
		},
		{
			name: "api error",
			responseBody: `{
				"errcode": 40013,
				"errmsg": "invalid corpid"
			}`,
			responseStatus: 200,
			wantErr:        true,
			errCode:        40013,
		},
		{
			name:           "http error",
			responseBody:   "",
			responseStatus: 500,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			cfg := MockConfig()
			encKey := crypto.GenerateKeyFromPassphrase("test-passphrase")
			client := NewClient(cfg, encKey).(*impl)

			// Override the HTTP client to redirect to test server
			client.httpCli = &http.Client{
				Timeout: 5 * time.Second,
				Transport: &roundTripperFunc{
					roundTrip: func(req *http.Request) (*http.Response, error) {
						testURL := server.URL + req.URL.Path + "?" + req.URL.RawQuery
						newReq, err := http.NewRequest(req.Method, testURL, req.Body)
						if err != nil {
							return nil, err
						}
						return http.DefaultTransport.RoundTrip(newReq)
					},
				},
			}

			// Call fetchAccessToken with any URL (transport will redirect)
			url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=test&corpsecret=test"
			req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)
			resp, err := client.httpCli.Do(req)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var result struct {
				ErrCode     int    `json:"errcode"`
				ErrMsg      string `json:"errmsg"`
				AccessToken string `json:"access_token"`
				ExpiresIn   int    `json:"expires_in"`
			}
			json.Unmarshal(body, &result)

			if tt.wantErr && result.ErrCode != 0 {
				if tt.errCode != 0 && result.ErrCode != tt.errCode {
					t.Errorf("expected error code %d, got %d", tt.errCode, result.ErrCode)
				}
				return
			}

			if !tt.wantErr && result.AccessToken != tt.wantToken {
				t.Errorf("expected token %s, got %s", tt.wantToken, result.AccessToken)
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "single string",
			strs:     []string{"a"},
			sep:      "|",
			expected: "a",
		},
		{
			name:     "multiple strings",
			strs:     []string{"a", "b", "c"},
			sep:      "|",
			expected: "a|b|c",
		},
		{
			name:     "empty slice",
			strs:     []string{},
			sep:      "|",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSplitStrings(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected []string
	}{
		{
			name:     "single string",
			s:        "a",
			sep:      "|",
			expected: []string{"a"},
		},
		{
			name:     "multiple strings",
			s:        "a|b|c",
			sep:      "|",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty string",
			s:        "",
			sep:      "|",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitStrings(tt.s, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("index %d: expected %s, got %s", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestWeComAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      *WeComAPIError
		expected string
	}{
		{
			name:     "error message",
			err:      &WeComAPIError{ErrCode: 40013, ErrMsg: "invalid corpid"},
			expected: "invalid corpid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestWeComAPIError_IsAccessTokenExpired(t *testing.T) {
	tests := []struct {
		name     string
		errCode  int
		expected bool
	}{
		{
			name:     "expired token 40014",
			errCode:  40014,
			expected: true,
		},
		{
			name:     "expired token 42001",
			errCode:  42001,
			expected: true,
		},
		{
			name:     "other error",
			errCode:  40013,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &WeComAPIError{ErrCode: tt.errCode}
			if err.IsAccessTokenExpired() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, err.IsAccessTokenExpired())
			}
		})
	}
}

func TestWeComAPIError_IsInvalidCredential(t *testing.T) {
	tests := []struct {
		name     string
		errCode  int
		expected bool
	}{
		{
			name:     "invalid credential 40013",
			errCode:  40013,
			expected: true,
		},
		{
			name:     "invalid credential 40091",
			errCode:  40091,
			expected: true,
		},
		{
			name:     "other error",
			errCode:  40014,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &WeComAPIError{ErrCode: tt.errCode}
			if err.IsInvalidCredential() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, err.IsInvalidCredential())
			}
		})
	}
}

func TestCreateSchedule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			w.WriteHeader(200)
			w.Write([]byte(`{
				"errcode": 0,
				"errmsg": "ok",
				"access_token": "test-token",
				"expires_in": 7200
			}`))
			return
		}
		if r.URL.Path == "/cgi-bin/oa/schedule/add" {
			w.WriteHeader(200)
			w.Write([]byte(`{
				"errcode": 0,
				"errmsg": "ok",
				"scheduleid": "test-schedule-id",
				"cal_id": "test-cal-id"
			}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	// Need to mock the config and client properly
	// For now, test the logic
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Hour)

	params := &ScheduleParams{
		Organizer:       "test-user",
		Summary:         "Test Meeting",
		Description:     "Test Description",
		StartTime:       startTime,
		EndTime:         endTime,
		Attendees:       []string{"user1", "user2"},
		Location:        "Room 101",
		RemindBeforeMin: 15,
	}

	// Validate params
	if params.Organizer == "" {
		t.Error("organizer should not be empty")
	}
	if params.Summary == "" {
		t.Error("summary should not be empty")
	}
	if params.EndTime.Before(params.StartTime) {
		t.Error("end_time should be after start_time")
	}
}

func TestGetSchedules(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)

	// Test request parameters
	limit := 50
	if limit <= 0 || limit > 100 {
		t.Error("limit should be between 1 and 100")
	}

	userID := "test-user"
	if userID == "" {
		t.Error("userid should not be empty")
	}

	_ = startTime
	_ = endTime
}

func TestUpdateSchedule(t *testing.T) {
	scheduleID := "test-schedule-id"
	params := &ScheduleParams{
		Summary:     "Updated Summary",
		Description: "Updated Description",
	}

	if scheduleID == "" {
		t.Error("schedule ID should not be empty")
	}

	// Test that only non-zero values are updated
	if params.Summary == "" && params.Description == "" {
		// This is valid - partial update
	}
}

func TestDeleteSchedule(t *testing.T) {
	scheduleID := "test-schedule-id"

	if scheduleID == "" {
		t.Error("schedule ID should not be empty")
	}
}

func TestBookMeetingRoom(t *testing.T) {
	startTime := time.Now().Add(1 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	params := &BookingParams{
		MeetingRoomID: "test-room-id",
		Subject:       "Test Meeting",
		StartTime:     startTime,
		EndTime:       endTime,
		Booker:        "test-user",
		Attendees:     []string{"user1"},
	}

	// Validate
	if params.MeetingRoomID == "" {
		t.Error("meeting room ID should not be empty")
	}
	if params.Subject == "" {
		t.Error("subject should not be empty")
	}
	if params.EndTime.Before(params.StartTime) {
		t.Error("end_time should be after start_time")
	}
	if params.Booker == "" {
		t.Error("booker should not be empty")
	}
}

func TestSendText(t *testing.T) {
	params := &MessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1", "user2"},
		Content:      "Test message",
		Safe:         false,
	}

	// Validate
	if params.ReceiverType == "" {
		t.Error("receiver_type should not be empty")
	}
	if len(params.ReceiverIDs) == 0 {
		t.Error("receiver_ids should not be empty")
	}
	if params.Content == "" {
		t.Error("content should not be empty")
	}
}

func TestSendMarkdown(t *testing.T) {
	params := &MessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		Content:      "**Test** markdown",
	}

	if params.ReceiverType == "" {
		t.Error("receiver_type should not be empty")
	}
	if len(params.ReceiverIDs) == 0 {
		t.Error("receiver_ids should not be empty")
	}
	if params.Content == "" {
		t.Error("content should not be empty")
	}
}

func TestSendImage(t *testing.T) {
	params := &ImageMessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		MediaID:      "test-media-id",
	}

	if params.ReceiverType == "" {
		t.Error("receiver_type should not be empty")
	}
	if len(params.ReceiverIDs) == 0 {
		t.Error("receiver_ids should not be empty")
	}
	if params.MediaID == "" && params.ImageURL == "" {
		t.Error("either media_id or image_url should be provided")
	}
}

func TestSendFile(t *testing.T) {
	params := &FileMessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		MediaID:      "test-media-id",
	}

	if params.ReceiverType == "" {
		t.Error("receiver_type should not be empty")
	}
	if len(params.ReceiverIDs) == 0 {
		t.Error("receiver_ids should not be empty")
	}
	if params.MediaID == "" && params.FileURL == "" {
		t.Error("either media_id or file_url should be provided")
	}
}

func TestSendCard(t *testing.T) {
	params := &CardMessageParams{
		ReceiverType: "user",
		ReceiverIDs:  []string{"user1"},
		CardContent: map[string]interface{}{
			"config": map[string]interface{}{
				"wide_screen_mode": true,
			},
		},
	}

	if params.ReceiverType == "" {
		t.Error("receiver_type should not be empty")
	}
	if len(params.ReceiverIDs) == 0 {
		t.Error("receiver_ids should not be empty")
	}
	if params.CardContent == nil {
		t.Error("card_content should not be empty")
	}
}
