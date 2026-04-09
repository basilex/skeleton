package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/transaction"
)

type RecordTransactionHandler struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	bus          eventbus.Bus
	txManager    transaction.Manager
}

func NewRecordTransactionHandler(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	bus eventbus.Bus,
	txManager transaction.Manager,
) *RecordTransactionHandler {
	return &RecordTransactionHandler{
		accounts:     accounts,
		transactions: transactions,
		bus:          bus,
		txManager:    txManager,
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
	var result RecordTransactionResult

	err := h.txManager.Execute(ctx, func(ctx context.Context) error {
		// Parse IDs
		fromAccountID, err := domain.ParseAccountID(cmd.FromAccountID)
		if err != nil {
			return fmt.Errorf("invalid from account id: %w", err)
		}

		toAccountID, err := domain.ParseAccountID(cmd.ToAccountID)
		if err != nil {
			return fmt.Errorf("invalid to account id: %w", err)
		}

		// Parse currency
		currency, err := domain.ParseCurrency(cmd.Currency)
		if err != nil {
			return fmt.Errorf("invalid currency: %w", err)
		}

		// Create money value
		money, err := domain.NewMoney(cmd.Amount, currency)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}

		// Load accounts
		fromAccount, err := h.accounts.FindByID(ctx, fromAccountID)
		if err != nil {
			return fmt.Errorf("find from account: %w", err)
		}

		toAccount, err := h.accounts.FindByID(ctx, toAccountID)
		if err != nil {
			return fmt.Errorf("find to account: %w", err)
		}

		// Perform double-entry bookkeeping
		if err := fromAccount.Credit(money); err != nil {
			return fmt.Errorf("credit from account: %w", err)
		}

		if err := toAccount.Debit(money); err != nil {
			return fmt.Errorf("debit to account: %w", err)
		}

		// Create transaction record
		transaction, err := domain.NewTransaction(fromAccountID, toAccountID, money, currency, cmd.Reference, cmd.Description, cmd.PostedBy)
		if err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		// Save all aggregates within transaction
		if err := h.accounts.Save(ctx, fromAccount); err != nil {
			return fmt.Errorf("save from account: %w", err)
		}

		if err := h.accounts.Save(ctx, toAccount); err != nil {
			return fmt.Errorf("save to account: %w", err)
		}

		if err := h.transactions.Save(ctx, transaction); err != nil {
			return fmt.Errorf("save transaction: %w", err)
		}

		// Publish domain events
		events := transaction.PullEvents()
		for _, e := range events {
			if err := h.bus.Publish(ctx, e); err != nil {
				return fmt.Errorf("publish event: %w", err)
			}
		}

		result = RecordTransactionResult{
			TransactionID: transaction.GetID().String(),
		}

		return nil
	})

	return result, err
}
