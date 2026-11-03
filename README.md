# ⏱️ Mini Time Agent

> **一个基于大模型与 MCP 协议的极简终端时间与日程管理智能体**

[![Go Version](https://img.shields.io/badge/Go-1.26.5+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Bubble Tea](https://img.shields.io/badge/TUI-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![MCP Supported](https://img.shields.io/badge/MCP-Supported-success?style=flat-square)](https://modelcontextprotocol.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)

通过自然语言对话，在终端里轻松记录、查询、管理日程，甚至**一键导出手机可用的日历文件**。

---

## ✨ 项目简介

**Mini Time Agent** 是一个运行在命令行终端中的全能 AI 日程助手。你不需要记忆复杂的命令，只需像聊天一样输入 *“帮我记一下下个月5号交房租”* 或 *“把这周的行程导出成日历文件”*，Agent 就会自主思考，并动态调用底层工具完成任务。

项目内置了高颜值的全屏 TUI 界面（基于 `bubbletea`）和极简的 REPL 交互模式。数据安全地保存在本地 `events.json` 中。

**🔌 最大的亮点在于 MCP 架构的引入**：项目原生支持 Model Context Protocol，你可以通过修改一个简单的 `mcp.json` 配置文件，无缝接入社区海量的第三方 MCP 工具，让 Agent 的能力无限延伸。

---

## 🚀 核心特性

### 🧠 智能与自主
- **自主决策 Agent 循环**：基于 LLM 的 Function Calling 能力，Agent 会自动分析用户意图。它不依赖死板的固定流程，而是通过阅读工具的“说明书（Description）”自主判断是否需要调用工具，以及调用哪一个。
- **精准的安全拦截**：在 Agent 层拦截大模型底层特殊 Token 泄漏，确保终端输出稳定干净。

### 🧩 强大的 MCP 扩展生态
- **动态插件系统**：通过 `mcp.json` 一键拉起后台服务进程，基于 stdio 协议与外部 MCP 服务通信。
- **黄历与时序查询**（内置接入 `chinese-almanac-mcp`）：精准查询农历、24 节气、法定节假日及调休安排。
- **日程一键导出**（内置接入 `syllabus-calendar-mcp`）：整理本地数据，自动生成标准的 `.ics` 日历文件。**完美兼容小米、苹果、华为手机以及 Windows 系统日历**，扫码或双击即可导入。

### 🛠️ 丰富的内置原生工具

| 工具名 | 功能描述 |
| :--- | :--- |
| `get_current_time` | 获取当前系统绝对真实时间。 |
| `record_event` | 记录/覆盖带日期的事件。 |
| `days_until` | 计算距离目标日期的天数（支持模糊匹配）。 |
| `update_event` | 修改事件的名称、日期或备注。 |
| `delete_event` | 安全删除事件（多结果时强制要求用户二次确认）。 |

### 🎨 优雅的终端 UI
- **全屏 TUI 模式**：包含动态 Spinner、工具调用实时推流展示（不再干等）、历史消息平滑滚动。
- **极简 REPL 模式**：适合极客的快速测试与高频交互。

### 💾 数据安全与持久化
- 所有日程使用本地 `events.json` 存储。
- 通过 `sync.Mutex` 保证多协程下的并发读写安全，防止文件损坏。

---

## 🛠️ 技术栈

| 分类 | 技术选型 |
| :--- | :--- |
| **开发语言** | Go (Golang) 1.26+ |
| **终端 UI** | Bubble Tea / Lip Gloss |
| **大模型 SDK** | openai-go/v3 |
| **模型提供方** | SiliconFlow (DeepSeek-V3) |
| **MCP 协议** | modelcontextprotocol/go-sdk |
| **外部插件环境** | Node.js (npx) |
| **数据存储** | 本地 JSON 文件 |

---

## 📦 快速开始

### 1. 环境准备
- 安装 [Go](https://go.dev/dl/) (推荐 1.26+)
- 安装 [Node.js](https://nodejs.org/) (推荐 18+，用于运行 MCP 服务)
- 获取 `SILICONFLOW_API_KEY` (用于调用大模型)

### 2. 克隆项目
```bash
git clone https://github.com/你的用户名/mini-time-agent.git
cd mini-time-agent
```

### 3. 配置环境变量
在终端中设置你的 API Key：

```bash
# macOS / Linux
export SILICONFLOW_API_KEY="sk-你的密钥"

# Windows (PowerShell)
$env:SILICONFLOW_API_KEY="sk-你的密钥"
```

### 4. 配置 MCP 插件
在项目根目录创建 `mcp.json` 文件，填入你想启用的扩展服务（项目推荐配置如下）：

```json
{
  "mcpServers": {
    "almanac": {
      "command": "npx",
      "args": ["-y", "chinese-almanac-mcp@latest"]
    },
    "syllabus-calendar": {
      "command": "npx",
      "args": ["-y", "syllabus-calendar-mcp"]
    }
  }
}
```

### 5. 编译与运行
```bash
# 下载依赖
go mod tidy

# 以 TUI 模式启动（推荐）
go run . tui

# 以极简 REPL 模式启动
go run .
```

---

## 💡 对话示例

- **日常查询**：
  - *“今天几号？是工作日吗？”*
  - *“距离我生日还有多少天？”*
- **日程管理**：
  - *“帮我记录一下，11月15号我要去北京出差。”*
  - *“把交房租的日程改到每个月的5号。”*
- **🤖 MCP 特色能力**：
  - *“查一下这周六的农历和节气，适合搬家吗？”*（Agent 自动调用黄历 MCP）
  - *“帮我把这周的行程生成日历文件，保存到当前目录下的 week.ics。”*（Agent 先查本地数据，再调用日历 MCP 生成文件，直接发到手机上即可导入）

---

## 🏗️ 架构设计

本项目采用了 **Agent 循环 + MCP 动态工具桥接** 的设计模式。大模型只需要阅读工具说明书，无需修改核心代码即可获得新能力。

```text
mini-time-agent/
├── main.go               # 程序入口，初始化 Agent 和 MCP 管理器
├── mcp.json              # 外部 MCP 插件配置文件
├── events.json           # 本地日程数据文件
├── agent/                # Agent 核心循环、Prompt 控制与安全拦截
├── core/                 # 会话上下文管理 (Session)
├── llm/                  # 大语言模型 API 封装
├── tools/                # 原生内置工具 (增删改查日程)
│   ├── record_event.go
│   └── days_until.go
├── mcp/                  # ✨ MCP 桥接模块 (核心亮点)
│   ├── manager.go        # 负责拉起 Node.js 子进程，建立 stdio 通信管道
│   └── tool_wrapper.go   # 协议适配器，将 MCP 工具伪装成 Go 原生 ToolDef
└── ui/                   # REPL 与 TUI 界面
```

### 🔌 MCP 模块工作原理

1. **发现 (Discovery)**：启动时，`manager.go` 读取 `mcp.json`，通过 `os/exec` 拉起子进程，利用 stdio 管道获取外部工具列表。
2. **适配 (Adaption)**：`tool_wrapper.go` 将外部 MCP 的 JSON Schema 转换为 Go 的 `FunctionParameters`，并在 Description 中加入如 `【黄历查询】` 的中文前缀，大幅提升大模型的理解命中率。
3. **执行 (Execution)**：大模型决策后，适配器将参数通过管道转发给子进程。外部进程算完后回传结果，整个过程对 Agent 主循环完全透明，实现了真正的“插件化”。

---

## 📄 开源协议

本项目基于 [MIT License](https://opensource.org/licenses/MIT) 协议开源。欢迎提交 Issue 和 Pull Request！
