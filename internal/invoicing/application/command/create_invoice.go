package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateInvoiceHandler struct {
	invoices domain.InvoiceRepository
	bus      eventbus.Bus
}

func NewCreateInvoiceHandler(invoices domain.InvoiceRepository, bus eventbus.Bus) *CreateInvoiceHandler {
	return &CreateInvoiceHandler{
		invoices: invoices,
		bus:      bus,
	}
}

type CreateInvoiceCommand struct {
	InvoiceNumber string
	OrderID       *string
	ContractID    *string
	CustomerID    string
	SupplierID    *string
	Currency      string
	DueDate       time.Time
	Notes         *string
}

type CreateInvoiceResult struct {
	InvoiceID string
}

func (h *CreateInvoiceHandler) Handle(ctx context.Context, cmd CreateInvoiceCommand) (*CreateInvoiceResult, error) {
	invoice, err := domain.NewInvoice(
		cmd.InvoiceNumber,
		cmd.CustomerID,
		cmd.Currency,
		cmd.DueDate,
	)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	if cmd.OrderID != nil {
		invoice.LinkOrder(*cmd.OrderID)
	}

	if cmd.ContractID != nil {
		invoice.LinkContract(*cmd.ContractID)
	}

	if cmd.SupplierID != nil {
		invoice.SetSupplier(*cmd.SupplierID)
	}

	if cmd.Notes != nil {
		invoice.SetNotes(*cmd.Notes)
	}

	if err := h.invoices.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("save invoice: %w", err)
	}

	// Publish domain events
	events := invoice.PullEvents()
	for _, event := range events {
		if err := h.bus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation
		}
	}

	return &CreateInvoiceResult{
		InvoiceID: invoice.GetID().String(),
	}, nil
}
