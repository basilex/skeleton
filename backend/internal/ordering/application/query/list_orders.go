package query

import (
	"context"

	"github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type ListOrdersHandler struct {
	orders domain.OrderRepository
}

func NewListOrdersHandler(orders domain.OrderRepository) *ListOrdersHandler {
	return &ListOrdersHandler{orders: orders}
}

type ListOrdersQuery struct {
	CustomerID *string
	SupplierID *string
	Status     *string
	StartDate  *string
	EndDate    *string
	Cursor     string
	Limit      int
}

func (h *ListOrdersHandler) Handle(ctx context.Context, q ListOrdersQuery) (pagination.PageResult[OrderDTO], error) {
	filter := domain.OrderFilter{
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}

	if q.CustomerID != nil {
		filter.CustomerID = q.CustomerID
	}
	if q.SupplierID != nil {
		filter.SupplierID = q.SupplierID
	}
	if q.Status != nil {
		status, err := domain.ParseOrderStatus(*q.Status)
		if err == nil {
			filter.Status = &status
		}
	}
	if q.StartDate != nil {
		filter.StartDate = q.StartDate
	}
	if q.EndDate != nil {
		filter.EndDate = q.EndDate
	}

	result, err := h.orders.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[OrderDTO]{}, err
	}

	dtos := make([]OrderDTO, 0, len(result.Items))
	for _, order := range result.Items {
		lines := make([]OrderLineDTO, 0, len(order.GetLines()))
		for _, line := range order.GetLines() {
			lines = append(lines, OrderLineDTO{
				ID:        line.GetID().String(),
				ItemID:    line.GetItemID(),
				ItemName:  line.GetItemName(),
				Quantity:  line.GetQuantity(),
				Unit:      line.GetUnit(),
				UnitPrice: line.GetUnitPrice().ToFloat64(),
				Discount:  line.GetDiscount().ToFloat64(),
				Total:     line.GetTotal().ToFloat64(),
				CreatedAt: line.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		var dueDate *string
		if order.GetDueDate() != nil {
			d := order.GetDueDate().Format("2006-01-02")
			dueDate = &d
		}

		var completedAt *string
		if order.GetCompletedAt() != nil {
			c := order.GetCompletedAt().Format("2006-01-02T15:04:05Z07:00")
			completedAt = &c
		}

		var cancelledAt *string
		if order.GetCancelledAt() != nil {
			c := order.GetCancelledAt().Format("2006-01-02T15:04:05Z07:00")
			cancelledAt = &c
		}

		dtos = append(dtos, OrderDTO{
			ID:          order.GetID().String(),
			OrderNumber: order.GetOrderNumber(),
			Status:      order.GetStatus().String(),
			CustomerID:  order.GetCustomerID(),
			SupplierID:  order.GetSupplierID(),
			ContractID:  order.GetContractID(),
			Subtotal:    order.GetSubtotal().ToFloat64(),
			TaxAmount:   order.GetTaxAmount().ToFloat64(),
			Discount:    order.GetDiscount().ToFloat64(),
			Total:       order.GetTotal().ToFloat64(),
			Currency:    order.GetCurrency(),
			Lines:       lines,
			OrderDate:   order.GetOrderDate().Format("2006-01-02T15:04:05Z07:00"),
			DueDate:     dueDate,
			CompletedAt: completedAt,
			CancelledAt: cancelledAt,
			Notes:       order.GetNotes(),
			CreatedBy:   order.GetCreatedBy(),
			CreatedAt:   order.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   order.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return pagination.PageResult[OrderDTO]{
		Items:      dtos,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		Limit:      result.Limit,
	}, nil
}
