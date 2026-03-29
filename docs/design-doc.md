# 设计文档：WeCom Gateway — 企业微信文档管理

| 项目 | 内容 |
|------|------|
| 项目名称 | wecom-gateway（文档管理模块） |
| 文档版本 | v1.0 |
| 创建日期 | 2026-03-29 |
| 作者 | M10S |
| 状态 | 草案 |
| 需求来源 | [需求文档](./requirements-doc.md) |

---

## 1. 架构设计

### 1.1 模块结构

```
internal/
├── document/                  # 新增：文档管理模块
│   ├── handler.go            # HTTP handler 层
│   ├── service.go            # 业务逻辑层
│   ├── client.go             # 企业微信文档 API 客户端
│   ├── types.go              # 请求/响应类型定义
│   ├── handler_test.go       # handler 单元测试
│   ├── service_test.go       # service 单元测试
│   └── mock_client.go        # mock 客户端（供测试使用）
```

### 1.2 调用链路

```
HTTP Request
    → API Key 认证中间件
    → 权限校验 (document:read / document:write)
    → DocumentHandler
    → DocumentService
    → DocumentClient (调用企业微信文档 API)
    → 企业微信服务端
```

---

## 2. 企业微信 API 对接

### 2.1 API 基础信息

- **Base URL**: `https://qyapi.weixin.qq.com`
- **认证方式**: access_token（通过 CorpID + Secret 获取）
- **Content-Type**: `application/json`

### 2.2 API 端点映射

| 功能 | HTTP Method | 企业微信 API 路径 |
|------|------------|------------------|
| 新建文档 | POST | `/cgi-bin/document/create` |
| 重命名文档 | POST | `/cgi-bin/document/rename` |
| 删除文档 | POST | `/cgi-bin/document/delete` |
| 获取文档信息 | POST | `/cgi-bin/document/get` |
| 分享文档 | POST | `/cgi-bin/document/share` |
| 获取文档权限 | POST | `/cgi-bin/document/permission/get` |
| 编辑文档内容 | POST | `/cgi-bin/document/content/edit` |
| 获取文档数据 | POST | `/cgi-bin/document/content/get` |
| 上传文档图片 | POST | `/cgi-bin/document/image/upload` |
| 编辑表格内容 | POST | `/cgi-bin/document/sheet/content/edit` |
| 获取表格行列 | POST | `/cgi-bin/document/sheet/rowcol/get` |
| 获取表格数据 | POST | `/cgi-bin/document/sheet/content/get` |
| 新建空间 | POST | `/cgi-bin/document/space/create` |
| 获取空间信息 | POST | `/cgi-bin/document/space/get` |
| 获取文件列表 | POST | `/cgi-bin/document/space/file/list` |

### 2.3 Client 接口定义

```go
// DocumentClient 企业微信文档 API 客户端接口
type DocumentClient interface {
    // 文档管理
    CreateDocument(ctx context.Context, corpName, appName string, req *CreateDocumentRequest) (*DocumentInfo, error)
    RenameDocument(ctx context.Context, corpName, appName string, docID string, name string) error
    DeleteDocument(ctx context.Context, corpName, appName string, docID string) error
    GetDocumentInfo(ctx context.Context, corpName, appName string, docID string) (*DocumentInfo, error)
    ShareDocument(ctx context.Context, corpName, appName string, docID string, req *ShareRequest) error
    GetDocumentPermissions(ctx context.Context, corpName, appName string, docID string) (*DocumentPermissions, error)

    // 文档内容
    EditDocumentContent(ctx context.Context, corpName, appName string, docID string, operations []ContentOperation) error
    GetDocumentData(ctx context.Context, corpName, appName string, docID string) (*DocumentData, error)
    UploadDocumentImage(ctx context.Context, corpName, appName string, docID string, imageData []byte, fileName string) (*ImageInfo, error)

    // 表格内容
    EditSheetContent(ctx context.Context, corpName, appName string, docID string, req *EditSheetRequest) error
    GetSheetRowCol(ctx context.Context, corpName, appName string, docID string) (*SheetRowColInfo, error)
    GetSheetData(ctx context.Context, corpName, appName string, docID string, dataRange string) (*SheetData, error)

    // 空间管理
    CreateSpace(ctx context.Context, corpName, appName string, req *CreateSpaceRequest) (*SpaceInfo, error)
    GetSpaceInfo(ctx context.Context, corpName, appName string, spaceID string) (*SpaceInfo, error)
    GetSpaceFileList(ctx context.Context, corpName, appName string, spaceID string, cursor string, limit int) (*FileListResult, error)
}
```

---

## 3. 网关 API 设计

### 3.1 路由注册

```go
// 在 cmd/server/main.go 中注册
docGroup := apiGroup.Group("/docs")
{
    docGroup.POST("", docHandler.CreateDocument)          // 新建文档
    docGroup.GET("/:docid", docHandler.GetDocument)        // 获取文档信息
    docGroup.PUT("/:docid/rename", docHandler.RenameDocument)  // 重命名
    docGroup.DELETE("/:docid", docHandler.DeleteDocument)   // 删除文档
    docGroup.POST("/:docid/share", docHandler.ShareDocument) // 分享
    docGroup.GET("/:docid/permissions", docHandler.GetPermissions) // 权限
    docGroup.PUT("/:docid/content", docHandler.EditContent) // 编辑内容
    docGroup.GET("/:docid/data", docHandler.GetDocumentData) // 获取数据
    docGroup.POST("/:docid/images", docHandler.UploadImage)  // 上传图片

    // 表格
    docGroup.POST("/sheets/:docid/content", docHandler.EditSheetContent)
    docGroup.GET("/sheets/:docid/rows", docHandler.GetSheetRowCol)
    docGroup.GET("/sheets/:docid/data", docHandler.GetSheetData)

    // 空间
    docGroup.POST("/spaces", docHandler.CreateSpace)
    docGroup.GET("/spaces/:spaceid", docHandler.GetSpaceInfo)
    docGroup.GET("/spaces/:spaceid/files", docHandler.GetSpaceFileList)
}
```

