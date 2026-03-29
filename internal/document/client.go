package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"wecom-gateway/internal/wecom"
)

const (
	docAPIBase = "https://qyapi.weixin.qq.com/cgi-bin/document"
)

// Client 企业微信文档 API 客户端
type Client struct {
	httpClient *http.Client
	getToken   func(ctx context.Context, corpName, appName string) (string, error)
}

// NewClient 创建文档客户端
func NewClient(wecomClient wecom.Client) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		getToken: func(ctx context.Context, corpName, appName string) (string, error) {
			return wecomClient.GetAccessToken(ctx, corpName, appName)
		},
	}
}

// wecomResponse 企业微信通用响应
type wecomResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// doRequest 执行企业微信 API 请求
func (c *Client) doRequest(ctx context.Context, corpName, appName, path string, reqBody interface{}) ([]byte, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s%s?access_token=%s", docAPIBase, path, token)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for API error
	var apiErr wecomResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrCode != 0 {
		return nil, fmt.Errorf("WeCom API error: [%d] %s", apiErr.ErrCode, apiErr.ErrMsg)
	}

	return body, nil
}

// doUpload 执行文件上传请求
func (c *Client) doUpload(ctx context.Context, corpName, appName, path string, fileData []byte, fileName string, extraFields map[string]string) ([]byte, error) {
	token, err := c.getToken(ctx, corpName, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("%s%s?access_token=%s", docAPIBase, path, token)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if fileData != nil {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		if _, err := part.Write(fileData); err != nil {
			return nil, fmt.Errorf("failed to write file data: %w", err)
		}
	}

	for key, val := range extraFields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiErr wecomResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrCode != 0 {
		return nil, fmt.Errorf("WeCom API error: [%d] %s", apiErr.ErrCode, apiErr.ErrMsg)
	}

	return body, nil
}

// --- 文档管理 ---

// CreateDocument 新建文档
func (c *Client) CreateDocument(ctx context.Context, corpName, appName string, req *CreateDocumentRequest) (*DocumentInfo, error) {
	body := map[string]interface{}{
		"owner_userid": req.OwnerUserID,
		"name":         req.Name,
		"doc_type":     req.Type,
	}
	if req.SpaceID != "" {
		body["space_id"] = req.SpaceID
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/create", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID  string `json:"document_id"`
		URL         string `json:"url"`
		CreatorID   string `json:"creator_id"`
		CreateTime  int64  `json:"create_time"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &DocumentInfo{
		DocID:      result.DocumentID,
		Name:       req.Name,
		Type:       req.Type,
		URL:        result.URL,
		CreatorID:  result.CreatorID,
		CreateTime: result.CreateTime,
	}, nil
}

// RenameDocument 重命名文档
func (c *Client) RenameDocument(ctx context.Context, corpName, appName, docID, name string) error {
	body := map[string]interface{}{
		"document_id": docID,
		"name":        name,
	}

	_, err := c.doRequest(ctx, corpName, appName, "/rename", body)
	if err != nil {
		return fmt.Errorf("failed to rename document: %w", err)
	}
	return nil
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, corpName, appName, docID string) error {
	body := map[string]interface{}{
		"document_id": docID,
	}

	_, err := c.doRequest(ctx, corpName, appName, "/delete", body)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// GetDocumentInfo 获取文档基础信息
func (c *Client) GetDocumentInfo(ctx context.Context, corpName, appName, docID string) (*DocumentInfo, error) {
	body := map[string]interface{}{
		"document_id": docID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get document info: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID  string `json:"document_id"`
		Name        string `json:"name"`
		DocType     string `json:"doc_type"`
		URL         string `json:"url"`
		CreatorID   string `json:"creator_id"`
		CreateTime  int64  `json:"create_time"`
		UpdateTime  int64  `json:"update_time"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &DocumentInfo{
		DocID:      result.DocumentID,
		Name:       result.Name,
		Type:       result.DocType,
		URL:        result.URL,
		CreatorID:  result.CreatorID,
		CreateTime: result.CreateTime,
		UpdateTime: result.UpdateTime,
	}, nil
}

// ShareDocument 分享文档
func (c *Client) ShareDocument(ctx context.Context, corpName, appName, docID string, req *ShareDocumentRequest) error {
	body := map[string]interface{}{
		"document_id": docID,
		"share_type":  req.ShareType,
	}
	if req.ExpireTime > 0 {
		body["expire_time"] = req.ExpireTime
	}

	_, err := c.doRequest(ctx, corpName, appName, "/share", body)
	if err != nil {
		return fmt.Errorf("failed to share document: %w", err)
	}
	return nil
}

// GetDocumentPermissions 获取文档权限信息
func (c *Client) GetDocumentPermissions(ctx context.Context, corpName, appName, docID string) (*DocumentPermissions, error) {
	body := map[string]interface{}{
		"document_id": docID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/permission/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get document permissions: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID   string   `json:"document_id"`
		OwnerUserID  string   `json:"owner_userid"`
		ReadOnly     bool     `json:"read_only"`
		MemberList   []string `json:"member_list"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &DocumentPermissions{
		DocID:       result.DocumentID,
		OwnerUserID: result.OwnerUserID,
		ReadOnly:    result.ReadOnly,
		MemberList:  result.MemberList,
	}, nil
}

// --- 文档内容管理 ---

// EditDocumentContent 编辑文档内容
func (c *Client) EditDocumentContent(ctx context.Context, corpName, appName, docID string, operations []ContentOperation) error {
	body := map[string]interface{}{
		"document_id": docID,
		"operations":  operations,
	}

	_, err := c.doRequest(ctx, corpName, appName, "/content/edit", body)
	if err != nil {
		return fmt.Errorf("failed to edit document content: %w", err)
	}
	return nil
}

// GetDocumentData 获取文档数据
func (c *Client) GetDocumentData(ctx context.Context, corpName, appName, docID string) (*DocumentData, error) {
	body := map[string]interface{}{
		"document_id": docID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/content/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get document data: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID string      `json:"document_id"`
		Title      string      `json:"title"`
		Body       interface{} `json:"body"`
		Hash       string      `json:"hash"`
		Version    int         `json:"version"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &DocumentData{
		DocID:  result.DocumentID,
		Title:  result.Title,
		Body:   result.Body,
		Hash:   result.Hash,
		Version: result.Version,
	}, nil
}

// UploadDocumentImage 上传文档图片
func (c *Client) UploadDocumentImage(ctx context.Context, corpName, appName, docID string, fileData []byte, fileName string) (*ImageInfo, error) {
	extraFields := map[string]string{
		"document_id": docID,
	}

	resp, err := c.doUpload(ctx, corpName, appName, "/image/upload", fileData, fileName, extraFields)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document image: %w", err)
	}

	var result struct {
		wecomResponse
		FileID string `json:"file_id"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ImageInfo{
		FileID: result.FileID,
		URL:    result.URL,
		Width:  result.Width,
		Height: result.Height,
	}, nil
}

// --- 表格内容管理 ---

// EditSheetContent 编辑表格内容
func (c *Client) EditSheetContent(ctx context.Context, corpName, appName string, docID string, req *EditSheetRequest) error {
	body := map[string]interface{}{
		"document_id": docID,
		"row":         req.Row,
		"col":         req.Col,
		"value":       req.Value,
	}

	_, err := c.doRequest(ctx, corpName, appName, "/sheet/content/edit", body)
	if err != nil {
		return fmt.Errorf("failed to edit sheet content: %w", err)
	}
	return nil
}

// GetSheetRowCol 获取表格行列信息
func (c *Client) GetSheetRowCol(ctx context.Context, corpName, appName, docID string) (*SheetRowColInfo, error) {
	body := map[string]interface{}{
		"document_id": docID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/sheet/rowcol/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet row/col info: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID string `json:"document_id"`
		RowCount   int    `json:"row_count"`
		ColCount   int    `json:"col_count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SheetRowColInfo{
		DocID:    result.DocumentID,
		RowCount: result.RowCount,
		ColCount: result.ColCount,
	}, nil
}

// GetSheetData 获取表格数据
func (c *Client) GetSheetData(ctx context.Context, corpName, appName, docID string, dataRange string) (*SheetData, error) {
	body := map[string]interface{}{
		"document_id": docID,
	}
	if dataRange != "" {
		body["range"] = dataRange
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/sheet/content/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	var result struct {
		wecomResponse
		DocumentID string      `json:"document_id"`
		RowCount   int         `json:"row_count"`
		ColCount   int         `json:"col_count"`
		Values     [][]string  `json:"values"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SheetData{
		DocID:    result.DocumentID,
		RowCount: result.RowCount,
		ColCount: result.ColCount,
		Values:   result.Values,
	}, nil
}

// --- 空间管理 ---

// CreateSpace 新建空间
func (c *Client) CreateSpace(ctx context.Context, corpName, appName string, req *CreateSpaceRequest) (*SpaceInfo, error) {
	body := map[string]interface{}{
		"name":         req.Name,
		"admin_userid": req.AdminUserID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/space/create", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create space: %w", err)
	}

	var result struct {
		wecomResponse
		SpaceID   string   `json:"space_id"`
		Name      string   `json:"name"`
		AdminList []string `json:"admin_list"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SpaceInfo{
		SpaceID:   result.SpaceID,
		Name:      result.Name,
		AdminList: result.AdminList,
	}, nil
}

// GetSpaceInfo 获取空间信息
func (c *Client) GetSpaceInfo(ctx context.Context, corpName, appName, spaceID string) (*SpaceInfo, error) {
	body := map[string]interface{}{
		"space_id": spaceID,
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/space/get", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get space info: %w", err)
	}

	var result struct {
		wecomResponse
		SpaceID     string   `json:"space_id"`
		Name        string   `json:"name"`
		AdminList   []string `json:"admin_list"`
		MemberCount int      `json:"member_count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SpaceInfo{
		SpaceID:     result.SpaceID,
		Name:        result.Name,
		AdminList:   result.AdminList,
		MemberCount: result.MemberCount,
	}, nil
}

// GetSpaceFileList 获取空间文件列表
func (c *Client) GetSpaceFileList(ctx context.Context, corpName, appName, spaceID, cursor string, limit int) (*FileListResult, error) {
	body := map[string]interface{}{
		"space_id": spaceID,
	}
	if cursor != "" {
		body["cursor"] = cursor
	}
	if limit > 0 {
		body["limit"] = limit
	}

	resp, err := c.doRequest(ctx, corpName, appName, "/space/file/list", body)
	if err != nil {
		return nil, fmt.Errorf("failed to get space file list: %w", err)
	}

	var result struct {
		wecomResponse
		HasMore  bool          `json:"has_more"`
		Cursor   string        `json:"cursor"`
		FileList []struct {
			DocumentID string `json:"document_id"`
			Name       string `json:"name"`
			DocType    string `json:"doc_type"`
			URL        string `json:"url"`
			CreatorID  string `json:"creator_id"`
			CreateTime int64  `json:"create_time"`
			UpdateTime int64  `json:"update_time"`
		} `json:"file_list"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	fileList := make([]DocumentInfo, 0, len(result.FileList))
	for _, f := range result.FileList {
		fileList = append(fileList, DocumentInfo{
			DocID:      f.DocumentID,
			Name:       f.Name,
			Type:       f.DocType,
			URL:        f.URL,
			CreatorID:  f.CreatorID,
			CreateTime: f.CreateTime,
			UpdateTime: f.UpdateTime,
		})
	}

	return &FileListResult{
		HasMore:  result.HasMore,
		Cursor:   result.Cursor,
		FileList: fileList,
	}, nil
}
