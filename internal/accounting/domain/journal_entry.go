package domain

import (
	"errors"
	"fmt"
	"time"
)

// JournalEntry represents a double-entry bookkeeping journal entry.
// It enforces the fundamental accounting principle: Debits must equal Credits.
type JournalEntry struct {
	id          JournalEntryID
	periodID    AccountingPeriodID
	description string
	lines       []JournalLine
	status      JournalEntryStatus
	postedAt    *time.Time
	createdAt   time.Time
	createdBy   string
	events      []DomainEvent
}

// JournalLine represents a single line in a journal entry.
// Each line debits or credits a specific account.
type JournalLine struct {
	accountID   AccountID
	debit       Money
	credit      Money
	description string
}

// JournalEntryStatus represents the lifecycle state of a journal entry.
type JournalEntryStatus string

const (
	JournalEntryStatusDraft  JournalEntryStatus = "draft"
	JournalEntryStatusPosted JournalEntryStatus = "posted"
	JournalEntryStatusVoided JournalEntryStatus = "voided"
)

// NewJournalEntry creates a new journal entry in draft status.
func NewJournalEntry(description string, createdBy string) (*JournalEntry, error) {
	if description == "" {
		return nil, errors.New("description is required")
	}
	if createdBy == "" {
		return nil, errors.New("created by is required")
	}

	now := time.Now().UTC()
	id := NewJournalEntryID()

	je := &JournalEntry{
		id:          id,
		description: description,
		lines:       make([]JournalLine, 0),
		status:      JournalEntryStatusDraft,
		createdAt:   now,
		createdBy:   createdBy,
		events:      make([]DomainEvent, 0),
	}

	je.events = append(je.events, JournalEntryCreated{
		JournalEntryID: id,
		Description:    description,
		CreatedBy:      createdBy,
		CreatedAt:      now,
	})

	return je, nil
}

// AddLine adds a line to the journal entry.
// A line can be either debit OR credit, but not both.
func (je *JournalEntry) AddLine(accountID AccountID, debit, credit Money, description string) error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("cannot add lines to non-draft journal entry")
	}

	// Validate that exactly one side is non-zero
	if debit.IsZero() && credit.IsZero() {
		return errors.New("line must have either debit or credit")
	}
	if !debit.IsZero() && !credit.IsZero() {
		return errors.New("line cannot have both debit and credit")
	}

	line := JournalLine{
		accountID:   accountID,
		debit:       debit,
		credit:      credit,
		description: description,
	}

	je.lines = append(je.lines, line)
	return nil
}

// Post validates and posts the journal entry.
// Business rules:
// - Must have at least 2 lines
// - Total debits must equal total credits
// - Must be in draft status
func (je *JournalEntry) Post() error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("only draft entries can be posted")
	}

	if len(je.lines) < 2 {
		return errors.New("journal entry must have at least 2 lines")
	}

	if !je.IsBalanced() {
		return errors.New("journal entry is not balanced: debits must equal credits")
	}

	now := time.Now().UTC()
	je.status = JournalEntryStatusPosted
	je.postedAt = &now

	je.events = append(je.events, JournalEntryPosted{
		JournalEntryID: je.id,
		PeriodID:       je.periodID,
		PostedAt:       now,
		PostedBy:       je.createdBy,
	})

	return nil
}

// Void voids a posted journal entry.
// Business rule: Only posted entries can be voided
func (je *JournalEntry) Void(reason string) error {
	if je.status != JournalEntryStatusPosted {
		return errors.New("only posted entries can be voided")
	}

	je.status = JournalEntryStatusVoided

	je.events = append(je.events, JournalEntryVoided{
		JournalEntryID: je.id,
		Reason:         reason,
		VoidedAt:       time.Now().UTC(),
	})

	return nil
}

// IsBalanced checks if total debits equal total credits.
// This is the fundamental rule of double-entry bookkeeping.
func (je *JournalEntry) IsBalanced() bool {
	if len(je.lines) == 0 {
		return true
	}

	totalDebits := Money{Amount: 0, Currency: je.lines[0].debit.Currency}
	totalCredits := Money{Amount: 0, Currency: je.lines[0].credit.Currency}

	for _, line := range je.lines {
		totalDebits = totalDebits.Add(line.debit)
		totalCredits = totalCredits.Add(line.credit)
	}

	return totalDebits.Equals(totalCredits)
}

// GetLineCount returns the number of lines in the journal entry.
func (je *JournalEntry) GetLineCount() int {
	return len(je.lines)
}

// GetTotalDebits returns the sum of all debit amounts.
func (je *JournalEntry) GetTotalDebits() Money {
	if len(je.lines) == 0 {
		return Money{Amount: 0, Currency: CurrencyUSD}
	}
	total := Money{Amount: 0, Currency: je.lines[0].debit.Currency}
	for _, line := range je.lines {
		total = total.Add(line.debit)
	}
	return total
}

// GetTotalCredits returns the sum of all credit amounts.
func (je *JournalEntry) GetTotalCredits() Money {
	if len(je.lines) == 0 {
		return Money{Amount: 0, Currency: CurrencyUSD}
	}
	total := Money{Amount: 0, Currency: je.lines[0].credit.Currency}
	for _, line := range je.lines {
		total = total.Add(line.credit)
	}
	return total
}

// SetPeriod assigns the journal entry to an accounting period.
func (je *JournalEntry) SetPeriod(periodID AccountingPeriodID) error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("cannot set period on non-draft entry")
	}
	je.periodID = periodID
	return nil
}

// GetID returns the journal entry ID.
func (je *JournalEntry) GetID() JournalEntryID {
	return je.id
}

// GetStatus returns the current status.
func (je *JournalEntry) GetStatus() JournalEntryStatus {
	return je.status
}

// GetLines returns all journal entry lines.
func (je *JournalEntry) GetLines() []JournalLine {
	return je.lines
}

// PullEvents returns all domain events and clears the event buffer.
func (je *JournalEntry) PullEvents() []DomainEvent {
	events := je.events
	je.events = make([]DomainEvent, 0)
	return events
}

// JournalLine methods

// GetAccountID returns the account ID for this line.
func (jl *JournalLine) GetAccountID() AccountID {
	return jl.accountID
}

// GetDebit returns the debit amount.
func (jl *JournalLine) GetDebit() Money {
	return jl.debit
}

// GetCredit returns the credit amount.
func (jl *JournalLine) GetCredit() Money {
	return jl.credit
}

// GetDescription returns the line description.
func (jl *JournalLine) GetDescription() string {
	return jl.description
}

// IsDebit returns true if this is a debit line.
func (jl *JournalLine) IsDebit() bool {
	return !jl.debit.IsZero()
}

// IsCredit returns true if this is a credit line.
func (jl *JournalLine) IsCredit() bool {
	return !jl.credit.IsZero()
}

// String returns a string representation of the journal entry.
func (je *JournalEntry) String() string {
	return fmt.Sprintf("JournalEntry{id=%s, status=%s, lines=%d, balanced=%v}",
		je.id, je.status, len(je.lines), je.IsBalanced())
}
