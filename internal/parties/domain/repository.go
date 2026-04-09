package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type PartyFilter struct {
	PartyType *PartyType
	Status    *PartyStatus
	Search    string
	TaxID     string
	Cursor    string
	Limit     int
}

type CustomerRepository interface {
	Save(ctx context.Context, customer *Customer) error
	FindByID(ctx context.Context, id PartyID) (*Customer, error)
	FindAll(ctx context.Context, filter PartyFilter) (pagination.PageResult[*Customer], error)
	FindByEmail(ctx context.Context, email string) (*Customer, error)
	Delete(ctx context.Context, id PartyID) error
}

type SupplierRepository interface {
	Save(ctx context.Context, supplier *Supplier) error
	FindByID(ctx context.Context, id PartyID) (*Supplier, error)
	FindAll(ctx context.Context, filter PartyFilter) (pagination.PageResult[*Supplier], error)
	FindByEmail(ctx context.Context, email string) (*Supplier, error)
	Delete(ctx context.Context, id PartyID) error
}

type PartnerRepository interface {
	Save(ctx context.Context, partner *Partner) error
	FindByID(ctx context.Context, id PartyID) (*Partner, error)
	FindAll(ctx context.Context, filter PartyFilter) (pagination.PageResult[*Partner], error)
	FindByEmail(ctx context.Context, email string) (*Partner, error)
	Delete(ctx context.Context, id PartyID) error
}

type EmployeeRepository interface {
	Save(ctx context.Context, employee *Employee) error
	FindByID(ctx context.Context, id PartyID) (*Employee, error)
	FindAll(ctx context.Context, filter PartyFilter) (pagination.PageResult[*Employee], error)
	FindByEmail(ctx context.Context, email string) (*Employee, error)
	Delete(ctx context.Context, id PartyID) error
}
