// Package domain provides domain entities and repository interfaces for the audit module.
// This package contains the core business logic types for audit trail tracking and
// repository contracts for persisting audit records.
package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// Action represents the type of action performed in an audit record.
type Action string

// Action type constants.
const (
	ActionCreate     Action = "create"
	ActionRead       Action = "read"
	ActionUpdate     Action = "update"
	ActionDelete     Action = "delete"
	ActionLogin      Action = "login"
	ActionLogout     Action = "logout"
	ActionAssignRole Action = "assign_role"
	ActionRevokeRole Action = "revoke_role"
	ActionRegister   Action = "register"
)

// String returns the string representation of the action.
func (a Action) String() string { return string(a) }

// ActorType represents the type of actor performing an action.
type ActorType string

// ActorType constants.
const (
	ActorUser   ActorType = "user"
	ActorSystem ActorType = "system"
)

// String returns the string representation of the actor type.
func (a ActorType) String() string { return string(a) }

// RecordID is a unique identifier for an audit record.
type RecordID uuid.UUID

// NewRecordID generates a new unique RecordID using UUID v7.
func NewRecordID() RecordID {
	return RecordID(uuid.NewV7())
}

// String returns the string representation of RecordID.
func (id RecordID) String() string {
	return uuid.UUID(id).String()
}

// ParseRecordID parses a string into a RecordID.
// Returns an error if the string is empty or invalid.
func ParseRecordID(s string) (RecordID, error) {
	if s == "" {
		return RecordID{}, ErrInvalidRecordID
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return RecordID{}, ErrInvalidRecordID
	}
	return RecordID(u), nil
}

// Record represents an audit trail entry capturing who did what to which resource.
type Record struct {
	id         RecordID
	actorID    string
	actorType  ActorType
	action     Action
	resource   string
	resourceID string
	metadata   string
	ip         string
	userAgent  string
	status     int
	createdAt  time.Time
}

// NewRecord creates a new audit record with the provided details.
func NewRecord(actorID string, actorType ActorType, action Action, resource, resourceID, metadata, ip, userAgent string, status int) *Record {
	return &Record{
		id:         NewRecordID(),
		actorID:    actorID,
		actorType:  actorType,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		metadata:   metadata,
		ip:         ip,
		userAgent:  userAgent,
		status:     status,
		createdAt:  time.Now().UTC(),
	}
}

// ReconstituteRecord reconstructs a Record entity from persisted state.
// This is used by repositories to hydrate record entities from storage.
func ReconstituteRecord(
	id RecordID,
	actorID string,
	actorType ActorType,
	action Action,
	resource, resourceID, metadata, ip, userAgent string,
	status int,
	createdAt time.Time,
) *Record {
	return &Record{
		id:         id,
		actorID:    actorID,
		actorType:  actorType,
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		metadata:   metadata,
		ip:         ip,
		userAgent:  userAgent,
		status:     status,
		createdAt:  createdAt,
	}
}

// ID returns the record's unique identifier.
func (r *Record) ID() RecordID { return r.id }

// ActorID returns the ID of the actor who performed the action.
func (r *Record) ActorID() string { return r.actorID }

// ActorType returns the type of actor (user or system).
func (r *Record) ActorType() ActorType { return r.actorType }

// Action returns the action that was performed.
func (r *Record) Action() Action { return r.action }

// Resource returns the resource type that was acted upon.
func (r *Record) Resource() string { return r.resource }

// ResourceID returns the ID of the specific resource instance.
func (r *Record) ResourceID() string { return r.resourceID }

// Metadata returns additional contextual information as a string.
func (r *Record) Metadata() string { return r.metadata }

// IP returns the IP address from which the action was performed.
func (r *Record) IP() string { return r.ip }

// UserAgent returns the user agent string of the client.
func (r *Record) UserAgent() string { return r.userAgent }

// Status returns the HTTP status code or result status of the action.
func (r *Record) Status() int { return r.status }

// CreatedAt returns the timestamp when the record was created.
func (r *Record) CreatedAt() time.Time { return r.createdAt }

// RecordFilter provides filtering options for querying audit records.
type RecordFilter struct {
	ActorID  string
	Resource string
	Action   string
	DateFrom time.Time
	DateTo   time.Time
	Cursor   string
	Limit    int
}
