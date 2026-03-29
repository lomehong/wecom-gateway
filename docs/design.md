# 设计文档：WeCom Gateway — 企业微信智能体网关

| 项目 | 内容 |
|------|------|
| 项目名称 | wecom-gateway |
| 文档版本 | v1.1 |
| 创建日期 | 2026-03-29 |
| 作者 | M10S |
| 需求来源 | [需求文档](./requirements.md) |
| 状态 | 草案 |

---

## 1. 系统架构

### 1.1 整体架构

```
┌─────────────┐    HTTPS    ┌──────────────────────────────────────────────────┐
│             │  ◄───────►  │                    wecom-gateway                   │
│  AI 智能体   │             │                                                    │
│  (REST API) │             │  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│             │             │  │  Auth    │→ │ Scheduler │→ │   Meeting    │  │
└─────────────┘             │  │ 中间件   │  │  日程服务  │  │    会议室     │  │
                            │  └──────────┘  └───────────┘  └──────────────┘  │
┌─────────────┐   gRPC     │  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│             │  ◄───────►  │  │  Message │  │   Audit   │  │   WeCom      │  │
│  AI 智能体   │             │  │  消息服务 │  │  审计日志  │  │   Client     │  │
│  (gRPC API) │             │  └──────────┘  └───────────┘  └──────┬───────┘  │
└─────────────┘             │  ┌──────────┐  ┌───────────┐         │          │
                            │  │   Key    │  │   Admin   │         │          │
┌─────────────┐             │  │  密钥管理 │  │  管理服务  │         │          │
│             │  ◄───────►  │  └──────────┘  └───────────┘         │          │
│  管理后台    │             │  ┌──────────┐  ┌───────────┐         │          │
│  (Web UI)   │             │  │  gRPC    │  │  Swagger  │         │          │
│             │             │  │  Server  │  │   UI      │         │          │
└─────────────┘             │  └──────────┘  └───────────┘         │          │
                            └─────────────────────────────────────┼──────────┘
                                                                 │ HTTPS
                                                                 ▼
                                                      ┌──────────────────┐
                                                      │   企业微信 API     │
                                                      │  (api.weixin.qq)  │
                                                      └──────────────────┘
                                                                 │
                                                                 ▼
                                                      ┌──────────────────┐
                                                      │   SQLite/PG      │
                                                      │   持久化存储       │
                                                      └──────────────────┘
```

### 1.2 核心数据流

**智能体调用流程（以创建日程为例）：**

```
AI 智能体 POST /v1/schedules
    Authorization: Bearer wgk_xxxx
    Body: { organizer: "zhangsan", summary: "周会", ... }
        │
        ▼
Auth 中间件：验证 API Key → 提取权限 → 注入 Context
        │
        ▼
Schedule Handler：参数校验 → 权限检查(calendar:write)
        │
        ▼
Schedule Service：调用 WeCom Client → 创建日程
        │
        ├── 成功 → 返回 201 + 日程详情
        │
        └── 失败 → 返回 4xx/5xx + 错误描述
        │
        ▼
Audit Logger（异步）：记录请求日志到数据库
```

### 1.3 多租户（多应用 + 多企业）模型

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────┐
│ API Key A   │────►│ Corp 1 / App 1  │────►│ WeCom API    │
│ (OA Agent)  │     │ (OA系统)        │     │              │
├─────────────┤     ├─────────────────┤     │              │
│ API Key B   │────►│ Corp 1 / App 2  │────►│  Corp 1      │
│ (HR Agent)  │     │ (HR系统)        │     │              │
├─────────────┤     └─────────────────┘     ├──────────────┤
│ Admin Key   │                             │              │
├─────────────┤     ┌─────────────────┐     │              │
│ API Key C   │────►│ Corp 2 / App 1  │────►│  Corp 2      │
│ (外部Agent) │     │ (外部合作)       │     │              │
└─────────────┘     └─────────────────┘     └──────────────┘
```

- 每个 API Key 绑定一个 **CorpID + App** 组合
- 不同企业的 access_token 完全隔离
- 管理员 Key 不绑定特定应用，可管理所有企业的资源

### 1.4 双协议架构（REST + gRPC）

```
                    ┌────────────────────────────┐
                    │      Business Logic        │
                    │  (schedule/message/...)    │
                    └──────────┬─────────────────┘
                               │
                 ┌─────────────┼─────────────┐
                 ▼                           ▼
        ┌─────────────┐             ┌─────────────┐
        │  HTTP Layer  │             │  gRPC Layer  │
        │  (Gin)       │             │  (grpc-go)   │
        └──────┬──────┘             └──────┬──────┘
               │                           │
               └───────────┬───────────────┘
                           │
                  ┌────────┴────────┐
                  │  Shared Middle │
                  │  Auth / Audit  │
                  │  / RateLimit   │
                  └─────────────────┘
