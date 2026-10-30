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
					fmt.Printf("  [result] %s\n", e.Content)
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