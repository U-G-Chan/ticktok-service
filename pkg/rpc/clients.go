package rpc

import (
	"sync"
	"ticktok-service/api/chatbot/v1"
	"ticktok-service/api/content/v1"
	"ticktok-service/api/message/v1"
	"ticktok-service/api/user/v1"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	once          sync.Once
	clientManager *ClientManager
)

// ClientManager manages all gRPC clients
type ClientManager struct {
	UserClient    user.UserServiceClient
	ContentClient content.ContentServiceClient
	MessageClient message.MessageServiceClient
	ChatbotClient chatbot.ChatbotServiceClient
}

func GetClientManager() *ClientManager {
	once.Do(func() {
		clientManager = &ClientManager{}
		clientManager.initClients()
	})
	return clientManager
}

func NewClientManager() *ClientManager {
	cm := &ClientManager{}
	cm.initClients()
	return cm
}

func (cm *ClientManager) initClients() {
	if config.Config.Microservices.User != "" {
		conn, err := connect(config.Config.Microservices.User)
		if err != nil {
			logger.Log.Error("failed to connect to user service: " + err.Error())
		} else {
			cm.UserClient = user.NewUserServiceClient(conn)
			logger.Log.Info("User RPC client initialized")
		}
	}

	if config.Config.Microservices.Content != "" {
		conn, err := connect(config.Config.Microservices.Content)
		if err != nil {
			logger.Log.Error("failed to connect to content service: " + err.Error())
		} else {
			cm.ContentClient = content.NewContentServiceClient(conn)
			logger.Log.Info("Content RPC client initialized")
		}
	}

	if config.Config.Microservices.Message != "" {
		conn, err := connect(config.Config.Microservices.Message)
		if err != nil {
			logger.Log.Error("failed to connect to message service: " + err.Error())
		} else {
			cm.MessageClient = message.NewMessageServiceClient(conn)
			logger.Log.Info("Message RPC client initialized")
		}
	}

	if config.Config.Microservices.Chatbot != "" {
		conn, err := connect(config.Config.Microservices.Chatbot)
		if err != nil {
			logger.Log.Error("failed to connect to chatbot service: " + err.Error())
		} else {
			cm.ChatbotClient = chatbot.NewChatbotServiceClient(conn)
			logger.Log.Info("Chatbot RPC client initialized")
		}
	}
}

func connect(target string) (*grpc.ClientConn, error) {
	// Centralized connection logic (Retry, Timeout, TLS, Interceptors can be added here)
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
