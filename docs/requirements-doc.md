# 需求文档：WeCom Gateway — 企业微信文档管理

| 项目 | 内容 |
|------|------|
| 项目名称 | wecom-gateway（文档管理模块） |
| 文档版本 | v1.0 |
| 创建日期 | 2026-03-29 |
| 作者 | M10S |
| 状态 | 草案 |
| 需求来源 | 洪岩指令：通过网关创建文档、编辑文档、删除文档等 |
| API 参考 | [企业微信文档 API](https://developer.work.weixin.qq.com/document/path/97392) |

---

## 1. 功能概述

在 WeCom Gateway 中新增企业微信文档管理能力，使 AI 智能体可以通过统一的 RESTful API 对企业微信文档进行 CRUD 操作，包括：

- **文档管理**：新建、重命名、删除、获取信息、分享、权限管理
- **文档内容管理**：编辑文档内容、获取文档数据
- **表格内容管理**：编辑表格、获取表格行列、获取表格数据
- **空间管理**：新建空间、获取空间信息、文件列表管理
- **文档图片上传**：上传文档中的图片

### 1.1 企业微信文档 API 对照

| 网关 API | 企业微信 API | 说明 |
|----------|-------------|------|
| POST /v1/docs | 新建文档 (97460) | 创建新文档 |
| PUT /v1/docs/:docid/rename | 重命名文档 (97736) | 修改文档名称 |
| DELETE /v1/docs/:docid | 删除文档 (97735) | 删除指定文档 |
| GET /v1/docs/:docid | 获取文档基础信息 (97734) | 查询文档详情 |
| POST /v1/docs/:docid/share | 分享文档 (97733) | 设置文档分享 |
| GET /v1/docs/:docid/permissions | 获取文档权限信息 (97461) | 查询权限 |
| PUT /v1/docs/:docid/content | 编辑文档内容 (97626) | 编辑文档 |
| GET /v1/docs/:docid/data | 获取文档数据 (101161) | 读取文档内容 |
| POST /v1/docs/sheets/:docid/content | 编辑表格内容 (101168) | 编辑表格 |
| GET /v1/docs/sheets/:docid/rows | 获取表格行列信息 (97711) | 查询行列 |
| GET /v1/docs/sheets/:docid/data | 获取表格数据 (97661) | 读取表格 |
| POST /v1/docs/spaces | 新建空间 (93655) | 创建空间 |
| GET /v1/docs/spaces/:spaceid | 获取空间信息 (97858) | 查询空间 |
| GET /v1/docs/spaces/:spaceid/files | 获取文件列表 (93657) | 列出文件 |
| POST /v1/docs/:docid/images | 上传文档图片 (99933) | 上传图片 |

---

## 2. 功能需求

### 2.1 文档生命周期管理

#### FR-D001：新建文档
- **描述**：创建一个新的企业微信文档
- **请求**：`POST /v1/docs`
- **参数**：
  - `owner_userid` (string, required)：文档所有者 userid
  - `name` (string, required)：文档标题
  - `type` (string, required)：文档类型，枚举：`doc`（文档）、`sheet`（表格）、`bitable`（智能表格）、`mindnote`（脑图）、`docx`（新版文档）
  - `space_id` (string, optional)：所属空间 ID
- **权限**：`document:write`

#### FR-D002：重命名文档
- **描述**：修改已有文档的名称
- **请求**：`PUT /v1/docs/:docid/rename`
- **参数**：
  - `name` (string, required)：新文档名称
- **权限**：`document:write`

#### FR-D003：删除文档
- **描述**：删除指定文档（移入回收站或彻底删除）
- **请求**：`DELETE /v1/docs/:docid`
- **参数**：
  - `docid` (path, required)：文档 ID
- **权限**：`document:write`

#### FR-D004：获取文档基础信息
- **描述**：获取文档的基本元信息
- **请求**：`GET /v1/docs/:docid`
- **参数**：
  - `docid` (path, required)：文档 ID
- **权限**：`document:read`

#### FR-D005：分享文档
- **描述**：设置文档的分享方式
- **请求**：`POST /v1/docs/:docid/share`
- **参数**：
  - `share_type` (int, required)：分享类型（1=仅企业内, 2=指定人, 3=链接分享）
  - `expire_time` (int, optional)：链接过期时间戳（秒）
- **权限**：`document:write`

#### FR-D006：获取文档权限信息
- **描述**：查询文档的权限设置
- **请求**：`GET /v1/docs/:docid/permissions`
- **参数**：
  - `docid` (path, required)：文档 ID
- **权限**：`document:read`

---

### 2.2 文档内容管理

#### FR-D010：编辑文档内容
- **描述**：编辑文档的正文内容（通过操作类型增量编辑）
- **请求**：`PUT /v1/docs/:docid/content`
- **参数**：
  - `docid` (path, required)：文档 ID
  - `operations` (array, required)：操作列表
    - `op_type` (int)：操作类型（1=插入, 2=删除, 3=替换）
    - `position` (object)：操作位置
    - `content` (string)：操作内容
- **权限**：`document:write`

#### FR-D011：获取文档数据
- **描述**：获取文档的完整内容数据
- **请求**：`GET /v1/docs/:docid/data`
- **参数**：
  - `docid` (path, required)：文档 ID
- **权限**：`document:read`

#### FR-D012：上传文档图片
- **描述**：上传图片到文档中
- **请求**：`POST /v1/docs/:docid/images`
- **参数**：
  - `image_url` (string, optional)：图片 URL（网关先下载再上传）
  - `image_base64` (string, optional)：图片 Base64 编码
  - `image_file` (file, optional)：图片文件上传
- **权限**：`document:write`

---

### 2.3 表格内容管理

#### FR-D020：编辑表格内容
- **描述**：编辑表格单元格内容
- **请求**：`POST /v1/docs/sheets/:docid/content`
- **参数**：
  - `docid` (path, required)：表格文档 ID
  - `row` (int, required)：行号（0-based）
  - `col` (int, required)：列号（0-based）
  - `value` (string, required)：单元格值
- **权限**：`document:write`

#### FR-D021：获取表格行列信息
- **描述**：获取表格的行列元信息
- **请求**：`GET /v1/docs/sheets/:docid/rows`
- **参数**：
  - `docid` (path, required)：表格文档 ID
- **权限**：`document:read`

#### FR-D022：获取表格数据
- **描述**：获取表格的完整数据
- **请求**：`GET /v1/docs/sheets/:docid/data`
- **参数**：
  - `docid` (path, required)：表格文档 ID
  - `range` (string, optional)：数据范围，如 "A1:C10"
- **权限**：`document:read`

---

### 2.4 空间管理

#### FR-D030：新建空间
- **描述**：创建一个新的文档空间
- **请求**：`POST /v1/docs/spaces`
- **参数**：
  - `name` (string, required)：空间名称
  - `admin_userid` (string, required)：空间管理员 userid
- **权限**：`document:write`

#### FR-D031：获取空间信息
- **描述**：获取空间的详细信息
- **请求**：`GET /v1/docs/spaces/:spaceid`
- **参数**：
  - `spaceid` (path, required)：空间 ID
- **权限**：`document:read`

#### FR-D032：获取文件列表
- **描述**：获取空间内的文件列表
- **请求**：`GET /v1/docs/spaces/:spaceid/files`
- **参数**：
  - `spaceid` (path, required)：空间 ID
  - `cursor` (string, optional)：分页游标
  - `limit` (int, optional)：每页数量，默认 50，最大 100
- **权限**：`document:read`

---

## 3. 权限设计

### 3.1 新增权限

| 权限 | 说明 |
|------|------|
| `document:read` | 读取文档/空间信息 |
| `document:write` | 创建/编辑/删除文档、分享设置 |

### 3.2 通配符权限

`*` 通配符权限自动包含 `document:read` 和 `document:write`。

---

## 4. API Key 权限更新

在现有的 API Key 创建接口中，`permissions` 字段新增可选值：

```
calendar:read      # 日程读取（已有）
calendar:write     # 日程写入（已有）
meetingroom:read   # 会议室读取（已有）
meetingroom:write  # 会议室写入（已有）
message:send       # 消息发送（已有）
document:read      # 文档读取（新增）
document:write     # 文档写入（新增）
*                  # 管理员通配符（已有）
```

---

## 5. 审计日志

所有文档操作纳入审计日志系统，记录：
- 操作类型（create/rename/delete/get/share/edit）
- 文档 ID
- 操作者（API Key 对应的智能体名称）
- 操作时间

---

## 6. 错误码

| 错误码 | 说明 |
|--------|------|
| 40101 | 未授权 / API Key 无效 |
| 40301 | 权限不足（缺少 document:read 或 document:write） |
| 40001 | 参数错误 |
| 40401 | 文档不存在 |
| 40901 | 文档名称冲突 |
| 50001 | 企业微信 API 调用失败 |
