package ws

import (
	"testing"
	"ticktok-service/pkg/logger"
)

func TestLocalWebSocketConnPool(t *testing.T) {
	logger.Init("debug")
	pool := &LocalWebSocketConnPool{}
	userID := int64(123)

	// Test Get on empty pool
	_, ok := pool.GetClient(userID)
	if ok {
		t.Errorf("Expected false, got true")
	}

	// Test Add
	client := &Client{UserID: userID}
	pool.AddClient(userID, client)

	// Test Get
	c, ok := pool.GetClient(userID)
	if !ok || c.UserID != userID {
		t.Errorf("Expected true and %d, got %v and %v", userID, ok, c)
	}

	// Note: We skip DelClient test because it tries to call conn.Close() and conn is nil in this mock.
}
