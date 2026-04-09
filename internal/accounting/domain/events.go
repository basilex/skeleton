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

// JournalEntryCreated is published when a new journal entry is created.
type JournalEntryCreated struct {
	JournalEntryID JournalEntryID
	Description    string
	CreatedBy      string
	CreatedAt      time.Time
}

func (e JournalEntryCreated) EventName() string {
	return "accounting.journal_entry_created"
}

func (e JournalEntryCreated) OccurredAt() time.Time {
	return e.CreatedAt
}

// JournalEntryPosted is published when a journal entry is posted.
type JournalEntryPosted struct {
	JournalEntryID JournalEntryID
	PeriodID       AccountingPeriodID
	PostedAt       time.Time
	PostedBy       string
}

func (e JournalEntryPosted) EventName() string {
	return "accounting.journal_entry_posted"
}

func (e JournalEntryPosted) OccurredAt() time.Time {
	return e.PostedAt
}

// JournalEntryVoided is published when a journal entry is voided.
type JournalEntryVoided struct {
	JournalEntryID JournalEntryID
	Reason         string
	VoidedAt       time.Time
}

func (e JournalEntryVoided) EventName() string {
	return "accounting.journal_entry_voided"
}

func (e JournalEntryVoided) OccurredAt() time.Time {
	return e.VoidedAt
}

// AccountPeriodClosed is published when an accounting period is closed.
type AccountPeriodClosed struct {
	PeriodID AccountingPeriodID
	ClosedAt time.Time
	ClosedBy string
}

func (e AccountPeriodClosed) EventName() string {
	return "accounting.account_period_closed"
}

func (e AccountPeriodClosed) OccurredAt() time.Time {
	return e.ClosedAt
}

// ReconciliationCompleted is published when a reconciliation is completed.
type ReconciliationCompleted struct {
	ReconciliationID ReconciliationID
	AccountID        AccountID
	CompletedAt      time.Time
}

func (e ReconciliationCompleted) EventName() string {
	return "accounting.reconciliation_completed"
}

func (e ReconciliationCompleted) OccurredAt() time.Time {
	return e.CompletedAt
}
