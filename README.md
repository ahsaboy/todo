# TODO 任务管理系统

一个轻量级的 TODO 任务管理服务，基于 Go + Gin + SQLite，支持 REST API CRUD、后台提醒推送、重复任务自动生成。

## 功能特性

- 任务 CRUD（创建、查询、更新、删除）
- 分页、排序、筛选、关键字搜索
- 任务优先级（高/中/低）
- 截止时间与提醒时间设定
- 后台定时扫描，自动 HTTP 推送提醒（Webhook，支持自定义模板，适配飞书/钉钉/企业微信等）
- 重复任务（daily/weekly/monthly/yearly），完成时自动生成下一次
- API Key 认证
- 健康检查端点（Docker 健康探测）
- Swagger API 文档
- 优雅退出
- Docker Compose 部署

## 快速开始

### 本地运行

```bash
# 编译
go build -o bin/server ./cmd/server

# 运行（默认读取 config.yaml）
./bin/server

# 或直接运行
go run ./cmd/server
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

**示例:**

```bash
./bin/server -c /etc/todo/config.yaml
./bin/server -p 9090 --mode release
./bin/server -c prod.yaml -p 80 --host 0.0.0.0
```

## API 接口

服务启动后访问 Swagger 文档：`http://localhost:8080/docs/index.html`

### Swagger 认证

打开 Swagger UI 后，点击页面右上角的 **Authorize** 按钮，输入你的 API Key（对应 `config.yaml` 中的 `auth.api_key`），之后所有请求会自动携带 `X-API-Key` 请求头。

### 端点一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查（无需认证） |
| POST | `/api/v1/tasks` | 创建任务 |
| GET | `/api/v1/tasks` | 获取任务列表 |
| GET | `/api/v1/tasks/:id` | 获取单个任务 |
| PUT | `/api/v1/tasks/:id` | 更新任务 |
| DELETE | `/api/v1/tasks/:id` | 删除任务 |
| PATCH | `/api/v1/tasks/:id/complete` | 切换完成状态 |

### 认证

除健康检查外，所有 API 需要在请求头中携带 `X-API-Key`：

```bash
curl -H "X-API-Key: your-secure-api-key" http://localhost:8080/api/v1/tasks
```

### 示例

**创建任务:**

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secure-api-key" \
  -d '{
    "title": "完成项目报告",
    "priority": 1,
    "due_at": "2026-05-20 14:00:00",
    "remind_at": "2026-05-19 09:00:00",
    "repeat_type": "weekly"
  }'
```

**列表查询（分页 + 筛选）:**

```bash
curl "http://localhost:8080/api/v1/tasks?page=1&limit=10&status=pending&priority=1&sort=due_at&order=asc" \
  -H "X-API-Key: your-secure-api-key"
```

**切换完成状态:**

```bash
curl -X PATCH http://localhost:8080/api/v1/tasks/1/complete \
  -H "X-API-Key: your-secure-api-key"
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
  host: "0.0.0.0"       # 监听地址
  port: 8080             # 监听端口
  mode: "debug"          # debug / release

database:
  path: "./data/tasks.db"  # SQLite 文件路径

auth:
  api_key: "your-secure-api-key"  # 留空则不启用认证

reminder:
  enabled: true
  scan_interval_seconds: 30        # 扫描间隔
  webhook_url: "http://localhost:9000/webhook"
  webhook_method: "POST"
  webhook_headers:
    Content-Type: "application/json"
  webhook_body_template: |         # 支持 Go template 变量
    {"msg_type":"text","content":{"text":"[TODO提醒] {{.Title}}\n截止: {{.DueAt}}"}}
  max_retries: 3
  retry_delay_seconds: 5

cors:
  enabled: true
  allowed_origins: ["*"]

logging:
  level: "info"          # debug / info / warn / error
  format: "json"         # json / console
```

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

**适配飞书:**

```yaml
webhook_body_template: |
  {"msg_type":"text","content":{"text":"[TODO] {{.Title}}\n优先级: {{.PriorityText}}\n截止: {{.DueAt}}"}}
```

**适配钉钉 (Markdown):**

```yaml
webhook_body_template: |
  {"msgtype":"markdown","markdown":{"title":"TODO提醒","text":"### {{.Title}}\n- 优先级: {{.PriorityText}}\n- 截止: {{.DueAt}}"}}
```

**适配企业微信:**

```yaml
webhook_body_template: |
  {"msgtype":"text","text":{"content":"[TODO提醒] {{.Title}}\n优先级: {{.PriorityText}}\n截止: {{.DueAt}}"}}
```

## 项目结构

```
TODO/
├── cmd/server/main.go              # 入口：配置加载、路由注册、优雅退出
├── internal/
│   ├── config/config.go            # YAML 配置加载
│   ├── database/database.go        # SQLite 连接 + WAL 模式 + 自动建表
│   ├── models/task.go              # 数据模型 + 模板变量映射
│   ├── handlers/task_handler.go    # HTTP 处理函数
│   ├── repository/task_repo.go     # 数据库操作层
│   ├── service/
│   │   ├── task_service.go         # 业务逻辑 + 重复任务自动生成
│   │   └── reminder_service.go     # 后台提醒扫描 + Webhook 推送
│   ├── middleware/auth.go          # API Key 认证中间件
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
make build          # 编译
make run            # 编译并运行
make test           # 运行测试
make dev            # 本地开发（go run，读取 config.yaml）
make clean          # 清理构建产物和数据库文件
make docker-build   # 构建 Docker 镜像
make docker-up      # Docker Compose 启动
make docker-down    # Docker Compose 停止
make docker-logs    # 查看 Docker 日志
```

> Makefile 自动检测操作系统，Windows 上编译出 `server.exe`，Linux/macOS 编译出 `server`。

## 技术栈

- **语言**: Go
- **Web 框架**: Gin
- **数据库**: SQLite (CGO)
- **日志**: zap
- **配置**: YAML
- **API 文档**: Swagger (swaggo)
