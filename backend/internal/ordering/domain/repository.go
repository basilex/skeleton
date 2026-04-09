package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type OrderFilter struct {
	CustomerID *string
	SupplierID *string
	Status     *OrderStatus
	StartDate  *string
	EndDate    *string
	Cursor     string
	Limit      int
}

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id OrderID) (*Order, error)
	FindByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	FindByCustomerID(ctx context.Context, customerID string, filter OrderFilter) (pagination.PageResult[*Order], error)
	FindBySupplierID(ctx context.Context, supplierID string, filter OrderFilter) (pagination.PageResult[*Order], error)
	FindAll(ctx context.Context, filter OrderFilter) (pagination.PageResult[*Order], error)
	Delete(ctx context.Context, id OrderID) error
}
