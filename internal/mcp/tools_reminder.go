package mcp

import (
	"context"
	"fmt"

	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/views"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// registerReminderTools 把 5 个提醒配置工具注册到 MCP server。
// 所有 handler 都通过 service 层访问数据,按 ctx 中的 user_id 隔离。
func registerReminderTools(s *mcpsrv.MCPServer, svc *service.ReminderConfigService) {
	if s == nil || svc == nil {
		return
	}

	s.AddTool(buildListReminderConfigsTool(), listReminderConfigsHandler(svc))
	s.AddTool(buildCreateReminderConfigTool(), createReminderConfigHandler(svc))
	s.AddTool(buildGetReminderConfigTool(), getReminderConfigHandler(svc))
	s.AddTool(buildUpdateReminderConfigTool(), updateReminderConfigHandler(svc))
	s.AddTool(buildDeleteReminderConfigTool(), deleteReminderConfigHandler(svc))
}

// ---------- 工具定义 ----------

func buildListReminderConfigsTool() mcpgo.Tool {
	return mcpgo.NewTool("list_reminder_configs",
		mcpgo.WithDescription("列出当前用户已配置的全部提醒渠道。无入参,返回 {configs:[...]},数组可能为空。"),
	)
}

func buildCreateReminderConfigTool() mcpgo.Tool {
	return mcpgo.NewTool("create_reminder_config",
		mcpgo.WithDescription("新增一个提醒渠道。支持原生 webhook,以及飞书/钉钉/企业微信/Slack 等机器人。max_retries 默认 3,retry_delay_seconds 默认 5,enabled 默认 true。"),
		mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("渠道名称(1-64 字符)"), mcpgo.MinLength(1), mcpgo.MaxLength(64)),
		mcpgo.WithString("channel_type", mcpgo.Required(), mcpgo.Description("渠道类型"), mcpgo.Enum("webhook", "feishu", "dingtalk", "wecom", "slack")),
		mcpgo.WithString("webhook_url", mcpgo.Required(), mcpgo.Description("Webhook 完整 URL(http/https)")),
		mcpgo.WithString("webhook_method", mcpgo.Description("HTTP 方法,默认 POST"), mcpgo.Enum("GET", "POST", "PUT")),
		mcpgo.WithObject("webhook_headers",
			mcpgo.Description("自定义 HTTP 请求头,键值均为字符串,例如 {\"X-Token\":\"abc\"}"),
			mcpgo.AdditionalProperties(map[string]any{"type": "string"}),
		),
		mcpgo.WithString("webhook_body_template", mcpgo.Description("请求体模板,支持 {{title}} {{message}} {{remind_at}} 等占位符;为空则使用全局默认模板")),
		mcpgo.WithNumber("max_retries", mcpgo.Description("失败重试次数(0-10,默认 3)"), mcpgo.Min(0), mcpgo.Max(10)),
		mcpgo.WithNumber("retry_delay_seconds", mcpgo.Description("重试间隔秒数(1-300,默认 5)"), mcpgo.Min(1), mcpgo.Max(300)),
		mcpgo.WithBoolean("enabled", mcpgo.Description("是否启用,默认 true")),
	)
}

