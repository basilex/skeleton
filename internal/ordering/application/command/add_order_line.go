package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/ordering/domain"
)

type AddOrderLineHandler struct {
	orders domain.OrderRepository
}

func NewAddOrderLineHandler(orders domain.OrderRepository) *AddOrderLineHandler {
	return &AddOrderLineHandler{orders: orders}
}

type AddOrderLineCommand struct {
	OrderID   string
	ItemID    string
	ItemName  string
	Quantity  float64
	Unit      string
	UnitPrice float64
	Discount  float64
}

func (h *AddOrderLineHandler) Handle(ctx context.Context, cmd AddOrderLineCommand) error {
	orderID, err := domain.ParseOrderID(cmd.OrderID)
	if err != nil {
		return fmt.Errorf("parse order id: %w", err)
	}

	order, err := h.orders.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}

	line, err := domain.NewOrderLine(orderID, cmd.ItemID, cmd.ItemName, cmd.Quantity, cmd.Unit, cmd.UnitPrice, cmd.Discount)
	if err != nil {
		return fmt.Errorf("create order line: %w", err)
	}

	if err := order.AddLine(line); err != nil {
		return fmt.Errorf("add line to order: %w", err)
	}

	if err := h.orders.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	return nil
}
