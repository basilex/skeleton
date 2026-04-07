package uuid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewV7(t *testing.T) {
	u := NewV7()
	require.Equal(t, 7, u.Version())
	require.False(t, u.Timestamp().IsZero())
}

func TestNewV7Sortability(t *testing.T) {
	u1 := NewV7()
	time.Sleep(time.Millisecond)
	u2 := NewV7()

	require.Less(t, u1.String(), u2.String())
}

func TestParseAndString(t *testing.T) {
	u := NewV7()
	s := u.String()

	parsed, err := Parse(s)
	require.NoError(t, err)
	require.Equal(t, u, parsed)
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "abc"},
		{"too long", "00000000-0000-7000-8000-0000000000000"},
		{"missing dashes", "00000000000070008000000000000000"},
		{"invalid hex", "xxxxxxxx-xxxx-7xxx-8xxx-xxxxxxxxxxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			require.Error(t, err)
		})
	}
}

func TestMustParse(t *testing.T) {
	u := NewV7()
	parsed := MustParse(u.String())
	require.Equal(t, u, parsed)
}

func TestMustParsePanics(t *testing.T) {
	require.Panics(t, func() {
		MustParse("invalid")
	})
}

func TestTimestamp(t *testing.T) {
	now := time.Now()
	u := NewV7()

	diff := u.Timestamp().Sub(now)
	require.Less(t, diff.Abs(), time.Second)
}

func TestVersion(t *testing.T) {
	u := NewV7()
	require.Equal(t, 7, u.Version())
}

func TestMarshalUnmarshalText(t *testing.T) {
	u := NewV7()

	data, err := u.MarshalText()
	require.NoError(t, err)

	var parsed UUID
	err = parsed.UnmarshalText(data)
	require.NoError(t, err)
	require.Equal(t, u, parsed)
}

func TestUUIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		u := NewV7()
		s := u.String()
		require.False(t, seen[s], "duplicate UUID: %s", s)
		seen[s] = true
	}
}
