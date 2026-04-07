// Package uuid provides UUID v7 generation and parsing utilities.
// UUID v7 is time-sortable and suitable for use as database primary keys
// and distributed system identifiers.
package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// uuidLen is the standard string representation length of a UUID (36 characters including hyphens).
const uuidLen = 36

// UUID represents a 128-bit Universally Unique Identifier.
type UUID [16]byte

// NewV7 generates a new UUID v7 using the current timestamp.
// UUID v7 is time-sortable and recommended for use as primary keys.
func NewV7() UUID {
	return newV7(time.Now())
}

// newV7 generates a UUID v7 for a specific timestamp, useful for testing.
func newV7(now time.Time) UUID {
	var u UUID

	ms := uint64(now.UnixMilli())

	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	u[6] = 0x70 | (byte(randUint16()) >> 4)

	_, _ = rand.Read(u[8:])

	u[8] = 0x80 | (u[8] & 0x3F)

	return u
}

// Parse converts a UUID string in the form "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" to a UUID.
// Returns an error if the string format is invalid.
func Parse(s string) (UUID, error) {
	var u UUID
	if len(s) != uuidLen {
		return u, fmt.Errorf("uuid: invalid length %d", len(s))
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return u, fmt.Errorf("uuid: invalid format")
	}

	b := make([]byte, 16)
	_, err := hex.Decode(b, []byte(s[0:8]+s[9:13]+s[14:18]+s[19:23]+s[24:36]))
	if err != nil {
		return u, fmt.Errorf("uuid: invalid hex: %w", err)
	}
	copy(u[:], b)
	return u, nil
}

// MustParse is like Parse but panics if the string cannot be parsed.
// Use this only when the input is guaranteed to be valid.
func MustParse(s string) UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// String returns the standard string representation of the UUID.
func (u UUID) String() string {
	var buf [uuidLen]byte
	encodeHex(buf[:], u)
	return string(buf[:])
}

// MarshalText implements encoding.TextMarshaler for JSON and other text encodings.
func (u UUID) MarshalText() ([]byte, error) {
	var buf [uuidLen]byte
	encodeHex(buf[:], u)
	return buf[:], nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON and other text encodings.
func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// Timestamp extracts the timestamp from a UUID v7. Returns the zero time for invalid UUIDs.
func (u UUID) Timestamp() time.Time {
	ms := uint64(u[0])<<40 | uint64(u[1])<<32 | uint64(u[2])<<24 | uint64(u[3])<<16 | uint64(u[4])<<8 | uint64(u[5])
	return time.UnixMilli(int64(ms))
}

// Version returns the UUID version number (should be 7 for UUID v7).
func (u UUID) Version() int {
	return int(u[6] >> 4)
}

// encodeHex writes the UUID as a hyphenated hex string to dst.
func encodeHex(dst []byte, u UUID) {
	hex.Encode(dst[0:8], u[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], u[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], u[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], u[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], u[10:])
}

// randMu protects concurrent access to randBuf.
var randMu sync.Mutex

// randBuf is a reusable buffer for random byte generation.
var randBuf [8]byte

// randUint16 generates a random uint16 using crypto/rand.
func randUint16() uint16 {
	randMu.Lock()
	defer randMu.Unlock()
	_, _ = rand.Read(randBuf[:])
	return uint16(randBuf[0])<<8 | uint16(randBuf[1])
}
