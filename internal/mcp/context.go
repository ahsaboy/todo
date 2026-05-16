package mcp

import "context"

// ctxKey 是包私有的 context key 类型,避免与其他包发生冲突。
type ctxKey struct{ name string }

var (
	userIDKey            = ctxKey{name: "userID"}
	structuredOutputKey  = ctxKey{name: "structuredOutput"}
	remindersEnabledKey  = ctxKey{name: "remindersEnabled"}
)

// WithUserID 把 user_id 写入 context,用于在 MCP 工具 handler 内取出。
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext 从 context 中读出 user_id;不存在时返回 (0, false)。
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

// WithStructuredOutput 把"是否启用结构化输出"开关写入 context。
// 启用时工具返回沿用 NewToolResultStructured(数据+fallback) 行为;
// 未启用时工具把完整 JSON 字符串塞进 content[0].text,不设置 structuredContent。
func WithStructuredOutput(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, structuredOutputKey, enabled)
}

// StructuredOutputEnabled 从 context 中读取结构化输出开关;未设置视为 false。
func StructuredOutputEnabled(ctx context.Context) bool {
	v, _ := ctx.Value(structuredOutputKey).(bool)
	return v
}

// WithRemindersEnabled 把"是否暴露提醒配置工具"开关写入 context。
// 启用时 tools/list 显示 5 个 reminder_config 工具且允许 tools/call;
// 未启用时 tools/list 隐藏它们且 tools/call 直接拒绝。
func WithRemindersEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, remindersEnabledKey, enabled)
}

// RemindersEnabled 从 context 中读取提醒工具开关;未设置视为 false。
func RemindersEnabled(ctx context.Context) bool {
	v, _ := ctx.Value(remindersEnabledKey).(bool)
	return v
}
