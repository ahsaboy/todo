// Package mcp 提供基于 mark3labs/mcp-go 的 MCP(Model Context Protocol)服务器。
// 通过 Streamable HTTP 传输(协议 2025-11-25),将 TODO 系统的任务/提醒/用户能力以 MCP 工具
// 形式暴露给 LLM 客户端。认证沿用 REST API 的 user_api_keys 表,按 user_id 隔离。
package mcp

import (
	"net/http"

	"todo/internal/repository"
	"todo/internal/service"

	mcpsrv "github.com/mark3labs/mcp-go/server"
)

const (
	// serverName / serverVersion 在 MCP initialize 阶段以 ServerInfo 形式返回给客户端。
	serverName    = "todo-mcp"
	serverVersion = "0.1.0"

	// endpointPath 仅作为 mcp-go 内部元信息;实际路由由 main.go 挂载到 /mcp(T04 接入)。
	endpointPath = "/mcp"
)

// Dependencies 收集 MCP 工具需要的 service / repo,由 main.go 在初始化阶段构造后传入。
type Dependencies struct {
	TaskSvc     *service.TaskService
	ReminderSvc *service.ReminderConfigService
	AuthSvc     *service.AuthService
	APIKeyRepo  *repository.APIKeyRepo
}

// NewMCPServer 构造一个挂载好工具与认证的 MCP HTTP handler。
// 外层包了 apiKeyAuthMiddleware,内部为 mcp-go 的 StreamableHTTPServer(实现 http.Handler)。
//
// 协议版本:遵循 MCP 最新规范 mcp.LATEST_PROTOCOL_VERSION("2025-11-25"),由 mcp-go 在
// initialize 应答中自动报告;ValidProtocolVersions 同时兼容 2025-06-18 / 2025-03-26。
//
// Origin 校验:本期暂跳过 mcp-go 自带的 CORS,由 main.go 的全局 CORS 中间件兜底(见 T04)。
//
// registerTaskTools / registerReminderTools / registerUserTools 的定义分别放在
// tools_task.go(T02)与 tools_reminder.go / tools_user.go(T03)中。
func NewMCPServer(deps Dependencies) http.Handler {
	s := mcpsrv.NewMCPServer(
		serverName,
		serverVersion,
		mcpsrv.WithInstructions("TODO MCP server: manage tasks, reminder configs and user profile for the authenticated user."),
		mcpsrv.WithToolCapabilities(false),
		mcpsrv.WithLogging(),
		mcpsrv.WithRecovery(),
	)

	registerTaskTools(s, deps.TaskSvc)
	registerReminderTools(s, deps.ReminderSvc)
	registerUserTools(s, deps.AuthSvc)

	stream := mcpsrv.NewStreamableHTTPServer(
		s,
		mcpsrv.WithEndpointPath(endpointPath),
	)

	return apiKeyAuthMiddleware(deps.APIKeyRepo)(stream)
}
