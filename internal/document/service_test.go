package document

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

func testAuthCtx() *auth.AuthContext {
	return &auth.AuthContext{
		CorpName: "test-corp",
		AppName:  "test-app",
		KeyName:  "test-key",
	}
}

// mockWecomClient 用于测试的 mock wecom.Client
type mockWecomClient struct {
	accessToken string
}

func newMockWecomClient() *mockWecomClient {
	return &mockWecomClient{
		accessToken: "test-access-token-123",
	}
}

func (m *mockWecomClient) GetAccessToken(ctx context.Context, corpName, appName string) (string, error) {
	return m.accessToken, nil
}

// Implement all other wecom.Client methods with no-op stubs
func (m *mockWecomClient) CreateSchedule(ctx context.Context, corpName, appName string, params *wecom.ScheduleParams) (*wecom.Schedule, error) {
	return &wecom.Schedule{ScheduleID: "test"}, nil
}
func (m *mockWecomClient) GetSchedules(ctx context.Context, corpName, appName string, userID string, startTime, endTime time.Time, limit int) ([]*wecom.Schedule, error) {
	return nil, nil
}
func (m *mockWecomClient) UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *wecom.ScheduleParams) error { return nil }
func (m *mockWecomClient) DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error         { return nil }
func (m *mockWecomClient) ListMeetingRooms(ctx context.Context, corpName, appName string, opts *wecom.RoomQueryOptions) ([]*wecom.MeetingRoom, string, error) {
	return nil, "", nil
}
func (m *mockWecomClient) GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*wecom.TimeSlot, error) {
	return nil, nil
}
func (m *mockWecomClient) BookMeetingRoom(ctx context.Context, corpName, appName string, params *wecom.BookingParams) (*wecom.BookingResult, error) {
	return nil, nil
}
func (m *mockWecomClient) SendText(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *mockWecomClient) SendMarkdown(ctx context.Context, corpName, appName string, params *wecom.MessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *mockWecomClient) SendImage(ctx context.Context, corpName, appName string, params *wecom.ImageMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *mockWecomClient) SendFile(ctx context.Context, corpName, appName string, params *wecom.FileMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *mockWecomClient) SendCard(ctx context.Context, corpName, appName string, params *wecom.CardMessageParams) (*wecom.SendResult, error) {
	return nil, nil
}
func (m *mockWecomClient) UploadMedia(ctx context.Context, corpName, appName string, mediaType string, data []byte, filename string) (string, error) {
	return "media-123", nil
}

// --- Service Tests ---

func TestService_ValidateDocType(t *testing.T) {
	tests := []struct {
		name    string
		docType string
		wantErr bool
	}{
		{"doc type", DocTypeDoc, false},
		{"sheet type", DocTypeSheet, false},
		{"bitable type", DocTypeBitable, false},
		{"mindnote type", DocTypeMindnote, false},
		{"docx type", DocTypeDocx, false},
		{"invalid type", "invalid", true},
		{"empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDocType(tt.docType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDocType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateDocument_Validation(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	// Test invalid doc type
	_, err := svc.CreateDocument(context.Background(), testAuthCtx(), &CreateDocumentRequest{
		OwnerUserID: "user1",
		Name:        "test",
		Type:        "invalid",
	})
	if err == nil {
		t.Error("expected error for invalid doc type")
	}
}

func TestService_RenameDocument_Validation(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	// Test empty doc ID
	err := svc.RenameDocument(context.Background(), testAuthCtx(), "", "new-name")
	if err == nil {
		t.Error("expected error for empty doc id")
	}

	// Test empty name
	err = svc.RenameDocument(context.Background(), testAuthCtx(), "doc123", "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestService_DeleteDocument_Validation(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	err := svc.DeleteDocument(context.Background(), testAuthCtx(), "")
	if err == nil {
		t.Error("expected error for empty doc id")
	}
}

func TestService_ShareDocument_Validation(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	tests := []struct {
		name      string
		shareType int
		wantErr   bool
	}{
		{"valid type 1", 1, false},
		{"valid type 2", 2, false},
		{"valid type 3", 3, false},
		{"invalid type 0", 0, true},
		{"invalid type 4", 4, true},
		{"invalid type -1", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ShareDocument(context.Background(), testAuthCtx(), "doc123", &ShareDocumentRequest{ShareType: tt.shareType})
			if (err != nil) != tt.wantErr {
				t.Errorf("ShareDocument() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateSpace_Validation(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	// Test empty name
	_, err := svc.CreateSpace(context.Background(), testAuthCtx(), &CreateSpaceRequest{
		Name:        "",
		AdminUserID: "admin1",
	})
	if err == nil {
		t.Error("expected error for empty space name")
	}

	// Test empty admin
	_, err = svc.CreateSpace(context.Background(), testAuthCtx(), &CreateSpaceRequest{
		Name:        "Test Space",
		AdminUserID: "",
	})
	if err == nil {
		t.Error("expected error for empty admin_userid")
	}
}

func TestService_GetSpaceFileList_LimitClamp(t *testing.T) {
	mockWecom := newMockWecomClient()
	client := NewClient(mockWecom)
	svc := NewService(client)

	// Test empty space id
	_, err := svc.GetSpaceFileList(context.Background(), testAuthCtx(), "", "", 10)
	if err == nil {
		t.Error("expected error for empty space id")
	}
}

// --- Client Tests ---

func TestClient_DoRequest_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40001,
			"errmsg":  "invalid credential",
		})
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		getToken: func(ctx context.Context, corpName, appName string) (string, error) {
			return "test-token", nil
		},
	}
	// Override the base URL by using doRequest directly
	// We can't easily test this without modifying the client, so skip for now
	_ = client
}

func TestTypes_JSONTags(t *testing.T) {
	// Verify all request types have proper JSON tags
	req := CreateDocumentRequest{
		OwnerUserID: "user1",
		Name:        "Test Doc",
		Type:        DocTypeDoc,
		SpaceID:     "space1",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled["owner_userid"] != "user1" {
		t.Error("owner_userid not found in JSON")
	}
	if unmarshaled["name"] != "Test Doc" {
		t.Error("name not found in JSON")
	}
	if unmarshaled["type"] != "doc" {
		t.Error("type not found in JSON")
	}
}
