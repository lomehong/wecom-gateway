package sheet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHandler()

	sheetGroup := r.Group("/v1/sheets")
	{
		sheetGroup.POST("", handler.CreateSheet)
		sheetGroup.GET("/:docid/sheets", handler.ListSheetTabs)
		sheetGroup.GET("/:docid/sheets/:sheetid/fields", handler.GetSheetFields)
		sheetGroup.POST("/:docid/sheets/:sheetid/fields", handler.AddSheetFields)
		sheetGroup.GET("/:docid/sheets/:sheetid/records", handler.GetSheetRecords)
		sheetGroup.POST("/:docid/sheets/:sheetid/records", handler.AddSheetRecords)
		sheetGroup.PUT("/:docid/sheets/:sheetid/records", handler.UpdateSheetRecords)
		sheetGroup.DELETE("/:docid/sheets/:sheetid/records", handler.DeleteSheetRecords)
	}

	return r
}

func TestCreateSheet(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest("POST", "/v1/sheets", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestListSheetTabs(t *testing.T) {
	r := setupRouter()

	// Missing docid
	req := httptest.NewRequest("GET", "/v1/sheets//sheets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing docid, got %d", w.Code)
	}

	// Valid docid
	req = httptest.NewRequest("GET", "/v1/sheets/doc123/sheets", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["doc_id"] != "doc123" {
		t.Errorf("expected doc_id doc123, got %v", data["doc_id"])
	}
}

func TestGetSheetFields(t *testing.T) {
	r := setupRouter()

	// Missing sheetid
	req := httptest.NewRequest("GET", "/v1/sheets/doc123/sheets//fields", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Valid
	req = httptest.NewRequest("GET", "/v1/sheets/doc123/sheets/sheet1/fields", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetSheetRecords(t *testing.T) {
	r := setupRouter()

	req := httptest.NewRequest("GET", "/v1/sheets/doc123/sheets/sheet1/records?limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["limit"] != float64(50) {
		t.Errorf("expected limit 50, got %v", data["limit"])
	}
}

func TestAddSheetRecords(t *testing.T) {
	r := setupRouter()

	// No body
	req := httptest.NewRequest("POST", "/v1/sheets/doc123/sheets/sheet1/records", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for no body, got %d", w.Code)
	}

	// With body
	body := `{"records": [{"field1": "value1"}]}`
	req = httptest.NewRequest("POST", "/v1/sheets/doc123/sheets/sheet1/records", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
