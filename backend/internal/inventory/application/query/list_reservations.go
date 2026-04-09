package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ListReservationsHandler struct {
	reservations domain.StockReservationRepository
}

func NewListReservationsHandler(reservations domain.StockReservationRepository) *ListReservationsHandler {
	return &ListReservationsHandler{
		reservations: reservations,
	}
}

type ListReservationsQuery struct {
	OrderID string
}

type ListReservationsResult struct {
	Reservations []StockReservationDTO `json:"reservations"`
}

func (h *ListReservationsHandler) Handle(ctx context.Context, query ListReservationsQuery) (*ListReservationsResult, error) {
	reservations, err := h.reservations.FindByOrder(ctx, query.OrderID)
	if err != nil {
		return nil, fmt.Errorf("find reservations: %w", err)
	}

	dtos := make([]StockReservationDTO, 0, len(reservations))
	for _, reservation := range reservations {
		dtos = append(dtos, *toStockReservationDTO(reservation))
	}

	return &ListReservationsResult{
		Reservations: dtos,
	}, nil
}
