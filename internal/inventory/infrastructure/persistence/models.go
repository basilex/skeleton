package persistence

import (
	"time"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type warehouseDTO struct {
	ID        string            `db:"id"`
	Name      string            `db:"name"`
	Code      string            `db:"code"`
	Location  string            `db:"location"`
	Capacity  float64           `db:"capacity"`
	Status    string            `db:"status"`
	Metadata  map[string]string `db:"metadata"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

func (dto *warehouseDTO) toDomain() (*domain.Warehouse, error) {
	id, err := domain.ParseWarehouseID(dto.ID)
	if err != nil {
		return nil, err
	}

	status := domain.WarehouseStatus(dto.Status)
	return domain.RestoreWarehouse(
		id,
		dto.Name,
		dto.Code,
		dto.Location,
		dto.Capacity,
		status,
		dto.Metadata,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}

type stockDTO struct {
	ID              string    `db:"id"`
	ItemID          string    `db:"item_id"`
	WarehouseID     string    `db:"warehouse_id"`
	Quantity        float64   `db:"quantity"`
	ReservedQty     float64   `db:"reserved_qty"`
	AvailableQty    float64   `db:"available_qty"`
	ReorderPoint    float64   `db:"reorder_point"`
	ReorderQuantity float64   `db:"reorder_quantity"`
	LastMovementID  string    `db:"last_movement_id"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (dto *stockDTO) toDomain() (*domain.Stock, error) {
	id, err := domain.ParseStockID(dto.ID)
	if err != nil {
		return nil, err
	}

	warehouseID, err := domain.ParseWarehouseID(dto.WarehouseID)
	if err != nil {
		return nil, err
	}

	var lastMovementID domain.StockMovementID
	if dto.LastMovementID != "" {
		var err error
		lastMovementID, err = domain.ParseStockMovementID(dto.LastMovementID)
		if err != nil {
			return nil, err
		}
	}

	return domain.RestoreStock(
		id,
		dto.ItemID,
		warehouseID,
		dto.Quantity,
		dto.ReservedQty,
		dto.AvailableQty,
		dto.ReorderPoint,
		dto.ReorderQuantity,
		lastMovementID,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}

type stockMovementDTO struct {
	ID            string    `db:"id"`
	MovementType  string    `db:"movement_type"`
	ItemID        string    `db:"item_id"`
	FromWarehouse string    `db:"from_warehouse"`
	ToWarehouse   string    `db:"to_warehouse"`
	Quantity      float64   `db:"quantity"`
	ReferenceID   string    `db:"reference_id"`
	ReferenceType string    `db:"reference_type"`
	Notes         string    `db:"notes"`
	OccurredAt    time.Time `db:"occurred_at"`
	CreatedAt     time.Time `db:"created_at"`
}

func (dto *stockMovementDTO) toDomain() (*domain.StockMovement, error) {
	id, err := domain.ParseStockMovementID(dto.ID)
	if err != nil {
		return nil, err
	}

	var fromWarehouse, toWarehouse domain.WarehouseID
	if dto.FromWarehouse != "" {
		fromWarehouse, err = domain.ParseWarehouseID(dto.FromWarehouse)
		if err != nil {
			return nil, err
		}
	}
	if dto.ToWarehouse != "" {
		toWarehouse, err = domain.ParseWarehouseID(dto.ToWarehouse)
		if err != nil {
			return nil, err
		}
	}

	return domain.RestoreStockMovement(
		id,
		domain.MovementType(dto.MovementType),
		dto.ItemID,
		fromWarehouse,
		toWarehouse,
		dto.Quantity,
		dto.ReferenceID,
		dto.ReferenceType,
		dto.Notes,
		dto.OccurredAt,
		dto.CreatedAt,
	), nil
}

type stockReservationDTO struct {
	ID          string     `db:"id"`
	StockID     string     `db:"stock_id"`
	OrderID     string     `db:"order_id"`
	Quantity    float64    `db:"quantity"`
	Status      string     `db:"status"`
	ReservedAt  time.Time  `db:"reserved_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	FulfilledAt *time.Time `db:"fulfilled_at"`
	CancelledAt *time.Time `db:"cancelled_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func (dto *stockReservationDTO) toDomain() (*domain.StockReservation, error) {
	id, err := domain.ParseStockReservationID(dto.ID)
	if err != nil {
		return nil, err
	}

	stockID, err := domain.ParseStockID(dto.StockID)
	if err != nil {
		return nil, err
	}

	return domain.RestoreStockReservation(
		id,
		stockID,
		dto.OrderID,
		dto.Quantity,
		domain.ReservationStatus(dto.Status),
		dto.ReservedAt,
		dto.ExpiresAt,
		dto.FulfilledAt,
		dto.CancelledAt,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}
