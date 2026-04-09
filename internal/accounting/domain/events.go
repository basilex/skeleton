package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type AccountCreated struct {
	AccountID   AccountID
	AccountCode string
	AccountName string
	AccountType AccountType
	OcurredAt   time.Time
}

func (e AccountCreated) EventName() string {
	return "accounting.account_created"
}

func (e AccountCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

type TransactionRecorded struct {
	TransactionID string
	FromAccount   AccountID
	ToAccount     AccountID
	Amount        Money
	OcurredAt     time.Time
}

func (e TransactionRecorded) EventName() string {
	return "accounting.transaction_recorded"
}

func (e TransactionRecorded) OccurredAt() time.Time {
	return e.OcurredAt
}

type InvoiceIssued struct {
	InvoiceID string
	PartyID   string
	Amount    Money
	Direction string
	OcurredAt time.Time
}

func (e InvoiceIssued) EventName() string {
	return "accounting.invoice_issued"
}

func (e InvoiceIssued) OccurredAt() time.Time {
	return e.OcurredAt
}

type PaymentRecorded struct {
	PaymentID string
	PayableID string
	Amount    Money
	OcurredAt time.Time
}

func (e PaymentRecorded) EventName() string {
	return "accounting.payment_recorded"
}

func (e PaymentRecorded) OccurredAt() time.Time {
	return e.OcurredAt
}
