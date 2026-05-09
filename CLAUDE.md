# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
make build    # 编译 (Windows 自动加 .exe)
make run      # 编译并运行
make dev      # go run 本地开发
make test     # 运行测试
make clean    # 清理 bin/ 和 data/*.db
make docker-up / docker-down / docker-logs
```

编译需要 `CGO_ENABLED=1`（go-sqlite3 依赖 CGO）。Dockerfile 中已显式设置。

`make build` 和 `make dev` 会自动重新生成 Swagger 文档。也可单独执行：
```bash
make swag     # 仅重新生成 Swagger 文档
```

## 架构

五层结构，依赖通过构造函数注入：

```
cmd/server/main.go              入口：配置 → DB → 注册路由 → 提醒 goroutine → 优雅退出
  ↓
handlers/auth_handler.go        认证 HTTP 处理（注册/登录/Key管理/用户管理）
handlers/task_handler.go        任务 HTTP 处理
handlers/reminder_config_handler.go  提醒配置 CRUD
  ↓
service/auth_service.go         认证逻辑（bcrypt 密码/SHA-256 API Key）
service/task_service.go         业务逻辑（重复任务自动生成）
service/reminder_service.go     后台提醒（定时扫描 + 按用户多渠道推送）
service/reminder_config_service.go 提醒配置管理
  ↓
repository/user_repo.go         用户数据库操作
repository/api_key_repo.go      API Key 数据库操作
repository/task_repo.go         任务数据库操作（按 user_id 隔离）
repository/reminder_config_repo.go 提醒配置数据库操作
  ↓
database/database.go            SQLite 连接（WAL 模式 + busy_timeout）
```

中间件链：Recovery → CORS（可选）→ AuthMiddleware（Bearer/api-key/X-API-Key）→ 路由

## 数据库表

- `users` — 用户表（username, email, password_hash）
- `user_api_keys` — 用户 API Key 表（存 SHA-256 哈希，不存明文）
- `user_reminder_configs` — 用户提醒配置表（支持多渠道 webhook）
- `tasks` — 任务表（含 user_id 外键，按用户隔离）

## 认证方式

支持两种请求头（任选其一）：
- `Authorization: Bearer <api_key>`
- `api-key: <api_key>`

API Key 通过 `crypto/rand` 生成 32 字节 base64url 编码（43 字符），SHA-256 哈希后存储在数据库。

## 关键约定

**统一响应格式**（`internal/utils/response.go`）：
- 成功：`{"success": true, "data": ...}`
- 错误：`{"success": false, "error": "...", "code": "NOT_FOUND|UNAUTHORIZED|INVALID_INPUT|INTERNAL_ERROR"}`
- 分页：额外返回 `{"meta": {"page", "limit", "total_items", "total_pages"}}`

**配置优先级**：`config.yaml` → CLI flag 覆盖（`-p`, `--host`, `--mode`）

**提醒模板**：`config.yaml` 中 `reminder.webhook_body_template` 为默认模板，用户可在 API 中自定义每个渠道的模板。

## 注意事项

- SQLite 数据库文件在 `data/tasks.db`，通过 Docker volume 挂载持久化
- `reminder_sent` 字段用于去重，更新 `remind_at` 时会自动 reset
- 重复任务在 `ToggleComplete` 标记完成时自动生成下一次
- docs/docs.go 是 `swag init` 自动生成的，不要手动编辑
- 密码使用 bcrypt(cost=14) 哈希，API Key 使用 SHA-256 哈希存储
