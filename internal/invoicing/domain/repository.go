package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type InvoiceFilter struct {
	CustomerID *string
	SupplierID *string
	OrderID    *string
	ContractID *string
	Status     *InvoiceStatus
	StartDate  *string
	EndDate    *string
	Overdue    *bool
	Cursor     string
	Limit      int
}

type InvoiceRepository interface {
	Save(ctx context.Context, invoice *Invoice) error
	FindByID(ctx context.Context, id InvoiceID) (*Invoice, error)
	FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*Invoice, error)
	FindByOrderID(ctx context.Context, orderID string) (*Invoice, error)
	FindByCustomerID(ctx context.Context, customerID string, filter InvoiceFilter) (pagination.PageResult[*Invoice], error)
	FindAll(ctx context.Context, filter InvoiceFilter) (pagination.PageResult[*Invoice], error)
	Delete(ctx context.Context, id InvoiceID) error
}

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id PaymentID) (*Payment, error)
	FindByInvoiceID(ctx context.Context, invoiceID InvoiceID) ([]*Payment, error)
	Delete(ctx context.Context, id PaymentID) error
}
