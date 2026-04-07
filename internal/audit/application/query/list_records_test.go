package query

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type mockAuditRepo struct {
	records []*domain.Record
	err     error
}

func (m *mockAuditRepo) Save(ctx context.Context, record *domain.Record) error {
	return nil
}

func (m *mockAuditRepo) FindAll(ctx context.Context, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	if m.err != nil {
		return pagination.PageResult[*domain.Record]{}, m.err
	}
	return pagination.NewPageResult(m.records, filter.Limit), nil
}

func (m *mockAuditRepo) FindByActorID(ctx context.Context, actorID string, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	if m.err != nil {
		return pagination.PageResult[*domain.Record]{}, m.err
	}
	return pagination.NewPageResult(m.records, filter.Limit), nil
}

func (m *mockAuditRepo) DeleteBefore(ctx context.Context, before time.Time) (int, error) {
	return 0, nil
}

func TestListRecordsHandler_Handle(t *testing.T) {
	record1 := domain.NewRecord("user-1", domain.ActorUser, domain.ActionLogin, "auth", "user-1", "", "127.0.0.1", "test", 200)
	record2 := domain.NewRecord("user-2", domain.ActorUser, domain.ActionRegister, "user", "user-2", "", "127.0.0.1", "test", 201)

	tests := []struct {
		name    string
		query   ListRecordsQuery
		records []*domain.Record
		wantLen int
	}{
		{
			name:    "list all records",
			query:   ListRecordsQuery{Limit: 20},
			records: []*domain.Record{record1, record2},
			wantLen: 2,
		},
		{
			name:    "list with actor filter",
			query:   ListRecordsQuery{ActorID: "user-1", Limit: 20},
			records: []*domain.Record{record1},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{records: tt.records}
			handler := NewListRecordsHandler(repo)

			result, err := handler.Handle(context.Background(), tt.query)
			require.NoError(t, err)
			require.Len(t, result.Items, tt.wantLen)
		})
	}
}
