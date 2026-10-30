package tools

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

const eventFile = "events.json"

// Event 一条事件记录。
type Event struct {
	Name        string `json:"name"`
	Date        string `json:"date"` // YYYY-MM-DD
	Description string `json:"description,omitempty"`
}

// 全局锁，防止并发读写 events.json。
var eventMu sync.Mutex

// loadEvents 从 events.json 读事件。
func loadEvents() ([]Event, error) {
	data, err := os.ReadFile(eventFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil // 文件不存在时返回空列表
		}
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// saveEvents 按日期排序后写入 events.json。
func saveEvents(events []Event) error {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Date < events[j].Date
	})

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(eventFile, data, 0644)
}