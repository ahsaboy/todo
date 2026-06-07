# TODO 任务管理系统

[![Version](https://img.shields.io/github/v/release/ahsaboy/todo?color=blue&label=version)](https://github.com/ahsaboy/todo/releases)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)](https://github.com/ahsaboy/todo/releases)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-orange.svg)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/ahsaboy/todo/total)](https://github.com/ahsaboy/todo/releases/latest)

一个轻量级的多用户 TODO 任务管理服务，基于 Go + Gin + SQLite，支持用户注册/登录、个人 API Key、多渠道提醒推送。

## 界面预览

![登录页](img/login.png)

![主页面](img/main_page.png)

## 功能特性

- 用户注册/登录，个人 API Key 管理（支持 `Authorization: Bearer`、`api-key`、`X-API-Key` 三种请求头）
- OAuth 社交登录（GitHub / Google / LinuxDo），支持账号绑定与解绑
- 任务按用户隔离，每人只能看到自己的任务
- 任务 CRUD、分页、排序、筛选、关键字搜索，支持优先级（高/中/低）与截止/提醒时间
- 标签管理（颜色/图标/排序）、看板视图、专注计时
- 后台定时扫描，按用户多渠道推送提醒（飞书/钉钉/企业微信/Gotify 等 webhook）
- 重复任务（daily/weekly/monthly/yearly），完成时自动生成下一次
- **MCP 服务器**：暴露 `/mcp` 端点，可被 LLM 客户端直接调用（12 个工具）
- **管理后台**（localhost-only）：仪表盘、用户/任务/提醒管理、系统配置、操作日志
- **国际化**（i18n）：中英文错误消息本地化，通过 Accept-Language 头自动切换
- 健康检查端点、Swagger API 文档、优雅退出、Docker Compose 部署

## 快速开始

### 配置文件

```bash
cp config.example.yaml config.yaml
# 按需编辑 config.yaml（端口、时区、管理员账号、提醒模板等）
```

> `config.yaml` 包含敏感信息（如管理员密码），已被添加到 `.gitignore`，不会被 Git 跟踪。

### 本地运行

```bash
make build   # 编译（自动生成 Swagger 文档）
make run     # 编译并运行
make dev     # 直接运行（go run，不生成二进制）
```

### Docker 部署

```bash
make docker-build              # 本地构建 Docker 镜像
docker compose up -d           # 使用已发布的 GHCR 镜像启动
```

发布镜像：`ghcr.io/ahsaboy/todo:latest` / `ghcr.io/ahsaboy/todo:vX.Y.Z`

## API 接口

> 服务启动后访问 Swagger 文档：`http://localhost:8080/docs/index.html`

## 技术栈

- **后端**: Go + Gin + SQLite（纯 Go 驱动，无需 CGO）+ zap 日志
- **前端**: Vue 3 + TypeScript + Pinia + Vite
- **API 文档**: Swagger (swaggo)
- **协议**: REST API + MCP (Model Context Protocol)

---

<details>
<summary><h3 style="display:inline">MCP 服务器</h3></summary>

除 REST API 外，本服务同时暴露一个基于 [Model Context Protocol](https://modelcontextprotocol.io/) 的端点，可被 LLM 客户端直接调用。

- **端点**：`POST /mcp`（同时支持 `GET` 用于 SSE 事件流，`DELETE` 用于关闭 session）
- **传输**：Streamable HTTP（MCP 协议版本 `2025-11-25`，由 [`github.com/mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) v0.54.0 实现）
- **认证**：请求头 `api-key: <key>`（也接受 `Authorization: Bearer <key>` 或 legacy `X-API-Key: <key>`）—— 与 REST API 共享同一份 `user_api_keys` 表，按 `user_id` 隔离
- **会话**：`initialize` 应答头会返回 `Mcp-Session-Id`，后续请求需通过 `Mcp-Session-Id` 请求头携带

### 客户端选项 Header

下面三个可选请求头按需启用(**任意非空字符串都视为开启**,例如 `1` / `true` / `on`;缺失或空字符串为关闭):

| Header | 默认 | 启用后 |
| --- | --- | --- |
| `X-MCP-Include-Reminders` | `tools/list` 隐藏 5 个 `*_reminder_config` 工具;`tools/call` 调用它们时返回 `tool not available` 错误。任务工具与 `get_user_profile` 不受影响。 | 5 个 reminder 工具正常列出与调用。 |
| `X-MCP-Structured-Output` | `tools/call` 的结果把完整 JSON 序列化后塞进 `content[0].text`,**不**返回 `structuredContent`,方便纯文本客户端。 | `content` 仅放简短摘要,`structuredContent` 放完整结构化对象(沿用 mcp-go 的 `NewToolResultStructured`)。 |
| `X-MCP-Timezone` | 工具结果中的时间字段按 `server.timezone` 配置输出。 | 改用 header 指定的时区(IANA 名 `Asia/Shanghai`、固定偏移 `+08:00`、`UTC`、`Local` 均可);非法值静默回退到全局配置。 |

每次请求都按当时的 header 重新判定，无需重新 `initialize`。示例：

```bash
# 让 tools/list 暴露 reminder_config 工具
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -H 'Mcp-Session-Id: <SESSION_ID>' \
  -H 'X-MCP-Include-Reminders: 1' \
  -d '{"jsonrpc":"2.0","id":99,"method":"tools/list"}'

# 让 list_tasks 的结果走 structuredContent
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -H 'Mcp-Session-Id: <SESSION_ID>' \
  -H 'X-MCP-Structured-Output: 1' \
  -d '{"jsonrpc":"2.0","id":100,"method":"tools/call","params":{"name":"list_tasks","arguments":{"limit":5}}}'
```

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
</details>

<details>
<summary><h3 style="display:inline">配置文件详解</h3></summary>

`config.yaml` 包含所有可配置项。首先复制示例配置：

```bash
cp config.example.yaml config.yaml
```

然后根据需要修改。核心配置项：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"
  timezone: "Asia/Shanghai"

database:
  path: "./data/tasks.db"

static_files: true

reminder:
  enabled: true
  scan_interval_seconds: 30
  webhook_body_template: |
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
  file_enabled: true
  path: "./logs"
  max_days: 7

# 管理后台配置（仅限 localhost 访问）
admin:
  enabled: true
  username: admin    # 管理员用户名（首次启动时创建）
  password: admin123 # 管理员密码（首次启动时创建）
  email: ""

# 国际化配置（错误消息本地化）
i18n:
  default_lang: zh-CN  # 默认语言: zh-CN(中文)、en(英文)

# OAuth 社交登录
oauth:
  enabled: false
  github:
    enabled: false
    client_id: ""
    client_secret: ""
  google:
    enabled: false
    client_id: ""
    client_secret: ""
  linuxdo:
    enabled: false
    client_id: ""
    client_secret: ""
```

说明：

- `static_files=true` 时，后端会返回 embed 的前端静态文件，并暴露 `/docs/index.html` Swagger 文档。
- `static_files=false` 时，后端只保留 API / MCP 路由，不再返回前端页面，也不暴露 Swagger。
- 环境变量 `HOST` / `PORT` / `STATIC_FILES` / `CORS` 的优先级高于 `config.yaml`；CLI 参数仍高于环境变量。
- 用户必须通过 `/api/v1/user/reminder-configs` 创建并启用自己的通知渠道。
- `default_templates` 只提供模板参考，不会直接作为发送目标。
- **管理后台认证**：首次启动时根据 `admin.username` 和 `admin.password` 自动创建管理员账号。后续登录使用用户名密码认证。
- **国际化**：错误消息支持中英文自动切换，客户端通过 `Accept-Language` 头指定语言偏好，服务端自动回退到配置的默认语言。
</details>

<details>
<summary><h3 style="display:inline">Webhook 模板变量</h3></summary>

| 变量                | 说明       | 示例(`server.timezone=Asia/Shanghai`) |
| ------------------- | ---------- | -------------------- |
| `{{.TaskID}}`       | 任务 ID    | `42`                 |
| `{{.Title}}`        | 任务标题   | `"完成报告"`         |
| `{{.Description}}`  | 任务描述   | `"Q2 报告"`          |
| `{{.Priority}}`     | 优先级数字 | `1`                  |
| `{{.PriorityText}}` | 优先级文字 | `"高"`               |
| `{{.DueAt}}`        | 截止时间   | `"5月20日 周二 14:00"` |
| `{{.RemindAt}}`     | 提醒时间   | `"5月19日 周一 09:00"` |
| `{{.RepeatType}}`   | 重复类型   | `"weekly"`               |
| `{{.CreatedAt}}`    | 创建时间   | `"5月9日 周日 10:00"` |

> `DueAt` / `RemindAt` / `CreatedAt` 已在服务端按 `server.timezone` 格式化为 `M月D日 周X HH:MM`,直接渲染到模板。其他字段原样输出。
</details>

<details>
<summary><h3 style="display:inline">时间契约</h3></summary>

所有时间字段统一遵循"存 UTC,传 RFC3339,展示按配置时区输出"。

- **数据库存储**:SQLite `TEXT`,内容为 UTC RFC3339,例如 `2026-05-10T10:30:00Z`
- **API 入参**:接受多种格式,内部一律 UTC 入库
  - RFC3339:`2026-05-10T10:30:00Z`、`2026-05-10T18:30:00+08:00`
  - ISO8601 无冒号时区:`2026-05-10T18:30:00+0800`、`2026-05-10 18:30:00-0500`
- **API 出参**:按 `server.timezone` 配置的目标时区输出 RFC3339 with offset
  - 默认服务器本地时区,例如 `2026-05-10T18:30:00+08:00`(UTC `2026-05-10T10:30:00Z` 在 `Asia/Shanghai` 下的展示)
  - `server.timezone: "UTC"` 时仍输出 `2026-05-10T10:30:00Z`
- **前端展示**:根据浏览器本地时区格式化(`new Date(...).toLocaleString()`)
- **涉及字段**:`due_at`、`remind_at`、`repeat_end_date`、`created_at`、`updated_at`、`reminder_sent_at`、`last_used_at`

### `server.timezone` 配置项

```yaml
server:
  timezone: "Asia/Shanghai"   # IANA 名;或固定偏移 "+08:00";空 / "Local" 表示服务器本地时区;"UTC" 显式 UTC
```

- 可执行文件内置了 `tzdata`，因此 `Asia/Shanghai` 等 IANA 时区名不依赖宿主机是否安装系统时区数据库。
非法值(例如 `"Mars/Olympus"`)启动期记 warn 日志后回退到 `Local`,服务可正常启动。

### MCP `X-MCP-Timezone` 请求头

MCP 调用支持 per-request 覆盖输出时区,优先级:**请求头 > `server.timezone` 配置 > Local**。

```bash
curl -X POST http://localhost:8080/mcp \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -H 'api-key: <YOUR_API_KEY>' \
  -H 'Mcp-Session-Id: <SESSION_ID>' \
  -H 'X-MCP-Timezone: America/New_York' \
  -d '{"jsonrpc":"2.0","id":101,"method":"tools/call","params":{"name":"list_tasks","arguments":{}}}'
```

非法值会被静默忽略,工具回落到全局配置(不返回 4xx,保证客户端健壮性)。

> 数据库中可能存在旧格式(`YYYY-MM-DD HH:MM:SS`)的历史数据,读取路径兼容,但新写入数据统一为 UTC RFC3339。
</details>

<details>
<summary><h3 style="display:inline">业务约束</h3></summary>

- 创建任务时如果设置了提醒时间（`remind_at`），则必须存在至少一个已启用的提醒渠道。
- 未设置 `remind_at` 的任务不需要提醒渠道，可直接创建。
- 禁用某个用户的全部提醒渠道后，该用户任务不会回退到任何全局 webhook。
- 提醒只有在该任务的所有已启用渠道都发送成功后，才会标记为已发送。
- 注册接口在并发下如果用户名冲突，会稳定返回 `409 Conflict`。
</details>

<details>
<summary><h3 style="display:inline">构建与部署</h3></summary>

### 交叉编译

编译输出文件名自动带 OS-ARCH 后缀（如 `server-linux-amd64`）。

```bash
make build           # 编译当前平台
make build-linux     # 交叉编译 Linux (amd64)
make build-windows   # 交叉编译 Windows (amd64)
make build-darwin    # 交叉编译 macOS (arm64)
```

> 项目使用纯 Go SQLite 驱动（modernc.org/sqlite），无需 CGO，所有平台均可直接交叉编译。

### UPX 压缩

编译时会自动检测系统是否安装了 [UPX](https://upx.github.io/)。如果已安装，会自动对二进制文件进行压缩（约减少 60% 体积）。

### Docker 部署详情

`Dockerfile` 会先构建 Vue 前端，再由 Go 后端把静态资源 embed 到同一个镜像和二进制里。

```bash
# 本地构建单体 Docker 镜像
make docker-build

# 直接使用已发布的 GHCR 镜像启动
docker compose up -d
```

发布镜像地址：

```text
ghcr.io/ahsaboy/todo:latest
ghcr.io/ahsaboy/todo:vX.Y.Z
```

如需临时切换 compose 使用的镜像 tag，可覆盖环境变量：

```bash
TODO_IMAGE=ghcr.io/ahsaboy/todo:v1.2.3 docker compose up -d
```

`docker-compose.yml` 还支持用环境变量覆盖配置文件中的部分运行参数，且优先级高于 `config.yaml`：

```bash
PORT=9090 \
HOST=0.0.0.0 \
STATIC_FILES=false \
CORS=https://todo.example.com,https://admin.example.com \
docker compose up -d
```

**Docker 部署配置文件说明**：

Docker 镜像默认使用 `config.example.yaml` 作为配置模板。生产环境需要通过 volume 挂载实际的 `config.yaml`：

```bash
# 方式1：通过 docker run 挂载
docker run -d \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -p 8080:8080 \
  ghcr.io/ahsaboy/todo:latest

# 方式2：在 docker-compose.yml 中配置
services:
  todo:
    image: ghcr.io/ahsaboy/todo:latest
    volumes:
      - ./config.yaml:/app/config.yaml
    ports:
      - "8080:8080"
```

### 独立前端构建

如果需要把前端单独部署到其他静态站点或 CDN，保留一个独立构建命令。它通过 `API_BASE_URL` 指定浏览器访问后端 API 的地址；该值会在 Vite 构建阶段写入前端产物，对应环境变量为 `VITE_API_BASE_URL`。

```bash
# 本地构建前端 dist，不复制到 web/dist
make frontend-build-standalone API_BASE_URL=http://localhost:8080/api/v1

# 构建 nginx 静态前端镜像
docker build -f frontend/Dockerfile \
  --build-arg API_BASE_URL=http://localhost:8080/api/v1 \
  -t todo-frontend:latest \
  ./frontend
```

如果生产环境把独立前端部署到非同源域名，需要在后端 `config.yaml` 中允许前端来源：

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
  --log-file-enabled   启用日志文件输出
  --log-file-disabled  禁用日志文件输出
  --static-files-enabled   启用前端静态文件与 Swagger 路由
  --static-files-disabled  禁用前端静态文件与 Swagger 路由
  -v, --version        显示版本号
  -h, --help           显示帮助信息
```

### Makefile 命令

```bash
make build          # 编译当前平台（自动生成 Swagger 文档，支持 UPX 压缩）
make build-linux    # 交叉编译 Linux amd64
make build-windows  # 交叉编译 Windows amd64
make build-darwin   # 交叉编译 macOS arm64
make frontend-build-standalone API_BASE_URL=http://localhost:8080/api/v1  # 仅构建独立前端 dist
make run            # 编译并运行
make test           # 运行测试
make dev            # 本地开发（go run，自动生成 Swagger 文档）
make swag           # 仅重新生成 Swagger 文档
make clean          # 清理构建产物和数据库文件
make docker-build   # 构建 Docker 镜像（默认标签 ghcr.io/ahsaboy/todo:latest）
make docker-up      # Docker Compose 启动
make docker-down    # Docker Compose 停止
make docker-logs    # 查看 Docker 日志
```

### 日志配置

- 服务端日志始终输出到终端。
- `logging.file_enabled=true` 时，服务端日志还会按天写入 `backend-YYYY-MM-DD.log`，文件位置由 `logging.path` 决定。
- `logging.max_days` 控制日志保留天数；仅在 `logging.file_enabled=true` 时，启动时会清理早于保留窗口的历史日志文件。
- 所有 HTTP 请求都会进入统一访问日志，包括 API、静态资源、404、401 和健康检查。
</details>

<details>
<summary><h3 style="display:inline">管理后台</h3></summary>

系统提供 localhost-only 的管理后台界面，用于系统管理和监控。

### 启用管理后台

1. **配置 admin 段**（在 `config.yaml` 中）：

```yaml
admin:
  enabled: true
  username: admin      # 管理员用户名
  password: admin123   # 管理员密码
  email: ""            # 可选邮箱
```

**注意**：首次启动时会根据配置自动创建管理员账号。后续登录使用用户名密码认证。

2. **访问管理后台**：

```
http://localhost:8080/admin/login
```

### 安全特性

- **仅限 localhost 访问**：所有 `/admin/api/*` 端点只能从服务器本机访问
- **用户名密码认证**：首次启动时自动创建管理员账号，后续使用用户名密码登录
- **会话管理**：令牌存储在浏览器 sessionStorage（非 localStorage）
- **敏感信息隐藏**：系统配置接口自动隐藏敏感信息

### 功能模块

| 模块 | 说明 |
|------|------|
| 仪表盘 | 系统统计信息（用户数、任务数、完成率、提醒配置/日志数）与趋势 |
| 用户管理 | 用户列表、搜索、删除、强制重置密码、切换管理员权限 |
| 任务管理 | 所有用户任务列表、多条件筛选、编辑、切换完成、删除 |
| 提醒配置 | 所有用户的提醒渠道列表、启停、删除 |
| 提醒日志 | 提醒发送记录、状态、错误信息 |
| 系统日志 | 按文件查看/下载服务端日志，可按级别筛选 |
| 操作日志 | 管理员操作审计记录 |
| 系统配置 | 当前系统配置查看与编辑（支持热更新） |

### API 端点

所有 `/admin/api/*` 端点仅限 localhost 访问。除登录外均需 `Authorization: Bearer <api_key>`（管理员账号的 API Key）：

```bash
# 登录（用户名密码），返回管理员 api_key
POST /admin/api/auth/login
Body: {"account": "admin", "password": "admin123"}

# 统计
GET /admin/api/stats
GET /admin/api/stats/trends

# 用户管理
GET    /admin/api/users?page=1&limit=20&search=<keyword>
GET    /admin/api/users/:id
DELETE /admin/api/users/:id
POST   /admin/api/users/:id/reset-password   Body: {"new_password": "..."}
PATCH  /admin/api/users/:id/admin            # 切换管理员权限

# 任务管理
GET    /admin/api/tasks?page=1&limit=20&user_id=<id>&status=<pending|completed>&priority=<1|2|3>
PATCH  /admin/api/tasks/:id/toggle
PUT    /admin/api/tasks/:id
DELETE /admin/api/tasks/:id

# 提醒管理
GET    /admin/api/reminder-configs?page=1&limit=20
PATCH  /admin/api/reminder-configs/:id/toggle
DELETE /admin/api/reminder-configs/:id
GET    /admin/api/reminder-logs?page=1&limit=20

# 系统配置 / 日志 / 审计
GET /admin/api/config
PUT /admin/api/config
DELETE /admin/api/config/*key
GET /admin/api/audit-logs?page=1&limit=20
GET /admin/api/system-logs
GET /admin/api/system-logs/:filename/entries?page=1&limit=50&level=<level>
GET /admin/api/system-logs/:filename/download
```
</details>

<details>
<summary><h3 style="display:inline">项目结构</h3></summary>

```
TODO/
├── cmd/server/main.go              # 入口：配置加载、路由注册、优雅退出
├── internal/
│   ├── config/                     # YAML 配置加载 + 动态注册表
│   ├── logging/                    # 日志初始化、日志路径和保留清理
│   ├── database/database.go        # SQLite 连接 + WAL 模式 + 自动建表
│   ├── i18n/                       # 国际化模块（错误消息本地化）
│   ├── models/                     # 数据模型（task/user/tag/reminder/oauth）
│   ├── handlers/                   # HTTP 处理层（auth/task/tag/reminder/admin/oauth/profile）
│   ├── repository/                 # 数据库操作层
│   ├── service/                    # 业务逻辑层
│   ├── middleware/                  # 中间件（认证/限流/localhost/admin）
│   ├── mcp/                        # MCP 服务器（Streamable HTTP，12 个工具）
│   ├── oauth/                      # OAuth provider（GitHub/Google/LinuxDo）
│   ├── views/                      # 视图层（API 出参转换）
│   ├── utils/                      # 响应格式/校验/时间工具
│   ├── timezone/                   # 时区处理模块
│   └── testutil/                   # 测试用 mock repo/service
├── frontend/                       # Vue 3 + TypeScript 前端
│   ├── src/app/                    # 路由、stores、guards
│   ├── src/entities/               # 领域实体（model + api + mapper）
│   ├── src/features/               # 业务组件 + composables
│   ├── src/pages/                  # 页面（含 admin/ 子目录）
│   ├── src/shared/                 # API client、composables、UI 组件
│   ├── src/widgets/                # 布局壳（sidebar、topbar、bottomNav）
│   └── src/styles/                 # 全局 CSS（variables、motion、layout、theme）
├── web/                            # Go embed 前端构建产物入口
├── docs/                           # Swagger 文档（自动生成）
├── config.example.yaml             # 配置示例文件
├── Dockerfile                      # 单体镜像多阶段构建
├── docker-compose.yml              # 容器编排
└── Makefile                        # 构建命令
```

**注意**：`config.yaml` 包含敏感信息，已被 Git 忽略。首次运行前需要复制 `config.example.yaml`。
</details>
