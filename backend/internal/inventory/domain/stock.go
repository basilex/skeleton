package domain

import (
	"errors"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Stock struct {
	id              StockID
	itemID          string
	warehouseID     WarehouseID
	quantity        float64
	reservedQty     float64
	availableQty    float64
	reorderPoint    float64
	reorderQuantity float64
	lastMovementID  StockMovementID
	createdAt       time.Time
	updatedAt       time.Time
	events          []eventbus.Event
}

func NewStock(
	itemID string,
	warehouseID WarehouseID,
) (*Stock, error) {
	if itemID == "" {
		return nil, errors.New("item ID cannot be empty")
	}

	now := time.Now().UTC()
	stock := &Stock{
		id:           NewStockID(),
		itemID:       itemID,
		warehouseID:  warehouseID,
		quantity:     0,
		reservedQty:  0,
		availableQty: 0,
		createdAt:    now,
		updatedAt:    now,
		events:       make([]eventbus.Event, 0),
	}

	return stock, nil
}

func RestoreStock(
	id StockID,
	itemID string,
	warehouseID WarehouseID,
	quantity float64,
	reservedQty float64,
	availableQty float64,
	reorderPoint float64,
	reorderQuantity float64,
	lastMovementID StockMovementID,
	createdAt time.Time,
	updatedAt time.Time,
) *Stock {
	return &Stock{
		id:              id,
		itemID:          itemID,
		warehouseID:     warehouseID,
		quantity:        quantity,
		reservedQty:     reservedQty,
		availableQty:    availableQty,
		reorderPoint:    reorderPoint,
		reorderQuantity: reorderQuantity,
		lastMovementID:  lastMovementID,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		events:          make([]eventbus.Event, 0),
	}
}

func (s *Stock) GetID() StockID {
	return s.id
}

func (s *Stock) GetItemID() string {
	return s.itemID
}

func (s *Stock) GetWarehouseID() WarehouseID {
	return s.warehouseID
}

func (s *Stock) GetQuantity() float64 {
	return s.quantity
}

func (s *Stock) GetReservedQty() float64 {
	return s.reservedQty
}

func (s *Stock) GetAvailableQty() float64 {
	return s.availableQty
}

func (s *Stock) GetReorderPoint() float64 {
	return s.reorderPoint
}

func (s *Stock) GetReorderQuantity() float64 {
	return s.reorderQuantity
}

func (s *Stock) GetLastMovementID() StockMovementID {
	return s.lastMovementID
}

func (s *Stock) GetCreatedAt() time.Time {
	return s.createdAt
}

func (s *Stock) GetUpdatedAt() time.Time {
	return s.updatedAt
}

// PullEvents returns all pending domain events and clears the buffer.
func (s *Stock) PullEvents() []eventbus.Event {
	events := s.events
	s.events = make([]eventbus.Event, 0)
	return events
}

// publishEvent adds an event to the aggregate's event buffer.
func (s *Stock) publishEvent(event eventbus.Event) {
	s.events = append(s.events, event)
	s.updatedAt = time.Now().UTC()
}

func (s *Stock) AdjustQuantity(quantity float64, movementID StockMovementID) {
	oldQty := s.quantity
	s.quantity += quantity
	s.availableQty = s.quantity - s.reservedQty
	s.lastMovementID = movementID
	s.updatedAt = time.Now().UTC()

	s.publishEvent(StockAdjusted{
		StockID:     s.id,
		ItemID:      s.itemID,
		WarehouseID: s.warehouseID,
		OldQty:      oldQty,
		NewQty:      s.quantity,
		MovementID:  movementID,
		Reason:      "adjustment",
		occurredAt:  s.updatedAt,
	})
}

func (s *Stock) Reserve(reservedQty float64, reservationID StockReservationID) error {
	if s.availableQty < reservedQty {
		return ErrInsufficientStock
	}

	oldReserved := s.reservedQty
	oldAvailable := s.availableQty

	s.reservedQty += reservedQty
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now().UTC()

	s.publishEvent(StockReserved{
		ReservationID: reservationID,
		StockID:       s.id,
		OrderID:       "",
		Quantity:      reservedQty,
		OldReserved:   oldReserved,
		NewReserved:   s.reservedQty,
		OldAvailable:  oldAvailable,
		NewAvailable:  s.availableQty,
		occurredAt:    s.updatedAt,
	})

	return nil
}

func (s *Stock) ReleaseReservation(reservedQty float64) {
	if s.reservedQty < reservedQty {
		reservedQty = s.reservedQty
	}

	oldReserved := s.reservedQty
	oldAvailable := s.availableQty

	s.reservedQty -= reservedQty
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now().UTC()

	s.publishEvent(StockReservationReleased{
		StockID:      s.id,
		Quantity:     reservedQty,
		OldReserved:  oldReserved,
		NewReserved:  s.reservedQty,
		OldAvailable: oldAvailable,
		NewAvailable: s.availableQty,
		occurredAt:   s.updatedAt,
	})
}

func (s *Stock) FulfillReservation(quantity float64) {
	if s.reservedQty < quantity {
		quantity = s.reservedQty
	}

	oldQty := s.quantity
	oldReserved := s.reservedQty
	oldAvailable := s.availableQty

	s.quantity -= quantity
	s.reservedQty -= quantity
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now().UTC()

	s.publishEvent(StockReservationFulfilled{
		StockID:      s.id,
		Quantity:     quantity,
		OldQty:       oldQty,
		NewQty:       s.quantity,
		OldReserved:  oldReserved,
		NewReserved:  s.reservedQty,
		OldAvailable: oldAvailable,
		NewAvailable: s.availableQty,
		occurredAt:   s.updatedAt,
	})
}

func (s *Stock) SetReorderPoint(reorderPoint float64) error {
	if reorderPoint < 0 {
		return errors.New("reorder point cannot be negative")
	}

	s.reorderPoint = reorderPoint
	s.updatedAt = time.Now()
	return nil
}

func (s *Stock) SetReorderQuantity(reorderQuantity float64) error {
	if reorderQuantity <= 0 {
		return errors.New("reorder quantity must be positive")
	}

	s.reorderQuantity = reorderQuantity
	s.updatedAt = time.Now()
	return nil
}

func (s *Stock) NeedsReorder() bool {
	return s.availableQty <= s.reorderPoint
}

func (s *Stock) IsAvailable(quantity float64) bool {
	return s.availableQty >= quantity
}
