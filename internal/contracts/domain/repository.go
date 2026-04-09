package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type ContractFilter struct {
	PartyID       *string
	ContractType  *ContractType
	Status        *ContractStatus
	ActiveOnly    bool
	ExpiresBefore *string
	Cursor        string
	Limit         int
}

type ContractRepository interface {
	Save(ctx context.Context, contract *Contract) error
	FindByID(ctx context.Context, id ContractID) (*Contract, error)
	FindByPartyID(ctx context.Context, partyID string, filter ContractFilter) (pagination.PageResult[*Contract], error)
	FindAll(ctx context.Context, filter ContractFilter) (pagination.PageResult[*Contract], error)
	Delete(ctx context.Context, id ContractID) error
}
