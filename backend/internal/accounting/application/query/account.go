package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/accounting/domain"
)

type GetAccountHandler struct {
	accounts domain.AccountRepository
}

func NewGetAccountHandler(accounts domain.AccountRepository) *GetAccountHandler {
	return &GetAccountHandler{
		accounts: accounts,
	}
}

type GetAccountQuery struct {
	AccountID string
}

type AccountDTO struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	AccountType string  `json:"account_type"`
	Currency    string  `json:"currency"`
	Balance     float64 `json:"balance"`
	ParentID    *string `json:"parent_id,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h *GetAccountHandler) Handle(ctx context.Context, q GetAccountQuery) (AccountDTO, error) {
	accountID, err := domain.ParseAccountID(q.AccountID)
	if err != nil {
		return AccountDTO{}, fmt.Errorf("parse account id: %w", err)
	}

	account, err := h.accounts.FindByID(ctx, accountID)
	if err != nil {
		return AccountDTO{}, fmt.Errorf("find account: %w", err)
	}

	var parentID *string
	if account.GetParentID() != nil {
		pid := account.GetParentID().String()
		parentID = &pid
	}

	return AccountDTO{
		ID:          account.GetID().String(),
		Code:        account.GetCode(),
		Name:        account.GetName(),
		AccountType: account.GetType().String(),
		Currency:    account.GetCurrency().String(),
		Balance:     account.GetBalance().ToFloat64(),
		ParentID:    parentID,
		IsActive:    account.IsActive(),
		CreatedAt:   account.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   account.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
