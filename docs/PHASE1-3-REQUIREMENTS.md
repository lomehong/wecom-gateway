# wecom-gateway Phase 1-3 能力补全需求文档

> 基于 wecom-cli 官方能力分析，补全 wecom-gateway 缺失功能，打造无企业规模限制的企业微信 AI Agent 网关。

## 项目架构约定

- 语言：Go 1.25
- Web 框架：Gin
- 数据库：SQLite（开发）/ PostgreSQL（生产）
- 认证：API Key Bearer Token + 权限体系
- 项目根目录：`/home/m10s/projects/wecom-gateway`
- Git remote：`https://github.com/lomehong/wecom-gateway.git`

## 现有权限标识

已有权限：`calendar:read`, `calendar:write`, `meetingroom:read`, `meetingroom:write`, `message:send`, `document:read`, `document:write`

---

## Phase 1: 能力补全

### 1.1 通讯录模块 (`internal/contact/`)

**企业微信 API 参考**：https://developer.work.weixin.qq.com/document/path/96065

#### 新增文件
- `internal/contact/types.go` — 类型定义
- `internal/contact/client.go` — 企业微信 API 调用
- `internal/contact/service.go` — 业务逻辑
- `internal/contact/handler.go` — HTTP handler
- `internal/contact/handler_test.go` — 单元测试
- `internal/contact/service_test.go` — 单元测试

#### API 接口

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/v1/contacts/users` | `contact:read` | 获取可见范围成员列表 |
| GET | `/v1/contacts/users/search` | `contact:read` | 按姓名/别名搜索成员 |

#### wecom.Client 接口新增
```go
// Contact operations
GetUserList(ctx context.Context, corpName, appName string, departmentID int) ([]*ContactUser, error)
SearchUser(ctx context.Context, corpName, appName string, query string) ([]*ContactUser, error)
```

#### 类型定义
```go
type ContactUser struct {
    UserID   string `json:"userid"`
    Name     string `json:"name"`
    Alias    string `json:"alias,omitempty"`
    Mobile   string `json:"mobile,omitempty"`
    Email    string `json:"email,omitempty"`
    Department []int `json:"department,omitempty"`
    Position string `json:"position,omitempty"`
    Gender   int    `json:"gender,omitempty"`
    Status   int    `json:"status,omitempty"`
    Avatar   string `json:"avatar,omitempty"`
}
```

#### 企业微信 API 调用
- `GET /cgi-bin/user/simplelist?access_token=TOKEN&department_id=DEPT_ID` — 获取部门成员
- `GET /cgi-bin/user/list?access_token=TOKEN&department_id=DEPT_ID` — 获取部门成员详情
- `POST /cgi-bin/user/list_id` — 按 UserID 批量查询

---

### 1.2 待办模块 (`internal/todo/`)

**企业微信 API 参考**：https://developer.work.weixin.qq.com/document/path/93670

#### 新增文件
- `internal/todo/types.go`
- `internal/todo/service.go`
- `internal/todo/handler.go`
- `internal/todo/handler_test.go`
- `internal/todo/service_test.go`

#### API 接口

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/v1/todos` | `todo:read` | 查询待办列表（分页+时间过滤） |
| GET | `/v1/todos/:id` | `todo:read` | 查询待办详情 |
| POST | `/v1/todos` | `todo:write` | 创建待办 |
| PUT | `/v1/todos/:id` | `todo:write` | 更新待办（内容/状态/提醒） |
| DELETE | `/v1/todos/:id` | `todo:write` | 删除待办 |
| PUT | `/v1/todos/:id/status` | `todo:write` | 变更用户处理状态 |

#### wecom.Client 接口新增
```go
// Todo operations
GetTodoList(ctx context.Context, corpName, appName string, opts *TodoListOptions) (*TodoListResult, error)
GetTodoDetail(ctx context.Context, corpName, appName string, todoIDs []string) ([]*TodoDetail, error)
CreateTodo(ctx context.Context, corpName, appName string, params *CreateTodoParams) (string, error)
UpdateTodo(ctx context.Context, corpName, appName string, todoID string, params *UpdateTodoParams) error
DeleteTodo(ctx context.Context, corpName, appName string, todoID string) error
ChangeTodoUserStatus(ctx context.Context, corpName, appName string, todoID string, status int) error
```

