package snowflake

import (
	"testing"
)

func TestSnowflake(t *testing.T) {
	Init()
	id1 := GenerateMsgID()
	id2 := GenerateMsgID()

	if id1 == id2 {
		t.Errorf("Expected unique IDs, got same: %d", id1)
	}
	if id1 <= 0 || id2 <= 0 {
		t.Errorf("Expected positive IDs, got %d and %d", id1, id2)
	}
}
