package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/ordering/domain"
)

type UpdateOrderStatusHandler struct {
	orders domain.OrderRepository
}

func NewUpdateOrderStatusHandler(orders domain.OrderRepository) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{orders: orders}
}

type UpdateOrderStatusCommand struct {
	OrderID string
	Status  string
	Reason  string
}

func (h *UpdateOrderStatusHandler) Handle(ctx context.Context, cmd UpdateOrderStatusCommand) error {
	orderID, err := domain.ParseOrderID(cmd.OrderID)
	if err != nil {
		return fmt.Errorf("parse order id: %w", err)
	}

	order, err := h.orders.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}

	status, err := domain.ParseOrderStatus(cmd.Status)
	if err != nil {
		return fmt.Errorf("parse status: %w", err)
	}

	switch status {
	case domain.OrderStatusConfirmed:
		if err := order.Confirm(); err != nil {
			return fmt.Errorf("confirm order: %w", err)
		}
	case domain.OrderStatusCompleted:
		if err := order.Complete(); err != nil {
			return fmt.Errorf("complete order: %w", err)
		}
	case domain.OrderStatusCancelled:
		if err := order.Cancel(cmd.Reason); err != nil {
			return fmt.Errorf("cancel order: %w", err)
		}
	default:
		return fmt.Errorf("unsupported status transition: %s", status)
	}

	if err := h.orders.Save(ctx, order); err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	return nil
}
