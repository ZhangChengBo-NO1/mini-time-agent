package tools

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v3"
)

type ToolDef struct {
	Name        string
	Description string
	Params      openai.FunctionParameters
	Fn          func(ctx context.Context, args map[string]any) (string, error)
}

func (t ToolDef) ToParam() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionToolUnionParam{
		OfFunction: &openai.ChatCompletionFunctionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  t.Params,
			},
		},
	}
}

// 把 LLM 给的 JSON 字符串参数解析成 map,再调 Fn
func (t ToolDef) Call(ctx context.Context, argsJSON string) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", err
	}
	return t.Fn(ctx, args)
}