package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

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
	Amount        money.Money
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
	Amount    money.Money
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
	Amount    money.Money
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

// AccountParentChanged is published when an account's parent is changed.
type AccountParentChanged struct {
	AccountID   AccountID
	OldParentID *AccountID
	NewParentID *AccountID
	OcurredAt   time.Time
}

func (e AccountParentChanged) EventName() string {
	return "accounting.account_parent_changed"
}

func (e AccountParentChanged) OccurredAt() time.Time {
	return e.OcurredAt
}

// AccountActivated is published when an account is activated.
type AccountActivated struct {
	AccountID AccountID
	OcurredAt time.Time
}

func (e AccountActivated) EventName() string {
	return "accounting.account_activated"
}

func (e AccountActivated) OccurredAt() time.Time {
	return e.OcurredAt
}

// AccountDeactivated is published when an account is deactivated.
type AccountDeactivated struct {
	AccountID AccountID
	OcurredAt time.Time
}

func (e AccountDeactivated) EventName() string {
	return "accounting.account_deactivated"
}

func (e AccountDeactivated) OccurredAt() time.Time {
	return e.OcurredAt
}
