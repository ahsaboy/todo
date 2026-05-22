package mcp

import (
	"context"
	"fmt"

	"todo/internal/service"
	"todo/internal/views"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"
)

// registerUserTools 把用户信息工具注册到 MCP server。
// 当前仅暴露 get_user_profile;其他写操作(改密码等)留在 REST API,不通过 MCP 工具开放。
func registerUserTools(s *mcpsrv.MCPServer, svc service.AuthServiceInterface) {
	if s == nil || svc == nil {
		return
	}
	s.AddTool(buildGetUserProfileTool(), getUserProfileHandler(svc))
}

func buildGetUserProfileTool() mcpgo.Tool {
	return mcpgo.NewTool("get_user_profile",
		mcpgo.WithDescription("获取当前已认证用户的个人资料(id/username/email/created_at)。不返回密码哈希,无需入参。"),
	)
}

func getUserProfileHandler(svc service.AuthServiceInterface) mcpsrv.ToolHandlerFunc {
	return func(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		userID, errResult := requireUserID(ctx)
		if errResult != nil {
			return errResult, nil
		}
		user, err := svc.GetUserByID(ctx, userID)
		if err != nil {
			return mcpgo.NewToolResultErrorf("failed to get user profile: %v", err), nil
		}
		if user == nil {
			return mcpgo.NewToolResultError("user not found"), nil
		}
		resp := views.UserResponseView(user.ToResponse(), resolveLoc(ctx))
		fallback := fmt.Sprintf("user #%d %s", resp.ID, resp.Username)
		return buildToolResult(ctx, resp, fallback)
	}
}
