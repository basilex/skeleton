package eventhandler

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/accounting/domain"
	invoicingDomain "github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type InvoiceEventHandler struct {
	accountRepo     domain.AccountRepository
	transactionRepo domain.TransactionRepository
}

func NewInvoiceEventHandler(
	accountRepo domain.AccountRepository,
	transactionRepo domain.TransactionRepository,
) *InvoiceEventHandler {
	return &InvoiceEventHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (h *InvoiceEventHandler) HandleInvoiceCreated(ctx context.Context, event invoicingDomain.InvoiceCreated) error {
	receivableAccount, err := h.accountRepo.FindByCode(ctx, "1300")
	if err != nil {
		return fmt.Errorf("find accounts receivable: %w", err)
	}

	revenueAccount, err := h.accountRepo.FindByCode(ctx, "4000")
	if err != nil {
		return fmt.Errorf("find revenue account: %w", err)
	}

	reference := fmt.Sprintf("INV-%s", event.InvoiceNumber)
	description := fmt.Sprintf("Invoice %s for customer %s", event.InvoiceNumber, event.CustomerID)

	transaction, err := domain.NewTransaction(
		receivableAccount.GetID(),
		revenueAccount.GetID(),
		domain.Money{Amount: event.Total, Currency: domain.Currency(event.Currency)},
		domain.Currency(event.Currency),
		reference,
		description,
		"SYSTEM",
	)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	if err := h.transactionRepo.Save(ctx, transaction); err != nil {
		return fmt.Errorf("save transaction: %w", err)
	}

	return nil
}

func (h *InvoiceEventHandler) Register(bus eventbus.Bus) {
	bus.Subscribe("invoicing.invoice_created", h.handleInvoiceCreated)
}

func (h *InvoiceEventHandler) handleInvoiceCreated(ctx context.Context, event eventbus.Event) error {
	e, ok := event.(invoicingDomain.InvoiceCreated)
	if !ok {
		return fmt.Errorf("invalid event type: expected InvoiceCreated")
	}
	return h.HandleInvoiceCreated(ctx, e)
}