#### 类型定义
```go
type TodoListOptions struct {
    CreateBeginTime *time.Time
    CreateEndTime   *time.Time
    RemindBeginTime *time.Time
    RemindEndTime   *time.Time
    Limit           int
    Cursor          string
}

type TodoListResult struct {
    IndexList []TodoIndex `json:"index_list"`
    NextCursor string    `json:"next_cursor"`
    HasMore   bool       `json:"has_more"`
}

type TodoIndex struct {
    TodoID     string     `json:"todo_id"`
    TodoStatus int        `json:"todo_status"`
    UserStatus int        `json:"user_status"`
    CreatorID  string     `json:"creator_id"`
    RemindTime *time.Time `json:"remind_time,omitempty"`
    CreateTime time.Time  `json:"create_time"`
    UpdateTime time.Time  `json:"update_time"`
}

type TodoDetail struct {
    TodoIndex
    Content    string   `json:"content"`
    Assignees  []string `json:"assignees,omitempty"`
}

type CreateTodoParams struct {
    Content    string     `json:"content"`
    Assignees  []string   `json:"assignees,omitempty"`
    RemindTime *time.Time `json:"remind_time,omitempty"`
}

type UpdateTodoParams struct {
    Content    *string     `json:"content,omitempty"`
    Status     *int        `json:"status,omitempty"`     // 0=完成, 1=进行中
    RemindTime *time.Time `json:"remind_time,omitempty"`
    Assignees  []string   `json:"assignees,omitempty"`
}
```

#### 企业微信 API 调用
- `POST /cgi-bin/oa/todo/list` — 查询待办列表
- `POST /cgi-bin/oa/todo/get` — 查询待办详情
- `POST /cgi-bin/oa/todo/add` — 创建待办
- `POST /cgi-bin/oa/todo/update` — 更新待办
- `POST /cgi-bin/oa/todo/del` — 删除待办
- `POST /cgi-bin/oa/todo/update_user_status` — 变更用户状态

---

### 1.3 会议预约模块 (扩展 `internal/meeting/`)

**注意**：现有 meeting 模块是"会议室管理"（booking），需要新增"会议预约"功能。

#### 新增文件
- `internal/meeting/appointment.go` — 会议预约相关类型和方法
- `internal/meeting/appointment_test.go`

#### 扩展 API 接口

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/v1/meetings` | `meeting:write` | 创建预约会议 |
| DELETE | `/v1/meetings/:id` | `meeting:write` | 取消会议 |
| PUT | `/v1/meetings/:id/invitees` | `meeting:write` | 更新受邀成员 |
| GET | `/v1/meetings` | `meeting:read` | 查询会议列表 |
| GET | `/v1/meetings/:id` | `meeting:read` | 获取会议详情 |

#### wecom.Client 接口新增
```go
// Meeting appointment operations
CreateMeeting(ctx context.Context, corpName, appName string, params *CreateMeetingParams) (*MeetingInfo, error)
CancelMeeting(ctx context.Context, corpName, appName string, meetingID string) error
UpdateMeetingInvitees(ctx context.Context, corpName, appName string, meetingID string, invitees *MeetingInvitees) error
ListMeetings(ctx context.Context, corpName, appName string, opts *MeetingListOptions) (*MeetingListResult, error)
GetMeetingInfo(ctx context.Context, corpName, appName string, meetingID string) (*MeetingInfo, error)
```

#### 类型定义
```go
type CreateMeetingParams struct {
    Title           string   `json:"title"`
    StartDateTime   time.Time `json:"meeting_start_datetime"`
    Duration        int      `json:"meeting_duration"`        // 秒
    Invitees        *MeetingInvitees `json:"invitees,omitempty"`
    MeetingType     int      `json:"meeting_type,omitempty"`   // 0=视频, 1=语音, ...
    Settings        *MeetingSettings `json:"settings,omitempty"`
}

type MeetingInvitees struct {
    UserIDs  []string `json:"userid,omitempty"`
    DeptIDs  []string `json:"department,omitempty"`
}

type MeetingSettings struct {
    MuteUponEntry   bool `json:"mute_upon_entry,omitempty"`
    WaitingRoom     bool `json:"waiting_room,omitempty"`
    EnableRecording bool `json:"enable_recording,omitempty"`
}

type MeetingInfo struct {
    MeetingID      string   `json:"meetingid"`
    Title          string   `json:"title"`
    Status         int      `json:"status"`
    StartDateTime  time.Time `json:"meeting_start_datetime"`
    Duration       int      `json:"meeting_duration"`
    Creator        string   `json:"creator"`
    Invitees       []MeetingInviteeInfo `json:"invitees,omitempty"`
    MeetingLink    string   `json:"meeting_link,omitempty"`
}

