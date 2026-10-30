package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mini-time-agent/agent"
)

type responseMsg struct {
	text string
	err  error
}

type Model struct {
	agent    *agent.Agent
	textarea textarea.Model
	spinner  spinner.Model
	viewport viewport.Model
	messages []string
	loading  bool
	status   string // ✅ 新增：当前状态文字，例如 "调用 get_current_time"
	width    int
	height   int
}

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	userStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	assistantStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	toolStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	resultStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	mutedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func NewTUI(a *agent.Agent) Model {
	ta := textarea.New()
	ta.Placeholder = "请输入你的问题..."
	ta.SetHeight(3)
	ta.SetWidth(70)
	ta.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(76, 15)
	vp.MouseWheelEnabled = true

	messages := []string{
		titleStyle.Render("Mini Time Agent"),
		mutedStyle.Render("输入问题，Ctrl+C 退出。"),
		"",
	}
	vp.SetContent(strings.Join(messages, "\n"))

	return Model{
		agent:    a,
		textarea: ta,
		spinner:  sp,
		viewport: vp,
		messages: messages,
		status:   "",
		width:    80,
		height:   24,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) appendMessage(s string) Model {
	wasAtBottom := m.viewport.AtBottom()
	m.messages = append(m.messages, s)
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = max(20, msg.Width-6)
		m.viewport.Height = max(3, msg.Height-11)
		m.textarea.SetWidth(max(10, msg.Width-10))
		return m, nil

	// ✅ 核心：处理 Agent 推送过来的事件
	case agent.Event:
		switch msg.Type {
		case "tool_call":
			m.status = "正在调用 " + msg.Content // 更新状态栏
			m = m.appendMessage(toolStyle.Render("  ⚙ [tool] ") + msg.Content)
		case "tool_result":
			m.status = "获取结果中..."
			m = m.appendMessage(resultStyle.Render("  ✓ [result] ") + msg.Content)
		case "thinking":
			m.status = "思考中..."
			m = m.appendMessage(mutedStyle.Render("  ✻ [think] ") + msg.Content)
		}
		return m, nil

	case responseMsg:
		m.loading = false
		m.status = ""
		if msg.err != nil {
			m = m.appendMessage(errorStyle.Render("✗ Error") + "\n" + msg.err.Error() + "\n")
		} else {
			m = m.appendMessage(assistantStyle.Render("◆ Assistant") + "\n" + msg.text + "\n")
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.loading {
				return m, nil
			}
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				return m, nil
			}
			m = m.appendMessage(userStyle.Render("❯ You") + "\n" + input + "\n")
			m.textarea.Reset()
			m.loading = true
			m.status = "思考中..." // ✅ 初始状态

			a := m.agent
			return m, tea.Batch(
				spinner.Tick,
				func() tea.Msg {
					answer, err := a.Run(context.Background(), input, nil)
					return responseMsg{text: answer, err: err}
				},
			)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	m.textarea, cmd = m.textarea.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	width := max(20, m.width-4)

	header := titleStyle.Render("Mini Time Agent") + "  " +
		mutedStyle.Render("AI Terminal Assistant")

	// ✅ 状态栏：显示当前工具调用，而不是固定的 "Thinking..."
	status := ""
	if m.loading {
		label := m.status
		if label == "" {
			label = "Thinking..."
		}
		status = fmt.Sprintf("%s %s", m.spinner.View(), mutedStyle.Render(label))
	} else {
		status = mutedStyle.Render("Enter 发送 • Ctrl+C 退出")
	}

	content := strings.Join([]string{
		header,
		"",
		m.viewport.View(),
		"",
		status,
		"",
		m.textarea.View(),
	}, "\n")

	return lipgloss.NewStyle().Width(width).Render(content)
}