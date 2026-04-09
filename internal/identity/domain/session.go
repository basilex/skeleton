// Package domain provides domain entities and repository interfaces for the identity module.
package domain

import (
	"fmt"
	"net"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// SessionID is a unique identifier for a session.
type SessionID string

// NewSessionID generates a new unique SessionID.
func NewSessionID() SessionID {
	return SessionID(uuid.NewV7().String())
}

// ParseSessionID validates and converts a string to SessionID.
func ParseSessionID(s string) (SessionID, error) {
	if s == "" {
		return "", fmt.Errorf("session id cannot be empty")
	}
	return SessionID(s), nil
}

// String returns the string representation of SessionID.
func (id SessionID) String() string {
	return string(id)
}

// SessionStatus represents the state of a session.
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusExpired   SessionStatus = "expired"
	SessionStatusRevoked   SessionStatus = "revoked"
	SessionStatusLoggedOut SessionStatus = "logged_out"
)

// String returns the string representation of SessionStatus.
func (s SessionStatus) String() string {
	return string(s)
}

// ParseSessionStatus converts a string to a SessionStatus value.
func ParseSessionStatus(s string) (SessionStatus, error) {
	switch s {
	case string(SessionStatusActive):
		return SessionStatusActive, nil
	case string(SessionStatusExpired):
		return SessionStatusExpired, nil
	case string(SessionStatusRevoked):
		return SessionStatusRevoked, nil
	case string(SessionStatusLoggedOut):
		return SessionStatusLoggedOut, nil
	default:
		return "", fmt.Errorf("invalid session status: %s", s)
	}
}

// DeviceInfo contains information about the device used for the session.
type DeviceInfo struct {
	userAgent  string
	deviceType string
	os         string
	browser    string
	deviceName string
}

// NewDeviceInfo creates a new DeviceInfo.
func NewDeviceInfo(userAgent, deviceType, os, browser, deviceName string) DeviceInfo {
	return DeviceInfo{
		userAgent:  userAgent,
		deviceType: deviceType,
		os:         os,
		browser:    browser,
		deviceName: deviceName,
	}
}

// UserAgent returns the user agent string.
func (d DeviceInfo) UserAgent() string {
	return d.userAgent
}

// DeviceType returns the device type (mobile, tablet, desktop).
func (d DeviceInfo) DeviceType() string {
	return d.deviceType
}

// OS returns the operating system.
func (d DeviceInfo) OS() string {
	return d.os
}

// Browser returns the browser name.
func (d DeviceInfo) Browser() string {
	return d.browser
}

// DeviceName returns the device name.
func (d DeviceInfo) DeviceName() string {
	return d.deviceName
}

// Session represents a user session aggregate.
type Session struct {
	id            SessionID
	userID        UserID
	status        SessionStatus
	device        DeviceInfo
	ipAddress     net.IP
	expiresAt     time.Time
	lastActivity  time.Time
	createdAt     time.Time
	revokedAt     *time.Time
	revokedReason string
	events        []DomainEvent
}

// NewSession creates a new session for a user.
func NewSession(userID UserID, device DeviceInfo, ipAddress net.IP, duration time.Duration) (*Session, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(duration)

	session := &Session{
		id:           NewSessionID(),
		userID:       userID,
		status:       SessionStatusActive,
		device:       device,
		ipAddress:    ipAddress,
		expiresAt:    expiresAt,
		lastActivity: now,
		createdAt:    now,
		events:       make([]DomainEvent, 0),
	}

	session.events = append(session.events, SessionCreated{
		SessionID:  session.id,
		UserID:     userID,
		DeviceType: device.DeviceType(),
		UserAgent:  device.UserAgent(),
		IPAddress:  ipAddress.String(),
		ExpiresAt:  expiresAt,
		occurredAt: now,
	})

	return session, nil
}

