package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// headerEnabled 实现宽松解析:任意非空(去掉首尾空格后)字符串都视为开启。
// 缺失或仅空白的值视为未启用。
func headerEnabled(value string) bool {
	return strings.TrimSpace(value) != ""
}

// buildToolResult 根据 ctx 中的 StructuredOutput 开关,决定返回格式:
//   - 开启:走 mcp-go 原生 NewToolResultStructured(data, fallback),
//     content 放摘要 fallback,structuredContent 放完整对象。
//   - 关闭(默认):把 data 序列化为带缩进的 JSON 字符串,塞进 content[0].text,
//     不设置 structuredContent。
func buildToolResult(ctx context.Context, data any, fallback string) (*mcpgo.CallToolResult, error) {
	if StructuredOutputEnabled(ctx) {
		return mcpgo.NewToolResultStructured(data, fallback), nil
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}
	return mcpgo.NewToolResultText(string(raw)), nil
}
