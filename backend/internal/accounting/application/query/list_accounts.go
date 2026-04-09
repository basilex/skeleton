package query

import (
	"context"

	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type ListAccountsHandler struct {
	accounts domain.AccountRepository
}

func NewListAccountsHandler(accounts domain.AccountRepository) *ListAccountsHandler {
	return &ListAccountsHandler{
		accounts: accounts,
	}
}

type ListAccountsQuery struct {
	AccountType *string
	IsActive    *bool
	Search      string
	Cursor      string
	Limit       int
}

func (h *ListAccountsHandler) Handle(ctx context.Context, q ListAccountsQuery) (pagination.PageResult[AccountDTO], error) {
	filter := domain.AccountFilter{
		Search: q.Search,
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}

	if q.AccountType != nil {
		accountType, err := domain.ParseAccountType(*q.AccountType)
		if err == nil {
			filter.AccountType = &accountType
		}
	}

	if q.IsActive != nil {
		filter.IsActive = q.IsActive
	}

	result, err := h.accounts.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[AccountDTO]{}, err
	}

	dtos := make([]AccountDTO, 0, len(result.Items))
	for _, account := range result.Items {
		var parentID *string
		if account.GetParentID() != nil {
			pid := account.GetParentID().String()
			parentID = &pid
		}

		dtos = append(dtos, AccountDTO{
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
		})
	}

	return pagination.PageResult[AccountDTO]{
		Items:      dtos,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		Limit:      result.Limit,
	}, nil
}