### 3.2 请求/响应类型

```go
// 新建文档请求
type CreateDocumentRequest struct {
    OwnerUserID string `json:"owner_userid" binding:"required"`
    Name        string `json:"name" binding:"required"`
    Type        string `json:"type" binding:"required"`  // doc/sheet/bitable/mindnote/docx
    SpaceID     string `json:"space_id"`                  // optional
}

// 重命名请求
type RenameDocumentRequest struct {
    Name string `json:"name" binding:"required"`
}

// 分享请求
type ShareDocumentRequest struct {
    ShareType int   `json:"share_type" binding:"required"`
    ExpireTime int  `json:"expire_time"`
}

// 编辑表格请求
type EditSheetRequest struct {
    Row   int    `json:"row" binding:"required"`
    Col   int    `json:"col" binding:"required"`
    Value string `json:"value" binding:"required"`
}

// 文档信息响应
type DocumentInfo struct {
    DocID      string `json:"doc_id"`
    Name       string `json:"name"`
    Type       string `json:"type"`
    URL        string `json:"url"`
    CreatorID  string `json:"creator_id"`
    CreateTime int64  `json:"create_time"`
    UpdateTime int64  `json:"update_time"`
}

// 空间信息响应
type SpaceInfo struct {
    SpaceID   string `json:"space_id"`
    Name      string `json:"name"`
    AdminList []string `json:"admin_list"`
    MemberCount int   `json:"member_count"`
}
```

---

## 4. 权限校验

### 4.1 权限映射

| API | 所需权限 |
|-----|---------|
| POST /v1/docs | `document:write` |
| GET /v1/docs/:docid | `document:read` |
| PUT /v1/docs/:docid/rename | `document:write` |
| DELETE /v1/docs/:docid | `document:write` |
| POST /v1/docs/:docid/share | `document:write` |
| GET /v1/docs/:docid/permissions | `document:read` |
| PUT /v1/docs/:docid/content | `document:write` |
| GET /v1/docs/:docid/data | `document:read` |
| POST /v1/docs/:docid/images | `document:write` |
| POST /v1/docs/sheets/:docid/content | `document:write` |
| GET /v1/docs/sheets/:docid/rows | `document:read` |
| GET /v1/docs/sheets/:docid/data | `document:read` |
| POST /v1/docs/spaces | `document:write` |
| GET /v1/docs/spaces/:spaceid | `document:read` |
| GET /v1/docs/spaces/:spaceid/files | `document:read` |

### 4.2 权限检查函数

在 `internal/auth/permissions.go` 中新增：

```go
// HasDocumentReadPermission 检查是否有文档读取权限
func HasDocumentReadPermission(permissions []string) bool {
    return hasPermission(permissions, "document:read")
}

// HasDocumentWritePermission 检查是否有文档写入权限
func HasDocumentWritePermission(permissions []string) bool {
    return hasPermission(permissions, "document:write")
}
```

---

## 5. 审计日志集成

所有文档操作记录到审计日志：

| 操作 | Action | 资源类型 |
|------|--------|---------|
| 新建文档 | `document.create` | `doc:{docid}` |
| 重命名 | `document.rename` | `doc:{docid}` |
| 删除 | `document.delete` | `doc:{docid}` |
| 获取信息 | `document.get` | `doc:{docid}` |
| 分享 | `document.share` | `doc:{docid}` |
| 编辑内容 | `document.edit` | `doc:{docid}` |
| 获取数据 | `document.read` | `doc:{docid}` |
| 上传图片 | `document.upload_image` | `doc:{docid}` |

---

## 6. 实现计划

### Phase 1：核心文档管理（P0）
- [ ] `internal/document/client.go` — DocumentClient 实现
- [ ] `internal/document/types.go` — 类型定义
- [ ] `internal/document/service.go` — 业务逻辑
- [ ] `internal/document/handler.go` — HTTP handler
- [ ] 新建/重命名/删除/获取文档信息
- [ ] 权限校验集成
- [ ] 审计日志集成
- [ ] 单元测试

### Phase 2：文档内容管理（P1）
- [ ] 编辑文档内容
- [ ] 获取文档数据
- [ ] 上传文档图片
- [ ] 表格编辑/读取

### Phase 3：空间管理（P2）
- [ ] 新建空间
- [ ] 获取空间信息
- [ ] 获取文件列表

### Phase 4：前端管理（P3）
- [ ] 管理后台增加文档管理页面

---

## 7. 注意事项

1. **企业微信文档 API 使用 POST 方法**：所有企业微信文档 API 均为 POST 请求（即使查询操作），网关内部将其映射为 RESTful 风格
2. **access_token 缓存**：复用现有的 `wecom.Client` token 管理机制
3. **企业微信应用需开启文档权限**：应用需在管理后台配置文档相关权限
4. **文档内容编辑格式**：企业微信文档内容编辑使用操作列表（operations），格式较复杂，需要详细参考官方文档
5. **文件大小限制**：上传文档图片有大小限制（参考企业微信官方限制）
