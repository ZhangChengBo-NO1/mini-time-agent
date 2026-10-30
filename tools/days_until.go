package tools

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
)

func NewDaysUntil() *ToolDef {
	return &ToolDef{
		Name: "days_until",
		Description:"查询某个已记录事件的具体日期（几号）、距离该事件还有多少天，或已经过了多少天。" +
    "当用户询问关于已记录事件的日期、天数时，必须调用此工具。支持模糊匹配。",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]string{
					"type":        "string",
					"description": "事件名称。必须从用户问题中提取。支持模糊匹配。不传则列出所有事件",
				},
				"date": map[string]string{
					"type":        "string",
					"description": "可选，直接传日期 YYYY-MM-DD 查询",
				},
			},
			"required": []string{},
		},
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			dateStr, _ := args["date"].(string)

			if dateStr != "" && name == "" {
				return formatDaysUntil("该日期", dateStr)
			}

			eventMu.Lock()
			events, err := loadEvents()
			eventMu.Unlock()
			if err != nil {
				return "", err
			}
			if len(events) == 0 {
				return "目前没有任何已记录的事件", nil
			}

			if name != "" {
				var matchedLines []string
				for _, ev := range events {
					// ✅ 核心：双向模糊匹配
					if strings.Contains(ev.Name, name) || strings.Contains(name, ev.Name) {
						line, _ := formatDaysUntil(ev.Name, ev.Date)
						matchedLines = append(matchedLines, line)
					}
				}

				if len(matchedLines) > 0 {
					return strings.Join(matchedLines, "\n"), nil
				}

				// 如果没匹配到，列出所有事件
				var allLines []string
				for _, ev := range events {
					line, _ := formatDaysUntil(ev.Name, ev.Date)
					allLines = append(allLines, line)
				}
				return fmt.Sprintf("未找到包含「%s」的事件。当前记录的所有事件如下，请确认您要查询的是哪一个：\n%s", name, strings.Join(allLines, "\n")), nil
			}

			// 列出全部
			lines := make([]string, 0, len(events))
			for _, ev := range events {
				line, _ := formatDaysUntil(ev.Name, ev.Date)
				lines = append(lines, line)
			}
			return strings.Join(lines, "\n"), nil
		},
	}
}

func formatDaysUntil(name, date string) (string, error) {
	target, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	diff := int(math.Round(target.Sub(today).Hours() / 24))

	switch {
	case diff > 0:
		return fmt.Sprintf("%s（%s）还有 %d 天", name, date, diff), nil
	case diff == 0:
		return fmt.Sprintf("%s（%s）就是今天", name, date), nil
	default:
		return fmt.Sprintf("%s（%s）已经过去 %d 天", name, date, -diff), nil
	}
}