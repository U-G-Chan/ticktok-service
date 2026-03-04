package service

import (
	"context"
	"fmt"
	"time"

	pb "ticktok-service/api/chatbot/v1"
	"ticktok-service/internal/chatbot/repository"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/mysql"
	redisPkg "ticktok-service/pkg/redis"
	"ticktok-service/pkg/util"

	openai "github.com/sashabaranov/go-openai"
)

type ChatbotService struct {
	pb.UnimplementedChatbotServiceServer
	conversationRepo *repository.ConversationRepo
	messageService   *MessageService
	llmAPI           *LLMAPIService
	lifecycleManager *LifecycleManager
}

func NewChatbotService() *ChatbotService {
	s := &ChatbotService{
		conversationRepo: repository.NewConversationRepo(mysql.DB),
		messageService:   NewMessageService(),
		llmAPI:           NewLLMAPIService(),
		lifecycleManager: NewLifecycleManager(),
	}

	// Register default hooks
	s.registerHooks()

	return s
}

// CreateConversation creates a new conversation session
func (s *ChatbotService) CreateConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
	sessionID := util.GenerateUUID()
	err := s.conversationRepo.CreateConversation(sessionID, req.UserId, req.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %v", err)
	}

	return &pb.CreateConversationResponse{
		SessionId: sessionID,
	}, nil
}

func (s *ChatbotService) registerHooks() {
	// Before Hook: Save user message
	s.lifecycleManager.RegisterBeforeHook(func(ctx context.Context, session *ChatSession) error {
		if session.Prompt != "" {
			return s.messageService.SaveMessage(ctx, session.SessionID, "user", session.Prompt)
		}
		return nil
	})

	// In Hook: Content Moderation (Example: Check for "forbidden")
	s.lifecycleManager.RegisterInHook(func(ctx context.Context, session *ChatSession, chunk string) error {
		// @TODO
		return nil
	})

	// After Hook: Save AI response
	s.lifecycleManager.RegisterAfterHook(func(ctx context.Context, session *ChatSession, genErr error) {
		if session.FullResponse.Len() > 0 {
			role := "assistant"
			if session.IsIntercepted {
				role = "system" // Or mark as intercepted
			}
			if err := s.messageService.SaveMessage(ctx, session.SessionID, role,
				 session.FullResponse.String()); err != nil {
				logger.Log.Error("failed to save assistant message: " + err.Error())
			}
		}
	})
}

// StreamChat implements the streaming chat logic with lifecycle management
func (s *ChatbotService) StreamChat(req *pb.ChatStreamRequest, stream pb.ChatbotService_ChatStreamServer) error {
	ctx := stream.Context()
	sessionID := req.SessionId
	lockKey := fmt.Sprintf("chat:session:%s:lock", sessionID)

	// 0. Redis Lock (@TODO Unsafe Lock)
	success, err := redisPkg.RDB.SetNX(ctx, lockKey, 1, 30*time.Second).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	if !success {
		return fmt.Errorf("session is busy, please try again later")
	}
	defer redisPkg.RDB.Del(ctx, lockKey)

	// Build Session Context
	// Extract the last message as the current prompt if available
	var prompt string
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}
	
	// Convert pb messages to openai messages for history
	// @TODO O(n) --->> O(1)? 
	var openaiMsgs []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		openaiMsgs = append(openaiMsgs, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	chatSession := &ChatSession{
		SessionID: sessionID,
		UserID:    req.UserId,
		Prompt:    prompt,
		Messages:  openaiMsgs,
		CreatedAt: time.Now(),
	}

	// 1. Before Hooks
	if err := s.lifecycleManager.ExecuteBefore(ctx, chatSession); err != nil {
		return err
	}

	// 2. Generate (Call LLM API)
	genErr := s.llmAPI.GenerateStreamWithMessages(ctx, chatSession.Messages, func(chunk string) error {
		// 3. In Hooks
		if err := s.lifecycleManager.ExecuteIn(ctx, chatSession, chunk); err != nil {
			chatSession.IsIntercepted = true
			return err
		}

		// Accumulate response
		chatSession.FullResponse.WriteString(chunk)

		// Send to gRPC stream
		if err := stream.Send(&pb.ChatStreamResponse{
			ContentChunk: chunk,
			IsFinish:     false,
		}); err != nil {
			return err
		}
		return nil
	})

	// Send Finish signal if no error
	if genErr == nil {
		stream.Send(&pb.ChatStreamResponse{IsFinish: true, FinishReason: "stop"})
	}

	// 4. After Hooks
	s.lifecycleManager.ExecuteAfter(ctx, chatSession, genErr)

	return genErr
}

// GetChatHistory implements retrieving chat history for a session
func (s *ChatbotService) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.GetChatHistoryResponse, error) {
	msgs, err := s.messageService.GetChatHistory(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	var pbMsgs []*pb.ChatMessage
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &pb.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return &pb.GetChatHistoryResponse{
		Messages: pbMsgs,
	}, nil
}

// ListConversations implements listing conversations for a user
func (s *ChatbotService) ListConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	conversations, err := s.conversationRepo.ListConversations(req.UserId)
	if err != nil {
		return nil, err
	}

	var pbConvs []*pb.ConversationInfo
	for _, c := range conversations {
		pbConvs = append(pbConvs, &pb.ConversationInfo{
			SessionId: c.SessionID,
			Title:     c.Title,
			CreatedAt: c.CreatedAt.Unix(),
		})
	}

	return &pb.ListConversationsResponse{
		Conversations: pbConvs,
	}, nil
}
