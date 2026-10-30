package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
)

func NewDeleteEvent() *ToolDef {
	return &ToolDef{
		Name: "delete_event",
		Description: "删除一个已记录的事件。支持模糊匹配（如删'买电脑'可以匹配到'买了电脑'）。" +
			"如果匹配到多个事件，会全部列出让用户确认，不会误删",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]string{
					"type":        "string",
					"description": "要删除的事件名称。必须严格从用户当前提问中提取（如：用户说'删掉游泳'，这里就填'游泳'）",
				},
				"confirm": map[string]string{
					"type":        "boolean",
					"description": "当匹配到多个事件时，需要用户确认后再传 true 执行删除；匹配唯一时可以不传",
				},
			},
			"required": []string{"name"},
		},
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			if name == "" {
				return "", fmt.Errorf("name 是必填项")
			}

			eventMu.Lock()
			defer eventMu.Unlock()

			events, err := loadEvents()
			if err != nil {
				return "", err
			}
			if len(events) == 0 {
				return "目前没有任何已记录的事件", nil
			}

			// 1. 模糊匹配
			var matched []Event
			for _, ev := range events {
				if strings.Contains(ev.Name, name) || strings.Contains(name, ev.Name) {
					matched = append(matched, ev)
				}
			}

			// 2. 没匹配到
			if len(matched) == 0 {
				var all []string
				for _, ev := range events {
					all = append(all, fmt.Sprintf("- %s（%s）", ev.Name, ev.Date))
				}
				return fmt.Sprintf("未找到包含「%s」的事件。当前所有事件：\n%s",
					name, strings.Join(all, "\n")), nil
			}

			// 3. 匹配到多个，列出让用户确认
			if len(matched) > 1 {
				var lines []string
				for _, ev := range matched {
					lines = append(lines, fmt.Sprintf("- %s（%s）", ev.Name, ev.Date))
				}
				return fmt.Sprintf("匹配到多个事件：\n%s\n\n请告诉用户具体删除哪一个（请用户提供完整名称）。",
					strings.Join(lines, "\n")), nil
			}

			// 4. 匹配到唯一一个，删除
			target := matched[0]
			var remaining []Event
			for _, ev := range events {
				if ev.Name != target.Name {
					remaining = append(remaining, ev)
				}
			}

			if err := saveEvents(remaining); err != nil {
				return "", err
			}

			return fmt.Sprintf("已删除事件：%s（%s），剩余 %d 个事件",
				target.Name, target.Date, len(remaining)), nil
		},
	}
}