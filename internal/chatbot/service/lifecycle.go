package service

import (
	"context"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type ChatSession struct {
	SessionID     string
	UserID        int64
	Prompt        string
	Messages      []openai.ChatCompletionMessage // To hold history
	FullResponse  strings.Builder
	TotalTokens   int
	TraceID       string
	IsIntercepted bool
	CreatedAt     time.Time
}

type BeforeHook func(ctx context.Context, session *ChatSession) error
type InHook func(ctx context.Context, session *ChatSession, chunk string) error
type AfterHook func(ctx context.Context, session *ChatSession, genErr error)

type LifecycleManager struct {
	BeforeHooks []BeforeHook
	InHooks     []InHook
	AfterHooks  []AfterHook
}

func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		BeforeHooks: make([]BeforeHook, 0),
		InHooks:     make([]InHook, 0),
		AfterHooks:  make([]AfterHook, 0),
	}
}

func (lm *LifecycleManager) RegisterBeforeHook(hook BeforeHook) {
	lm.BeforeHooks = append(lm.BeforeHooks, hook)
}

func (lm *LifecycleManager) RegisterInHook(hook InHook) {
	lm.InHooks = append(lm.InHooks, hook)
}

func (lm *LifecycleManager) RegisterAfterHook(hook AfterHook) {
	lm.AfterHooks = append(lm.AfterHooks, hook)
}

func (lm *LifecycleManager) ExecuteBefore(ctx context.Context, session *ChatSession) error {
	for _, hook := range lm.BeforeHooks {
		if err := hook(ctx, session); err != nil {
			return err
		}
	}
	return nil
}

func (lm *LifecycleManager) ExecuteIn(ctx context.Context, session *ChatSession, chunk string) error {
	for _, hook := range lm.InHooks {
		if err := hook(ctx, session, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (lm *LifecycleManager) ExecuteAfter(ctx context.Context, session *ChatSession, genErr error) {
	for _, hook := range lm.AfterHooks {
		hook(ctx, session, genErr)
	}
}
