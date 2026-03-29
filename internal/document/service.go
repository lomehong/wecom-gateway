package document

import (
	"context"
	"fmt"

	"wecom-gateway/internal/auth"
)

// Service 文档管理业务逻辑
type Service struct {
	client *Client
}

// NewService 创建文档服务
func NewService(client *Client) *Service {
	return &Service{client: client}
}

// CreateDocument 新建文档
func (s *Service) CreateDocument(ctx context.Context, authCtx *auth.AuthContext, req *CreateDocumentRequest) (*DocumentInfo, error) {
	if err := validateDocType(req.Type); err != nil {
		return nil, err
	}

	info, err := s.client.CreateDocument(ctx, authCtx.CorpName, authCtx.AppName, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return info, nil
}

// RenameDocument 重命名文档
func (s *Service) RenameDocument(ctx context.Context, authCtx *auth.AuthContext, docID string, name string) error {
	if docID == "" {
		return fmt.Errorf("document id is required")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}

	return s.client.RenameDocument(ctx, authCtx.CorpName, authCtx.AppName, docID, name)
}

// DeleteDocument 删除文档
func (s *Service) DeleteDocument(ctx context.Context, authCtx *auth.AuthContext, docID string) error {
	if docID == "" {
		return fmt.Errorf("document id is required")
	}

	return s.client.DeleteDocument(ctx, authCtx.CorpName, authCtx.AppName, docID)
}

// GetDocumentInfo 获取文档信息
func (s *Service) GetDocumentInfo(ctx context.Context, authCtx *auth.AuthContext, docID string) (*DocumentInfo, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}

	return s.client.GetDocumentInfo(ctx, authCtx.CorpName, authCtx.AppName, docID)
}

// ShareDocument 分享文档
func (s *Service) ShareDocument(ctx context.Context, authCtx *auth.AuthContext, docID string, req *ShareDocumentRequest) error {
	if docID == "" {
		return fmt.Errorf("document id is required")
	}
	if req.ShareType < 1 || req.ShareType > 3 {
		return fmt.Errorf("invalid share_type, must be 1, 2, or 3")
	}

	return s.client.ShareDocument(ctx, authCtx.CorpName, authCtx.AppName, docID, req)
}

// GetDocumentPermissions 获取文档权限
func (s *Service) GetDocumentPermissions(ctx context.Context, authCtx *auth.AuthContext, docID string) (*DocumentPermissions, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}

	return s.client.GetDocumentPermissions(ctx, authCtx.CorpName, authCtx.AppName, docID)
}

// EditDocumentContent 编辑文档内容
func (s *Service) EditDocumentContent(ctx context.Context, authCtx *auth.AuthContext, docID string, operations []ContentOperation) error {
	if docID == "" {
		return fmt.Errorf("document id is required")
	}
	if len(operations) == 0 {
		return fmt.Errorf("operations is required")
	}

	return s.client.EditDocumentContent(ctx, authCtx.CorpName, authCtx.AppName, docID, operations)
}

// GetDocumentData 获取文档数据
func (s *Service) GetDocumentData(ctx context.Context, authCtx *auth.AuthContext, docID string) (*DocumentData, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}

	return s.client.GetDocumentData(ctx, authCtx.CorpName, authCtx.AppName, docID)
}

// UploadDocumentImage 上传文档图片
func (s *Service) UploadDocumentImage(ctx context.Context, authCtx *auth.AuthContext, docID string, fileData []byte, fileName string) (*ImageInfo, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}
	if len(fileData) == 0 {
		return nil, fmt.Errorf("file data is required")
	}
	if fileName == "" {
		fileName = "image.png"
	}

	return s.client.UploadDocumentImage(ctx, authCtx.CorpName, authCtx.AppName, docID, fileData, fileName)
}

// EditSheetContent 编辑表格内容
func (s *Service) EditSheetContent(ctx context.Context, authCtx *auth.AuthContext, docID string, req *EditSheetRequest) error {
	if docID == "" {
		return fmt.Errorf("document id is required")
	}

	return s.client.EditSheetContent(ctx, authCtx.CorpName, authCtx.AppName, docID, req)
}

// GetSheetRowCol 获取表格行列信息
func (s *Service) GetSheetRowCol(ctx context.Context, authCtx *auth.AuthContext, docID string) (*SheetRowColInfo, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}

	return s.client.GetSheetRowCol(ctx, authCtx.CorpName, authCtx.AppName, docID)
}

// GetSheetData 获取表格数据
func (s *Service) GetSheetData(ctx context.Context, authCtx *auth.AuthContext, docID string, dataRange string) (*SheetData, error) {
	if docID == "" {
		return nil, fmt.Errorf("document id is required")
	}

	return s.client.GetSheetData(ctx, authCtx.CorpName, authCtx.AppName, docID, dataRange)
}

// CreateSpace 新建空间
func (s *Service) CreateSpace(ctx context.Context, authCtx *auth.AuthContext, req *CreateSpaceRequest) (*SpaceInfo, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("space name is required")
	}
	if req.AdminUserID == "" {
		return nil, fmt.Errorf("admin_userid is required")
	}

	return s.client.CreateSpace(ctx, authCtx.CorpName, authCtx.AppName, req)
}

// GetSpaceInfo 获取空间信息
func (s *Service) GetSpaceInfo(ctx context.Context, authCtx *auth.AuthContext, spaceID string) (*SpaceInfo, error) {
	if spaceID == "" {
		return nil, fmt.Errorf("space id is required")
	}

	return s.client.GetSpaceInfo(ctx, authCtx.CorpName, authCtx.AppName, spaceID)
}

// GetSpaceFileList 获取空间文件列表
func (s *Service) GetSpaceFileList(ctx context.Context, authCtx *auth.AuthContext, spaceID string, cursor string, limit int) (*FileListResult, error) {
	if spaceID == "" {
		return nil, fmt.Errorf("space id is required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.client.GetSpaceFileList(ctx, authCtx.CorpName, authCtx.AppName, spaceID, cursor, limit)
}

// validateDocType 校验文档类型
func validateDocType(docType string) error {
	validTypes := map[string]bool{
		DocTypeDoc: true, DocTypeSheet: true,
		DocTypeBitable: true, DocTypeMindnote: true,
		DocTypeDocx: true,
	}
	if !validTypes[docType] {
		return fmt.Errorf("invalid doc_type, must be one of: doc, sheet, bitable, mindnote, docx")
	}
	return nil
}
