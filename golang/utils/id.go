package utils

import (
	"crypto/rand"
	"fmt"
)

func GenerateUUID() string {
	// Create a byte slice for the UUID
	uuid := make([]byte, 16)

	// Fill with random bytes
	rand.Read(uuid)

	// Set version bits (4th byte, highest 4 bits)
	uuid[6] = (uuid[6] & 0x0f) | 0x40

	// Set variant bits (8th byte, highest 2 bits)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// Format as UUID string
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
