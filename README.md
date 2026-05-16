# TODO 任务管理系统

[![Version](https://img.shields.io/github/v/release/ahsaboy/todo?color=blue&label=version)](https://github.com/ahsaboy/todo/releases)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](https://github.com/ahsaboy/todo/releases)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-orange.svg)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/ahsaboy/todo/total)](https://github.com/ahsaboy/todo/releases/latest)

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
# 编译（自动生成 Swagger 文档）
make build

# 运行
make run

# 或直接运行（go run，不生成二进制）
make dev
```

### 交叉编译

编译输出文件名自动带 OS-ARCH 后缀（如 `server-linux-amd64`）。

```bash
make build           # 编译当前平台
make build-linux     # 交叉编译 Linux (amd64)
make build-windows   # 交叉编译 Windows (amd64)
make build-darwin    # 交叉编译 macOS (arm64)
```

`make build` 会先构建 `frontend/dist`，复制到 `web/dist`，再由 Go `embed` 打包进最终二进制。前端构建产物里可能包含 `_` 开头的文件，因此嵌入规则使用 `//go:embed all:dist`。

> 项目使用纯 Go SQLite 驱动（modernc.org/sqlite），无需 CGO，所有平台均可直接交叉编译。

### UPX 压缩

编译时会自动检测系统是否安装了 [UPX](https://upx.github.io/)。如果已安装，会自动对二进制文件进行压缩（约减少 60% 体积）。

```bash
# 安装 UPX（可选）
# Windows: choco install upx
# macOS:   brew install upx
# Linux:   apt install upx   或  yum install upx
```

### Docker 部署

单体部署仍使用当前 `Dockerfile`：Docker 会先构建 Vue 前端，复制到 `web/dist`，再由 Go 后端 embed 到同一个镜像中。

```bash
# 本地二进制单体构建
make build

# 单体 Docker 镜像
docker build -t todo-app:latest .

# 单体 Compose 启动
docker compose up --build -d
```

### 前后端分离部署

后端纯 API 模式使用 `separate_frontend` 构建标签，不再注册前端静态资源路由：

```bash
# 本地构建纯 API 后端二进制
make build-backend

# 构建后端 Docker 镜像
docker build -f Dockerfile.backend -t todo-backend:latest .
```

前端独立构建时通过 `API_BASE_URL` 指定浏览器访问后端 API 的地址。该值会在 Vite 构建阶段写入前端产物，对应环境变量为 `VITE_API_BASE_URL`。

```bash
# 本地构建前端 dist，不复制到 web/dist
make frontend-build-standalone API_BASE_URL=http://localhost:8080/api/v1

# 构建 nginx 静态前端镜像
docker build -f frontend/Dockerfile \
  --build-arg API_BASE_URL=http://localhost:8080/api/v1 \
  -t todo-frontend:latest \
  ./frontend
```

也可以直接使用分离部署 Compose 示例：

```bash
docker compose -f docker-compose.separated.yml up --build -d
```

默认示例中后端监听 `http://localhost:8080`，前端监听 `http://localhost:3000`。如果生产环境前端域名不是同源地址，需要在后端 `config.yaml` 中允许前端来源：

```yaml
cors:
  enabled: true
  allowed_origins:
    - "https://todo.example.com"
```

### CLI 参数

```
todo-server [选项]

选项:
  -c, --config <path>  配置文件路径 (默认: config.yaml)
  -p, --port <port>    覆盖服务端口号
  --host <addr>        覆盖监听地址
  --mode <mode>        覆盖运行模式 (debug/release)
  --log-path <path>    覆盖日志存储路径
  --log-max-days <n>   覆盖日志保留天数
  --backend-log <mode> 覆盖后端日志输出模式 (console/file/both/off)
  --frontend-log <mode> 覆盖前端日志输出模式 (console/file/both/off)
  --frontend-log-level <level>  覆盖前端日志级别
  -v, --version        显示版本号
  -h, --help           显示帮助信息
```

## API 接口

> 服务启动后访问 Swagger 文档：`http://localhost:8080/docs/index.html`
>
> `GET /api/v1/runtime-config` 和 `POST /api/v1/logs/frontend` 属于内部运行时接口，供前端日志初始化和上报使用，不作为公开业务 API。


## MCP 服务器

除 REST API 外，本服务同时暴露一个基于 [Model Context Protocol](https://modelcontextprotocol.io/) 的端点，可被 LLM 客户端直接调用。

- **端点**：`POST /mcp`（同时支持 `GET` 用于 SSE 事件流，`DELETE` 用于关闭 session）
- **传输**：Streamable HTTP（MCP 协议版本 `2025-11-25`，由 [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) v0.54.0 实现）
- **认证**：请求头 `api-key: <key>`（也接受 `Authorization: Bearer <key>` 或 legacy `X-API-Key: <key>`）—— 与 REST API 共享同一份 `user_api_keys` 表，按 `user_id` 隔离
- **会话**：`initialize` 应答头会返回 `Mcp-Session-Id`，后续请求需通过 `Mcp-Session-Id` 请求头携带

### 工具列表（共 12 个）

任务管理（6 个，复用 `TaskService`）：

| 工具 | 说明 |
| --- | --- |
| `create_task` | 创建任务，支持标题/描述/优先级/截止/提醒/重复规则 |
| `list_tasks` | 分页查询当前用户任务，支持筛选与排序 |
| `get_task` | 按 ID 获取任务详情 |
| `update_task` | 局部更新任务字段（指针化的可选字段） |
| `delete_task` | 按 ID 删除任务 |
| `toggle_task_complete` | 切换任务完成状态（自动生成下一个重复任务） |

提醒配置（5 个，复用 `ReminderConfigService`）：

| 工具 | 说明 |
| --- | --- |
| `list_reminder_configs` | 列出当前用户的所有提醒渠道 |
| `create_reminder_config` | 新建提醒渠道（webhook/feishu/dingtalk/wecom/slack） |
| `get_reminder_config` | 按 ID 获取提醒渠道详情 |
| `update_reminder_config` | 局部更新提醒渠道字段 |
| `delete_reminder_config` | 按 ID 删除提醒渠道 |

用户信息（1 个，复用 `AuthService`）：

| 工具 | 说明 |
| --- | --- |
| `get_user_profile` | 获取当前认证用户的个人资料（不含密码哈希） |

### 调用示例（curl）

`initialize` —— 拿到 `Mcp-Session-Id`：

```bash
curl -i -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -d '{
    "jsonrpc":"2.0",
    "id":1,
    "method":"initialize",
    "params":{
      "protocolVersion":"2025-11-25",
      "capabilities":{},
      "clientInfo":{"name":"demo","version":"0.0.1"}
    }
  }'
```

完成 `initialize` 后必须先发一个 `notifications/initialized` 通知（MCP 协议要求），再发 `tools/list`：

```bash
curl -i -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -H 'Mcp-Session-Id: <SESSION_ID_FROM_INITIALIZE>' \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -H 'Mcp-Session-Id: <SESSION_ID_FROM_INITIALIZE>' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
```

无 `api-key` 直接访问会返回 `401`，错误结构与 REST API 一致（`success:false, code:"UNAUTHORIZED"`）。


## 配置文件

`config.yaml` 包含所有可配置项：

```
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  path: "./data/tasks.db"

reminder:
  enabled: true
  scan_interval_seconds: 30
  webhook_body_template: | # 用户渠道未自定义模板时使用的默认消息模板
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
  path: "./logs"
  max_days: 7
  backend:
    console_enabled: true
    file_enabled: false
  frontend:
    console_enabled: false
    file_enabled: false
    level: "warn"
```

说明：

- 用户必须通过 `/api/v1/user/reminder-configs` 创建并启用自己的通知渠道。
- `default_templates` 只提供模板参考，不会直接作为发送目标。

### 日志配置

- 后端日志按天写入 `backend-YYYY-MM-DD.log`，文件位置由 `logging.path` 决定。
- 前端日志先通过 `POST /api/v1/logs/frontend` 上报，再按天写入 `frontend-YYYY-MM-DD.log`。
- `logging.max_days` 控制日志保留天数，启动时会清理早于保留窗口的历史日志文件。
- `GET /api/v1/runtime-config` 只下发前端所需的日志开关和级别，属于内部运行时接口。
- 前端日志只记录必要字段，不包含认证头、密码、请求体或完整响应体。

### Webhook 模板变量

| 变量                | 说明       | 示例                 |
| ------------------- | ---------- | -------------------- |
| `{{.TaskID}}`       | 任务 ID    | `42`                 |
| `{{.Title}}`        | 任务标题   | `"完成报告"`         |
| `{{.Description}}`  | 任务描述   | `"Q2 报告"`          |
| `{{.Priority}}`     | 优先级数字 | `1`                  |
| `{{.PriorityText}}` | 优先级文字 | `"高"`               |
| `{{.DueAt}}`        | 截止时间   | `"2026-05-20T06:00:00Z"` |
| `{{.RemindAt}}`     | 提醒时间   | `"2026-05-19T01:00:00Z"` |
| `{{.RepeatType}}`   | 重复类型   | `"weekly"`               |
| `{{.CreatedAt}}`    | 创建时间   | `"2026-05-09T02:00:00Z"` |

## 时间契约

所有时间字段统一使用 **UTC RFC3339** 格式存储和传输。

- **API 入参/出参**：`2026-05-10T10:30:00Z`
- **数据库存储**：SQLite `TEXT`，内容为 UTC RFC3339
- **前端展示**：根据浏览器本地时区格式化
- **涉及字段**：`due_at`、`remind_at`、`repeat_end_date`、`created_at`、`updated_at`、`reminder_sent_at`、`last_used_at`

> 数据库中可能存在旧格式（`YYYY-MM-DD HH:MM:SS`）的历史数据，新代码路径兼容读取旧格式，但新写入的数据统一为 UTC RFC3339。
> 历史数据迁移可运行：`go run scripts/migrate_time_format.go -db data/tasks.db -dry-run`

## 业务约束

- 创建任务时如果设置了提醒时间（`remind_at`），则必须存在至少一个已启用的提醒渠道。
- 未设置 `remind_at` 的任务不需要提醒渠道，可直接创建。
- 禁用某个用户的全部提醒渠道后，该用户任务不会回退到任何全局 webhook。
- 提醒只有在该任务的所有已启用渠道都发送成功后，才会标记为已发送。
- 注册接口在并发下如果用户名冲突，会稳定返回 `409 Conflict`。

## 项目结构

```
TODO/
├── cmd/server/main.go              # 入口：配置加载、路由注册、优雅退出
├── internal/
│   ├── config/config.go            # YAML 配置加载
│   ├── logging/                    # 日志初始化、日志路径和保留清理
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
├── frontend/                       # Vue 前端源码
│   ├── Dockerfile                  # 分离部署前端 nginx 镜像
│   ├── nginx.conf                  # 前端 SPA fallback 配置
│   └── src/shared/logger/          # 前端日志封装和上报
├── web/                            # Go embed 前端构建产物入口
│   ├── embed.go                    # 使用 //go:embed all:dist
│   └── dist/                       # make build 生成并复制的静态文件
├── config.yaml                     # 配置文件
├── Dockerfile                      # 单体镜像多阶段构建
├── Dockerfile.backend              # 分离部署纯 API 后端镜像
├── docker-compose.yml              # 单体容器编排
├── docker-compose.separated.yml    # 前后端分离容器编排示例
└── Makefile                        # 构建命令
```

## Makefile 命令

```bash
make build          # 编译当前平台（自动生成 Swagger 文档，支持 UPX 压缩）
make build-linux    # 交叉编译 Linux amd64
make build-windows  # 交叉编译 Windows amd64
make build-darwin   # 交叉编译 macOS arm64
make build-backend  # 编译当前平台纯 API 后端（-tags separate_frontend）
make build-separated API_BASE_URL=http://localhost:8080/api/v1  # 纯 API 后端 + 独立前端 dist
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
- **数据库**: SQLite（纯 Go 驱动，无需 CGO）
- **日志**: zap
- **配置**: YAML
- **API 文档**: Swagger (swaggo)
