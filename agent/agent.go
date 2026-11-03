package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mini-time-agent/core"
	"mini-time-agent/llm"
	"mini-time-agent/tools"

	"github.com/openai/openai-go/v3"
)

const MaxLoop = 10

// ✅ 更新了 prompt，加入了 almanac 和 syllabus-calendar 的说明
var systemPrompt = fmt.Sprintf(`你是一个智能的时间与日程管理助手。回答简洁、中文输出。

【当前系统真实时间】：%s。

你可以使用外部工具来辅助完成任务。请根据用户的意图**自主决定**是否需要调用工具，以及调用哪个工具。

请遵循以下原则：
1. **直接回答**：简单的日期时间问题（如“今天几号”），直接使用【系统真实时间】回答。
2. **数据获取与导出**：当用户要求整理、导出、查看行程时，**绝对禁止向用户索要数据**。你必须先阅读工具描述，找到能“查询所有事件”的工具去获取数据，然后再找能“生成日历文件”的工具去导出。
3. **参数严格性**：参数必须严格从用户当前提问中提取，绝对不要盲目使用历史事件。
4. **日期推算**：必须严格基于系统真实时间推算 YYYY-MM-DD。
5. **确认机制**：删除/修改多个匹配结果时，必须列出让用户确认。
6. **农历/黄历判断**：当用户询问农历日期、节气、法定节假日、调休或工作日时，请寻找描述中包含“农历”、“黄历”或“中国日历”等字眼的工具。
7. **导出日历文件**：当需要生成 .ics 文件时，寻找描述中包含“生成”、“导出”或“ICS”字眼的工具。**极其重要**：调用此类工具时，必须传入 output_path 参数！如果用户没有指定具体路径，你必须自动将 output_path 设为 ics/日程.ics。绝对不要把大段的 ICS 代码直接输出给用户。`,
	time.Now().Format("2006-01-02 15:04:05"))

type Event struct {
	Type    string
	Content string
}

type OnEvent func(Event)

type Agent struct {
	llm      *llm.Client
	tools    map[string]*tools.ToolDef
	toolDefs []openai.ChatCompletionToolUnionParam
	session  *core.Session
	program  *tea.Program
}

func NewAgent(llmClient *llm.Client) *Agent {
	toolMap := tools.GetAllTools()
	defs := tools.BuildToolDefs(toolMap)

	s := core.NewSession()
	s.AddMessage(openai.SystemMessage(systemPrompt))

	return &Agent{
		llm:      llmClient,
		tools:    toolMap,
		toolDefs: defs,
		session:  s,
	}
}

// ✅ 新增：动态注册外部工具（MCP 工具）
func (a *Agent) RegisterTools(extra []*tools.ToolDef) {
	merged := tools.MergeExternal(a.tools, extra)
	a.tools = merged
	a.toolDefs = tools.BuildToolDefs(merged)
}

func (a *Agent) SetProgram(p *tea.Program) {
	a.program = p
}

func (a *Agent) emit(e Event, onEvent OnEvent) {
	if a.program != nil {
		a.program.Send(e)
	}
	if onEvent != nil {
		onEvent(e)
	}
}

func (a *Agent) Run(ctx context.Context, userInput string, onEvent OnEvent) (string, error) {
	a.session.AddMessage(openai.UserMessage(userInput))

	for i := 0; i <= MaxLoop; i++ {
		resp, err := a.llm.CreateChatCompletion(ctx, a.session.GetMessages(), a.toolDefs)
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("大模型没有返回回答")
		}

		msg := resp.Choices[0].Message
		content := strings.TrimSpace(msg.Content)

		if strings.Contains(content, "<|tool_call_begin|>") || strings.Contains(content, "</tool_call>") {
			return "（系统拦截：大模型输出了底层乱码，请重新输入）", nil
		}

		assistantMsg := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(msg.Content),
				},
			},
		}

		if len(msg.ToolCalls) > 0 {
			var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
			for _, tc := range msg.ToolCalls {
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					},
				})
			}
			assistantMsg.OfAssistant.ToolCalls = toolCalls
		}

		a.session.AddMessage(assistantMsg)

		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		for _, call := range msg.ToolCalls {
			a.emit(Event{
				Type:    "tool_call",
				Content: fmt.Sprintf("%s(%s)", call.Function.Name, call.Function.Arguments),
			}, onEvent)

			var result string
			if t, ok := a.tools[call.Function.Name]; ok {
				r, err := t.Call(ctx, call.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("错误: %v", err)
				} else {
					result = r
				}
			} else {
				result = fmt.Sprintf("未找到工具: %s", call.Function.Name)
			}

			a.emit(Event{Type: "tool_result", Content: result}, onEvent)

			a.session.AddMessage(openai.ToolMessage(result, call.ID))
		}
	}

	return "", fmt.Errorf("达到最大循环次数 %d", MaxLoop)
}