package domain

import (
	"testing"
	"time"
)

func TestNewRecord(t *testing.T) {
	actorID := "user-123"
	actorType := ActorUser
	action := ActionCreate
	resource := "document"
	resourceID := "doc-456"
	metadata := `{"title":"Test"}`
	ip := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	status := 201

	record := NewRecord(actorID, actorType, action, resource, resourceID, metadata, ip, userAgent, status)

	if record.ActorID() != actorID {
		t.Errorf("expected actor ID %s, got %s", actorID, record.ActorID())
	}
	if record.ActorType() != actorType {
		t.Errorf("expected actor type %s, got %s", actorType, record.ActorType())
	}
	if record.Action() != action {
		t.Errorf("expected action %s, got %s", action, record.Action())
	}
	if record.Resource() != resource {
		t.Errorf("expected resource %s, got %s", resource, record.Resource())
	}
	if record.ResourceID() != resourceID {
		t.Errorf("expected resource ID %s, got %s", resourceID, record.ResourceID())
	}
	if record.Metadata() != metadata {
		t.Errorf("expected metadata %s, got %s", metadata, record.Metadata())
	}
	if record.IP() != ip {
		t.Errorf("expected IP %s, got %s", ip, record.IP())
	}
	if record.UserAgent() != userAgent {
		t.Errorf("expected user agent %s, got %s", userAgent, record.UserAgent())
	}
	if record.Status() != status {
		t.Errorf("expected status %d, got %d", status, record.Status())
	}
	if record.ID() == (RecordID{}) {
		t.Error("expected record ID to be set")
	}
	if record.CreatedAt().IsZero() {
		t.Error("expected created at to be set")
	}
}

func TestReconstituteRecord(t *testing.T) {
	id := NewRecordID()
	actorID := "user-456"
	actorType := ActorSystem
	action := ActionDelete
	resource := "file"
	resourceID := "file-789"
	metadata := `{"reason":"cleanup"}`
	ip := "10.0.0.1"
	userAgent := "CLI/1.0"
	status := 200
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	record := ReconstituteRecord(id, actorID, actorType, action, resource, resourceID, metadata, ip, userAgent, status, createdAt)

	if record.ID() != id {
		t.Errorf("expected ID %s, got %s", id, record.ID())
	}
	if record.ActorID() != actorID {
		t.Errorf("expected actor ID %s, got %s", actorID, record.ActorID())
	}
	if record.CreatedAt() != createdAt {
		t.Errorf("expected created at %v, got %v", createdAt, record.CreatedAt())
	}
}

func TestActionString(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionCreate, "create"},
		{ActionRead, "read"},
		{ActionUpdate, "update"},
		{ActionDelete, "delete"},
		{ActionLogin, "login"},
		{ActionLogout, "logout"},
		{ActionAssignRole, "assign_role"},
		{ActionRevokeRole, "revoke_role"},
		{ActionRegister, "register"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.action.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.action.String())
			}
		})
	}
}

func TestActorTypeString(t *testing.T) {
	tests := []struct {
		actorType ActorType
		expected  string
	}{
		{ActorUser, "user"},
		{ActorSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.actorType.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.actorType.String())
			}
		})
	}
}
