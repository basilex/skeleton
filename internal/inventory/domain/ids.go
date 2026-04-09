package domain

import (
	"github.com/basilex/skeleton/pkg/uuid"
)

type WarehouseID uuid.UUID

func NewWarehouseID() WarehouseID {
	return WarehouseID(uuid.NewV7())
}

func ParseWarehouseID(s string) (WarehouseID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return WarehouseID{}, err
	}
	return WarehouseID(id), nil
}

func (id WarehouseID) String() string {
	return uuid.UUID(id).String()
}

type StockID uuid.UUID

func NewStockID() StockID {
	return StockID(uuid.NewV7())
}

func ParseStockID(s string) (StockID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return StockID{}, err
	}
	return StockID(id), nil
}

func (id StockID) String() string {
	return uuid.UUID(id).String()
}

type StockMovementID uuid.UUID

func NewStockMovementID() StockMovementID {
	return StockMovementID(uuid.NewV7())
}

func ParseStockMovementID(s string) (StockMovementID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return StockMovementID{}, err
	}
	return StockMovementID(id), nil
}

func (id StockMovementID) String() string {
	return uuid.UUID(id).String()
}

type StockReservationID uuid.UUID

func NewStockReservationID() StockReservationID {
	return StockReservationID(uuid.NewV7())
}

func ParseStockReservationID(s string) (StockReservationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return StockReservationID{}, err
	}
	return StockReservationID(id), nil
}

func (id StockReservationID) String() string {
	return uuid.UUID(id).String()
}
