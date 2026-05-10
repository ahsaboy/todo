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
| `{{.DueAt}}`        | 截止时间   | `"2026-05-20 14:00"` |
| `{{.RemindAt}}`     | 提醒时间   | `"2026-05-19 09:00"` |
| `{{.RepeatType}}`   | 重复类型   | `"weekly"`           |
| `{{.CreatedAt}}`    | 创建时间   | `"2026-05-09 10:00"` |

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
│   └── src/shared/logger/           # 前端日志封装和上报
├── web/                            # Go embed 前端构建产物入口
│   ├── embed.go                    # 使用 //go:embed all:dist
│   └── dist/                       # make build 生成并复制的静态文件
├── config.yaml                     # 配置文件
├── Dockerfile                      # 多阶段构建
├── docker-compose.yml              # 容器编排
└── Makefile                        # 构建命令
```

## Makefile 命令

```bash
make build          # 编译当前平台（自动生成 Swagger 文档，支持 UPX 压缩）
make build-linux    # 交叉编译 Linux amd64
make build-windows  # 交叉编译 Windows amd64
make build-darwin   # 交叉编译 macOS arm64
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
