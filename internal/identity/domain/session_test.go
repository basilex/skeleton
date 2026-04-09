package domain

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")
	duration := 24 * time.Hour

	session, err := NewSession(userID, device, ip, duration)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.NotEmpty(t, session.ID())
	require.Equal(t, userID, session.UserID())
	require.Equal(t, SessionStatusActive, session.Status())
	require.Equal(t, device, session.Device())
	require.True(t, session.IPAddress().Equal(ip))
	require.True(t, session.ExpiresAt().After(time.Now()))
	require.True(t, session.IsActive())
	require.False(t, session.IsExpired())
	require.False(t, session.IsRevoked())

	events := session.PullEvents()
	require.Len(t, events, 1)
	_, ok := events[0].(SessionCreated)
	require.True(t, ok)
}

func TestSession_Refresh(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	t.Run("refresh active session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		_ = session.PullEvents()
		originalExpiry := session.ExpiresAt()

		err := session.Refresh(2 * time.Hour)
		require.NoError(t, err)
		require.True(t, session.ExpiresAt().After(originalExpiry))

		events := session.PullEvents()
		_, ok := events[0].(SessionRefreshed)
		require.True(t, ok)
	})

	t.Run("cannot refresh expired session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Millisecond)
		_ = session.PullEvents()
		time.Sleep(2 * time.Millisecond)

		err := session.Refresh(1 * time.Hour)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrSessionExpired)
	})

	t.Run("cannot refresh revoked session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		_ = session.PullEvents()
		_ = session.Revoke("security breach")

		err := session.Refresh(1 * time.Hour)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrSessionRevoked)
	})
}

func TestSession_Revoke(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	t.Run("revoke active session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		_ = session.PullEvents()

		err := session.Revoke("user logout")
		require.NoError(t, err)
		require.Equal(t, SessionStatusRevoked, session.Status())
		require.True(t, session.IsRevoked())
		require.NotNil(t, session.RevokedAt())
		require.Equal(t, "user logout", session.RevokedReason())

		events := session.PullEvents()
		_, ok := events[0].(SessionRevoked)
		require.True(t, ok)
	})

	t.Run("cannot revoke already revoked session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		_ = session.Revoke("reason 1")

		err := session.Revoke("reason 2")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrSessionAlreadyRevoked)
	})
}

func TestSession_Expire(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	session, _ := NewSession(userID, device, ip, 1*time.Hour)
	_ = session.PullEvents()

	session.Expire()

	require.Equal(t, SessionStatusExpired, session.Status())

	events := session.PullEvents()
	_, ok := events[0].(SessionExpired)
	require.True(t, ok)
}

func TestSession_Logout(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	session, _ := NewSession(userID, device, ip, 1*time.Hour)
	_ = session.PullEvents()

	session.Logout()

	require.Equal(t, SessionStatusLoggedOut, session.Status())

	events := session.PullEvents()
	_, ok := events[0].(SessionLoggedOut)
	require.True(t, ok)
}

func TestSession_CheckExpiration(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	t.Run("session not expired", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		expired := session.CheckExpiration()
		require.False(t, expired)
		require.Equal(t, SessionStatusActive, session.Status())
	})

	t.Run("session expired after check", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Millisecond)
		time.Sleep(2 * time.Millisecond)

		expired := session.CheckExpiration()
		require.True(t, expired)
		require.Equal(t, SessionStatusExpired, session.Status())
	})
}

func TestSession_UpdateActivity(t *testing.T) {
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "desktop", "macOS", "Chrome", "MacBook Pro")
	ip := net.ParseIP("192.168.1.1")

	t.Run("update active session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		originalActivity := session.LastActivity()

		time.Sleep(10 * time.Millisecond)
		err := session.UpdateActivity()
		require.NoError(t, err)
		require.True(t, session.LastActivity().After(originalActivity))
	})

	t.Run("cannot update revoked session", func(t *testing.T) {
		session, _ := NewSession(userID, device, ip, 1*time.Hour)
		_ = session.Revoke("security")

		err := session.UpdateActivity()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrSessionRevoked)
	})
}

func TestReconstituteSession(t *testing.T) {
	id := NewSessionID()
	userID := NewUserID()
	device := NewDeviceInfo("Mozilla/5.0", "mobile", "iOS", "Safari", "iPhone")
	ip := net.ParseIP("10.0.0.1")
	now := time.Now().UTC()
	expiresAt := now.Add(1 * time.Hour)
	revokedAt := now.Add(30 * time.Minute)

	session := ReconstituteSession(
		id,
		userID,
		SessionStatusActive,
		device,
		ip,
		expiresAt,
		now,
		now,
		&revokedAt,
		"test revocation",
	)

	require.Equal(t, id, session.ID())
	require.Equal(t, userID, session.UserID())
	require.Equal(t, SessionStatusActive, session.Status())
	require.Equal(t, device, session.Device())
	require.True(t, session.IPAddress().Equal(ip))
	require.Equal(t, expiresAt, session.ExpiresAt())
	require.NotNil(t, session.RevokedAt())
}

func TestParseSessionStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    SessionStatus
		wantErr bool
	}{
		{"active", SessionStatusActive, false},
		{"expired", SessionStatusExpired, false},
		{"revoked", SessionStatusRevoked, false},
		{"logged_out", SessionStatusLoggedOut, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSessionStatus(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDeviceInfo(t *testing.T) {
	device := NewDeviceInfo(
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"desktop",
		"macOS",
		"Chrome",
		"MacBook Pro",
	)

	require.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", device.UserAgent())
	require.Equal(t, "desktop", device.DeviceType())
	require.Equal(t, "macOS", device.OS())
	require.Equal(t, "Chrome", device.Browser())
	require.Equal(t, "MacBook Pro", device.DeviceName())
}