type MeetingListOptions struct {
    BeginDatetime string `json:"begin_datetime"`
    EndDatetime   string `json:"end_datetime"`
    Limit         int    `json:"limit,omitempty"`
    Cursor        string `json:"cursor,omitempty"`
}
```

#### 企业微信 API 调用
- `POST /cgi-bin/meeting/create_meeting` — 创建会议
- `POST /cgi-bin/meeting/cancel_meeting` — 取消会议
- `POST /cgi-bin/meeting/set_invite_meeting_members` — 更新受邀人
- `POST /cgi-bin/meeting/list_user_meetings` — 查询会议列表
- `POST /cgi-bin/meeting/get_meeting_info` — 获取会议详情

---

## Phase 2: Agent 生态

### 2.1 MCP 协议端点 (`internal/mcp/`)

实现 Model Context Protocol (MCP) 端点，让 AI Agent 能自动发现和调用网关能力。

#### 新增文件
- `internal/mcp/server.go` — MCP JSON-RPC 处理器
- `internal/mcp/tools.go` — 工具注册和描述
- `internal/mcp/handler.go` — Gin HTTP handler
- `internal/mcp/server_test.go`
- `internal/mcp/tools_test.go`

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/mcp` | MCP SSE 连接（Server-Sent Events） |
| POST | `/mcp` | MCP JSON-RPC 请求 |

#### MCP 工具列表（自动从路由注册）
```
wecom_create_schedule, wecom_get_schedules, wecom_update_schedule, wecom_delete_schedule,
wecom_list_meeting_rooms, wecom_book_meeting_room,
wecom_create_meeting, wecom_cancel_meeting, wecom_list_meetings,
wecom_send_text, wecom_send_markdown, wecom_send_image, wecom_send_file, wecom_send_card,
wecom_create_document, wecom_edit_document, wecom_get_document,
wecom_get_contacts, wecom_search_contact,
wecom_get_todos, wecom_create_todo, wecom_update_todo, wecom_delete_todo,
wecom_check_availability
```

每个工具需提供 `name`, `description`, `inputSchema`（JSON Schema）。

### 2.2 OpenAPI 3.0 规范 (`internal/openapi/`)

#### 新增文件
- `internal/openapi/spec.go` — OpenAPI 规范生成器
- `internal/openapi/handler.go` — HTTP handler

#### API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/openapi.json` | OpenAPI 3.0 JSON 规范 |
| GET | `/openapi.yaml` | OpenAPI 3.0 YAML 规范 |
| GET | `/docs` | Swagger UI 页面 |

### 2.3 Agent Skills 包

为每个品类生成 `SKILL.md` 文件（参考 wecom-cli 的格式），发布到：
- `~/.agents/skills/wecom-gateway-*/`
- 同时 symlink 到 Claude Code 和 OpenClaw

---

## Phase 3: 能力增强

### 3.1 消息拉取 (扩展 `internal/message/`)

#### 新增 API

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/v1/messages/chats` | `message:read` | 获取会话列表 |
| GET | `/v1/messages/chats/:chatid/messages` | `message:read` | 拉取会话消息 |
| GET | `/v1/messages/media/:mediaid` | `message:read` | 下载多媒体文件 |

#### 新增权限
- `message:read` — 消息读取权限

#### 企业微信 API
- `POST /cgi-bin/message/get_msg_chat_list` — 会话列表（最近7天）
- `POST /cgi-bin/message/get_message` — 消息记录
- `POST /cgi-bin/media/get` — 下载媒体

### 3.2 闲忙查询 (扩展 `internal/schedule/`)

#### 新增 API

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/v1/schedules/availability` | `calendar:read` | 查询多成员闲忙 |

#### wecom.Client 接口新增
```go
CheckAvailability(ctx context.Context, corpName, appName string, opts *AvailabilityOptions) ([]*UserAvailability, error)
```

### 3.3 智能表格 (扩展 `internal/document/`)