```

- 业务逻辑层与传输协议解耦，通过接口定义
- Auth、Audit、RateLimit 逻辑抽取为公共中间件/拦截器
- gRPC 拦截器与 HTTP 中间件调用同一套底层实现

---

## 2. 模块设计

### 2.1 包结构

```
wecom-gateway/
├── cmd/
│   └── server/
│       └── main.go                 # 入口（启动 HTTP + gRPC）
├── internal/
│   ├── config/
│   │   └── config.go               # 配置加载（环境变量 + YAML）
│   ├── auth/
│   │   ├── middleware.go            # HTTP 认证中间件
│   │   ├── apikey.go               # API Key 认证逻辑
│   │   └── apikey_test.go
│   ├── apikey/
│   │   ├── store.go                # API Key 存储接口 + 实现
│   │   ├── service.go              # API Key 管理服务
│   │   └── service_test.go
│   ├── wecom/
│   │   ├── client.go               # 企业微信 API 客户端接口
│   │   ├── token.go                # access_token 管理（自动刷新）
│   │   ├── schedule.go             # 日程 API 封装
│   │   ├── meetingroom.go          # 会议室 API 封装
│   │   ├── message.go              # 消息 API 封装
│   │   ├── media.go                # 素材上传 API 封装
│   │   └── types.go                # 企业微信 API 类型定义
│   ├── schedule/
│   │   ├── handler.go              # 日程 HTTP handler
│   │   ├── service.go              # 日程业务逻辑
│   │   └── service_test.go
│   ├── meeting/
│   │   ├── handler.go              # 会议室 HTTP handler
│   │   ├── service.go              # 会议室业务逻辑
│   │   └── service_test.go
│   ├── message/
│   │   ├── handler.go              # 消息 HTTP handler
│   │   ├── service.go              # 消息业务逻辑
│   │   └── service_test.go
│   ├── audit/
│   │   ├── logger.go               # 审计日志记录器
│   │   ├── store.go                # 审计日志存储
│   │   └── query.go                # 审计日志查询 handler
│   ├── admin/
│   │   ├── handler.go              # 管理接口 handler
│   │   └── service.go              # 管理业务逻辑
│   ├── store/
│   │   ├── db.go                   # 数据库抽象层
│   │   ├── sqlite.go               # SQLite 实现
│   │   └── postgres.go             # PostgreSQL 实现
│   ├── crypto/
│   │   └── crypto.go               # 加密工具（AES-256-GCM, bcrypt）
│   ├── httputil/
│   │   └── response.go             # 统一响应格式、错误码
│   ├── ratelimit/
│   │   └── ratelimit.go            # 令牌桶限流器
│   ├── grpcserver/
│   │   ├── server.go               # gRPC 服务端
│   │   ├── interceptors.go         # 认证/审计/限流/recovery 拦截器
│   │   └── server_test.go
│   └── ui/
│       └── embed.go                # 管理后台静态资源（go:embed）
├── api/
│   └── proto/
│       ├── wecom_gateway.proto     # gRPC 服务定义
│       └── wecom_gateway.pb.go     # 生成的 Go 代码
├── ui/                             # 管理后台前端源码
│   ├── index.html
│   ├── css/
│   ├── js/
│   └── dist/                       # 构建产物
├── docs/
│   ├── requirements.md
│   └── design.md
├── config.example.yaml
├── go.mod
└── go.sum
```

### 2.2 各模块职责

#### 2.2.1 config — 配置管理

**职责**：加载和验证系统配置。

**配置结构**：
```yaml
server:
  http_listen: ":8080"
  grpc_listen: ":9090"
  mode: "release"          # debug / release
  tls_cert: ""
  tls_key: ""

