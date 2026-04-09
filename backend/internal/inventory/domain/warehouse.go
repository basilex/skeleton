package domain

import (
	"errors"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Warehouse struct {
	id        WarehouseID
	name      string
	code      string
	location  string
	capacity  float64
	status    WarehouseStatus
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	events    []eventbus.Event
}

func NewWarehouse(
	name string,
	code string,
	location string,
) (*Warehouse, error) {
	if name == "" {
		return nil, ErrWarehouseNameEmpty
	}

	now := time.Now()
	warehouse := &Warehouse{
		id:        NewWarehouseID(),
		name:      name,
		code:      code,
		location:  location,
		status:    WarehouseStatusActive,
		metadata:  make(map[string]string),
		createdAt: now,
		updatedAt: now,
		events:    make([]eventbus.Event, 0),
	}

	warehouse.events = append(warehouse.events, WarehouseCreated{
		WarehouseID:   warehouse.id,
		WarehouseName: name,
		Location:      location,
		occurredAt:    now,
	})

	return warehouse, nil
}

func RestoreWarehouse(
	id WarehouseID,
	name string,
	code string,
	location string,
	capacity float64,
	status WarehouseStatus,
	metadata map[string]string,
	createdAt time.Time,
	updatedAt time.Time,
) *Warehouse {
	return &Warehouse{
		id:        id,
		name:      name,
		code:      code,
		location:  location,
		capacity:  capacity,
		status:    status,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
		events:    make([]eventbus.Event, 0),
	}
}

func (w *Warehouse) GetID() WarehouseID {
	return w.id
}

func (w *Warehouse) GetName() string {
	return w.name
}

func (w *Warehouse) GetCode() string {
	return w.code
}

func (w *Warehouse) GetLocation() string {
	return w.location
}

func (w *Warehouse) GetCapacity() float64 {
	return w.capacity
}

func (w *Warehouse) GetStatus() WarehouseStatus {
	return w.status
}

func (w *Warehouse) GetMetadata() map[string]string {
	return w.metadata
}

func (w *Warehouse) GetCreatedAt() time.Time {
	return w.createdAt
}

func (w *Warehouse) GetUpdatedAt() time.Time {
	return w.updatedAt
}

func (w *Warehouse) Activate() error {
	if w.status == WarehouseStatusActive {
		return nil
	}
	if !w.status.CanTransitionTo(WarehouseStatusActive) {
		return errors.New("cannot activate warehouse from current status")
	}
	w.status = WarehouseStatusActive
	w.updatedAt = time.Now()
	return nil
}

func (w *Warehouse) Deactivate() error {
	if w.status == WarehouseStatusInactive {
		return nil
	}
	if !w.status.CanTransitionTo(WarehouseStatusInactive) {
		return errors.New("cannot deactivate warehouse from current status")
	}
	w.status = WarehouseStatusInactive
	w.updatedAt = time.Now()
	return nil
}

func (w *Warehouse) SetMaintenance() error {
	if w.status == WarehouseStatusMaintenance {
		return nil
	}
	if !w.status.CanTransitionTo(WarehouseStatusMaintenance) {
		return errors.New("cannot set warehouse to maintenance from current status")
	}
	w.status = WarehouseStatusMaintenance
	w.updatedAt = time.Now()
	return nil
}

func (w *Warehouse) SetCapacity(capacity float64) error {
	if capacity < 0 {
		return errors.New("capacity cannot be negative")
	}
	w.capacity = capacity
	w.updatedAt = time.Now()
	return nil
}

func (w *Warehouse) IsActive() bool {
	return w.status == WarehouseStatusActive
}

func (w *Warehouse) PullEvents() []eventbus.Event {
	events := w.events
	w.events = make([]eventbus.Event, 0)
	return events
}
