package service

import (
	"context"
	"encoding/json"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/internal/message/repository"
	"ticktok-service/internal/message/ws"
	"ticktok-service/pkg/logger"
)

type MessageServiceServer struct {
	messagev1.UnimplementedMessageServiceServer
}

func NewMessageServiceServer() *MessageServiceServer {
	return &MessageServiceServer{}
}

// SendMessage is generally called by other services to send a message.
// Here we just wrap our SendMessage business logic.
func (s *MessageServiceServer) SendMessage(ctx context.Context, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	// Normally we parse req.Token to get fromUserID, but we can assume gateway passed it somehow
	// For now we'll just mock fromUserID as 0 if not provided or extract it from metadata.
	// We'll skip complex auth here as WS handles it.
	// @TODO 改造为“系统级消息发送口”
	return &messagev1.SendMessageResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

// PushMsgToClient pushes a message to the client connected to this local node
func (s *MessageServiceServer) PushMsgToClient(ctx context.Context, req *messagev1.PushMsgRequest) (*messagev1.PushMsgResponse, error) {
	msg := req.Message
	if client, ok := ws.GlobalPool.GetClient(msg.ToUserId); ok {
		// Found local connection, push the message
		wsMsg := ws.WsMessage{
			Action: ws.ActionSendMsg,
		}

		// Re-encode message for client
		dataBytes, _ := json.Marshal(msg)
		wsMsg.Data = dataBytes

		finalBytes, _ := json.Marshal(wsMsg)
		client.Send <- finalBytes

		return &messagev1.PushMsgResponse{
			Code: 0,
			Msg:  "success",
		}, nil
	}

	return &messagev1.PushMsgResponse{
		Code: 404,
		Msg:  "user not connected to this node",
	}, nil
}

// SyncMessageList handles pull sync requests for offline messages
func (s *MessageServiceServer) SyncMessageList(ctx context.Context, req *messagev1.SyncMessageListRequest) (*messagev1.SyncMessageListResponse, error) {
	missedIDs, err := repository.GetMissedMessageIDs(ctx, req.UserId, req.SyncKey)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to get missed IDs: %v", err)
		return &messagev1.SyncMessageListResponse{Code: 500, Msg: "failed"}, nil
	}

	if len(missedIDs) == 0 {
		return &messagev1.SyncMessageListResponse{
			Code:        0,
			Msg:         "success",
			MessageList: nil,
			NextSyncKey: req.SyncKey,
		}, nil
	}

	messages, err := repository.GetMessagesByIDs(ctx, missedIDs)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to get messages from DB: %v", err)
		return &messagev1.SyncMessageListResponse{Code: 500, Msg: "failed"}, nil
	}

	var pbMessages []*messagev1.Message
	var maxID int64 = req.SyncKey

	for _, m := range messages {
		pbMessages = append(pbMessages, &messagev1.Message{
			Id:         m.ID,
			ToUserId:   m.ToUserID,
			FromUserId: m.FromUserID,
			Content:    m.Content,
			CreateTime: m.CreateTime,
		})
		if m.ID > maxID {
			maxID = m.ID
		}
	}

	return &messagev1.SyncMessageListResponse{
		Code:        0,
		Msg:         "success",
		MessageList: pbMessages,
		NextSyncKey: maxID,
	}, nil
}
