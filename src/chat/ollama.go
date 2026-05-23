package chat

import (
	"context"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	client *api.Client
	model  string
}

func NewOllamaClient(model string) (*OllamaClient, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &OllamaClient{
		client: client,
		model:  model,
	}, nil
}

func (oc *OllamaClient) ChatStream(ctx context.Context, messages []api.Message, onChunk func(string) error) error {
	req := &api.ChatRequest{
		Model:    oc.model,
		Messages: messages,
		Stream:   ptr(true),
	}

	respFunc := func(resp api.ChatResponse) error {
		return onChunk(resp.Message.Content)
	}

	return oc.client.Chat(ctx, req, respFunc)
}

func (oc *OllamaClient) SinglePrompt(ctx context.Context, prompt string) (string, error) {
	var fullResponse string
	req := &api.GenerateRequest{
		Model:  oc.model,
		Prompt: prompt,
		Stream: ptr(false),
	}

	respFunc := func(resp api.GenerateResponse) error {
		fullResponse = resp.Response
		return nil
	}

	err := oc.client.Generate(ctx, req, respFunc)
	return fullResponse, err
}

func ptr[T any](v T) *T {
	return &v
}
