package query

import (
	"context"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type ListInvoicesHandler struct {
	invoices domain.InvoiceRepository
}

func NewListInvoicesHandler(invoices domain.InvoiceRepository) *ListInvoicesHandler {
	return &ListInvoicesHandler{
		invoices: invoices,
	}
}

type ListInvoicesQuery struct {
	CustomerID *string
	SupplierID *string
	OrderID    *string
	ContractID *string
	Status     *string
	StartDate  *string
	EndDate    *string
	Overdue    *bool
	Cursor     string
	Limit      int
}

func (h *ListInvoicesHandler) Handle(ctx context.Context, query ListInvoicesQuery) (pagination.PageResult[*InvoiceDTO], error) {
	var status *domain.InvoiceStatus
	if query.Status != nil {
		s := domain.InvoiceStatus(*query.Status)
		status = &s
	}

	filter := domain.InvoiceFilter{
		CustomerID: query.CustomerID,
		SupplierID: query.SupplierID,
		OrderID:    query.OrderID,
		ContractID: query.ContractID,
		Status:     status,
		StartDate:  query.StartDate,
		EndDate:    query.EndDate,
		Overdue:    query.Overdue,
		Cursor:     query.Cursor,
		Limit:      query.Limit,
	}

	result, err := h.invoices.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[*InvoiceDTO]{}, err
	}

	dtos := make([]*InvoiceDTO, 0, len(result.Items))
	for _, invoice := range result.Items {
		dtos = append(dtos, toDTO(invoice))
	}

	return pagination.PageResult[*InvoiceDTO]{
		Items:      dtos,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}, nil
}
