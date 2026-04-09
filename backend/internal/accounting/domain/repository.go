package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type AccountFilter struct {
	AccountType *AccountType
	IsActive    *bool
	Search      string
	Cursor      string
	Limit       int
}

type AccountRepository interface {
	Save(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id AccountID) (*Account, error)
	FindByCode(ctx context.Context, code string) (*Account, error)
	FindAll(ctx context.Context, filter AccountFilter) (pagination.PageResult[*Account], error)
	Delete(ctx context.Context, id AccountID) error
}
