// Package mcp 提供基于 mark3labs/mcp-go 的 MCP(Model Context Protocol)服务器。
// 通过 Streamable HTTP 传输(协议 2025-11-25),将 TODO 系统的任务/提醒/用户能力以 MCP 工具
// 形式暴露给 LLM 客户端。认证沿用 REST API 的 user_api_keys 表,按 user_id 隔离。
package mcp

import (
	"context"
	"net/http"
	"strings"

	"todo/internal/repository"
	"todo/internal/service"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

const (
	// serverName / serverVersion 在 MCP initialize 阶段以 ServerInfo 形式返回给客户端。
	serverName    = "todo-mcp"
	serverVersion = "0.1.0"

	// endpointPath 仅作为 mcp-go 内部元信息;实际路由由 main.go 挂载到 /mcp(T04 接入)。
	endpointPath = "/mcp"
)

// reminderToolNames 罗列所有受 X-MCP-Include-Reminders 控制的工具名。
// 与 registerReminderTools 中注册的 5 个工具保持同步。
var reminderToolNames = map[string]struct{}{
	"list_reminder_configs":  {},
	"create_reminder_config": {},
	"get_reminder_config":    {},
	"update_reminder_config": {},
	"delete_reminder_config": {},
}

// Dependencies 收集 MCP 工具需要的 service / repo,由 main.go 在初始化阶段构造后传入。
type Dependencies struct {
	TaskSvc     service.TaskServiceInterface
	ReminderSvc service.ReminderConfigServiceInterface
	AuthSvc     service.AuthServiceInterface
	APIKeyRepo  repository.APIKeyRepository
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
//
// WithToolFilter 在 tools/list 阶段按 ctx 中的 RemindersEnabled 开关过滤掉
// reminder 系列工具;tools/call 阶段则由各 reminder handler 入口的 requireRemindersEnabled
// 兜底拒绝,即使客户端绕过 list 直接调用也不会泄露能力。
func NewMCPServer(deps Dependencies) http.Handler {
	s := mcpsrv.NewMCPServer(
		serverName,
		serverVersion,
		mcpsrv.WithInstructions("TODO MCP server: manage tasks, reminder configs and user profile for the authenticated user."),
		mcpsrv.WithToolCapabilities(false),
		mcpsrv.WithLogging(),
		mcpsrv.WithRecovery(),
		mcpsrv.WithToolFilter(filterRemindersByContext),
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

// filterRemindersByContext 根据 ctx 中的 RemindersEnabled 决定是否在 tools/list 输出
// 中保留 reminder 工具。开关关闭时(默认)从 tools 切片中剔除 5 个 reminder 工具。
// 因为 mcp-go 的 toolFilters 在 handleListTools 中按 per-request ctx 调用,
// 所以每次 tools/list 都会拿到当时请求头的状态,无需关心 session 缓存。
func filterRemindersByContext(ctx context.Context, tools []mcpgo.Tool) []mcpgo.Tool {
	if RemindersEnabled(ctx) {
		return tools
	}
	filtered := make([]mcpgo.Tool, 0, len(tools))
	for _, t := range tools {
		if _, ok := reminderToolNames[strings.TrimSpace(t.Name)]; ok {
			continue
		}
		filtered = append(filtered, t)
	}
	return filtered
}
