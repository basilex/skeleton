package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateWarehouseHandler struct {
	warehouses domain.WarehouseRepository
	bus        eventbus.Bus
}

func NewCreateWarehouseHandler(warehouses domain.WarehouseRepository, bus eventbus.Bus) *CreateWarehouseHandler {
	return &CreateWarehouseHandler{
		warehouses: warehouses,
		bus:        bus,
	}
}

type CreateWarehouseCommand struct {
	Name     string
	Code     string
	Location string
}

type CreateWarehouseResult struct {
	WarehouseID string
}

func (h *CreateWarehouseHandler) Handle(ctx context.Context, cmd CreateWarehouseCommand) (*CreateWarehouseResult, error) {
	warehouse, err := domain.NewWarehouse(cmd.Name, cmd.Code, cmd.Location)
	if err != nil {
		return nil, fmt.Errorf("create warehouse: %w", err)
	}

	if err := h.warehouses.Save(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("save warehouse: %w", err)
	}

	events := warehouse.PullEvents()
	for _, event := range events {
		if err := h.bus.Publish(ctx, event); err != nil {
		}
	}

	return &CreateWarehouseResult{
		WarehouseID: warehouse.GetID().String(),
	}, nil
}