func buildGetReminderConfigTool() mcpgo.Tool {
	return mcpgo.NewTool("get_reminder_config",
		mcpgo.WithDescription("根据 ID 获取单个提醒渠道配置详情。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("提醒配置 ID"), mcpgo.Min(1)),
	)
}

func buildUpdateReminderConfigTool() mcpgo.Tool {
	return mcpgo.NewTool("update_reminder_config",
		mcpgo.WithDescription("部分更新提醒渠道,只传需要修改的字段;未传字段保留原值。webhook_headers 传 {} 时清空已有请求头。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("提醒配置 ID"), mcpgo.Min(1)),
		mcpgo.WithString("name", mcpgo.Description("新的渠道名称(1-64 字符)"), mcpgo.MinLength(1), mcpgo.MaxLength(64)),
		mcpgo.WithString("channel_type", mcpgo.Description("新的渠道类型"), mcpgo.Enum("webhook", "feishu", "dingtalk", "wecom", "slack")),
		mcpgo.WithString("webhook_url", mcpgo.Description("新的 Webhook URL")),
		mcpgo.WithString("webhook_method", mcpgo.Description("新的 HTTP 方法"), mcpgo.Enum("GET", "POST", "PUT")),
		mcpgo.WithObject("webhook_headers",
			mcpgo.Description("新的自定义请求头;传 {} 表示清空"),
			mcpgo.AdditionalProperties(map[string]any{"type": "string"}),
		),
		mcpgo.WithString("webhook_body_template", mcpgo.Description("新的请求体模板;传空字符串表示清空")),
		mcpgo.WithNumber("max_retries", mcpgo.Description("失败重试次数(0-10)"), mcpgo.Min(0), mcpgo.Max(10)),
		mcpgo.WithNumber("retry_delay_seconds", mcpgo.Description("重试间隔秒数(1-300)"), mcpgo.Min(1), mcpgo.Max(300)),
		mcpgo.WithBoolean("enabled", mcpgo.Description("是否启用")),
	)
}

func buildDeleteReminderConfigTool() mcpgo.Tool {
	return mcpgo.NewTool("delete_reminder_config",
		mcpgo.WithDescription("根据 ID 删除提醒渠道。成功返回 {deleted:true,id};不存在返回 not found 错误。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("提醒配置 ID"), mcpgo.Min(1)),
	)
}

// ---------- 公共工具 ----------

// extractWebhookHeaders 从工具参数中取出 webhook_headers 字段(对象类型),转换为 map[string]string。
// 字段缺失返回 nil;存在但为空对象返回非 nil 的空 map(配合 hasArg 判定语义)。
// 非字符串值会通过 fmt.Sprint 兜底转字符串,避免静默丢失。
func extractWebhookHeaders(request mcpgo.CallToolRequest) map[string]string {
	raw, ok := request.GetArguments()["webhook_headers"]
	if !ok || raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	headers := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			headers[k] = s
			continue
		}
		headers[k] = fmt.Sprint(v)
	}
	return headers
}

// structuredReminderConfig 把提醒渠道对象转为结构化结果,nil 时返回 not found 错误。
// 出口前用 views.UserReminderConfigView 把时间字段转成 per-request / 全局时区。
func structuredReminderConfig(ctx context.Context, c *models.UserReminderConfig) (*mcpgo.CallToolResult, error) {
	if c == nil {
		return mcpgo.NewToolResultError("reminder config not found"), nil
	}
	fallback := fmt.Sprintf("reminder config #%d %q (%s, enabled=%v)", c.ID, c.Name, c.ChannelType, c.Enabled)
	return buildToolResult(ctx, views.UserReminderConfigView(c, resolveLoc(ctx)), fallback)
}

// requireRemindersEnabled 检查 ctx 中的开关;未启用时返回标准 "tool not available" 错误结果。
// 命中开关关闭时返回非 nil 的 *CallToolResult,handler 应立即 return 它。
func requireRemindersEnabled(ctx context.Context) *mcpgo.CallToolResult {
	if !RemindersEnabled(ctx) {
		return mcpgo.NewToolResultError("tool not available: enable X-MCP-Include-Reminders header to access reminder config tools")
	}
	return nil
}

// ---------- handler 实现 ----------

func listReminderConfigsHandler(svc *service.ReminderConfigService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if blocked := requireRemindersEnabled(ctx); blocked != nil {
			return blocked, nil
		}
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		configs, err := svc.List(ctx, userID)
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to list reminder configs: %v", err), nil
		}
		if configs == nil {
			configs = []models.UserReminderConfig{}
		}
		payload := map[string]any{
			"configs": views.UserReminderConfigsView(configs, resolveLoc(ctx)),
			"total":   len(configs),
		}
		fallback := fmt.Sprintf("returned %d reminder config(s)", len(configs))
		return buildToolResult(ctx, payload, fallback)
	}
}

