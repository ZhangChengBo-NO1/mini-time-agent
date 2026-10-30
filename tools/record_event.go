package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
)

func NewRecordEvent() *ToolDef {
	return &ToolDef{
		Name: "record_event",
		Description: "记录一个带日期的事件（生日、纪念日、考试、会议、旅行等）。" +
			"同名事件会被覆盖。记录后可用 days_until 查询距离还有多少天",
		Params: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]string{
					"type":        "string",
					"description": "事件名称，例如：交房租、小明生日、期末考试",
				},
				"date": map[string]string{
					"type":        "string",
					"description": "事件日期，格式 YYYY-MM-DD，例如 2026-11-01",
				},
				"description": map[string]string{
					"type":        "string",
					"description": "可选，事件备注",
				},
			},
			"required": []string{"name", "date"},
		},
		Fn: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			date, _ := args["date"].(string)
			desc, _ := args["description"].(string)

			if name == "" || date == "" {
				return "", fmt.Errorf("name 和 date 都是必填项")
			}
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return "", fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
			}

			eventMu.Lock()
			defer eventMu.Unlock()

			events, err := loadEvents()
			if err != nil {
				return "", err
			}

			// 同名事件覆盖，否则新增
			updated := false
			for i := range events {
				if events[i].Name == name {
					events[i].Date = date
					events[i].Description = desc
					updated = true
					break
				}
			}
			if !updated {
				events = append(events, Event{Name: name, Date: date, Description: desc})
			}

			if err := saveEvents(events); err != nil {
				return "", err
			}

			action := "已记录"
			if updated {
				action = "已更新"
			}
			return fmt.Sprintf("%s事件：%s（%s）", action, name, date), nil
		},
	}
}