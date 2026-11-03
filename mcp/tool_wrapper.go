package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	openai "github.com/openai/openai-go/v3"

	"mini-time-agent/tools"
)

func WrapAsToolDef(
	session *mcp.ClientSession,
	originalName string,
	fnName string,
	tool *mcp.Tool,
) *tools.ToolDef {
	// 把 MCP 提供的 JSON Schema 转成 OpenAI 要求的格式
	// ✅ 核心修改：增强工具描述
	desc := tool.Description
	if strings.Contains(fnName, "almanac") {
		desc = "【黄历/农历查询工具】" + desc + "。可用于查询某天的农历、节气、是否节假日、是否工作日（调休）。"
	} else if strings.Contains(fnName, "build_ics") || strings.Contains(strings.ToLower(tool.Description), "ics") {
		desc = "【日历文件生成工具】" + desc + "。传入事件列表，生成标准的 .ics 日历文件。必须传入 output_path（保存路径，如 ics/日程.ics）。"
	}

	params := convertInputSchema(tool.InputSchema)
	return &tools.ToolDef{
		Name:        fnName,
		Description: desc,
		Params:      params,
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			// 收到大模型的调用，转发给 MCP 服务
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      originalName,
				Arguments: args,
			})
			if err != nil {
				return "", fmt.Errorf("MCP 调用 %s 失败: %w", originalName, err)
			}
			if result.IsError {
				return "", fmt.Errorf("MCP 工具 %s 返回错误: %s", originalName, mcpResultToText(result))
			}
			return mcpResultToText(result), nil
		},
	}
}

func convertInputSchema(schema any) openai.FunctionParameters {
	if schema == nil {
		return openai.FunctionParameters{"type": "object", "properties": map[string]any{}}
	}
	data, err := json.Marshal(schema)
	if err != nil {
		return openai.FunctionParameters{"type": "object", "properties": map[string]any{}}
	}
	var params openai.FunctionParameters
	if err := json.Unmarshal(data, &params); err != nil {
		return openai.FunctionParameters{"type": "object", "properties": map[string]any{}}
	}
	if _, ok := params["type"]; !ok {
		params["type"] = "object"
	}
	return params
}

func mcpResultToText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return "（MCP 工具没有返回内容）"
	}
	var out string
	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			out += text.Text
		}
	}
	if out == "" {
		return "（返回了非文本内容，暂不支持展示）"
	}
	return out
}