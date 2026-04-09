package eventhandler

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
	orderingDomain "github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type OrderEventHandler struct {
	stockRepo       domain.StockRepository
	reservationRepo domain.StockReservationRepository
}

func NewOrderEventHandler(
	stockRepo domain.StockRepository,
	reservationRepo domain.StockReservationRepository,
) *OrderEventHandler {
	return &OrderEventHandler{
		stockRepo:       stockRepo,
		reservationRepo: reservationRepo,
	}
}

func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event orderingDomain.OrderConfirmed) error {
	for _, line := range event.Lines {
		_ = line // TODO: Implement stock reservation when warehouse integration is ready
	}

	return nil
}

func (h *OrderEventHandler) HandleOrderCancelled(ctx context.Context, event orderingDomain.OrderCancelled) error {
	reservations, err := h.reservationRepo.FindByOrder(ctx, event.OrderID.String())
	if err != nil {
		return fmt.Errorf("find reservations for order %s: %w", event.OrderID, err)
	}

	for _, reservation := range reservations {
		stock, err := h.stockRepo.FindByID(ctx, reservation.GetStockID())
		if err != nil {
			continue
		}

		if err := reservation.Cancel(); err != nil {
			continue
		}

		stock.ReleaseReservation(reservation.GetQuantity())

		h.reservationRepo.Save(ctx, reservation)
		h.stockRepo.Save(ctx, stock)
	}

	return nil
}

func (h *OrderEventHandler) HandleOrderCompleted(ctx context.Context, event orderingDomain.OrderCompleted) error {
	reservations, err := h.reservationRepo.FindByOrder(ctx, event.OrderID.String())
	if err != nil {
		return fmt.Errorf("find reservations for order %s: %w", event.OrderID, err)
	}

	for _, reservation := range reservations {
		stock, err := h.stockRepo.FindByID(ctx, reservation.GetStockID())
		if err != nil {
			return fmt.Errorf("find stock %s: %w", reservation.GetStockID(), err)
		}

		if err := reservation.Fulfill(); err != nil {
			return fmt.Errorf("fulfill reservation: %w", err)
		}

		stock.FulfillReservation(reservation.GetQuantity())

		h.reservationRepo.Save(ctx, reservation)
		h.stockRepo.Save(ctx, stock)
	}

	return nil
}

func (h *OrderEventHandler) Register(bus eventbus.Bus) {
	bus.Subscribe("ordering.order_confirmed", h.handleOrderConfirmed)
	bus.Subscribe("ordering.order_cancelled", h.handleOrderCancelled)
	bus.Subscribe("ordering.order_completed", h.handleOrderCompleted)
}

func (h *OrderEventHandler) handleOrderConfirmed(ctx context.Context, event eventbus.Event) error {
	e, ok := event.(orderingDomain.OrderConfirmed)
	if !ok {
		return fmt.Errorf("invalid event type: expected OrderConfirmed")
	}
	return h.HandleOrderConfirmed(ctx, e)
}

func (h *OrderEventHandler) handleOrderCancelled(ctx context.Context, event eventbus.Event) error {
	e, ok := event.(orderingDomain.OrderCancelled)
	if !ok {
		return fmt.Errorf("invalid event type: expected OrderCancelled")
	}
	return h.HandleOrderCancelled(ctx, e)
}

func (h *OrderEventHandler) handleOrderCompleted(ctx context.Context, event eventbus.Event) error {
	e, ok := event.(orderingDomain.OrderCompleted)
	if !ok {
		return fmt.Errorf("invalid event type: expected OrderCompleted")
	}
	return h.HandleOrderCompleted(ctx, e)
}
