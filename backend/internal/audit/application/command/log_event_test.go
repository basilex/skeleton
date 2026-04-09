package command

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type mockRepository struct {
	savedRecord *domain.Record
	saveErr     error
}

func (m *mockRepository) Save(ctx context.Context, record *domain.Record) error {
	m.savedRecord = record
	return m.saveErr
}

func (m *mockRepository) FindAll(ctx context.Context, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	return pagination.PageResult[*domain.Record]{}, nil
}

func (m *mockRepository) FindByActorID(ctx context.Context, actorID string, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	return pagination.PageResult[*domain.Record]{}, nil
}

func (m *mockRepository) DeleteBefore(ctx context.Context, before time.Time) (int, error) {
	return 0, nil
}

func TestLogEventHandler_Handle(t *testing.T) {
	repo := &mockRepository{}
	handler := NewLogEventHandler(repo)

	cmd := LogEventCommand{
		ActorID:    "user-123",
		ActorType:  domain.ActorUser,
		Action:     domain.ActionLogin,
		Resource:   "auth",
		ResourceID: "user-123",
		Metadata:   `{"ip":"192.168.1.1"}`,
		IP:         "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Status:     200,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if repo.savedRecord == nil {
		t.Fatal("expected record to be saved")
	}

	if repo.savedRecord.ActorID() != cmd.ActorID {
		t.Errorf("expected actor ID %s, got %s", cmd.ActorID, repo.savedRecord.ActorID())
	}
	if repo.savedRecord.Action() != cmd.Action {
		t.Errorf("expected action %s, got %s", cmd.Action, repo.savedRecord.Action())
	}
}
