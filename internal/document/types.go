package document

// 文档类型常量
const (
	DocTypeDoc      = "doc"      // 传统文档
	DocTypeSheet    = "sheet"    // 表格
	DocTypeBitable  = "bitable"  // 智能表格
	DocTypeMindnote = "mindnote" // 脑图
	DocTypeDocx     = "docx"     // 新版文档
)

// CreateDocumentRequest 新建文档请求
type CreateDocumentRequest struct {
	OwnerUserID string `json:"owner_userid" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"` // doc/sheet/bitable/mindnote/docx
	SpaceID     string `json:"space_id"`                // 所属空间 ID
}

// RenameDocumentRequest 重命名文档请求
type RenameDocumentRequest struct {
	Name string `json:"name" binding:"required"`
}

// ShareDocumentRequest 分享文档请求
type ShareDocumentRequest struct {
	ShareType  int `json:"share_type" binding:"required"` // 1=仅企业内, 2=指定人, 3=链接分享
	ExpireTime int `json:"expire_time"`                  // 链接过期时间戳（秒）
}

// EditSheetRequest 编辑表格内容请求
type EditSheetRequest struct {
	Row   int    `json:"row" binding:"gte=0"`     // 行号（0-based）
	Col   int    `json:"col" binding:"gte=0"`     // 列号（0-based）
	Value string `json:"value" binding:"required"` // 单元格值
}

// CreateSpaceRequest 新建空间请求
type CreateSpaceRequest struct {
	Name        string `json:"name" binding:"required"`
	AdminUserID string `json:"admin_userid" binding:"required"`
}

// GetSpaceFilesRequest 获取空间文件列表请求（query params）
type GetSpaceFilesRequest struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit"`
}

// DocumentInfo 文档信息
type DocumentInfo struct {
	DocID      string `json:"doc_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	CreatorID  string `json:"creator_id"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// DocumentPermissions 文档权限信息
type DocumentPermissions struct {
	DocID       string   `json:"doc_id"`
	OwnerUserID string   `json:"owner_userid"`
	ReadOnly    bool     `json:"read_only"`
	MemberList  []string `json:"member_list"`
}

// DocumentData 文档数据
type DocumentData struct {
	DocID  string      `json:"doc_id"`
	Title  string      `json:"title"`
	Body   interface{} `json:"body"` // 文档内容（结构化数据）
	Hash   string      `json:"hash"`
	Version int        `json:"version"`
}

// SheetRowColInfo 表格行列信息
type SheetRowColInfo struct {
	DocID    string `json:"doc_id"`
	RowCount int    `json:"row_count"`
	ColCount int    `json:"col_count"`
}

// SheetData 表格数据
type SheetData struct {
	DocID    string        `json:"doc_id"`
	RowCount int           `json:"row_count"`
	ColCount int           `json:"col_count"`
	Values   [][]string    `json:"values"`
}

// ImageInfo 文档图片信息
type ImageInfo struct {
	FileID   string `json:"file_id"`
	URL      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// SpaceInfo 空间信息
type SpaceInfo struct {
	SpaceID     string   `json:"space_id"`
	Name        string   `json:"name"`
	AdminList   []string `json:"admin_list"`
	MemberCount int      `json:"member_count"`
}

// FileListResult 文件列表结果
type FileListResult struct {
	HasMore  bool          `json:"has_more"`
	Cursor   string        `json:"cursor"`
	FileList []DocumentInfo `json:"file_list"`
}

// ContentOperation 文档内容操作
type ContentOperation struct {
	OpType   int                    `json:"op_type"`   // 1=插入, 2=删除, 3=替换
	Position map[string]interface{} `json:"position"`  // 操作位置
	Content  interface{}            `json:"content"`   // 操作内容
}

// UploadImageRequest 上传文档图片请求
type UploadImageRequest struct {
	ImageURL    string `json:"image_url"`     // 图片 URL（可选）
	ImageBase64 string `json:"image_base64"`  // 图片 Base64（可选）
	FileName    string `json:"file_name"`     // 文件名（可选）
}
