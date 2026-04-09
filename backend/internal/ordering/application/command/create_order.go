package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateOrderHandler struct {
	orders domain.OrderRepository
	bus    eventbus.Bus
}

func NewCreateOrderHandler(orders domain.OrderRepository, bus eventbus.Bus) *CreateOrderHandler {
	return &CreateOrderHandler{
		orders: orders,
		bus:    bus,
	}
}

type CreateOrderCommand struct {
	OrderNumber string
	CustomerID  string
	SupplierID  string
	ContractID  string
	Currency    string
	CreatedBy   string
}

type CreateOrderResult struct {
	OrderID string
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (CreateOrderResult, error) {
	order, err := domain.NewOrder(
		cmd.OrderNumber,
		cmd.CustomerID,
		cmd.SupplierID,
		cmd.ContractID,
		cmd.Currency,
		cmd.CreatedBy,
	)
	if err != nil {
		return CreateOrderResult{}, fmt.Errorf("create order: %w", err)
	}

	if err := h.orders.Save(ctx, order); err != nil {
		return CreateOrderResult{}, fmt.Errorf("save order: %w", err)
	}

	events := order.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateOrderResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateOrderResult{
		OrderID: order.GetID().String(),
	}, nil
}
