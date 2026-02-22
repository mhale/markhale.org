package main

import (
	"crypto/rand"
	"fmt"
	"time"
)

const (
	// UUID structure constants
	UUIDSize         = 16
	TimestampBytes   = 6
	RandomBytesStart = 6
	RandomBytesCount = 10

	// Bit manipulation constants
	VersionByte    = 6
	VariantByte    = 8
	VersionMask    = 0x0f
	VersionV7      = 0x70
	VariantMask    = 0x3f
	VariantRFC4122 = 0x80

	// Bit shift constants for timestamp
	TimestampShift40 = 40
	TimestampShift32 = 32
	TimestampShift24 = 24
	TimestampShift16 = 16
	TimestampShift8  = 8
)

// GenerateUUIDv7 generates a version 7 UUID according to RFC 4122
// V7 UUIDs are time-ordered and contain a 48-bit timestamp followed by random data
func GenerateUUIDv7() string {
	// Get current Unix timestamp in milliseconds
	timestamp := time.Now().UnixMilli()

	// Create UUID byte array
	var uuid [UUIDSize]byte

	// Set timestamp (48 bits = 6 bytes) in big-endian format
	uuid[0] = byte(timestamp >> TimestampShift40)
	uuid[1] = byte(timestamp >> TimestampShift32)
	uuid[2] = byte(timestamp >> TimestampShift24)
	uuid[3] = byte(timestamp >> TimestampShift16)
	uuid[4] = byte(timestamp >> TimestampShift8)
	uuid[5] = byte(timestamp)

	// Fill remaining bytes with random data
	rand.Read(uuid[RandomBytesStart:])

	// Set version (4 bits) - version 7
	uuid[VersionByte] = (uuid[VersionByte] & VersionMask) | VersionV7

	// Set variant (2 bits) - RFC 4122 variant
	uuid[VariantByte] = (uuid[VariantByte] & VariantMask) | VariantRFC4122

	// Format as string with hyphens
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16])
}