// ReconstituteSession reconstructs a Session from persisted state.
func ReconstituteSession(
	id SessionID,
	userID UserID,
	status SessionStatus,
	device DeviceInfo,
	ipAddress net.IP,
	expiresAt time.Time,
	lastActivity time.Time,
	createdAt time.Time,
	revokedAt *time.Time,
	revokedReason string,
) *Session {
	return &Session{
		id:            id,
		userID:        userID,
		status:        status,
		device:        device,
		ipAddress:     ipAddress,
		expiresAt:     expiresAt,
		lastActivity:  lastActivity,
		createdAt:     createdAt,
		revokedAt:     revokedAt,
		revokedReason: revokedReason,
		events:        make([]DomainEvent, 0),
	}
}

// ID returns the session ID.
func (s *Session) ID() SessionID {
	return s.id
}

// UserID returns the user ID.
func (s *Session) UserID() UserID {
	return s.userID
}

// Status returns the session status.
func (s *Session) Status() SessionStatus {
	return s.status
}

// Device returns the device information.
func (s *Session) Device() DeviceInfo {
	return s.device
}

// IPAddress returns the IP address.
func (s *Session) IPAddress() net.IP {
	return s.ipAddress
}

// ExpiresAt returns the expiration time.
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// LastActivity returns the last activity time.
func (s *Session) LastActivity() time.Time {
	return s.lastActivity
}

// CreatedAt returns the creation time.
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// RevokedAt returns the revocation time (if revoked).
func (s *Session) RevokedAt() *time.Time {
	return s.revokedAt
}

// RevokedReason returns the reason for revocation.
func (s *Session) RevokedReason() string {
	return s.revokedReason
}

// IsExpired returns true if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.expiresAt)
}

// IsActive returns true if the session is active and not expired.
func (s *Session) IsActive() bool {
	return s.status == SessionStatusActive && !s.IsExpired()
}

// IsRevoked returns true if the session was revoked.
func (s *Session) IsRevoked() bool {
	return s.status == SessionStatusRevoked
}

// Refresh extends the session expiration and updates last activity.
func (s *Session) Refresh(duration time.Duration) error {
	if s.IsRevoked() {
		return ErrSessionRevoked
	}
	if s.IsExpired() {
		return ErrSessionExpired
	}

	now := time.Now().UTC()
	s.expiresAt = now.Add(duration)
	s.lastActivity = now

	s.events = append(s.events, SessionRefreshed{
		SessionID:  s.id,
		UserID:     s.userID,
		ExpiresAt:  s.expiresAt,
		occurredAt: now,
	})

	return nil
}

// UpdateActivity updates the last activity timestamp.
func (s *Session) UpdateActivity() error {
	if s.IsRevoked() {
		return ErrSessionRevoked
	}
	if s.IsExpired() {
		return ErrSessionExpired
	}

	s.lastActivity = time.Now().UTC()
	return nil
}

// Revoke revokes the session with a reason.
func (s *Session) Revoke(reason string) error {
	if s.status == SessionStatusRevoked {
		return ErrSessionAlreadyRevoked
	}

	now := time.Now().UTC()
	s.status = SessionStatusRevoked
	s.revokedAt = &now
	s.revokedReason = reason

	s.events = append(s.events, SessionRevoked{
		SessionID:  s.id,
		UserID:     s.userID,
		Reason:     reason,
		occurredAt: now,
	})

	return nil
}

// Expire marks the session as expired.
func (s *Session) Expire() {
	now := time.Now().UTC()
	s.status = SessionStatusExpired
	s.events = append(s.events, SessionExpired{
		SessionID:  s.id,
		UserID:     s.userID,
		occurredAt: now,
	})
}

// Logout marks the session as logged out.
func (s *Session) Logout() {
	now := time.Now().UTC()
	s.status = SessionStatusLoggedOut
	s.events = append(s.events, SessionLoggedOut{
		SessionID:  s.id,
		UserID:     s.userID,
		occurredAt: now,
	})
}

// CheckExpiration checks if the session has expired and updates status if needed.
func (s *Session) CheckExpiration() bool {
	if s.status == SessionStatusActive && s.IsExpired() {
		s.Expire()
		return true
	}
	return false
}

// PullEvents returns all pending domain events and clears the buffer.
func (s *Session) PullEvents() []DomainEvent {
	events := s.events
	s.events = make([]DomainEvent, 0)
	return events
}
