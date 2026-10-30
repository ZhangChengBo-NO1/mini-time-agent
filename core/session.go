package core

import "github.com/openai/openai-go/v3"

type Session struct {
	Messages []openai.ChatCompletionMessageParamUnion
	Len      int
}

func NewSession() *Session {
	return &Session{
		Messages: []openai.ChatCompletionMessageParamUnion{},
		Len:      0,
	}
}

func (s *Session) GetMessages() []openai.ChatCompletionMessageParamUnion {
	return s.Messages
}

func (s *Session) AddMessage(item openai.ChatCompletionMessageParamUnion) {
	s.Messages = append(s.Messages, item)
	s.Len++
}