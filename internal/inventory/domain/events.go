package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type WarehouseCreated struct {
	WarehouseID   WarehouseID
	WarehouseName string
	Location      string
	occurredAt    time.Time
}

func (e WarehouseCreated) EventName() string {
	return "inventory.warehouse_created"
}

func (e WarehouseCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type StockAdjusted struct {
	StockID     StockID
	ItemID      string
	WarehouseID WarehouseID
	OldQty      float64
	NewQty      float64
	MovementID  StockMovementID
	Reason      string
	occurredAt  time.Time
}

func (e StockAdjusted) EventName() string {
	return "inventory.stock_adjusted"
}

func (e StockAdjusted) OccurredAt() time.Time {
	return e.occurredAt
}

type StockReserved struct {
	ReservationID StockReservationID
	StockID       StockID
	OrderID       string
	Quantity      float64
	OldReserved   float64
	NewReserved   float64
	OldAvailable  float64
	NewAvailable  float64
	occurredAt    time.Time
}

func (e StockReserved) EventName() string {
	return "inventory.stock_reserved"
}

func (e StockReserved) OccurredAt() time.Time {
	return e.occurredAt
}

type StockReservationReleased struct {
	StockID      StockID
	Quantity     float64
	OldReserved  float64
	NewReserved  float64
	OldAvailable float64
	NewAvailable float64
	occurredAt   time.Time
}

func (e StockReservationReleased) EventName() string {
	return "inventory.stock_reservation_released"
}

func (e StockReservationReleased) OccurredAt() time.Time {
	return e.occurredAt
}

type StockReservationFulfilled struct {
	StockID      StockID
	Quantity     float64
	OldQty       float64
	NewQty       float64
	OldReserved  float64
	NewReserved  float64
	OldAvailable float64
	NewAvailable float64
	occurredAt   time.Time
}

func (e StockReservationFulfilled) EventName() string {
	return "inventory.stock_reservation_fulfilled"
}

func (e StockReservationFulfilled) OccurredAt() time.Time {
	return e.occurredAt
}

type StockMoved struct {
	MovementID    StockMovementID
	MovementType  MovementType
	FromWarehouse WarehouseID
	ToWarehouse   WarehouseID
	occurredAt    time.Time
}

func (e StockMoved) EventName() string {
	return "inventory.stock_moved"
}

func (e StockMoved) OccurredAt() time.Time {
	return e.occurredAt
}

// Lot events
type LotCreated struct {
	LotID      LotID
	ItemID     string
	LotNumber  string
	Quantity   float64
	occurredAt time.Time
}

func (e LotCreated) EventName() string {
	return "inventory.lot_created"
}

func (e LotCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type LotQuantityAdjusted struct {
	LotID      LotID
	OldQty     float64
	NewQty     float64
	occurredAt time.Time
}

func (e LotQuantityAdjusted) EventName() string {
	return "inventory.lot_quantity_adjusted"
}

func (e LotQuantityAdjusted) OccurredAt() time.Time {
	return e.occurredAt
}

type LotExpired struct {
	LotID      LotID
	occurredAt time.Time
}

func (e LotExpired) EventName() string {
	return "inventory.lot_expired"
}

func (e LotExpired) OccurredAt() time.Time {
	return e.occurredAt
}

type LotRecalled struct {
	LotID      LotID
	Reason     string
	occurredAt time.Time
}

func (e LotRecalled) EventName() string {
	return "inventory.lot_recalled"
}

func (e LotRecalled) OccurredAt() time.Time {
	return e.occurredAt
}

// StockTake events
type StockTakeCreated struct {
	StockTakeID StockTakeID
	WarehouseID WarehouseID
	Reference   string
	occurredAt  time.Time
}

func (e StockTakeCreated) EventName() string {
	return "inventory.stock_take_created"
}

func (e StockTakeCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type StockTakeStarted struct {
	StockTakeID StockTakeID
	StartedBy   string
	occurredAt  time.Time
}

func (e StockTakeStarted) EventName() string {
	return "inventory.stock_take_started"
}

func (e StockTakeStarted) OccurredAt() time.Time {
	return e.occurredAt
}

type StockTakeItemCounted struct {
	StockTakeID StockTakeID
	ItemID      string
	SystemQty   float64
	CountedQty  float64
	Variance    float64
	occurredAt  time.Time
}

func (e StockTakeItemCounted) EventName() string {
	return "inventory.stock_take_item_counted"
}

func (e StockTakeItemCounted) OccurredAt() time.Time {
	return e.occurredAt
}

type StockTakeCompleted struct {
	StockTakeID StockTakeID
	CompletedBy string
	occurredAt  time.Time
}

func (e StockTakeCompleted) EventName() string {
	return "inventory.stock_take_completed"
}

func (e StockTakeCompleted) OccurredAt() time.Time {
	return e.occurredAt
}

type StockTakeCancelled struct {
	StockTakeID StockTakeID
	Reason      string
	occurredAt  time.Time
}

func (e StockTakeCancelled) EventName() string {
	return "inventory.stock_take_cancelled"
}

func (e StockTakeCancelled) OccurredAt() time.Time {
	return e.occurredAt
}
