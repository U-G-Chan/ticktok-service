package rpc

import (
	"errors"
	"sync"
	"ticktok-service/api/chatbot/v1"
	"ticktok-service/api/content/v1"
	"ticktok-service/api/message/v1"
	"ticktok-service/api/user/v1"
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

// InitClientManager initializes the singleton instance of ClientManager with specific services
func InitClientManager(serviceAddrs map[string]string) (*ClientManager, error) {
	var err error
	once.Do(func() {
		clientManager = &ClientManager{}
		err = clientManager.initClients(serviceAddrs)
	})
	return clientManager, err
}

// GetClientManager returns the singleton instance of ClientManager
func GetClientManager() *ClientManager {
	if clientManager == nil {
		panic("ClientManager not initialized. Call InitClientManager first.")
	}
	return clientManager
}

func (cm *ClientManager) initClients(serviceAddrs map[string]string) error {
	for name, addr := range serviceAddrs {
		if addr == "" {
			continue
		}

		conn, err := connect(addr)
		if err != nil {
			logger.Log.Error("failed to connect to " + name + " service: " + err.Error())
			return err
		}

		switch name {
		case "user":
			cm.UserClient = user.NewUserServiceClient(conn)
			logger.Log.Info("User RPC client initialized")
		case "content":
			cm.ContentClient = content.NewContentServiceClient(conn)
			logger.Log.Info("Content RPC client initialized")
		case "message":
			cm.MessageClient = message.NewMessageServiceClient(conn)
			logger.Log.Info("Message RPC client initialized")
		case "chatbot":
			cm.ChatbotClient = chatbot.NewChatbotServiceClient(conn)
			logger.Log.Info("Chatbot RPC client initialized")
		default:
			return errors.New("unknown service name: " + name)
		}
	}
	return nil
}

func connect(target string) (*grpc.ClientConn, error) {
	// Centralized connection logic (Retry, Timeout, TLS, Interceptors can be added here)
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
