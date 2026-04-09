package eventhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	orderingDomain "github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type OrderEventHandler struct {
	invoiceRepo domain.InvoiceRepository
}

func NewOrderEventHandler(invoiceRepo domain.InvoiceRepository) *OrderEventHandler {
	return &OrderEventHandler{
		invoiceRepo: invoiceRepo,
	}
}

func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event orderingDomain.OrderConfirmed) error {
	invoiceNumber := fmt.Sprintf("INV-%d", time.Now().Unix())

	invoice, err := domain.NewInvoice(
		invoiceNumber,
		event.CustomerID,
		event.Currency,
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}

	invoice.LinkOrder(event.OrderID.String())

	for _, line := range event.Lines {
		if err := invoice.AddLine(
			line.ItemName,
			line.Quantity,
			line.UnitPrice,
			line.Unit,
			line.Discount,
		); err != nil {
			return fmt.Errorf("add invoice line: %w", err)
		}
	}

	if err := h.invoiceRepo.Save(ctx, invoice); err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	return nil
}

func (h *OrderEventHandler) Register(bus eventbus.Bus) {
	bus.Subscribe("ordering.order_confirmed", h.handleOrderConfirmed)
}

func (h *OrderEventHandler) handleOrderConfirmed(ctx context.Context, event eventbus.Event) error {
	e, ok := event.(orderingDomain.OrderConfirmed)
	if !ok {
		return fmt.Errorf("invalid event type: expected OrderConfirmed")
	}
	return h.HandleOrderConfirmed(ctx, e)
}
