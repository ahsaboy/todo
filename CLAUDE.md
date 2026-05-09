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

修改 API 注解后需手动重新生成 Swagger：
```bash
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go  # 修复 swag 版本兼容问题
```

## 架构

四层结构，依赖通过构造函数注入：

```
cmd/server/main.go          入口：配置 → DB → 注册路由 → 提醒 goroutine → 优雅退出
  ↓
handlers/task_handler.go    HTTP 处理（Gin handler）
  ↓
service/task_service.go     业务逻辑（重复任务自动生成）
service/reminder_service.go 后台提醒（定时扫描 + Webhook 模板渲染 + 重试）
  ↓
repository/task_repo.go     数据库操作（手写 SQL）
  ↓
database/database.go        SQLite 连接（WAL 模式 + busy_timeout）
```

中间件链：Recovery → CORS（可选）→ AuthMiddleware（X-API-Key）→ 路由

## 关键约定

**统一响应格式**（`internal/utils/response.go`）：
- 成功：`{"success": true, "data": ...}`
- 错误：`{"success": false, "error": "...", "code": "NOT_FOUND|UNAUTHORIZED|INVALID_INPUT|INTERNAL_ERROR"}`
- 分页：额外返回 `{"meta": {"page", "limit", "total_items", "total_pages"}}`

**配置优先级**：`config.yaml` → CLI flag 覆盖（`-p`, `--host`, `--mode`）

**提醒模板**：`config.yaml` 中 `reminder.webhook_body_template` 使用 Go `text/template`，可用 `{{.Title}}` `{{.PriorityText}}` 等变量。修改模板后无需重新编译。

## 注意事项

- SQLite 数据库文件在 `data/tasks.db`，通过 Docker volume 挂载持久化
- `reminder_sent` 字段用于去重，更新 `remind_at` 时会自动 reset
- 重复任务在 `ToggleComplete` 标记完成时自动生成下一次
- docs/docs.go 是 `swag init` 自动生成的，不要手动编辑
