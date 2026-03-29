# wecom-gateway 管理后台重构设计

## 问题分析

### 当前问题
1. **管理功能缺失**：缺少企业微信企业信息管理（CorpID/Secret/应用配置）
2. **认证体系混淆**：人类管理员使用 API Key 登录，应改为账号密码
3. **API Key 定位错误**：API Key 应仅用于 AI 智能体调用，不应作为管理登录凭证

### 角色分离
| 角色 | 认证方式 | 用途 |
|------|----------|------|
| 人类管理员 | 账号密码 + Session | 登录管理后台 |
| AI 智能体 | API Key（Bearer Token） | 调用业务 API |

## 重新设计

### 1. 管理员认证系统

#### 数据库表：admin_users
```sql
CREATE TABLE admin_users (
    id          TEXT PRIMARY KEY,
    username    TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,        -- bcrypt hash
    display_name TEXT,
    is_active   INTEGER DEFAULT 1,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);
```

#### 登录流程
1. POST `/v1/admin/login` → 验证用户名密码 → 返回 JWT Token
2. 前端存储 JWT Token（localStorage 或 Cookie）
3. 后续管理 API 请求携带 JWT Token
4. 支持修改密码、退出登录

#### 默认管理员
- 用户名：`admin`
- 密码：`admin123`（首次登录提示修改）

### 2. 企业管理功能

#### 数据库表：wecom_corps（已存在，需补充字段）
已有字段：id, name, corp_id, created_at, updated_at

#### 数据库表：wecom_apps（已存在，需补充字段）
已有字段：id, name, corp_name, agent_id, secret_enc, nonce, created_at, updated_at

#### 新增 API 端点

**企业管理：**
- `GET    /v1/admin/corps`          — 企业列表
- `POST   /v1/admin/corps`          — 添加企业（CorpID、名称）
- `PUT    /v1/admin/corps/:id`      — 修改企业信息
- `DELETE /v1/admin/corps/:id`      — 删除企业
- `GET    /v1/admin/corps/:id`      — 企业详情

**应用管理：**
- `GET    /v1/admin/corps/:corp_name/apps`      — 应用列表
- `POST   /v1/admin/corps/:corp_name/apps`      — 添加应用（AgentID、Secret）
- `PUT    /v1/admin/corps/:corp_name/apps/:id`  — 修改应用
- `DELETE /v1/admin/corps/:corp_name/apps/:id`  — 删除应用

### 3. 前端页面

#### 登录页 `/ui/login.html`
- 用户名 + 密码输入
- 登录按钮
- JWT Token 存储到 localStorage

#### 管理后台（已有，需扩展）

**导航栏：**
- 仪表盘（Dashboard）
- 企业管理（新增）
- 应用管理（新增）
- API Keys（已有，标注为"智能体调用凭证"）
- 审计日志（已有）
- 修改密码（新增）

**企业管理页面：**
- 企业列表表格（名称、CorpID、应用数量、操作）
- 添加/编辑企业弹窗（名称、CorpID）
- 删除确认

**应用管理页面：**
- 按企业分组的列表
- 添加/编辑应用弹窗（名称、AgentID、Secret）
- Secret 显示为星号，支持显示/隐藏切换

### 4. JWT Token 设计

```
Header:  {"alg": "HS256", "typ": "JWT"}
Payload: {"user_id": "xxx", "username": "admin", "exp": 1700000000}
```

- 有效期：24 小时
- 签名密钥：使用 config.yaml 中的 `auth.jwt_secret`
- 如果未配置 jwt_secret，自动生成并保存

### 5. 中间件调整

```
/v1/admin/login         → 无需认证
/v1/admin/health        → 无需认证
/v1/admin/initialize   → 无需认证（首次初始化）
/v1/admin/*             → JWT 认证（人类管理员）
/v1/schedules/*         → API Key 认证（AI 智能体）
/v1/meeting-rooms/*     → API Key 认证（AI 智能体）
/v1/messages/*          → API Key 认证（AI 智能体）
```

## 实现计划

### Phase 1: 后端
1. 添加 admin_users 表和 CRUD
2. 实现 JWT 认证（生成/验证/中间件）
3. 实现管理员登录 API
4. 实现企业管理 CRUD API
5. 实现应用管理 CRUD API
6. 调整中间件：管理接口用 JWT，业务接口用 API Key

### Phase 2: 前端
1. 创建登录页面
2. 修改 app.js：登录流程、JWT 存储
3. 添加企业管理页面
4. 添加应用管理页面
5. 添加修改密码功能
6. API Keys 页面标注为"智能体调用凭证"
