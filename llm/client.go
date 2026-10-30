package llm

import (
	"context"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client struct {
	client openai.Client
	model  string
}

func NewClient(apiKey, baseURL, model string, timeout time.Duration) *Client {
	c := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithRequestTimeout(timeout),
	)
	return &Client{client: c, model: model}
}

func (c *Client) CreateChatCompletion(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []openai.ChatCompletionToolUnionParam,
) (*openai.ChatCompletion, error) {
	return c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Tools:    tools,
		Model:    c.model,
	})
}