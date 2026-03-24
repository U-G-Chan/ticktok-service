package ws

import (
	"encoding/json"
)

type ActionType string

const (
	ActionPing    ActionType = "ping"
	ActionPong    ActionType = "pong"
	ActionOffline ActionType = "offline"
	ActionSendMsg ActionType = "send_msg"
	ActionAck     ActionType = "ack_server"
)

type WsMessage struct {
	Action ActionType      `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type PingData struct {
	Timestamp int64 `json:"timestamp"`
}

type PongData struct {
	Timestamp int64 `json:"timestamp"`
}

type OfflineData struct {
	Reason string `json:"reason"`
}

type SendMsgData struct {
	ToUserID int64  `json:"to_user_id"`
	Content  string `json:"content"`
}

type AckData struct {
	MsgID int64 `json:"msg_id"`
}
