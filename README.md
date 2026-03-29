# WeCom Gateway

企业微信智能体网关 - 为所有 AI 智能体提供统一的企业微信 API 接口

## 项目简介

WeCom Gateway 是一个企业微信智能体网关，作为所有 AI 智能体与企业微信交互的统一入口。它提供简洁的 RESTful API 和 gRPC 接口，支持日程管理、会议室预订、消息发送等核心功能。

### 主要特性

- **统一 API**：提供简洁的 RESTful API，智能体只需调用 HTTP 接口即可完成所有企业微信操作
- **凭证托管**：所有企业微信密钥由网关统一管理，智能体不接触任何密钥
- **鉴权体系**：智能体通过 API Key 认证身份，不同智能体拥有不同权限
- **审计日志**：所有操作完整记录，支持查询和追溯
- **多租户支持**：支持多个企业微信应用，统一管理
- **管理后台**：内置 Web 管理界面，方便管理 API Keys 和查看系统状态

## 技术栈

- **开发语言**：Golang 1.23+
- **HTTP 框架**：Gin
- **数据库**：SQLite（开发）/ PostgreSQL（生产）
- **认证方式**：Bearer Token（API Key）
- **加密方式**：AES-256-GCM

## 快速开始

### 前置要求

- Go 1.23+
- SQLite3

### 1. 克隆项目

```bash
git clone <repository-url>
cd wecom-gateway
```

### 2. 配置

复制示例配置文件：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，配置企业微信信息：

```yaml
wecom:
  corps:
    - name: "main"
      corp_id: "your_corp_id"
      apps:
        - name: "oa"
          agent_id: 1000001
          secret: "your_app_secret"
```

### 3. 运行

```bash
# 开发环境
go run cmd/server/main.go config.yaml

# 或编译后运行
go build -o wecom-gateway cmd/server/main.go
./wecom-gateway config.yaml
```

服务将在 `http://localhost:8080` 启动。

## Docker 部署

### 构建镜像

```bash
docker build -t wecom-gateway:latest .
```

### 运行容器

```bash
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name wecom-gateway \
  wecom-gateway:latest
```

## API 使用示例

### 1. 创建日程

```bash
curl -X POST http://localhost:8080/v1/schedules \
  -H "Authorization: Bearer wgk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "organizer": "zhangsan",
    "summary": "产品评审会",
    "description": "Q2 产品路线图评审",
    "start_time": "2026-04-01T14:00:00+08:00",
    "end_time": "2026-04-01T15:00:00+08:00",
    "attendees": ["lisi", "wangwu"],
    "location": "3F-会议室A"
  }'
```

### 2. 查询会议室

```bash
curl -X GET "http://localhost:8080/v1/meeting-rooms?city=北京&capacity=10" \
  -H "Authorization: Bearer wgk_your_api_key"
```

### 3. 发送消息

```bash
curl -X POST http://localhost:8080/v1/messages/text \
  -H "Authorization: Bearer wgk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "receiver_type": "user",
    "receiver_ids": ["zhangsan", "lisi"],
    "content": "明天下午 3 点开会，请准时参加。"
  }'
```

## 管理后台

访问 `http://localhost:8080/admin/` 可以使用管理后台界面：

- **仪表盘**：查看系统概览和统计信息
- **API Keys**：创建、查看、禁用、删除 API Keys
- **审计日志**：查询所有 API 操作记录

默认情况下，需要管理员权限才能访问管理后台。

## 配置说明

### 环境变量

可以通过环境变量覆盖配置文件中的值：

```bash
# 服务器配置
export WECOM_HTTP_LISTEN=:8080
export WECOM_SERVER_MODE=release

# 数据库配置
export WECOM_DB_DRIVER=sqlite
export WECOM_DB_DSN=data/wecom.db

# 企业微信配置
export WECOM_CORP_MAIN_ID=your_corp_id
export WECOM_APP_MAIN_OA_SECRET=your_app_secret

# 认证配置
export WECOM_ADMIN_API_KEY=wgk_admin_xxxxx
export WECOM_RATE_LIMIT=100
```

### API Key 权限

支持的权限类型：

- `calendar:read` - 读取日程
- `calendar:write` - 创建/修改/删除日程
- `meetingroom:read` - 读取会议室信息
- `meetingroom:write` - 预订会议室
- `message:send` - 发送消息
- `*` - 管理员权限（所有权限）

## 项目结构

```
wecom-gateway/
├── cmd/server/          # 应用入口
├── internal/            # 内部包
│   ├── config/         # 配置管理
│   ├── auth/           # 认证中间件
│   ├── apikey/         # API Key 管理
│   ├── wecom/          # 企业微信客户端
│   ├── schedule/       # 日程服务
│   ├── meeting/        # 会议室服务
│   ├── message/        # 消息服务
│   ├── admin/          # 管理服务
│   ├── audit/          # 审计日志
│   ├── ratelimit/      # 限流器
│   ├── store/          # 数据库抽象层
│   ├── crypto/         # 加密工具
│   ├── httputil/       # HTTP 工具
│   └── ui/             # UI 资源嵌入
├── api/proto/          # gRPC Protobuf 定义
├── ui/                 # 管理后台前端
├── docs/               # 文档
└── config.example.yaml # 示例配置
```

## 开发指南

### 运行测试

```bash
go test ./...
```

### 生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 代码检查

```bash
go vet ./...
go fmt ./...
```

## 生产部署建议

1. **数据库**：生产环境建议使用 PostgreSQL 而非 SQLite
2. **密钥管理**：使用环境变量或密钥管理系统存储敏感信息
3. **HTTPS**：生产环境务必启用 HTTPS
4. **监控**：配置日志收集和监控告警
5. **备份**：定期备份数据库和配置文件

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查数据库文件路径是否正确
   - 确保目录存在且有写入权限

2. **企业微信 API 调用失败**
   - 验证 CorpID、AgentID、Secret 是否正确
   - 检查网络连接和企业微信服务状态
   - 查看审计日志获取详细错误信息

3. **API Key 认证失败**
   - 确认 API Key 格式正确（以 `wgk_` 开头）
   - 检查 API Key 是否已禁用或过期

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 GitHub Issue
- 发送邮件至项目维护者

---

**注意**：本项目为生产级别实现，已包含完整的错误处理、日志记录和安全性考虑。请确保在使用前仔细阅读配置说明并根据实际环境进行调整。