database:
  driver: "sqlite"          # sqlite / postgres
  dsn: "data/wecom.db"     # SQLite 文件路径 或 PostgreSQL DSN

wecom:
  corps:
    - name: "main"
      corp_id: "ww1234567890"
      apps:
        - name: "oa"
          agent_id: 1000001
          secret: "encrypted_secret_here"
        - name: "hr"
          agent_id: 1000002
          secret: "encrypted_secret_here"
    - name: "partner"
      corp_id: "ww0987654321"
      apps:
        - name: "external"
          agent_id: 2000001
          secret: "encrypted_secret_here"

auth:
  admin_api_key: "wgk_admin_xxxx"
  key_expiry_days: 365
  rate_limit: 100

ui:
  enabled: true
```

**环境变量覆盖**：`WECOM_CORP_<NAME>_ID`、`WECOM_APP_<NAME>_SECRET` 等支持环境变量覆盖。

**对应需求**：FR-050, FR-051, FR-060

#### 2.2.2 auth — 认证与鉴权

**职责**：API Key 认证和权限校验。同时为 HTTP 中间件和 gRPC 拦截器提供底层实现。

**核心接口**：
```go
// Authenticator 认证接口，HTTP 中间件和 gRPC 拦截器共用
type Authenticator interface {
    Authenticate(ctx context.Context, rawKey string) (*AuthContext, error)
}

