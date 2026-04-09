package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
)

type AddInvoiceLineHandler struct {
	invoices domain.InvoiceRepository
}

func NewAddInvoiceLineHandler(invoices domain.InvoiceRepository) *AddInvoiceLineHandler {
	return &AddInvoiceLineHandler{
		invoices: invoices,
	}
}

type AddInvoiceLineCommand struct {
	InvoiceID   string
	Description string
	Quantity    float64
	UnitPrice   float64
	Unit        string
	Discount    float64
}

func (h *AddInvoiceLineHandler) Handle(ctx context.Context, cmd AddInvoiceLineCommand) error {
	invoiceID, err := domain.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("parse invoice ID: %w", err)
	}

	invoice, err := h.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("find invoice: %w", err)
	}

	if err := invoice.AddLine(cmd.Description, cmd.Quantity, cmd.UnitPrice, cmd.Unit, cmd.Discount); err != nil {
		return fmt.Errorf("add invoice line: %w", err)
	}

	if err := h.invoices.Save(ctx, invoice); err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	return nil
}