#### 新增 API

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/v1/sheets` | `document:write` | 创建智能表格 |
| GET | `/v1/sheets/:docid/sheets` | `document:read` | 查询子表列表 |
| POST | `/v1/sheets/:docid/sheets` | `document:write` | 添加子表 |
| GET | `/v1/sheets/:docid/sheets/:sheetid/fields` | `document:read` | 查询字段信息 |
| POST | `/v1/sheets/:docid/sheets/:sheetid/fields` | `document:write` | 添加字段 |
| PUT | `/v1/sheets/:docid/sheets/:sheetid/fields` | `document:write` | 更新字段 |
| DELETE | `/v1/sheets/:docid/sheets/:sheetid/fields/:fieldid` | `document:write` | 删除字段 |
| GET | `/v1/sheets/:docid/sheets/:sheetid/records` | `document:read` | 查询记录 |
| POST | `/v1/sheets/:docid/sheets/:sheetid/records` | `document:write` | 添加记录 |
| PUT | `/v1/sheets/:docid/sheets/:sheetid/records` | `document:write` | 更新记录 |
| DELETE | `/v1/sheets/:docid/sheets/:sheetid/records` | `document:write` | 删除记录 |

### 3.4 凭证轮换 + 外部密钥管理

#### 扩展 `internal/admin/`
- 凭证轮换 API：`POST /v1/admin/apps/:id/rotate-secret`
- 密钥提供器接口：`internal/secret/provider.go`
  - 默认：本地数据库
  - 扩展：环境变量
  - 扩展预留：HashiCorp Vault / AWS Secrets Manager

---

## 实施顺序

1. **Phase 1.1** — contact 模块（最简单，无依赖）
2. **Phase 1.2** — todo 模块
3. **Phase 1.3** — meeting 扩展
4. **Phase 3.1** — 消息拉取
5. **Phase 3.2** — 闲忙查询
6. **Phase 3.3** — 智能表格
7. **Phase 3.4** — 凭证轮换
8. **Phase 2.2** — OpenAPI 规范
9. **Phase 2.1** — MCP 协议
10. **Phase 2.3** — Agent Skills 包

## 路由注册模板

在 `cmd/server/main.go` 中按以下模式注册新路由：

```go
// Phase 1: New modules
contactGroup := v1.Group("/contacts")
contactGroup.Use(auth.GinMiddleware(authenticator))
    contactGroup.GET("/users", auth.RequirePermission("contact:read"), contactHandler.GetUserList)
    contactGroup.GET("/users/search", auth.RequirePermission("contact:read"), contactHandler.SearchUser)

todoGroup := v1.Group("/todos")
todoGroup.Use(auth.GinMiddleware(authenticator))
    todoGroup.GET("", auth.RequirePermission("todo:read"), todoHandler.GetTodoList)
    todoGroup.GET("/:id", auth.RequirePermission("todo:read"), todoHandler.GetTodoDetail)
    todoGroup.POST("", auth.RequirePermission("todo:write"), todoHandler.CreateTodo)
    todoGroup.PUT("/:id", auth.RequirePermission("todo:write"), todoHandler.UpdateTodo)
    todoGroup.DELETE("/:id", auth.RequirePermission("todo:write"), todoHandler.DeleteTodo)
    todoGroup.PUT("/:id/status", auth.RequirePermission("todo:write"), todoHandler.ChangeUserStatus)

meetingApptGroup := v1.Group("/meetings")
meetingApptGroup.Use(auth.GinMiddleware(authenticator))
    meetingApptGroup.POST("", auth.RequirePermission("meeting:write"), meetingHandler.CreateMeeting)
    meetingApptGroup.DELETE("/:id", auth.RequirePermission("meeting:write"), meetingHandler.CancelMeeting)
    meetingApptGroup.PUT("/:id/invitees", auth.RequirePermission("meeting:write"), meetingHandler.UpdateInvitees)
    meetingApptGroup.GET("", auth.RequirePermission("meeting:read"), meetingHandler.ListMeetings)
    meetingApptGroup.GET("/:id", auth.RequirePermission("meeting:read"), meetingHandler.GetMeetingInfo)
```

## 新增权限标识汇总

| 权限标识 | 说明 |
|---------|------|
| `contact:read` | 读取通讯录 |
| `todo:read` | 读取待办 |
| `todo:write` | 创建/编辑/删除待办 |
| `meeting:read` | 查询会议 |
| `meeting:write` | 创建/取消/修改会议 |
| `message:read` | 读取消息记录 |

## 单元测试要求

每个模块必须包含：
1. `*_test.go` 文件，覆盖所有 handler 和 service 方法
2. 使用 `internal/wecom/mock_client.go` 中的 MockClient
3. 测试覆盖率目标 ≥ 80%
4. 所有测试通过 `go test ./...`
