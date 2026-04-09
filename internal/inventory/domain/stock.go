package domain

import (
	"errors"
	"time"
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
}

func NewStock(
	itemID string,
	warehouseID WarehouseID,
) (*Stock, error) {
	if itemID == "" {
		return nil, errors.New("item ID cannot be empty")
	}

	now := time.Now()
	return &Stock{
		id:           NewStockID(),
		itemID:       itemID,
		warehouseID:  warehouseID,
		quantity:     0,
		reservedQty:  0,
		availableQty: 0,
		createdAt:    now,
		updatedAt:    now,
	}, nil
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

func (s *Stock) AdjustQuantity(quantity float64, movementID StockMovementID) {
	s.quantity += quantity
	s.availableQty = s.quantity - s.reservedQty
	s.lastMovementID = movementID
	s.updatedAt = time.Now()
}

func (s *Stock) Reserve(reservedQty float64) error {
	if s.availableQty < reservedQty {
		return ErrInsufficientStock
	}

	s.reservedQty += reservedQty
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now()
	return nil
}

func (s *Stock) ReleaseReservation(reservedQty float64) {
	if s.reservedQty < reservedQty {
		reservedQty = s.reservedQty
	}

	s.reservedQty -= reservedQty
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now()
}

func (s *Stock) FulfillReservation(quantity float64) {
	if s.reservedQty < quantity {
		quantity = s.reservedQty
	}

	s.quantity -= quantity
	s.reservedQty -= quantity
	s.availableQty = s.quantity - s.reservedQty
	s.updatedAt = time.Now()
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