func createReminderConfigHandler(svc *service.ReminderConfigService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if blocked := requireRemindersEnabled(ctx); blocked != nil {
			return blocked, nil
		}
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}

		name, err := request.RequireString("name")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}
		channelType, err := request.RequireString("channel_type")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}
		webhookURL, err := request.RequireString("webhook_url")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		req := models.CreateReminderConfigRequest{
			Name:                name,
			ChannelType:         channelType,
			WebhookURL:          webhookURL,
			WebhookMethod:       request.GetString("webhook_method", ""),
			WebhookHeaders:      extractWebhookHeaders(request),
			WebhookBodyTemplate: request.GetString("webhook_body_template", ""),
		}

		if hasArg(request, "max_retries") {
			v := request.GetInt("max_retries", 0)
			req.MaxRetries = &v
		}
		if hasArg(request, "retry_delay_seconds") {
			v := request.GetInt("retry_delay_seconds", 0)
			req.RetryDelaySeconds = &v
		}
		if hasArg(request, "enabled") {
			v := request.GetBool("enabled", true)
			req.Enabled = &v
		}

		cfg, err := svc.Create(ctx, userID, req)
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to create reminder config: %v", err), nil
		}
		return structuredReminderConfig(ctx, cfg)
	}
}

func getReminderConfigHandler(svc *service.ReminderConfigService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if blocked := requireRemindersEnabled(ctx); blocked != nil {
			return blocked, nil
		}
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}
		cfg, err := svc.GetByID(ctx, userID, int64(id))
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to get reminder config: %v", err), nil
		}
		return structuredReminderConfig(ctx, cfg)
	}
}

func updateReminderConfigHandler(svc *service.ReminderConfigService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if blocked := requireRemindersEnabled(ctx); blocked != nil {
			return blocked, nil
		}
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		var req models.UpdateReminderConfigRequest

		if hasArg(request, "name") {
			v := request.GetString("name", "")
			req.Name = &v
		}
		if hasArg(request, "channel_type") {
			v := request.GetString("channel_type", "")
			req.ChannelType = &v
		}
		if hasArg(request, "webhook_url") {
			v := request.GetString("webhook_url", "")
			req.WebhookURL = &v
		}
		if hasArg(request, "webhook_method") {
			v := request.GetString("webhook_method", "")
			req.WebhookMethod = &v
		}
		if hasArg(request, "webhook_headers") {
			headers := extractWebhookHeaders(request)
			if headers == nil {
				// 调用方显式传 webhook_headers:{} 时落地为空 map,
				// 让 repo 的 "req.WebhookHeaders != nil" 判定能命中并清空请求头。
				headers = map[string]string{}
			}
			req.WebhookHeaders = headers
		}
		if hasArg(request, "webhook_body_template") {
			v := request.GetString("webhook_body_template", "")
			req.WebhookBodyTemplate = &v
		}
		if hasArg(request, "max_retries") {
			v := request.GetInt("max_retries", 0)
			req.MaxRetries = &v
		}
		if hasArg(request, "retry_delay_seconds") {
			v := request.GetInt("retry_delay_seconds", 0)
			req.RetryDelaySeconds = &v
		}
		if hasArg(request, "enabled") {
			v := request.GetBool("enabled", false)
			req.Enabled = &v
		}

		cfg, err := svc.Update(ctx, userID, int64(id), req)
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to update reminder config: %v", err), nil
		}
		return structuredReminderConfig(ctx, cfg)
	}
}

func deleteReminderConfigHandler(svc *service.ReminderConfigService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if blocked := requireRemindersEnabled(ctx); blocked != nil {
			return blocked, nil
		}
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}
		deleted, err := svc.Delete(ctx, userID, int64(id))
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to delete reminder config: %v", err), nil
		}
		if !deleted {
			return mcpgo.NewToolResultError("reminder config not found"), nil
		}
		payload := map[string]any{"deleted": true, "id": int64(id)}
		fallback := fmt.Sprintf("reminder config #%d deleted", id)
		return buildToolResult(ctx, payload, fallback)
	}
}
