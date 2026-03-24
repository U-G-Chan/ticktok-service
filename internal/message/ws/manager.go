package ws

import (
	"sync"
	"ticktok-service/pkg/logger"
)

// LocalWebSocketConnPool manages local websocket connections
type LocalWebSocketConnPool struct {
	conns sync.Map
}

var GlobalPool = &LocalWebSocketConnPool{}

// AddClient adds a new client
func (p *LocalWebSocketConnPool) AddClient(userID int64, client *Client) {
	p.conns.Store(userID, client)
	logger.Log.Sugar().Infof("User %d connected. Local connection added.", userID)
}

// GetClient retrieves an existing client
func (p *LocalWebSocketConnPool) GetClient(userID int64) (*Client, bool) {
	if val, ok := p.conns.Load(userID); ok {
		return val.(*Client), true
	}
	return nil, false
}

// DelClient deletes a client
func (p *LocalWebSocketConnPool) DelClient(userID int64) {
	if client, ok := p.GetClient(userID); ok {
		client.Conn.Close()
		p.conns.Delete(userID)
		logger.Log.Sugar().Infof("User %d disconnected. Local connection removed.", userID)
	}
}
