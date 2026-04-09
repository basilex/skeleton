package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type RecordTransactionHandler struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	bus          eventbus.Bus
}

func NewRecordTransactionHandler(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	bus eventbus.Bus,
) *RecordTransactionHandler {
	return &RecordTransactionHandler{
		accounts:     accounts,
		transactions: transactions,
		bus:          bus,
	}
}

type RecordTransactionCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        float64
	Currency      string
	Reference     string
	Description   string
	PostedBy      string
}

type RecordTransactionResult struct {
	TransactionID string
}

func (h *RecordTransactionHandler) Handle(ctx context.Context, cmd RecordTransactionCommand) (RecordTransactionResult, error) {
	fromAccountID, err := domain.ParseAccountID(cmd.FromAccountID)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("invalid from account id: %w", err)
	}

	toAccountID, err := domain.ParseAccountID(cmd.ToAccountID)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("invalid to account id: %w", err)
	}

	currency, err := domain.ParseCurrency(cmd.Currency)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("invalid currency: %w", err)
	}

	money, err := domain.NewMoney(cmd.Amount, currency)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("invalid amount: %w", err)
	}

	fromAccount, err := h.accounts.FindByID(ctx, fromAccountID)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("find from account: %w", err)
	}

	toAccount, err := h.accounts.FindByID(ctx, toAccountID)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("find to account: %w", err)
	}

	if err := fromAccount.Credit(money); err != nil {
		return RecordTransactionResult{}, fmt.Errorf("credit from account: %w", err)
	}

	if err := toAccount.Debit(money); err != nil {
		return RecordTransactionResult{}, fmt.Errorf("debit to account: %w", err)
	}

	transaction, err := domain.NewTransaction(fromAccountID, toAccountID, money, currency, cmd.Reference, cmd.Description, cmd.PostedBy)
	if err != nil {
		return RecordTransactionResult{}, fmt.Errorf("create transaction: %w", err)
	}

	if err := h.accounts.Save(ctx, fromAccount); err != nil {
		return RecordTransactionResult{}, fmt.Errorf("save from account: %w", err)
	}

	if err := h.accounts.Save(ctx, toAccount); err != nil {
		return RecordTransactionResult{}, fmt.Errorf("save to account: %w", err)
	}

	if err := h.transactions.Save(ctx, transaction); err != nil {
		return RecordTransactionResult{}, fmt.Errorf("save transaction: %w", err)
	}

	events := transaction.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return RecordTransactionResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return RecordTransactionResult{
		TransactionID: transaction.GetID().String(),
	}, nil
}
