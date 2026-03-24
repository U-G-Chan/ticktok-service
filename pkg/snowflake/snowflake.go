package snowflake

import (
	"log"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// InitSnowflake initializes the snowflake node with a specific node number
func InitSnowflake(nodeNumber int64) {
	var err error
	node, err = snowflake.NewNode(nodeNumber)
	if err != nil {
		log.Fatalf("Failed to initialize snowflake: %v", err)
	}
}

// GenerateMsgID generates a unique ID
func GenerateMsgID() int64 {
	if node == nil {
		InitSnowflake(1)
	}
	return node.Generate().Int64()
}
