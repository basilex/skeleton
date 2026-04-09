package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type GetStockMovementHandler struct {
	movements domain.StockMovementRepository
}

func NewGetStockMovementHandler(movements domain.StockMovementRepository) *GetStockMovementHandler {
	return &GetStockMovementHandler{
		movements: movements,
	}
}

type GetStockMovementQuery struct {
	MovementID string
}

type StockMovementDTO struct {
	ID            string  `json:"id"`
	MovementType  string  `json:"movement_type"`
	ItemID        string  `json:"item_id"`
	FromWarehouse string  `json:"from_warehouse"`
	ToWarehouse   string  `json:"to_warehouse"`
	Quantity      float64 `json:"quantity"`
	ReferenceID   string  `json:"reference_id"`
	ReferenceType string  `json:"reference_type"`
	Notes         string  `json:"notes"`
	OccurredAt    string  `json:"occurred_at"`
	CreatedAt     string  `json:"created_at"`
}

func (h *GetStockMovementHandler) Handle(ctx context.Context, query GetStockMovementQuery) (*StockMovementDTO, error) {
	movementID, err := domain.ParseStockMovementID(query.MovementID)
	if err != nil {
		return nil, fmt.Errorf("parse movement ID: %w", err)
	}

	movement, err := h.movements.FindByID(ctx, movementID)
	if err != nil {
		return nil, fmt.Errorf("find movement: %w", err)
	}

	return toStockMovementDTO(movement), nil
}

func toStockMovementDTO(movement *domain.StockMovement) *StockMovementDTO {
	return &StockMovementDTO{
		ID:            movement.GetID().String(),
		MovementType:  movement.GetMovementType().String(),
		ItemID:        movement.GetItemID(),
		FromWarehouse: movement.GetFromWarehouse().String(),
		ToWarehouse:   movement.GetToWarehouse().String(),
		Quantity:      movement.GetQuantity(),
		ReferenceID:   movement.GetReferenceID(),
		ReferenceType: movement.GetReferenceType(),
		Notes:         movement.GetNotes(),
		OccurredAt:    movement.GetOccurredAt().Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:     movement.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
