package domain

import (
	"time"
)

type StockMovement struct {
	id            StockMovementID
	movementType  MovementType
	itemID        string
	fromWarehouse WarehouseID
	toWarehouse   WarehouseID
	quantity      float64
	referenceID   string
	referenceType string
	notes         string
	occurredAt    time.Time
	createdAt     time.Time
}

func NewStockMovement(
	movementType MovementType,
	itemID string,
	quantity float64,
) (*StockMovement, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	now := time.Now()
	return &StockMovement{
		id:           NewStockMovementID(),
		movementType: movementType,
		itemID:       itemID,
		quantity:     quantity,
		occurredAt:   now,
		createdAt:    now,
	}, nil
}

func NewReceipt(
	itemID string,
	warehouseID WarehouseID,
	quantity float64,
	referenceID string,
) (*StockMovement, error) {
	movement, err := NewStockMovement(MovementTypeReceipt, itemID, quantity)
	if err != nil {
		return nil, err
	}

	movement.toWarehouse = warehouseID
	movement.referenceID = referenceID
	movement.referenceType = "receipt"
	return movement, nil
}

func NewIssue(
	itemID string,
	warehouseID WarehouseID,
	quantity float64,
	orderID string,
) (*StockMovement, error) {
	movement, err := NewStockMovement(MovementTypeIssue, itemID, quantity)
	if err != nil {
		return nil, err
	}

	movement.fromWarehouse = warehouseID
	movement.referenceID = orderID
	movement.referenceType = "order"
	return movement, nil
}

func NewTransfer(
	itemID string,
	fromWarehouse WarehouseID,
	toWarehouse WarehouseID,
	quantity float64,
) (*StockMovement, error) {
	if fromWarehouse == toWarehouse {
		return nil, ErrSameWarehouseTransfer
	}

	movement, err := NewStockMovement(MovementTypeTransfer, itemID, quantity)
	if err != nil {
		return nil, err
	}

	movement.fromWarehouse = fromWarehouse
	movement.toWarehouse = toWarehouse
	movement.referenceType = "transfer"
	return movement, nil
}

func NewAdjustment(
	itemID string,
	warehouseID WarehouseID,
	quantity float64,
	reason string,
) (*StockMovement, error) {
	movement, err := NewStockMovement(MovementTypeAdjustment, itemID, quantity)
	if err != nil {
		return nil, err
	}

	movement.fromWarehouse = warehouseID
	movement.toWarehouse = warehouseID
	movement.notes = reason
	movement.referenceType = "adjustment"
	return movement, nil
}

func RestoreStockMovement(
	id StockMovementID,
	movementType MovementType,
	itemID string,
	fromWarehouse WarehouseID,
	toWarehouse WarehouseID,
	quantity float64,
	referenceID string,
	referenceType string,
	notes string,
	occurredAt time.Time,
	createdAt time.Time,
) *StockMovement {
	return &StockMovement{
		id:            id,
		movementType:  movementType,
		itemID:        itemID,
		fromWarehouse: fromWarehouse,
		toWarehouse:   toWarehouse,
		quantity:      quantity,
		referenceID:   referenceID,
		referenceType: referenceType,
		notes:         notes,
		occurredAt:    occurredAt,
		createdAt:     createdAt,
	}
}

func (m *StockMovement) GetID() StockMovementID {
	return m.id
}

func (m *StockMovement) GetMovementType() MovementType {
	return m.movementType
}

func (m *StockMovement) GetItemID() string {
	return m.itemID
}

func (m *StockMovement) GetFromWarehouse() WarehouseID {
	return m.fromWarehouse
}

func (m *StockMovement) GetToWarehouse() WarehouseID {
	return m.toWarehouse
}

func (m *StockMovement) GetQuantity() float64 {
	return m.quantity
}

func (m *StockMovement) GetReferenceID() string {
	return m.referenceID
}

func (m *StockMovement) GetReferenceType() string {
	return m.referenceType
}

func (m *StockMovement) GetNotes() string {
	return m.notes
}

func (m *StockMovement) GetOccurredAt() time.Time {
	return m.occurredAt
}

func (m *StockMovement) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *StockMovement) IsInbound() bool {
	return m.movementType.IsInbound()
}

func (m *StockMovement) IsOutbound() bool {
	return m.movementType.IsOutbound()
}
