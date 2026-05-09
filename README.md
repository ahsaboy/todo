# TODO 任务管理系统

一个轻量级的多用户 TODO 任务管理服务，基于 Go + Gin + SQLite，支持用户注册/登录、个人 API Key、多渠道提醒推送。

## 功能特性

- 用户注册/登录，个人 API Key 管理
- 任务按用户隔离，每人只能看到自己的任务
- 任务 CRUD（创建、查询、更新、删除）
- 创建任务前必须先配置至少一个已启用的提醒渠道
- 分页、排序、筛选、关键字搜索
- 任务优先级（高/中/低）
- 截止时间与提醒时间设定
- 后台定时扫描，按用户多渠道推送提醒（飞书/钉钉/企业微信等）
- 重复任务（daily/weekly/monthly/yearly），完成时自动生成下一次
- 支持 `Authorization: Bearer` 和 `api-key` 两种认证方式
- 健康检查端点（Docker 健康探测）
- Swagger API 文档
- 优雅退出
- Docker Compose 部署

## 快速开始

### 本地运行

```bash
# 编译
make build

# 运行
make run

# 或直接运行
make dev
```

### Docker 部署

```bash
docker compose up --build -d
```

### CLI 参数

```
todo-server [选项]

选项:
  -c, --config <path>  配置文件路径 (默认: config.yaml)
  -p, --port <port>    覆盖服务端口号
  --host <addr>        覆盖监听地址
  --mode <mode>        覆盖运行模式 (debug/release)
  -v, --version        显示版本号
  -h, --help           显示帮助信息
```

## API 接口

服务启动后访问 Swagger 文档：`http://localhost:8080/docs/index.html`

### 认证方式

支持两种请求头（任选其一）：

```bash
# 方式一：Bearer Token（推荐）
curl -H "Authorization: Bearer <api_key>" http://localhost:8080/api/v1/tasks

# 方式二：自定义 Header
curl -H "api-key: <api_key>" http://localhost:8080/api/v1/tasks
```

### 端点一览

#### 公开端点（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| POST | `/api/v1/auth/register` | 用户注册（返回 API Key） |
| POST | `/api/v1/auth/login` | 用户登录（返回 API Key） |
| GET | `/api/v1/templates` | 查看预置提醒模板列表（仅供创建用户自己的渠道配置时参考） |

#### 需认证端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/tasks` | 创建任务（要求至少一个已启用提醒渠道） |
| GET | `/api/v1/tasks` | 获取任务列表 |
| GET | `/api/v1/tasks/:id` | 获取单个任务 |
| PUT | `/api/v1/tasks/:id` | 更新任务 |
| DELETE | `/api/v1/tasks/:id` | 删除任务 |
| PATCH | `/api/v1/tasks/:id/complete` | 切换完成状态 |
| GET | `/api/v1/user/profile` | 获取用户信息 |
| PUT | `/api/v1/user/profile` | 更新用户信息 |
| PUT | `/api/v1/user/password` | 修改密码 |
| GET | `/api/v1/user/keys` | 列出所有 API Key |
| POST | `/api/v1/user/keys` | 生成新 API Key |
| DELETE | `/api/v1/user/keys/:id` | 撤销 API Key |
| GET | `/api/v1/user/reminder-configs` | 列出提醒配置 |
| POST | `/api/v1/user/reminder-configs` | 创建提醒配置 |
| GET | `/api/v1/user/reminder-configs/:id` | 获取单个提醒配置 |
| PUT | `/api/v1/user/reminder-configs/:id` | 更新提醒配置 |
| DELETE | `/api/v1/user/reminder-configs/:id` | 删除提醒配置 |

### 示例

**注册用户：**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "password": "123456",
    "email": "demo@example.com"
  }'
# 返回 user 信息和 api_key（明文，仅此一次显示）
```

**登录：**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "demo", "password": "123456"}'
```

**先配置提醒渠道：**

```bash
curl -X POST http://localhost:8080/api/v1/user/reminder-configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <api_key>" \
  -d '{
    "name": "feishu",
    "channel_type": "feishu",
    "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    "webhook_body_template": "{\"msg_type\":\"text\",\"content\":{\"text\":\"[TODO] {{.Title}}\\n优先级: {{.PriorityText}}\\n截止: {{.DueAt}}\"}}"
  }'
```

如果当前用户没有任何已启用渠道，`POST /api/v1/tasks` 会返回 `400 INVALID_INPUT`。

**再创建任务：**

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <api_key>" \
  -d '{
    "title": "完成项目报告",
    "priority": 1,
    "due_at": "2026-05-20 14:00:00",
    "remind_at": "2026-05-19 09:00:00",
    "repeat_type": "weekly"
  }'
