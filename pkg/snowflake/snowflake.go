package snowflake

import (
	"log"

	"ticktok-service/pkg/config"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init initializes the snowflake node with the node ID from config
func Init() {
	var err error
	nodeID := int64(1) // default fallback
	
	if config.Config != nil && config.Config.Snowflake.NodeID != 0 {
		nodeID = config.Config.Snowflake.NodeID
	}

	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		log.Fatalf("Failed to initialize snowflake with node ID %d: %v", nodeID, err)
	}
	log.Printf("Snowflake initialized successfully with node ID: %d", nodeID)
}

// GenerateMsgID generates a unique ID
func GenerateMsgID() int64 {
	if node == nil {
		Init()
	}
	return node.Generate().Int64()
}
