package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type SendInvoiceHandler struct {
	invoices domain.InvoiceRepository
	bus      eventbus.Bus
}

func NewSendInvoiceHandler(invoices domain.InvoiceRepository, bus eventbus.Bus) *SendInvoiceHandler {
	return &SendInvoiceHandler{
		invoices: invoices,
		bus:      bus,
	}
}

type SendInvoiceCommand struct {
	InvoiceID string
}

func (h *SendInvoiceHandler) Handle(ctx context.Context, cmd SendInvoiceCommand) error {
	invoiceID, err := domain.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("parse invoice ID: %w", err)
	}

	invoice, err := h.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("find invoice: %w", err)
	}

	if err := invoice.Send(); err != nil {
		return fmt.Errorf("send invoice: %w", err)
	}

	if err := h.invoices.Save(ctx, invoice); err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	// Publish domain events
	events := invoice.PullEvents()
	for _, event := range events {
		if err := h.bus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
		}
	}

	return nil
}
