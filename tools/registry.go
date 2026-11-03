package tools

import (
	openai "github.com/openai/openai-go/v3"
)

func GetAllTools() map[string]*ToolDef {
	all := []*ToolDef{
		NewGetCurrentTime(),
		NewRecordEvent(),
		NewDaysUntil(),
		NewDeleteEvent(),
		NewUpdateEvent(),
	}

	m := make(map[string]*ToolDef, len(all))
	for _, t := range all {
		m[t.Name] = t
	}
	return m
}

// MergeExternal 将 MCP 工具合并到内置工具中
func MergeExternal(base map[string]*ToolDef, extra []*ToolDef) map[string]*ToolDef {
	merged := make(map[string]*ToolDef, len(base)+len(extra))
	for k, v := range base {
		merged[k] = v
	}
	for _, t := range extra {
		merged[t.Name] = t
	}
	return merged
}

// BuildToolDefs 统一生成给大模型看的工具描述
func BuildToolDefs(toolMap map[string]*ToolDef) []openai.ChatCompletionToolUnionParam {
	defs := make([]openai.ChatCompletionToolUnionParam, 0, len(toolMap))
	for _, t := range toolMap {
		defs = append(defs, t.ToParam())
	}
	return defs
}