// Package uid generates random (version 4) UUIDs using the standard library.
package uid

import (
	"crypto/rand"
	"encoding/hex"
)

// UUID is a 128-bit (16-byte) RFC 4122 universally unique identifier.
type UUID [16]byte

// NewV4 returns a randomly generated version 4 UUID.
func NewV4() (UUID, error) {
	var u UUID
	if _, err := rand.Read(u[:]); err != nil {
		return UUID{}, err
	}
	// set version (4) and variant bits
	u[6] = (u[6] & 0x0f) | 0x40
	u[8] = (u[8] & 0x3f) | 0x80
	return u, nil
}

// String returns the canonical 8-4-4-4-12 hyphenated form of the UUID.
func (u UUID) String() string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], u[10:16])
	return string(buf)
}
