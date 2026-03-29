package document

import (
	"github.com/gin-gonic/gin"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/httputil"
)

// Handler 文档管理 HTTP handler
type Handler struct {
	service *Service
}

// NewHandler 创建文档 handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateDocument handles POST /v1/docs
// @Summary 新建文档
// @Description 创建一个新的企业微信文档
// @Tags documents
// @Accept json
// @Produce json
// @Param request body CreateDocumentRequest true "文档参数"
// @Success 200 {object} DocumentInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs [post]
func (h *Handler) CreateDocument(c *gin.Context) {
	var req CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	info, err := h.service.CreateDocument(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// GetDocument handles GET /v1/docs/:docid
// @Summary 获取文档信息
// @Description 获取文档的基础信息
// @Tags documents
// @Produce json
// @Param docid path string true "文档 ID"
// @Success 200 {object} DocumentInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid} [get]
func (h *Handler) GetDocument(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	info, err := h.service.GetDocumentInfo(c.Request.Context(), authCtx, docID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// RenameDocument handles PUT /v1/docs/:docid/rename
// @Summary 重命名文档
// @Description 修改已有文档的名称
// @Tags documents
// @Accept json
// @Produce json
// @Param docid path string true "文档 ID"
// @Param request body RenameDocumentRequest true "重命名参数"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/rename [put]
func (h *Handler) RenameDocument(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	var req RenameDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	if err := h.service.RenameDocument(c.Request.Context(), authCtx, docID, req.Name); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{"doc_id": docID})
}

// DeleteDocument handles DELETE /v1/docs/:docid
// @Summary 删除文档
// @Description 删除指定文档
// @Tags documents
// @Produce json
// @Param docid path string true "文档 ID"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid} [delete]
func (h *Handler) DeleteDocument(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	if err := h.service.DeleteDocument(c.Request.Context(), authCtx, docID); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{"doc_id": docID})
}

// ShareDocument handles POST /v1/docs/:docid/share
// @Summary 分享文档
// @Description 设置文档的分享方式
// @Tags documents
// @Accept json
// @Produce json
// @Param docid path string true "文档 ID"
// @Param request body ShareDocumentRequest true "分享参数"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/share [post]
func (h *Handler) ShareDocument(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	var req ShareDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	if err := h.service.ShareDocument(c.Request.Context(), authCtx, docID, &req); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{"doc_id": docID})
}

// GetPermissions handles GET /v1/docs/:docid/permissions
// @Summary 获取文档权限
// @Description 查询文档的权限设置
// @Tags documents
// @Produce json
// @Param docid path string true "文档 ID"
// @Success 200 {object} DocumentPermissions
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/permissions [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	perms, err := h.service.GetDocumentPermissions(c.Request.Context(), authCtx, docID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, perms)
}

// EditContent handles PUT /v1/docs/:docid/content
// @Summary 编辑文档内容
// @Description 编辑文档的正文内容
// @Tags documents
// @Accept json
// @Produce json
// @Param docid path string true "文档 ID"
// @Param request body object true "操作列表"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/content [put]
func (h *Handler) EditContent(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	var req struct {
		Operations []ContentOperation `json:"operations" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	if err := h.service.EditDocumentContent(c.Request.Context(), authCtx, docID, req.Operations); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{"doc_id": docID})
}

// GetDocumentData handles GET /v1/docs/:docid/data
// @Summary 获取文档数据
// @Description 获取文档的完整内容数据
// @Tags documents
// @Produce json
// @Param docid path string true "文档 ID"
// @Success 200 {object} DocumentData
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/data [get]
func (h *Handler) GetDocumentData(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	data, err := h.service.GetDocumentData(c.Request.Context(), authCtx, docID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, data)
}

// UploadImage handles POST /v1/docs/:docid/images
// @Summary 上传文档图片
// @Description 上传图片到文档中
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param docid path string true "文档 ID"
// @Param file formData file true "图片文件"
// @Success 200 {object} ImageInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/{docid}/images [post]
func (h *Handler) UploadImage(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httputil.BadRequest(c, "file is required: "+err.Error())
		return
	}
	defer file.Close()

	authCtx, _ := auth.GetAuthContext(c)

	fileData := make([]byte, header.Size)
	if _, err := file.Read(fileData); err != nil {
		httputil.InternalError(c, "failed to read file: "+err.Error())
		return
	}

	info, err := h.service.UploadDocumentImage(c.Request.Context(), authCtx, docID, fileData, header.Filename)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// EditSheetContent handles POST /v1/docs/sheets/:docid/content
// @Summary 编辑表格内容
// @Description 编辑表格单元格内容
// @Tags documents
// @Accept json
// @Produce json
// @Param docid path string true "表格文档 ID"
// @Param request body EditSheetRequest true "编辑参数"
// @Success 200 {object} httputil.Response
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/sheets/{docid}/content [post]
func (h *Handler) EditSheetContent(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	var req EditSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	if err := h.service.EditSheetContent(c.Request.Context(), authCtx, docID, &req); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, gin.H{"doc_id": docID})
}

// GetSheetRowCol handles GET /v1/docs/sheets/:docid/rows
// @Summary 获取表格行列信息
// @Description 获取表格的行列元信息
// @Tags documents
// @Produce json
// @Param docid path string true "表格文档 ID"
// @Success 200 {object} SheetRowColInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/sheets/{docid}/rows [get]
func (h *Handler) GetSheetRowCol(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	info, err := h.service.GetSheetRowCol(c.Request.Context(), authCtx, docID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// GetSheetData handles GET /v1/docs/sheets/:docid/data
// @Summary 获取表格数据
// @Description 获取表格的完整数据
// @Tags documents
// @Produce json
// @Param docid path string true "表格文档 ID"
// @Param range query string false "数据范围"
// @Success 200 {object} SheetData
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/sheets/{docid}/data [get]
func (h *Handler) GetSheetData(c *gin.Context) {
	docID := c.Param("docid")
	if docID == "" {
		httputil.BadRequest(c, "document id is required")
		return
	}

	dataRange := c.Query("range")

	authCtx, _ := auth.GetAuthContext(c)

	data, err := h.service.GetSheetData(c.Request.Context(), authCtx, docID, dataRange)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, data)
}

// CreateSpace handles POST /v1/docs/spaces
// @Summary 新建空间
// @Description 创建一个新的文档空间
// @Tags documents
// @Accept json
// @Produce json
// @Param request body CreateSpaceRequest true "空间参数"
// @Success 200 {object} SpaceInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/spaces [post]
func (h *Handler) CreateSpace(c *gin.Context) {
	var req CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid request parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	info, err := h.service.CreateSpace(c.Request.Context(), authCtx, &req)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// GetSpaceInfo handles GET /v1/docs/spaces/:spaceid
// @Summary 获取空间信息
// @Description 获取空间的详细信息
// @Tags documents
// @Produce json
// @Param spaceid path string true "空间 ID"
// @Success 200 {object} SpaceInfo
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/spaces/{spaceid} [get]
func (h *Handler) GetSpaceInfo(c *gin.Context) {
	spaceID := c.Param("spaceid")
	if spaceID == "" {
		httputil.BadRequest(c, "space id is required")
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	info, err := h.service.GetSpaceInfo(c.Request.Context(), authCtx, spaceID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, info)
}

// GetSpaceFileList handles GET /v1/docs/spaces/:spaceid/files
// @Summary 获取空间文件列表
// @Description 获取空间内的文件列表
// @Tags documents
// @Produce json
// @Param spaceid path string true "空间 ID"
// @Param cursor query string false "分页游标"
// @Param limit query int false "每页数量"
// @Success 200 {object} FileListResult
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Failure 500 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/docs/spaces/{spaceid}/files [get]
func (h *Handler) GetSpaceFileList(c *gin.Context) {
	spaceID := c.Param("spaceid")
	if spaceID == "" {
		httputil.BadRequest(c, "space id is required")
		return
	}

	var req GetSpaceFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	authCtx, _ := auth.GetAuthContext(c)

	result, err := h.service.GetSpaceFileList(c.Request.Context(), authCtx, spaceID, req.Cursor, req.Limit)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.Success(c, result)
}
