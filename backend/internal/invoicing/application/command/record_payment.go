package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/money"
)

type RecordPaymentHandler struct {
	invoices domain.InvoiceRepository
	bus      eventbus.Bus
}

func NewRecordPaymentHandler(invoices domain.InvoiceRepository, bus eventbus.Bus) *RecordPaymentHandler {
	return &RecordPaymentHandler{
		invoices: invoices,
		bus:      bus,
	}
}

type RecordPaymentCommand struct {
	InvoiceID string
	Amount    float64
	Method    string
	Reference string
	Notes     string
}

type RecordPaymentResult struct {
	PaymentID string
}

func (h *RecordPaymentHandler) Handle(ctx context.Context, cmd RecordPaymentCommand) (*RecordPaymentResult, error) {
	invoiceID, err := domain.ParseInvoiceID(cmd.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("parse invoice ID: %w", err)
	}

	invoice, err := h.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("find invoice: %w", err)
	}

	amount, err := money.NewFromFloat(cmd.Amount, invoice.GetCurrency())
	if err != nil {
		return nil, fmt.Errorf("create payment amount: %w", err)
	}

	payment, err := invoice.RecordPayment(
		amount,
		domain.PaymentMethod(cmd.Method),
		cmd.Reference,
	)
	if err != nil {
		return nil, fmt.Errorf("record payment: %w", err)
	}

	if cmd.Notes != "" {
		payment.AddNotes(cmd.Notes)
	}

	if err := h.invoices.Save(ctx, invoice); err != nil {
		return nil, fmt.Errorf("save invoice: %w", err)
	}

	events := invoice.PullEvents()
	for _, event := range events {
		if err := h.bus.Publish(ctx, event); err != nil {
		}
	}

	return &RecordPaymentResult{
		PaymentID: payment.GetID().String(),
	}, nil
}
