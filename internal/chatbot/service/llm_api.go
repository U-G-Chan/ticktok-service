package service

import (
	"context"
	"errors"
	"io"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"

	openai "github.com/sashabaranov/go-openai"
)

type LLMAPIService struct {
	client *openai.Client
	model  string
}

func NewLLMAPIService() *LLMAPIService {
	conf := openai.DefaultConfig(config.Config.LLM.APIKey)
	if config.Config.LLM.BaseURL != "" {
		conf.BaseURL = config.Config.LLM.BaseURL
	}
	client := openai.NewClientWithConfig(conf)

	return &LLMAPIService{
		client: client,
		model:  config.Config.LLM.Model,
	}
}

// GenerateStream wraps the OpenAI streaming completion API
func (s *LLMAPIService) GenerateStream(ctx context.Context, session *ChatSession, onChunk func(string) error) error {
	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    s.buildMessages(session),
		Stream:      true,
		Temperature: 0.7, // Default temperature, can be parameterized if needed
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.Log.Error("ChatCompletionStream error: " + err.Error())
		return err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			logger.Log.Error("Stream receive error: " + err.Error())
			return err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				if err := onChunk(content); err != nil {
					return err
				}
			}
		}
	}
}

func (s *LLMAPIService) buildMessages(session *ChatSession) []openai.ChatCompletionMessage {
	// Use the full conversation history stored in the session
	if len(session.Messages) > 0 {
		return session.Messages
	}

	// Fallback to just the prompt if no history is present (though session.Messages should ideally be populated)
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: session.Prompt,
		},
	}
}

// Update: To support history, we should probably allow passing messages or have them in session
func (s *LLMAPIService) GenerateStreamWithMessages(ctx context.Context, messages []openai.ChatCompletionMessage, onChunk func(string) error) error {
	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.7,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				if err := onChunk(content); err != nil {
					return err
				}
			}
		}
	}
}
