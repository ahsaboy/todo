package mcp

import "context"

// ctxKey 是包私有的 context key 类型,避免与其他包发生冲突。
type ctxKey struct{}

var userIDKey = ctxKey{}

// WithUserID 把 user_id 写入 context,用于在 MCP 工具 handler 内取出。
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext 从 context 中读出 user_id;不存在时返回 (0, false)。
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}
