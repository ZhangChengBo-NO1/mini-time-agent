package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"mini-time-agent/agent"
)

// RunREPL 启动一个最简单的命令行交互循环。
func RunREPL(a *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("========================================")
	fmt.Println(" Mini Time Agent (REPL)")
	fmt.Println(" 输入 exit 退出")
	fmt.Println("========================================")

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println("再见！")
			return
		}

		// ui/repl.go
		answer, err := a.Run(
			context.Background(),
			input,
			func(e agent.Event) {
				switch e.Type {
				case "thinking":
					fmt.Printf("  [think] %s\n", e.Content)
				case "tool_call":
					fmt.Printf("  [tool ] %s\n", e.Content)
				case "tool_result":
					// ✅ 修改这里：拦截过长的 JSON 或对象结构
					content := strings.TrimSpace(e.Content)
					if len(content) > 200 || strings.HasPrefix(content, "{") {
						fmt.Printf("  [result] 获取成功 (数据较长，已隐式传给 AI 处理)\n")
					} else {
						fmt.Printf("  [result] %s\n", content)
					}
				}
			},
		)

		if err != nil {
			fmt.Printf("\n[错误] %v\n", err)
			continue
		}

		fmt.Printf("\n%s\n", answer)
	}
}