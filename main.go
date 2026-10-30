package main

import (
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mini-time-agent/agent"
	"mini-time-agent/llm"
	"mini-time-agent/ui"
)

func main() {
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		log.Fatal("请先设置环境变量 SILICONFLOW_API_KEY")
	}

	llmClient := llm.NewClient(
		apiKey,
		"https://api.siliconflow.cn/v1",
		"deepseek-ai/DeepSeek-V3", // 或你正在用的模型
		30*time.Second,
	)

	a := agent.NewAgent(llmClient)

	if len(os.Args) > 1 && os.Args[1] == "tui" {
		p := tea.NewProgram(ui.NewTUI(a), tea.WithAltScreen())
		a.SetProgram(p) // ✅ 注入 program
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
		return
	}

	ui.RunREPL(a)
}