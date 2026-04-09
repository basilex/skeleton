package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type UpdateWarehouseHandler struct {
	warehouses domain.WarehouseRepository
	bus        eventbus.Bus
}

func NewUpdateWarehouseHandler(warehouses domain.WarehouseRepository, bus eventbus.Bus) *UpdateWarehouseHandler {
	return &UpdateWarehouseHandler{
		warehouses: warehouses,
		bus:        bus,
	}
}

type UpdateWarehouseCommand struct {
	WarehouseID string
	Name        *string
	Location    *string
	Capacity    *float64
	Activate    bool
	Deactivate  bool
	Maintenance bool
}

type UpdateWarehouseResult struct {
	WarehouseID string
}

func (h *UpdateWarehouseHandler) Handle(ctx context.Context, cmd UpdateWarehouseCommand) (*UpdateWarehouseResult, error) {
	warehouseID, err := domain.ParseWarehouseID(cmd.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse ID: %w", err)
	}

	warehouse, err := h.warehouses.FindByID(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("find warehouse: %w", err)
	}

	if cmd.Activate {
		if err := warehouse.Activate(); err != nil {
			return nil, fmt.Errorf("activate warehouse: %w", err)
		}
	}

	if cmd.Deactivate {
		if err := warehouse.Deactivate(); err != nil {
			return nil, fmt.Errorf("deactivate warehouse: %w", err)
		}
	}

	if cmd.Maintenance {
		if err := warehouse.SetMaintenance(); err != nil {
			return nil, fmt.Errorf("set maintenance: %w", err)
		}
	}

	if cmd.Capacity != nil {
		if err := warehouse.SetCapacity(*cmd.Capacity); err != nil {
			return nil, fmt.Errorf("set capacity: %w", err)
		}
	}

	if err := h.warehouses.Save(ctx, warehouse); err != nil {
		return nil, fmt.Errorf("save warehouse: %w", err)
	}

	events := warehouse.PullEvents()
	for _, event := range events {
		if err := h.bus.Publish(ctx, event); err != nil {
		}
	}

	return &UpdateWarehouseResult{
		WarehouseID: warehouse.GetID().String(),
	}, nil
}
