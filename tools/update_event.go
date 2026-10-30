package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
)

func NewUpdateEvent() *ToolDef {
	return &ToolDef{
		Name: "update_event",
		Description: "修改一个已记录事件的信息（名称、日期、备注）。支持模糊匹配目标事件。" +
			"如果要改名称，用 new_name；如果要改日期，用 new_date；如果要改备注，用 new_description。" +
			"只需传需要修改的字段，未传的字段保持原样。如果匹配到多个事件，会全部列出让用户确认",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]string{
					"type":        "string",
					"description": "要修改的目标事件名称（模糊匹配）。必须严格从用户当前提问中提取",
				},
				"new_name": map[string]string{
					"type":        "string",
					"description": "可选，新的名称。不传则不改名称",
				},
				"new_date": map[string]string{
					"type":        "string",
					"description": "可选，新的日期，格式 YYYY-MM-DD。不传则不改日期",
				},
				"new_description": map[string]string{
					"type":        "string",
					"description": "可选，新的备注。不传则不改备注",
				},
			},
			"required": []string{"name"},
		},
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			newName, _ := args["new_name"].(string)
			newDate, _ := args["new_date"].(string)
			newDesc, _ := args["new_description"].(string)

			if name == "" {
				return "", fmt.Errorf("name 是必填项")
			}
			if newName == "" && newDate == "" && newDesc == "" {
				return "", fmt.Errorf("至少要传 new_name、new_date 或 new_description 中的一个")
			}
			if newDate != "" {
				if _, err := time.Parse("2006-01-02", newDate); err != nil {
					return "", fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
				}
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

			// 1. 模糊匹配目标
			var matchedIdx []int
			for i, ev := range events {
				if strings.Contains(ev.Name, name) || strings.Contains(name, ev.Name) {
					matchedIdx = append(matchedIdx, i)
				}
			}

			// 2. 没匹配到
			if len(matchedIdx) == 0 {
				var all []string
				for _, ev := range events {
					all = append(all, fmt.Sprintf("- %s（%s）", ev.Name, ev.Date))
				}
				return fmt.Sprintf("未找到包含「%s」的事件。当前所有事件：\n%s",
					name, strings.Join(all, "\n")), nil
			}

			// 3. 匹配到多个，列出让用户确认
			if len(matchedIdx) > 1 {
				var lines []string
				for _, i := range matchedIdx {
					lines = append(lines, fmt.Sprintf("- %s（%s）", events[i].Name, events[i].Date))
				}
				return fmt.Sprintf("匹配到多个事件：\n%s\n\n请告诉用户具体修改哪一个（请用户提供完整名称）。",
					strings.Join(lines, "\n")), nil
			}

			// 4. 匹配到唯一一个，执行修改
			idx := matchedIdx[0]
			old := events[idx]

			// 检查新名称是否与其他事件冲突
			if newName != "" && newName != old.Name {
				for i, ev := range events {
					if i != idx && ev.Name == newName {
						return "", fmt.Errorf("已存在同名事件「%s」，请换一个名称", newName)
					}
				}
				events[idx].Name = newName
			}
			if newDate != "" {
				events[idx].Date = newDate
			}
			if newDesc != "" {
				events[idx].Description = newDesc
			}

			if err := saveEvents(events); err != nil {
				return "", err
			}

			// 组织反馈
			var changes []string
			if newName != "" && newName != old.Name {
				changes = append(changes, fmt.Sprintf("名称：%s → %s", old.Name, newName))
			}
			if newDate != "" && newDate != old.Date {
				changes = append(changes, fmt.Sprintf("日期：%s → %s", old.Date, newDate))
			}
			if newDesc != "" && newDesc != old.Description {
				changes = append(changes, fmt.Sprintf("备注：%q → %q", old.Description, newDesc))
			}

			return fmt.Sprintf("已更新事件「%s」：%s",
				events[idx].Name, strings.Join(changes, "；")), nil
		},
	}
}