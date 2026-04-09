package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type TransactionID string

func NewTransactionID() TransactionID {
	return TransactionID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id TransactionID) String() string {
	return string(id)
}

type Transaction struct {
	id          TransactionID
	fromAccount AccountID
	toAccount   AccountID
	amount      money.Money
	currency    Currency
	reference   string
	description string
	occurredAt  time.Time
	postedAt    time.Time
	postedBy    string
	events      []DomainEvent
}

func NewTransaction(
	fromAccount, toAccount AccountID,
	amount money.Money,
	currency Currency,
	reference, description, postedBy string,
) (*Transaction, error) {
	if fromAccount == toAccount {
		return nil, fmt.Errorf("cannot create transaction to same account")
	}

	now := time.Now().UTC()
	t := &Transaction{
		id:          NewTransactionID(),
		fromAccount: fromAccount,
		toAccount:   toAccount,
		amount:      amount,
		currency:    currency,
		reference:   reference,
		description: description,
		occurredAt:  now,
		postedAt:    now,
		postedBy:    postedBy,
		events:      make([]DomainEvent, 0),
	}

	t.events = append(t.events, TransactionRecorded{
		TransactionID: t.id.String(),
		FromAccount:   t.fromAccount,
		ToAccount:     t.toAccount,
		Amount:        t.amount,
		OcurredAt:     now,
	})

	return t, nil
}

func (t *Transaction) GetID() TransactionID      { return t.id }
func (t *Transaction) GetFromAccount() AccountID { return t.fromAccount }
func (t *Transaction) GetToAccount() AccountID   { return t.toAccount }
func (t *Transaction) GetAmount() money.Money    { return t.amount }
func (t *Transaction) GetCurrency() Currency     { return t.currency }
func (t *Transaction) GetReference() string      { return t.reference }
func (t *Transaction) GetDescription() string    { return t.description }
func (t *Transaction) GetOccurredAt() time.Time  { return t.occurredAt }
func (t *Transaction) GetPostedAt() time.Time    { return t.postedAt }
func (t *Transaction) GetPostedBy() string       { return t.postedBy }

func (t *Transaction) PullEvents() []DomainEvent {
	events := t.events
	t.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteTransaction(
	id TransactionID,
	fromAccount, toAccount AccountID,
	amount money.Money,
	currency Currency,
	reference, description string,
	occurredAt, postedAt time.Time,
	postedBy string,
) *Transaction {
	return &Transaction{
		id:          id,
		fromAccount: fromAccount,
		toAccount:   toAccount,
		amount:      amount,
		currency:    currency,
		reference:   reference,
		description: description,
		occurredAt:  occurredAt,
		postedAt:    postedAt,
		postedBy:    postedBy,
		events:      make([]DomainEvent, 0),
	}
}

type TransactionRepository interface {
	Save(ctx context.Context, transaction *Transaction) error
	FindByID(ctx context.Context, id TransactionID) (*Transaction, error)
	FindByAccount(ctx context.Context, accountID AccountID, limit int) ([]*Transaction, error)
	FindByReference(ctx context.Context, reference string) ([]*Transaction, error)
}
