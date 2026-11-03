<div align="center">

# ⏱️ Mini Time Agent

**一个基于大模型的极简终端时间与日程管理智能体**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Bubble Tea](https://img.shields.io/badge/Bubble_Tea-TUI-FF75B7?style=for-the-badge&logo=charm&logoColor=white)](https://github.com/charmbracelet/bubbletea)
[![OpenAI](https://img.shields.io/badge/OpenAI-SDK-412991?style=for-the-badge&logo=openai&logoColor=white)](https://github.com/openai/openai-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge)](LICENSE)

*通过自然语言对话，在终端里轻松记录、查询和管理你的重要日程。*

</div>

---

## ✨ 项目简介

**Mini Time Agent** 是一个运行在命令行终端中的 AI 助手。你不需要记忆复杂的命令，只需像聊天一样输入“*帮我记一下下个月5号交房租*”或“*距离我生日还有多少天？*”，Agent 就会**自主思考**并**调用底层工具**来完成你的需求。

项目内置了高颜值的全屏 TUI 界面（基于 `bubbletea`）和极简的 REPL 命令行交互模式，并将数据安全、轻量地保存在本地的 `events.json` 文件中。

## 🚀 核心特性

- **🧠 自主决策 Agent 循环**：基于 LLM 的 Function Calling 能力，Agent 会自动判断是否需要调用工具，以及调用哪个工具（内置最多 10 次循环防止死循环）。
- **🎨 优雅的终端 UI**：支持全屏 TUI 模式和 REPL 模式。TUI 模式包含动态 Spinner、工具调用状态实时推流展示、历史消息滚动等。
- **🛠️ 强大的内置工具集**：
  - `get_current_time`：获取当前系统真实时间。
  - `record_event`：记录/更新带日期的事件。
  - `days_until`：计算距离目标日期的天数（支持模糊匹配）。
  - `update_event`：修改事件的名称、日期或备注。
  - `delete_event`：安全删除事件（多结果时要求二次确认）。
- **💾 本地持久化与并发安全**：使用 `events.json` 存储数据，并通过 `sync.Mutex` 保证多协程下的读写安全。
- **🛡️ 精准的安全拦截**：在 Agent 层拦截底层特殊 Token 泄漏，确保输出稳定干净。

## 🛠️ 技术栈

| 分类 | 技术选型 |
| :--- | :--- |
| **开发语言** | [Go (Golang)](https://go.dev/) |
| **终端 UI** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) / [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| **大模型 SDK** | [openai-go/v3](https://github.com/openai/openai-go) |
| **模型提供方** | SiliconFlow (DeepSeek-V3) |
| **数据存储** | 本地 JSON 文件 |

---

## 📦 快速开始

### 1. 克隆项目
```bash
git clone https://github.com/你的用户名/mini-time-agent.git
cd mini-time-agent
