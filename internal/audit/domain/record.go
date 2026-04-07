package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

type Action string

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

func (a Action) String() string { return string(a) }

type ActorType string

const (
	ActorUser   ActorType = "user"
	ActorSystem ActorType = "system"
)

func (a ActorType) String() string { return string(a) }

type RecordID string

func NewRecordID() RecordID {
	return RecordID(uuid.NewV7().String())
}

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

func (r *Record) ID() RecordID         { return r.id }
func (r *Record) ActorID() string      { return r.actorID }
func (r *Record) ActorType() ActorType { return r.actorType }
func (r *Record) Action() Action       { return r.action }
func (r *Record) Resource() string     { return r.resource }
func (r *Record) ResourceID() string   { return r.resourceID }
func (r *Record) Metadata() string     { return r.metadata }
func (r *Record) IP() string           { return r.ip }
func (r *Record) UserAgent() string    { return r.userAgent }
func (r *Record) Status() int          { return r.status }
func (r *Record) CreatedAt() time.Time { return r.createdAt }

type RecordFilter struct {
	ActorID  string
	Resource string
	Action   string
	DateFrom time.Time
	DateTo   time.Time
	Cursor   string
	Limit    int
}
