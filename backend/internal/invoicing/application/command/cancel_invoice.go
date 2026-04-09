package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
)

type CancelInvoiceHandler struct {
	invoices domain.InvoiceRepository
}

func NewCancelInvoiceHandler(invoices domain.InvoiceRepository) *CancelInvoiceHandler {
	return &CancelInvoiceHandler{
		invoices: invoices,
	}
}

type CancelInvoiceCommand struct {
	InvoiceID string
	Reason    string
}

func (h *CancelInvoiceHandler) Handle(ctx context.Context, cmd CancelInvoiceCommand) error {
	invoiceID, err := domain.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return fmt.Errorf("parse invoice ID: %w", err)
	}

	invoice, err := h.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("find invoice: %w", err)
	}

	if err := invoice.Cancel(cmd.Reason); err != nil {
		return fmt.Errorf("cancel invoice: %w", err)
	}

	if err := h.invoices.Save(ctx, invoice); err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	return nil
}