```

### 列表查询参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `page` | 页码（默认 1） | `?page=2` |
| `limit` | 每页数量（默认 20，最大 100） | `?limit=50` |
| `sort` | 排序字段 | `?sort=due_at` |
| `order` | 排序方向 | `?order=asc` |
| `status` | 筛选状态 | `?status=pending` |
| `priority` | 筛选优先级 | `?priority=1` |
| `due_before` | 截止时间上限 | `?due_before=2026-05-31` |
| `due_after` | 截止时间下限 | `?due_after=2026-05-01` |
| `search` | 关键字搜索 | `?search=报告` |

## 配置文件

`config.yaml` 包含所有可配置项：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  path: "./data/tasks.db"

reminder:
  enabled: true
  scan_interval_seconds: 30
  webhook_body_template: |        # 用户渠道未自定义模板时使用的默认消息模板
    {"msg_type":"text","content":{"text":"[TODO] {{.Title}}"}}
  webhook_timeout_seconds: 10
  max_retries: 3
  retry_delay_seconds: 5
  default_templates:
    feishu:
      channel_type: "webhook"
      webhook_method: "POST"
      webhook_headers:
        Content-Type: "application/json"
      webhook_body_template: |
        {"msg_type":"text","content":{"text":"[TODO] {{.Title}}\n优先级: {{.PriorityText}}\n截止: {{.DueAt}}"}}

cors:
  enabled: true
  allowed_origins: ["*"]

logging:
  level: "info"
  format: "json"
```

说明：

- 系统不再支持全局 webhook 发送地址。
- 用户必须通过 `/api/v1/user/reminder-configs` 创建并启用自己的通知渠道。
- `default_templates` 只提供模板参考，不会直接作为发送目标。

### Webhook 模板变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{.TaskID}}` | 任务 ID | `42` |
| `{{.Title}}` | 任务标题 | `"完成报告"` |
| `{{.Description}}` | 任务描述 | `"Q2 报告"` |
| `{{.Priority}}` | 优先级数字 | `1` |
| `{{.PriorityText}}` | 优先级文字 | `"高"` |
| `{{.DueAt}}` | 截止时间 | `"2026-05-20 14:00"` |
| `{{.RemindAt}}` | 提醒时间 | `"2026-05-19 09:00"` |
| `{{.RepeatType}}` | 重复类型 | `"weekly"` |
| `{{.CreatedAt}}` | 创建时间 | `"2026-05-09 10:00"` |

## 业务约束

- 用户必须先创建并启用至少一个提醒渠道，才能创建任务。
- 禁用某个用户的全部提醒渠道后，该用户任务不会回退到任何全局 webhook。
- 提醒只有在该任务的所有已启用渠道都发送成功后，才会标记为已发送。
- 注册接口在并发下如果用户名冲突，会稳定返回 `409 Conflict`。

## 项目结构

```
TODO/
├── cmd/server/main.go              # 入口：配置加载、路由注册、优雅退出
├── internal/
│   ├── config/config.go            # YAML 配置加载
│   ├── database/database.go        # SQLite 连接 + WAL 模式 + 自动建表
│   ├── models/
│   │   ├── task.go                 # 任务数据模型
│   │   ├── user.go                 # 用户数据模型
│   │   ├── api_key.go              # API Key 数据模型
│   │   └── reminder_config.go      # 提醒配置数据模型
│   ├── handlers/
│   │   ├── auth_handler.go         # 认证 HTTP 处理（注册/登录/Key管理）
│   │   ├── task_handler.go         # 任务 HTTP 处理
│   │   └── reminder_config_handler.go  # 提醒配置 CRUD
│   ├── repository/
│   │   ├── user_repo.go            # 用户数据库操作
│   │   ├── api_key_repo.go         # API Key 数据库操作
│   │   ├── task_repo.go            # 任务数据库操作
│   │   └── reminder_config_repo.go # 提醒配置数据库操作
│   ├── service/
│   │   ├── auth_service.go         # 认证逻辑
│   │   ├── task_service.go         # 业务逻辑
│   │   ├── reminder_service.go     # 后台提醒（按用户多渠道）
│   │   └── reminder_config_service.go  # 提醒配置管理
│   ├── middleware/auth.go          # 认证中间件（Bearer/api-key/X-API-Key）
│   └── utils/
│       ├── response.go             # 统一响应格式
│       └── validator.go            # 参数校验
├── docs/                           # Swagger 文档（自动生成）
├── config.yaml                     # 配置文件
├── Dockerfile                      # 多阶段构建
├── docker-compose.yml              # 容器编排
└── Makefile                        # 构建命令
```

## Makefile 命令

```bash
make build          # 编译（自动生成 Swagger 文档）
make run            # 编译并运行
make test           # 运行测试
make dev            # 本地开发（go run，自动生成 Swagger 文档）
make swag           # 仅重新生成 Swagger 文档
make clean          # 清理构建产物和数据库文件
make docker-build   # 构建 Docker 镜像
make docker-up      # Docker Compose 启动
make docker-down    # Docker Compose 停止
make docker-logs    # 查看 Docker 日志
```

## 技术栈

- **语言**: Go
- **Web 框架**: Gin
- **数据库**: SQLite (CGO)
- **日志**: zap
- **配置**: YAML
- **API 文档**: Swagger (swaggo)
