package mcp

import (
	"context"
	"errors"
	"fmt"

	"todo/internal/models"
	"todo/internal/service"
	"todo/internal/views"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// registerTaskTools 把 6 个任务管理工具注册到 MCP server。
// 所有 handler 都从 ctx 取 user_id,经 service 层调用,避免直接访问 repository。
func registerTaskTools(s *mcpsrv.MCPServer, svc *service.TaskService) {
	if s == nil || svc == nil {
		return
	}

	s.AddTool(buildCreateTaskTool(), createTaskHandler(svc))
	s.AddTool(buildListTasksTool(), listTasksHandler(svc))
	s.AddTool(buildGetTaskTool(), getTaskHandler(svc))
	s.AddTool(buildUpdateTaskTool(), updateTaskHandler(svc))
	s.AddTool(buildDeleteTaskTool(), deleteTaskHandler(svc))
	s.AddTool(buildToggleTaskCompleteTool(), toggleTaskCompleteHandler(svc))
}

// ---------- 工具定义 ----------

func buildCreateTaskTool() mcpgo.Tool {
	return mcpgo.NewTool("create_task",
		mcpgo.WithDescription("创建一个新任务。支持标题、描述、优先级(1高/2中/3低)、截止时间、提醒时间、重复规则。若设置了 remind_at,需先存在至少一个已启用的提醒渠道。"),
		mcpgo.WithString("title", mcpgo.Required(), mcpgo.Description("任务标题(1-255 字符,必填)"), mcpgo.MinLength(1), mcpgo.MaxLength(255)),
		mcpgo.WithString("description", mcpgo.Description("任务描述(可选,最多 1000 字符)"), mcpgo.MaxLength(1000)),
		mcpgo.WithNumber("priority", mcpgo.Description("优先级:1=高 / 2=中 / 3=低"), mcpgo.Min(1), mcpgo.Max(3)),
		mcpgo.WithString("due_at", mcpgo.Description("截止时间,推荐格式 \"2026-05-10 18:30:00\"")),
		mcpgo.WithString("remind_at", mcpgo.Description("提醒时间,推荐格式 \"2026-05-10 18:30:00\"")),
		mcpgo.WithString("repeat_type", mcpgo.Description("重复类型"), mcpgo.Enum("none", "daily", "weekly", "monthly", "yearly")),
		mcpgo.WithNumber("repeat_interval", mcpgo.Description("重复间隔(1-365)"), mcpgo.Min(1), mcpgo.Max(365)),
		mcpgo.WithString("repeat_end_date", mcpgo.Description("重复结束时间,推荐格式 \"2026-05-10 18:30:00\"")),
	)
}

func buildListTasksTool() mcpgo.Tool {
	return mcpgo.NewTool("list_tasks",
		mcpgo.WithDescription("分页查询当前用户的任务列表,支持按状态/优先级/截止时间区间/关键字筛选,以及自定义排序。"),
		mcpgo.WithString("status", mcpgo.Description("任务状态筛选"), mcpgo.Enum("pending", "completed", "all")),
		mcpgo.WithNumber("priority", mcpgo.Description("优先级筛选:1/2/3"), mcpgo.Min(1), mcpgo.Max(3)),
		mcpgo.WithString("due_before", mcpgo.Description("截止时间上限,推荐格式 \"2026-05-10 18:30:00\"")),
		mcpgo.WithString("due_after", mcpgo.Description("截止时间下限,推荐格式 \"2026-05-10 18:30:00\"")),
		mcpgo.WithString("search", mcpgo.Description("搜索关键字(匹配标题/描述)")),
		mcpgo.WithNumber("page", mcpgo.Description("页码,默认 1"), mcpgo.Min(1), mcpgo.DefaultNumber(1)),
		mcpgo.WithNumber("limit", mcpgo.Description("每页数量,默认 20,最大 100"), mcpgo.Min(1), mcpgo.Max(100), mcpgo.DefaultNumber(20)),
		mcpgo.WithString("sort", mcpgo.Description("排序字段,默认 created_at"), mcpgo.Enum("created_at", "updated_at", "due_at", "priority"), mcpgo.DefaultString("created_at")),
		mcpgo.WithString("order", mcpgo.Description("排序方向,默认 desc"), mcpgo.Enum("asc", "desc"), mcpgo.DefaultString("desc")),
	)
}

func buildGetTaskTool() mcpgo.Tool {
	return mcpgo.NewTool("get_task",
		mcpgo.WithDescription("根据任务 ID 获取单个任务的详细信息。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("任务 ID"), mcpgo.Min(1)),
	)
}

func buildUpdateTaskTool() mcpgo.Tool {
	return mcpgo.NewTool("update_task",
		mcpgo.WithDescription("部分更新指定任务,只传需要修改的字段;未传或传空字符串的字段保持原值不变(MCP 不支持清空时间字段,如需清空请使用 REST API)。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("任务 ID"), mcpgo.Min(1)),
		mcpgo.WithString("title", mcpgo.Description("新的任务标题"), mcpgo.MinLength(1), mcpgo.MaxLength(255)),
		mcpgo.WithString("description", mcpgo.Description("新的任务描述"), mcpgo.MaxLength(1000)),
		mcpgo.WithNumber("priority", mcpgo.Description("优先级:1/2/3"), mcpgo.Min(1), mcpgo.Max(3)),
		mcpgo.WithString("due_at", mcpgo.Description("新的截止时间,推荐格式 \"2026-05-10 18:30:00\";未传或留空则不修改")),
		mcpgo.WithString("remind_at", mcpgo.Description("新的提醒时间,推荐格式 \"2026-05-10 18:30:00\";未传或留空则不修改")),
		mcpgo.WithString("repeat_type", mcpgo.Description("新的重复类型"), mcpgo.Enum("none", "daily", "weekly", "monthly", "yearly")),
		mcpgo.WithNumber("repeat_interval", mcpgo.Description("新的重复间隔(1-365)"), mcpgo.Min(1), mcpgo.Max(365)),
		mcpgo.WithString("repeat_end_date", mcpgo.Description("新的重复结束时间,推荐格式 \"2026-05-10 18:30:00\";未传或留空则不修改")),
	)
}

func buildDeleteTaskTool() mcpgo.Tool {
	return mcpgo.NewTool("delete_task",
		mcpgo.WithDescription("根据 ID 永久删除任务。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("任务 ID"), mcpgo.Min(1)),
	)
}

func buildToggleTaskCompleteTool() mcpgo.Tool {
	return mcpgo.NewTool("toggle_task_complete",
		mcpgo.WithDescription("切换任务完成/未完成状态。重复任务被标记为完成时,自动生成下一次实例。"),
		mcpgo.WithNumber("id", mcpgo.Required(), mcpgo.Description("任务 ID"), mcpgo.Min(1)),
	)
}

// ---------- 公共工具 ----------

// requireUserID 从 ctx 取出 user_id,未取到时返回标准未授权错误。
func requireUserID(ctx context.Context) (int64, *mcpgo.CallToolResult) {
	userID, ok := UserIDFromContext(ctx)
	if !ok || userID <= 0 {
		return 0, mcpgo.NewToolResultError("unauthorized: missing user context")
	}
	return userID, nil
}

// mapTaskServiceError 把 service 层错误映射为 MCP 工具错误结果。
// 命中已知错误时返回非 nil 的 *CallToolResult;否则返回 nil(由调用方决定如何处理)。
func mapTaskServiceError(err error) *mcpgo.CallToolResult {
	switch {
	case errors.Is(err, service.ErrReminderChannelMissing):
		return mcpgo.NewToolResultError("reminder channel required: please configure at least one enabled reminder channel before setting remind_at")
	case errors.Is(err, service.ErrInvalidTime):
		return mcpgo.NewToolResultError("invalid time format: use \"2026-05-10 18:30:00\"")
	}
	return nil
}

// hasArg 检查工具调用是否传入了该字段(用于区分"未传"与"传零值")。
func hasArg(req mcpgo.CallToolRequest, key string) bool {
	_, ok := req.GetArguments()[key]
	return ok
}

// structuredTask 用 task 对象做结构化返回,fallback 文本写成简短摘要,便于纯文本客户端显示。
// 出口前用 views.TaskView 把时间字段转成 per-request / 全局时区的带偏移格式。
func structuredTask(ctx context.Context, t *models.Task) (*mcpgo.CallToolResult, error) {
	if t == nil {
		return mcpgo.NewToolResultError("task not found"), nil
	}
	fallback := fmt.Sprintf("task #%d %q (priority=%d, completed=%v)", t.ID, t.Title, t.Priority, t.Completed)
	return buildToolResult(ctx, views.TaskView(t, resolveLoc(ctx)), fallback)
}

// ---------- handler 实现 ----------

func createTaskHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}

		title, err := request.RequireString("title")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		req := models.CreateTaskRequest{
			Title:       title,
			Description: request.GetString("description", ""),
		}

		if hasArg(request, "priority") {
			p := request.GetInt("priority", 0)
			req.Priority = &p
		}
		if hasArg(request, "due_at") {
			v := request.GetString("due_at", "")
			req.DueAt = &v
		}
		if hasArg(request, "remind_at") {
			v := request.GetString("remind_at", "")
			req.RemindAt = &v
		}
		if hasArg(request, "repeat_type") {
			v := request.GetString("repeat_type", "")
			req.RepeatType = &v
		}
		if hasArg(request, "repeat_interval") {
			v := request.GetInt("repeat_interval", 0)
			req.RepeatInterval = &v
		}
		if hasArg(request, "repeat_end_date") {
			v := request.GetString("repeat_end_date", "")
			req.RepeatEndDate = &v
		}

		task, err := svc.Create(ctx, userID, req)
		if err != nil {
			if mapped := mapTaskServiceError(err); mapped != nil {
				return mapped, nil
			}
			return mcpgo.NewToolResultErrorf("failed to create task: %v", err), nil
		}
		return structuredTask(ctx, task)
	}
}

func listTasksHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}

		page := request.GetInt("page", 1)
		if page < 1 {
			page = 1
		}
		limit := request.GetInt("limit", 20)
		if limit < 1 || limit > 100 {
			limit = 20
		}

		filters := models.TaskFilters{
			Status:    request.GetString("status", ""),
			Priority:  request.GetInt("priority", 0),
			DueBefore: request.GetString("due_before", ""),
			DueAfter:  request.GetString("due_after", ""),
			Search:    request.GetString("search", ""),
		}

		sortField := request.GetString("sort", "created_at")
		sortOrder := request.GetString("order", "desc")

		tasks, total, err := svc.List(ctx, userID, filters, page, limit, sortField, sortOrder)
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to list tasks: %v", err), nil
		}
		if tasks == nil {
			tasks = []models.Task{}
		}

		totalPages := 0
		if limit > 0 {
			totalPages = int((total + int64(limit) - 1) / int64(limit))
		}

		payload := map[string]any{
			"tasks": views.TasksView(tasks, resolveLoc(ctx)),
			"meta": map[string]any{
				"page":        page,
				"limit":       limit,
				"total_items": total,
				"total_pages": totalPages,
			},
		}

		fallback := fmt.Sprintf("returned %d task(s), page %d/%d, total %d", len(tasks), page, totalPages, total)
		return buildToolResult(ctx, payload, fallback)
	}
}

func getTaskHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		task, err := svc.GetByID(ctx, userID, int64(id))
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to get task: %v", err), nil
		}
		return structuredTask(ctx, task)
	}
}

func updateTaskHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		var req models.UpdateTaskRequest

		if hasArg(request, "title") {
			v := request.GetString("title", "")
			req.Title = &v
		}
		if hasArg(request, "description") {
			v := request.GetString("description", "")
			req.Description = &v
		}
		if hasArg(request, "priority") {
			v := request.GetInt("priority", 0)
			req.Priority = &v
		}
		if hasArg(request, "due_at") {
			if v := request.GetString("due_at", ""); v != "" {
				req.DueAt = &v
			}
		}
		if hasArg(request, "remind_at") {
			if v := request.GetString("remind_at", ""); v != "" {
				req.RemindAt = &v
			}
		}
		if hasArg(request, "repeat_type") {
			v := request.GetString("repeat_type", "")
			req.RepeatType = &v
		}
		if hasArg(request, "repeat_interval") {
			v := request.GetInt("repeat_interval", 0)
			req.RepeatInterval = &v
		}
		if hasArg(request, "repeat_end_date") {
			if v := request.GetString("repeat_end_date", ""); v != "" {
				req.RepeatEndDate = &v
			}
		}

		task, err := svc.Update(ctx, userID, int64(id), req)
		if err != nil {
			if mapped := mapTaskServiceError(err); mapped != nil {
				return mapped, nil
			}
			return mcpgo.NewToolResultErrorf("failed to update task: %v", err), nil
		}
		return structuredTask(ctx, task)
	}
}

func deleteTaskHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
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
			return mcpgo.NewToolResultErrorf("failed to delete task: %v", err), nil
		}
		if !deleted {
			return mcpgo.NewToolResultError("task not found"), nil
		}

		payload := map[string]any{"deleted": true, "id": int64(id)}
		fallback := fmt.Sprintf("task #%d deleted", id)
		return buildToolResult(ctx, payload, fallback)
	}
}

func toggleTaskCompleteHandler(svc *service.TaskService) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		id, err := request.RequireInt("id")
		if err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		task, err := svc.ToggleComplete(ctx, userID, int64(id))
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to toggle task: %v", err), nil
		}
		return structuredTask(ctx, task)
	}
}
