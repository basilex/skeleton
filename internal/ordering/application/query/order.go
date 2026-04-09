package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/ordering/domain"
)

type GetOrderHandler struct {
	orders domain.OrderRepository
}

func NewGetOrderHandler(orders domain.OrderRepository) *GetOrderHandler {
	return &GetOrderHandler{orders: orders}
}

type GetOrderQuery struct {
	OrderID string
}

type OrderLineDTO struct {
	ID        string  `json:"id"`
	ItemID    string  `json:"item_id"`
	ItemName  string  `json:"item_name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unit_price"`
	Discount  float64 `json:"discount"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"created_at"`
}

type OrderDTO struct {
	ID          string         `json:"id"`
	OrderNumber string         `json:"order_number"`
	Status      string         `json:"status"`
	CustomerID  string         `json:"customer_id"`
	SupplierID  string         `json:"supplier_id"`
	ContractID  string         `json:"contract_id,omitempty"`
	Subtotal    float64        `json:"subtotal"`
	TaxAmount   float64        `json:"tax_amount"`
	Discount    float64        `json:"discount"`
	Total       float64        `json:"total"`
	Currency    string         `json:"currency"`
	Lines       []OrderLineDTO `json:"lines"`
	OrderDate   string         `json:"order_date"`
	DueDate     *string        `json:"due_date,omitempty"`
	CompletedAt *string        `json:"completed_at,omitempty"`
	CancelledAt *string        `json:"cancelled_at,omitempty"`
	Notes       string         `json:"notes,omitempty"`
	CreatedBy   string         `json:"created_by,omitempty"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

func (h *GetOrderHandler) Handle(ctx context.Context, q GetOrderQuery) (OrderDTO, error) {
	orderID, err := domain.ParseOrderID(q.OrderID)
	if err != nil {
		return OrderDTO{}, fmt.Errorf("parse order id: %w", err)
	}

	order, err := h.orders.FindByID(ctx, orderID)
	if err != nil {
		return OrderDTO{}, fmt.Errorf("find order: %w", err)
	}

	lines := make([]OrderLineDTO, 0, len(order.GetLines()))
	for _, line := range order.GetLines() {
		lines = append(lines, OrderLineDTO{
			ID:        line.GetID().String(),
			ItemID:    line.GetItemID(),
			ItemName:  line.GetItemName(),
			Quantity:  line.GetQuantity(),
			Unit:      line.GetUnit(),
			UnitPrice: line.GetUnitPrice(),
			Discount:  line.GetDiscount(),
			Total:     line.GetTotal(),
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

	return OrderDTO{
		ID:          order.GetID().String(),
		OrderNumber: order.GetOrderNumber(),
		Status:      order.GetStatus().String(),
		CustomerID:  order.GetCustomerID(),
		SupplierID:  order.GetSupplierID(),
		ContractID:  order.GetContractID(),
		Subtotal:    order.GetSubtotal(),
		TaxAmount:   order.GetTaxAmount(),
		Discount:    order.GetDiscount(),
		Total:       order.GetTotal(),
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
	}, nil
}
