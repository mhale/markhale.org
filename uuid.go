package main

import (
	"crypto/rand"
	"encoding/binary"
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
)

// GenerateUUIDv7 generates a version 7 UUID according to RFC 4122
// V7 UUIDs contain a 48-bit timestamp followed by 80 bits of random data
func GenerateUUIDv7() string {
	var uuid [UUIDSize]byte
	var tsBytes [8]byte

	timestamp := time.Now().UnixMilli()

	// Set timestamp (48 bits = 6 bytes) in big-endian format
	// timestamp = 1772295802342 will look like this in binary:
	// 00000000 00000000 00000001 10011100 10100101 00001111 11001101 11100110
	binary.BigEndian.PutUint64(tsBytes[:], uint64(timestamp))

	// Copy last 6 bytes (48 bits) from the 64-bit big-endian buffer
	copy(uuid[0:TimestampBytes], tsBytes[8-TimestampBytes:])

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
