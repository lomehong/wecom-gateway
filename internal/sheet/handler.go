package sheet

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler handles HTTP requests for smart sheet operations
type Handler struct {
	// Smart sheets use the same document client under the hood
	// but expose a cleaner REST API at /v1/sheets
}

// NewHandler creates a new sheet handler
func NewHandler() *Handler {
	return &Handler{}
}

// CreateSheet handles POST /v1/sheets
// @Summary 创建智能表格
// @Tags sheets
func (h *Handler) CreateSheet(c *gin.Context) {
	authCtx, _ := auth.GetAuthContext(c)
	corpName := ""
	appName := ""
	if authCtx != nil {
		corpName = authCtx.CorpName
		appName = authCtx.AppName
	}

	httputil.Success(c, gin.H{
		"message":   "use POST /v1/docs with doc_type=10 to create smart sheets via document API",
		"corp_name": corpName,
		"app_name":  appName,
	})
}

// ListSheetTabs handles GET /v1/sheets/:docid/sheets
// @Summary 查询智能表格子表列表
// @Tags sheets
func (h *Handler) ListSheetTabs(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "docid is required")
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": "",
		"message":  "smart sheet sub-table listing requires bot MCP integration",
	})
}

// GetSheetFields handles GET /v1/sheets/:docid/sheets/:sheetid/fields
// @Summary 查询子表字段信息
// @Tags sheets
func (h *Handler) GetSheetFields(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"fields":   []interface{}{},
	})
}

// AddSheetFields handles POST /v1/sheets/:docid/sheets/:sheetid/fields
// @Summary 添加字段
// @Tags sheets
func (h *Handler) AddSheetFields(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"message":  "field addition requires bot MCP integration",
	})
}

// GetSheetRecords handles GET /v1/sheets/:docid/sheets/:sheetid/records
// @Summary 查询记录
// @Tags sheets
func (h *Handler) GetSheetRecords(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"records":   []interface{}{},
		"total":     0,
		"limit":     limit,
	})
}

// AddSheetRecords handles POST /v1/sheets/:docid/sheets/:sheetid/records
// @Summary 添加记录
// @Tags sheets
func (h *Handler) AddSheetRecords(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"message":  "record addition requires bot MCP integration",
	})
}

// UpdateSheetRecords handles PUT /v1/sheets/:docid/sheets/:sheetid/records
// @Summary 更新记录
// @Tags sheets
func (h *Handler) UpdateSheetRecords(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"message":  "record update requires bot MCP integration",
	})
}

// DeleteSheetRecords handles DELETE /v1/sheets/:docid/sheets/:sheetid/records
// @Summary 删除记录
// @Tags sheets
func (h *Handler) DeleteSheetRecords(c *gin.Context) {
	docID := c.Param("docid")
	sheetID := c.Param("sheetid")
	if docID == "" || sheetID == "" {
		httputil.BadRequest(c, "docid and sheetid are required")
		return
	}

	httputil.Success(c, gin.H{
		"doc_id":   docID,
		"sheet_id": sheetID,
		"message":  "record deletion requires bot MCP integration",
	})
}
