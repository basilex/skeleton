package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateAccountHandler struct {
	accounts domain.AccountRepository
	bus      eventbus.Bus
}

func NewCreateAccountHandler(
	accounts domain.AccountRepository,
	bus eventbus.Bus,
) *CreateAccountHandler {
	return &CreateAccountHandler{
		accounts: accounts,
		bus:      bus,
	}
}

type CreateAccountCommand struct {
	Code        string
	Name        string
	AccountType string
	Currency    string
	ParentID    *string
}

type CreateAccountResult struct {
	AccountID string
}

func (h *CreateAccountHandler) Handle(ctx context.Context, cmd CreateAccountCommand) (CreateAccountResult, error) {
	accountType, err := domain.ParseAccountType(cmd.AccountType)
	if err != nil {
		return CreateAccountResult{}, fmt.Errorf("invalid account type: %w", err)
	}

	currency, err := domain.ParseCurrency(cmd.Currency)
	if err != nil {
		return CreateAccountResult{}, fmt.Errorf("invalid currency: %w", err)
	}

	var parentID *domain.AccountID
	if cmd.ParentID != nil {
		pid, err := domain.ParseAccountID(*cmd.ParentID)
		if err != nil {
			return CreateAccountResult{}, fmt.Errorf("invalid parent id: %w", err)
		}
		parentID = &pid
	}

	account, err := domain.NewAccount(cmd.Code, cmd.Name, accountType, currency, parentID)
	if err != nil {
		return CreateAccountResult{}, fmt.Errorf("create account: %w", err)
	}

	if err := h.accounts.Save(ctx, account); err != nil {
		return CreateAccountResult{}, fmt.Errorf("save account: %w", err)
	}

	events := account.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateAccountResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateAccountResult{
		AccountID: account.GetID().String(),
	}, nil
}
