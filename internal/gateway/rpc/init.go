package rpc

import (
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
	UserClient    user.UserServiceClient
	ContentClient content.ContentServiceClient
	MessageClient message.MessageServiceClient
	ChatbotClient chatbot.ChatbotServiceClient
)

func Init() {
	initUserClient()
	initContentClient()
	initMessageClient()
	initChatbotClient()
}

func initUserClient() {
	conn, err := connect(config.Config.Microservices.User)
	if err != nil {
		logger.Log.Fatal("failed to connect to user service: " + err.Error())
	}
	UserClient = user.NewUserServiceClient(conn)
	logger.Log.Info("User RPC client initialized")
}

func initContentClient() {
	conn, err := connect(config.Config.Microservices.Content)
	if err != nil {
		logger.Log.Fatal("failed to connect to content service: " + err.Error())
	}
	ContentClient = content.NewContentServiceClient(conn)
	logger.Log.Info("Content RPC client initialized")
}

func initMessageClient() {
	conn, err := connect(config.Config.Microservices.Message)
	if err != nil {
		logger.Log.Fatal("failed to connect to message service: " + err.Error())
	}
	MessageClient = message.NewMessageServiceClient(conn)
	logger.Log.Info("Message RPC client initialized")
}

func initChatbotClient() {
	conn, err := connect(config.Config.Microservices.Chatbot)
	if err != nil {
		logger.Log.Fatal("failed to connect to chatbot service: " + err.Error())
	}
	ChatbotClient = chatbot.NewChatbotServiceClient(conn)
	logger.Log.Info("Chatbot RPC client initialized")
}

func connect(target string) (*grpc.ClientConn, error) {
	// Use NewClient for modern gRPC versions (non-blocking by default)
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
