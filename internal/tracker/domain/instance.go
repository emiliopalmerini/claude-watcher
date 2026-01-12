package domain

import (
	"crypto/sha256"
	"fmt"
)

// GenerateInstanceID creates a unique identifier for a Claude Code instance
// based on hostname and home directory hash
func GenerateInstanceID(hostname, homeDir string) string {
	hash := sha256.Sum256([]byte(homeDir))
	hashInt := int(hash[0])<<8 | int(hash[1])
	return fmt.Sprintf("%s:%04d", hostname, hashInt%10000)
}
