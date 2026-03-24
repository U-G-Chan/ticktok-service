package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ticktok-service/internal/message/producer"
	"ticktok-service/internal/message/router"
	"ticktok-service/pkg/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// Client represents a single websocket connection
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

func ServeWebSocket(w http.ResponseWriter, r *http.Request, nodeIP string) {
	userIDStr := r.Header.Get("X-User-Id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logger.Log.Error("Invalid user ID in header")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error("Failed to upgrade to websocket: " + err.Error())
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Register in local pool
	GlobalPool.AddClient(userID, client)

	// Set routing in Redis
	ctx := context.Background()
	if err := router.SetRoute(ctx, userID, nodeIP); err != nil {
		logger.Log.Sugar().Errorf("Failed to set route for user %d: %v", userID, err)
	}

	// Start pump goroutines
	go client.writePump()
	go client.readPump(ctx)
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		GlobalPool.DelClient(c.UserID)
		router.DeleteRoute(ctx, c.UserID)
		c.Conn.Close()
	}()

	// c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Sugar().Errorf("error: %v", err)
			}
			break
		}

		var wsMsg WsMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			logger.Log.Sugar().Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		switch wsMsg.Action {
		case ActionPing:
			// Refresh route TTL
			router.RenewRoute(ctx, c.UserID)
			// Reply pong
			pongMsg := WsMessage{
				Action: ActionPong,
				Data:   wsMsg.Data,
			}
			pongBytes, _ := json.Marshal(pongMsg)
			c.Send <- pongBytes

		case ActionOffline:
			// Client actively disconnects
			return

		case ActionSendMsg:
			var sendData SendMsgData
			if err := json.Unmarshal(wsMsg.Data, &sendData); err != nil {
				logger.Log.Sugar().Errorf("Failed to unmarshal sendData: %v", err)
				continue
			}

			// Call producer to send message (Producer)
			msgID, err := producer.SendMessage(ctx, c.UserID, sendData.ToUserID, sendData.Content)
			if err == nil {
				// Reply Ack
				ackData := AckData{MsgID: msgID}
				ackBytes, _ := json.Marshal(ackData)
				ackMsg := WsMessage{
					Action: ActionAck,
					Data:   ackBytes,
				}
				finalAckBytes, _ := json.Marshal(ackMsg)
				c.Send <- finalAckBytes
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