// AuthContext 认证后注入上下文的信息
type AuthContext struct {
    KeyID       string
    KeyName     string
    Permissions []string
    CorpName    string
    AppName     string
    IsAdmin     bool
}
```

**对应需求**：FR-001, FR-002

#### 2.2.3 apikey — API Key 管理

**职责**：管理员对 API Key 的生命周期管理。

**接口定义**：
```go
type Store interface {
    Create(ctx context.Context, key *APIKey) (string, error)
    GetByHash(ctx context.Context, hash string) (*APIKey, error)
    List(ctx context.Context, opts ListOptions) ([]*APIKey, string, error)
    Disable(ctx context.Context, id string) error
    Enable(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
}
```

**对应需求**：FR-003

#### 2.2.4 wecom — 企业微信客户端

**职责**：封装企业微信 API 调用，管理 access_token。支持多 CorpID 隔离。

**客户端接口**：
```go
type Client interface {
    // 日程
    CreateSchedule(ctx context.Context, corpName, appName string, params *ScheduleParams) (*Schedule, error)
    GetSchedules(ctx context.Context, corpName, appName string, userID string, opts QueryOptions) ([]*Schedule, string, error)
    UpdateSchedule(ctx context.Context, corpName, appName string, scheduleID string, params *ScheduleParams) error
    DeleteSchedule(ctx context.Context, corpName, appName string, scheduleID string) error

    // 会议室
    ListMeetingRooms(ctx context.Context, corpName, appName string, opts RoomQueryOptions) ([]*MeetingRoom, string, error)
    GetRoomAvailability(ctx context.Context, corpName, appName string, roomID string, start, end time.Time) ([]*TimeSlot, error)
    BookMeetingRoom(ctx context.Context, corpName, appName string, params *BookingParams) (*BookingResult, error)

    // 消息
    SendText(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error)
    SendMarkdown(ctx context.Context, corpName, appName string, params *MessageParams) (*SendResult, error)
    SendImage(ctx context.Context, corpName, appName string, params *ImageMessageParams) (*SendResult, error)
    SendFile(ctx context.Context, corpName, appName string, params *FileMessageParams) (*SendResult, error)
    SendCard(ctx context.Context, corpName, appName string, params *CardMessageParams) (*SendResult, error)
}
```

**access_token 管理**：
- 按 `corpName + appName` 维度独立管理 token
- token 有效期 7200 秒，提前 5 分钟自动刷新
- 使用 `sync.RWMutex` 保护并发读写

**对应需求**：FR-010~FR-013, FR-020~FR-022, FR-030~FR-034, FR-060

#### 2.2.5 grpcserver — gRPC 服务端

**职责**：提供 gRPC 协议接口，复用业务逻辑层。

**Protobuf 服务定义**（`api/proto/wecom_gateway.proto`）：
```protobuf
service WeComGateway {
    // 日程
    rpc CreateSchedule(CreateScheduleRequest) returns (Schedule);
    rpc GetSchedules(GetSchedulesRequest) returns (ScheduleList);
    rpc UpdateSchedule(UpdateScheduleRequest) returns (Schedule);
    rpc DeleteSchedule(DeleteScheduleRequest) returns (DeleteResponse);

    // 会议室
    rpc ListMeetingRooms(ListMeetingRoomsRequest) returns (MeetingRoomList);
    rpc GetRoomAvailability(AvailabilityRequest) returns (TimeSlotList);
    rpc BookMeetingRoom(BookMeetingRoomRequest) returns (BookingResult);

    // 消息
    rpc SendText(TextMessageRequest) returns (SendResult);
    rpc SendMarkdown(MarkdownMessageRequest) returns (SendResult);
    rpc SendImage(ImageMessageRequest) returns (SendResult);
    rpc SendFile(FileMessageRequest) returns (SendResult);
    rpc SendCard(CardMessageRequest) returns (SendResult);

    // 流式接口
    rpc WatchScheduleChanges(WatchRequest) returns (stream ScheduleEvent);
}

service WeComGatewayAdmin {
    // API Key 管理
    rpc CreateAPIKey(CreateAPIKeyRequest) returns (APIKeyInfo);
    rpc ListAPIKeys(ListAPIKeysRequest) returns (APIKeyList);
    rpc DisableAPIKey(DisableAPIKeyRequest) returns (Empty);
    rpc DeleteAPIKey(DeleteAPIKeyRequest) returns (Empty);

    // 审计日志
    rpc QueryAuditLogs(QueryAuditLogsRequest) returns (AuditLogList);
}
```

**拦截器链**：
```
Unary Interceptor Chain:
  Recovery → RateLimit → Auth → Audit → Handler
Stream Interceptor Chain:
  Recovery → RateLimit → Auth → Audit → Handler
```

**对应需求**：FR-070, FR-071

#### 2.2.6 ui — 管理后台

**职责**：提供 Web 管理界面，通过 `go:embed` 嵌入到二进制中。

**实现方式**：
- 前端使用纯 HTML + CSS + Vanilla JS（无构建依赖，保持轻量）
- 通过 `go:embed` 嵌入 `ui/dist/` 目录
- Gin 路由 `GET /admin/*` 提供静态文件服务
- 前端通过 fetch 调用后端管理 API，无需额外 API

**页面结构**：
```
/admin/                → 仪表盘（总览）
/admin/keys            → API Key 管理
/admin/apps            → 企业微信应用管理
/admin/audit           → 审计日志
```

**仪表盘数据**：
- 总请求数 / 成功率 / 错误率（最近 24h）
- 各 API Key 调用量排行
- 请求趋势图（最近 24h，每小时一个数据点）
- 活跃企业微信应用状态

**对应需求**：FR-090

#### 2.2.7 audit — 审计日志

**职责**：记录和查询所有 API 操作日志（HTTP + gRPC 统一格式）。

**实现方式**：
- HTTP：作为 Gin 中间件
- gRPC：作为 unary/server interceptor
- 两者共用同一个 `AuditLogger` 实例

**对应需求**：FR-040, FR-041

#### 2.2.8 OpenAPI 文档

**实现方式**：
- 使用 swaggo/swag，通过 handler 注释自动生成
- Gin 中间件 `swaggerFiles.Handler` 提供 Swagger UI
- 路由：`GET /v1/docs` → Swagger UI，`GET /v1/openapi.json` → OpenAPI spec

**注释示例**：
```go
// CreateSchedule godoc
// @Summary 创建日程
// @Description 为指定员工创建企业微信日程
// @Tags 日程
// @Accept json
// @Produce json
// @Param request body CreateScheduleRequest true "日程参数"
// @Success 201 {object} Schedule
// @Failure 400 {object} httputil.Response
// @Failure 403 {object} httputil.Response
// @Security BearerAuth
// @Router /v1/schedules [post]
```

**对应需求**：FR-080

---

## 3. API 设计

### 3.1 REST 路由总览

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| **系统** | | | |
| GET | `/health` | 健康检查 | 无需认证 |
| GET | `/v1/docs` | Swagger UI | 无需认证 |
| GET | `/v1/openapi.json` | OpenAPI spec | 无需认证 |
| **日程** | | | |
| POST | `/v1/schedules` | 创建日程 | `calendar:write` |
| GET | `/v1/schedules` | 查询日程列表 | `calendar:read` |
| GET | `/v1/schedules/:id` | 查询单个日程 | `calendar:read` |
| PATCH | `/v1/schedules/:id` | 更新日程 | `calendar:write` |
| DELETE | `/v1/schedules/:id` | 取消日程 | `calendar:write` |
| **会议室** | | | |
| GET | `/v1/meeting-rooms` | 查询会议室列表 | `meetingroom:read` |
| GET | `/v1/meeting-rooms/:id/availability` | 查询空闲时段 | `meetingroom:read` |
| POST | `/v1/meeting-rooms/:id/bookings` | 预订会议室 | `meetingroom:write` |
| **消息** | | | |
| POST | `/v1/messages/text` | 发送文本消息 | `message:send` |
| POST | `/v1/messages/markdown` | 发送 Markdown 消息 | `message:send` |
| POST | `/v1/messages/image` | 发送图片消息 | `message:send` |
| POST | `/v1/messages/file` | 发送文件消息 | `message:send` |
| POST | `/v1/messages/card` | 发送卡片消息 | `message:send` |
| **管理** | | | |
| POST | `/v1/admin/api-keys` | 创建 API Key | 管理员 |
| GET | `/v1/admin/api-keys` | 查询 API Key 列表 | 管理员 |
| PATCH | `/v1/admin/api-keys/:id` | 更新 API Key | 管理员 |
| DELETE | `/v1/admin/api-keys/:id` | 删除 API Key | 管理员 |
| GET | `/v1/admin/audit-logs` | 查询审计日志 | 管理员 |
| POST | `/v1/admin/apps` | 添加企业微信应用 | 管理员 |
| GET | `/v1/admin/apps` | 查询应用列表 | 管理员 |
| PATCH | `/v1/admin/apps/:id` | 更新应用配置 | 管理员 |
| **管理后台** | | | |
| GET | `/admin/*` | 管理后台静态文件 | 管理员 |

### 3.2 请求/响应示例

**创建日程**：
```json
// POST /v1/schedules
// Authorization: Bearer wgk_xxxx
{
  "organizer": "zhangsan",
  "summary": "产品评审会",
  "description": "Q2 产品路线图评审",
  "start_time": "2026-04-01T14:00:00+08:00",
  "end_time": "2026-04-01T15:00:00+08:00",
  "attendees": ["lisi", "wangwu"],
  "location": "3F-会议室A",
  "remind_before_minutes": 15
}

// Response 201
{
  "code": 0,
  "message": "ok",
  "data": {
    "schedule_id": "sched_abc123",
    "organizer": "zhangsan",
    "summary": "产品评审会",
    "start_time": "2026-04-01T14:00:00+08:00",
    "end_time": "2026-04-01T15:00:00+08:00",
    "attendees": ["zhangsan", "lisi", "wangwu"],
    "location": "3F-会议室A"
  }
}
```

**发送文本消息**：
```json
// POST /v1/messages/text
{
  "receiver_type": "user",
  "receiver_ids": ["zhangsan", "lisi"],
  "content": "明天下午 3 点开会，请准时参加。"
}

// Response 200
{
  "code": 0,
  "message": "ok",
  "data": {
    "invalid_receiver_ids": [],
    "failed_receiver_ids": []
  }
}
```

**创建 API Key**：
```json
// POST /v1/admin/api-keys
{
  "name": "oa-scheduler-agent",
  "permissions": ["calendar:read", "calendar:write", "meetingroom:read", "meetingroom:write"],
  "corp_name": "main",
  "app_name": "oa",
  "expires_days": 365
}

// Response 201
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "key_abc123",
    "name": "oa-scheduler-agent",
    "api_key": "wgk_a1b2c3d4e5f6...（仅此一次可见）",
    "permissions": ["calendar:read", "calendar:write", "meetingroom:read", "meetingroom:write"],
    "corp_name": "main",
    "app_name": "oa",
    "expires_at": "2027-03-29T00:00:00Z"
  }
}
```

---

## 4. 数据库设计

### 4.1 核心表

```sql
-- API Key 表
CREATE TABLE api_keys (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    key_hash    TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL,           -- JSON array: ["calendar:read","message:send"]
    corp_name   TEXT NOT NULL,           -- 绑定的企业名
    app_name    TEXT,                    -- 绑定的应用名（管理员 Key 为 NULL）
    expires_at  TIMESTAMP,
    disabled    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 企业微信企业表
CREATE TABLE wecom_corps (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    corp_id     TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 企业微信应用表
CREATE TABLE wecom_apps (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    corp_name   TEXT NOT NULL REFERENCES wecom_corps(name),
    agent_id    INTEGER NOT NULL,
    secret_enc  TEXT NOT NULL,
    nonce       TEXT NOT NULL,
    access_token TEXT,
    token_expires_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, corp_name)
);

-- 审计日志表
CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    timestamp   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    protocol    TEXT NOT NULL DEFAULT 'http',  -- 'http' / 'grpc'
    api_key_id  TEXT,
    api_key_name TEXT,
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    query       TEXT,
    body        TEXT,                    -- 脱敏后的请求体
    status_code INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    client_ip   TEXT,
    error_msg   TEXT
);

-- 仪表盘统计快照（每小时聚合一次）
CREATE TABLE stats_hourly (
    id          BIGSERIAL PRIMARY KEY,
    hour        TIMESTAMP NOT NULL,
    total_count BIGINT NOT NULL DEFAULT 0,
    error_count BIGINT NOT NULL DEFAULT 0,
    key_name    TEXT NOT NULL DEFAULT ''
);
```

**索引**：
- `api_keys.key_hash` — UNIQUE
- `api_keys.name` — INDEX
- `wecom_corps.name` — UNIQUE
- `wecom_apps(corp_name, name)` — UNIQUE
- `audit_logs.timestamp` — INDEX
- `audit_logs.api_key_name` — INDEX
- `audit_logs.path` — INDEX
- `stats_hourly(hour, key_name)` — UNIQUE

---

## 5. 关键设计决策

### 5.1 为什么采用 HTTP + gRPC 双协议

**决策**：同时提供 REST HTTP API 和 gRPC API，共用业务逻辑层。

**原因**：
- REST：简单易用，AI 智能体（特别是 Python/JS 生态）接入成本最低
- gRPC：强类型、高性能、支持流式，适合 Go/Java 等编译型语言的智能体
- 共享业务逻辑避免代码重复

**代价**：
- 需要维护 Protobuf 定义与 HTTP handler 的同步
- gRPC 端口需要额外配置和防火墙规则

### 5.2 为什么管理后台使用嵌入式 SPA 而非独立部署

**决策**：管理后台前端通过 `go:embed` 嵌入二进制。

**原因**：
- 单二进制分发，部署零额外步骤
- 无需 nginx 等静态文件服务器
- 轻量级（纯 HTML/CSS/JS，无构建依赖）

### 5.3 为什么多 CorpID 用名称而非 ID 作为标识

**决策**：API Key 和配置中通过 `corp_name`（如 "main"、"partner"）标识企业。

**原因**：
- 人类可读，配置和调试时一目了然
- 避免暴露企业内部 CorpID 给智能体开发者
- 通过名称映射表隔离内外部标识

### 5.4 为什么 API Key 存哈希而非加密

**决策**：API Key 存储 SHA-256 哈希。

**原因**：
- API Key 类似密码，只需验证不需还原
- 哈希比加密更安全（无密钥泄露风险）
- Key 明文仅在创建时返回一次

### 5.5 为什么审计日志覆盖 HTTP 和 gRPC

**决策**：两种协议的调用统一记录到同一张审计表。

**原因**：
- 管理员无需区分协议即可查看所有操作
- 统一的审计查询界面
- `protocol` 字段区分来源

---

## 6. 错误处理策略

| 场景 | 处理方式 | HTTP | gRPC Code |
|------|----------|------|-----------|
| API Key 缺失 | 中间件/拦截器拦截 | 401 | UNAUTHENTICATED |
| API Key 无效 | 中间件/拦截器拦截 | 401 | UNAUTHENTICATED |
| API Key 已禁用 | 中间件/拦截器拦截 | 403 | PERMISSION_DENIED |
| API Key 已过期 | 中间件/拦截器拦截 | 401 | UNAUTHENTICATED |
| 权限不足 | handler 检查 | 403 | PERMISSION_DENIED |
| 参数校验失败 | handler 检查 | 400 | INVALID_ARGUMENT |
| 资源不存在 | service 层 | 404 | NOT_FOUND |
| 会议室时间冲突 | service 层 | 409 | ALREADY_EXISTS |
| 请求过于频繁 | 限流中间件 | 429 | RESOURCE_EXHAUSTED |
| 企业微信 API 错误 | wecom client | 502 | UNAVAILABLE |
| access_token 获取失败 | wecom client | 503 | UNAVAILABLE |
| 内部错误 | 全局 recovery | 500 | INTERNAL |

---

## 7. 部署架构

### 7.1 运行要求

- Go 1.23+（编译后静态二进制）
- SQLite3 开发库（仅 SQLite 模式）
- 可选：PostgreSQL 客户端库
- 外部网络：`api.weixin.qq.com`

### 7.2 端口规划

| 端口 | 协议 | 用途 |
|------|------|------|
| 8080 | HTTP | REST API + Swagger UI + 管理后台 |
| 9090 | gRPC | gRPC API |

### 7.3 启动方式

```bash
# 开发环境（SQLite）
./wecom-gateway -config config.yaml

# 生产环境（PostgreSQL + 环境变量）
WECOM_DB_DRIVER=postgres \
WECOM_DB_DSN="postgres://user:pass@localhost/wecom" \
WECOM_MASTER_KEY="base64-encoded-32-byte-key" \
./wecom-gateway -config /etc/wecom-gateway/config.yaml
```

---

## 8. 测试策略

### 8.1 测试分层

| 层级 | 范围 | 方式 |
|------|------|------|
| 单元测试 | wecom client、crypto、ratelimit、各 service | mock 企业微信 API |
| 集成测试 | handler → service → wecom client 链路 | httptest + mock server |
| gRPC 测试 | gRPC handler → service 链路 | grpc-go test 工具 |
| 端到端测试 | 完整请求流程 | 真实企业微信 API（仅 CI 手动触发） |

### 8.2 Mock 策略

- `wecom.Client` 抽象为接口，业务 service 依赖接口
- gRPC 测试使用 `bufconn`（内存中的 gRPC 连接，无需真实端口）
- HTTP 测试使用 `httptest.NewServer`

### 8.3 覆盖率目标

- 总体覆盖率 ≥ 80%
- auth、apikey、crypto 模块 ≥ 90%
- grpcserver 拦截器 ≥ 85%
- handler 层 ≥ 70%

---

## 9. 未来演进方向

以下为当前版本未覆盖但可能需要的方向，**不作为当前实现依据**：

- **Webhook 事件接收**：接收企业微信回调事件（如日程变更通知）
- **消息模板管理**：预定义消息模板，智能体按模板发送
- **批量操作**：批量创建日程、批量发送消息
- **API 版本管理**：`/v2/` 路径，支持不兼容变更
- **多语言 SDK**：为常用语言提供官方 SDK
