package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ListStockMovementsHandler struct {
	movements domain.StockMovementRepository
}

func NewListStockMovementsHandler(movements domain.StockMovementRepository) *ListStockMovementsHandler {
	return &ListStockMovementsHandler{
		movements: movements,
	}
}

type ListStockMovementsQuery struct {
	ItemID        *string
	WarehouseID   *string
	MovementType  *string
	ReferenceType *string
	StartDate     *string
	EndDate       *string
	Cursor        string
	Limit         int
}

type ListStockMovementsResult struct {
	Movements  []StockMovementDTO `json:"movements"`
	NextCursor string             `json:"next_cursor"`
}

func (h *ListStockMovementsHandler) Handle(ctx context.Context, query ListStockMovementsQuery) (*ListStockMovementsResult, error) {
	var warehouseID *domain.WarehouseID
	if query.WarehouseID != nil {
		id, err := domain.ParseWarehouseID(*query.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("parse warehouse ID: %w", err)
		}
		warehouseID = &id
	}

	var movementType *domain.MovementType
	if query.MovementType != nil {
		mt := domain.MovementType(*query.MovementType)
		movementType = &mt
	}

	filter := domain.StockMovementFilter{
		ItemID:        query.ItemID,
		WarehouseID:   warehouseID,
		MovementType:  movementType,
		ReferenceType: query.ReferenceType,
		StartDate:     query.StartDate,
		EndDate:       query.EndDate,
		Cursor:        query.Cursor,
		Limit:         query.Limit,
	}

	result, err := h.movements.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find movements: %w", err)
	}

	movements := make([]StockMovementDTO, 0, len(result.Items))
	for _, movement := range result.Items {
		movements = append(movements, *toStockMovementDTO(movement))
	}

	return &ListStockMovementsResult{
		Movements:  movements,
		NextCursor: result.NextCursor,
	}, nil
}
