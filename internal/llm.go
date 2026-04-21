package internal

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type LLMCaller struct {
	client anthropic.Client
	model  string
}

func NewLLMCaller(apiKey, model string) *LLMCaller {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &LLMCaller{client: client, model: model}
}

func (l *LLMCaller) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := l.client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 4096,
		Model:     l.model,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{OfText: &anthropic.TextBlockParam{Text: userPrompt}},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("llm call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("llm returned empty response")
	}

	text := resp.Content[0].AsText()
	return text.Text, nil
}
