package persistence

import (
	"time"

	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/money"
)

type accountDTO struct {
	ID          string    `db:"id"`
	Code        string    `db:"code"`
	Name        string    `db:"name"`
	AccountType string    `db:"account_type"`
	Currency    string    `db:"currency"`
	Balance     int64     `db:"balance"`
	ParentID    *string   `db:"parent_id"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (dto *accountDTO) toDomain() (*domain.Account, error) {
	id, err := domain.ParseAccountID(dto.ID)
	if err != nil {
		return nil, err
	}

	accountType, err := domain.ParseAccountType(dto.AccountType)
	if err != nil {
		return nil, err
	}

	currency, err := domain.ParseCurrency(dto.Currency)
	if err != nil {
		return nil, err
	}

	var parentID *domain.AccountID
	if dto.ParentID != nil {
		pid, err := domain.ParseAccountID(*dto.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &pid
	}

	balance, _ := money.New(dto.Balance, string(currency))

	return domain.ReconstituteAccount(
		id,
		dto.Code,
		dto.Name,
		accountType,
		currency,
		balance,
		parentID,
		dto.IsActive,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

type transactionDTO struct {
	ID          string    `db:"id"`
	FromAccount string    `db:"from_account"`
	ToAccount   string    `db:"to_account"`
	Amount      int64     `db:"amount"`
	Currency    string    `db:"currency"`
	Reference   string    `db:"reference_type"`
	Description string    `db:"description"`
	OccurredAt  time.Time `db:"occurred_at"`
	PostedAt    time.Time `db:"posted_at"`
	PostedBy    string    `db:"posted_by"`
}

func (dto *transactionDTO) toDomain() (*domain.Transaction, error) {
	fromAccount, err := domain.ParseAccountID(dto.FromAccount)
	if err != nil {
		return nil, err
	}

	toAccount, err := domain.ParseAccountID(dto.ToAccount)
	if err != nil {
		return nil, err
	}

	currency, err := domain.ParseCurrency(dto.Currency)
	if err != nil {
		return nil, err
	}

	amount, _ := money.New(dto.Amount, string(currency))

	return domain.ReconstituteTransaction(
		domain.TransactionID(dto.ID),
		fromAccount,
		toAccount,
		amount,
		currency,
		dto.Reference,
		dto.Description,
		dto.OccurredAt,
		dto.PostedAt,
		dto.PostedBy,
	), nil
}
