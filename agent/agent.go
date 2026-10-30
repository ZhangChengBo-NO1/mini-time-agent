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

var systemPrompt = fmt.Sprintf(`你是一个智能的时间与日程管理助手。回答简洁、中文输出。

【当前系统真实时间】：%s。

你可以使用外部工具来辅助完成任务。请根据用户的意图**自主决定**是否需要调用工具，以及调用哪个工具。你的决策依据是每个工具的描述（Description）。

请遵循以下原则：
1. **直接回答**：对于非常简单的问题（如“今天几号”、“现在几点”），你可以直接使用上方提供的【系统真实时间】回答，无需调用工具。
2. **记忆查询**：当用户询问与已记录事件相关的天数（如“还有多少天”、“过了多少天”），**绝对禁止向用户索要日期**，必须主动调用相关工具去查询长期记忆。只有工具明确返回“未找到”时，才告知用户无记录。
3. **参数严格性**：在调用工具时，参数必须严格从用户【当前这一句】问题中提取，**绝对不要**盲目使用对话历史中提到的事件名称作为默认参数。
4. **日期推算**：如果用户表述相对日期（如“今年11月1号”），必须严格基于上方系统真实时间的年份进行推算，转换为标准的 YYYY-MM-DD 格式后再调用工具记录。
5. **确认机制**：在执行删除或修改操作时，如果工具返回匹配到多个结果，必须将候选结果原样列给用户，让用户明确指定要操作的对象，得到用户确认后再进行下一步工具调用。`,
	time.Now().Format("2006-01-02 15:04:05"))

// Event 是 Agent 执行过程中通知给 UI 的事件。
type Event struct {
	Type    string // "thinking" / "tool_call" / "tool_result"
	Content string
}

type OnEvent func(Event)

type Agent struct {
	llm      *llm.Client
	tools    map[string]*tools.ToolDef
	toolDefs []openai.ChatCompletionToolUnionParam
	session  *core.Session
	program  *tea.Program // ✅ 新增：用于向 TUI 推送事件
}

func NewAgent(llmClient *llm.Client) *Agent {
	toolMap := tools.GetAllTools()
	defs := make([]openai.ChatCompletionToolUnionParam, 0, len(toolMap))
	for _, t := range toolMap {
		defs = append(defs, t.ToParam())
	}

	s := core.NewSession()
	s.AddMessage(openai.SystemMessage(systemPrompt))

	return &Agent{
		llm:      llmClient,
		tools:    toolMap,
		toolDefs: defs,
		session:  s,
	}
}

// ✅ 新增：注入 Bubble Tea 程序，用于推送事件
func (a *Agent) SetProgram(p *tea.Program) {
	a.program = p
}

// emit 统一发送事件：优先推送到 TUI，同时回调 onEvent
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

		// 拦截底层特殊 Token 泄漏
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
			// ✅ 推送工具调用事件
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

			// ✅ 推送工具结果事件
			a.emit(Event{Type: "tool_result", Content: result}, onEvent)

			a.session.AddMessage(openai.ToolMessage(result, call.ID))
		}
	}

	return "", fmt.Errorf("达到最大循环次数 %d", MaxLoop)
}