package main

import (
	"context"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mini-time-agent/agent"
	"mini-time-agent/llm"
	"mini-time-agent/mcp"
	"mini-time-agent/tools"
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
		"deepseek-ai/DeepSeek-V3",
		30*time.Second,
	)

	// 1. 初始化基础 Agent（只包含内置工具）
	a := agent.NewAgent(llmClient)

	// 2. 初始化 MCP 管理器
	mgr, err := mcp.NewManager(context.Background(), "mcp.json")
	if err != nil {
		log.Printf("[mcp] 初始化失败（已忽略）: %v", err)
	} else {
		defer mgr.Close() // 确保程序退出时关闭子进程

		// 3. 把 MCP 发现的工具包装并注册给 Agent
		var mcpTools []*tools.ToolDef
		for _, dt := range mgr.AllTools() {
			// 加上前缀，例如 almanac_get_date_info
			fnName := mcp.SanitizeName(dt.Server, dt.Tool.Name)
			mcpTools = append(mcpTools,
				mcp.WrapAsToolDef(dt.Session, dt.Tool.Name, fnName, dt.Tool),
			)
		}
		a.RegisterTools(mcpTools)
		log.Printf("[mcp] 已加载 %d 个 MCP 工具", len(mcpTools))
	}

	// 4. 启动 UI 或 REPL
	if len(os.Args) > 1 && os.Args[1] == "tui" {
		p := tea.NewProgram(ui.NewTUI(a), tea.WithAltScreen())
		a.SetProgram(p)
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
		return
	}

	ui.RunREPL(a)
}