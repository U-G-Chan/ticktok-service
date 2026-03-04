package util

import "github.com/google/uuid"

// GenerateUUID generates a random UUID string
func GenerateUUID() string {
	return uuid.New().String()
}
