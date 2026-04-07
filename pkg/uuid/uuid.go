package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const uuidLen = 36

type UUID [16]byte

func NewV7() UUID {
	return newV7(time.Now())
}

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

func MustParse(s string) UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func (u UUID) String() string {
	var buf [uuidLen]byte
	encodeHex(buf[:], u)
	return string(buf[:])
}

func (u UUID) MarshalText() ([]byte, error) {
	var buf [uuidLen]byte
	encodeHex(buf[:], u)
	return buf[:], nil
}

func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

func (u UUID) Timestamp() time.Time {
	ms := uint64(u[0])<<40 | uint64(u[1])<<32 | uint64(u[2])<<24 | uint64(u[3])<<16 | uint64(u[4])<<8 | uint64(u[5])
	return time.UnixMilli(int64(ms))
}

func (u UUID) Version() int {
	return int(u[6] >> 4)
}

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

var randMu sync.Mutex
var randBuf [8]byte

func randUint16() uint16 {
	randMu.Lock()
	defer randMu.Unlock()
	_, _ = rand.Read(randBuf[:])
	return uint16(randBuf[0])<<8 | uint16(randBuf[1])
}
