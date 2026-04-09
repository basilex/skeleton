package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

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

type JournalLine struct {
	accountID   AccountID
	debit       money.Money
	credit      money.Money
	description string
}

type JournalEntryStatus string

const (
	JournalEntryStatusDraft  JournalEntryStatus = "draft"
	JournalEntryStatusPosted JournalEntryStatus = "posted"
	JournalEntryStatusVoided JournalEntryStatus = "voided"
)

func NewJournalEntry(description, createdBy string) *JournalEntry {
	now := time.Now().UTC()
	je := &JournalEntry{
		id:          NewJournalEntryID(),
		description: description,
		lines:       make([]JournalLine, 0),
		status:      JournalEntryStatusDraft,
		createdAt:   now,
		createdBy:   createdBy,
		events:      make([]DomainEvent, 0),
	}

	je.events = append(je.events, JournalEntryCreated{
		JournalEntryID: je.id,
		Description:    description,
		CreatedBy:      createdBy,
		CreatedAt:      now,
	})

	return je
}

func (je *JournalEntry) AddLine(accountID AccountID, debit, credit money.Money, description string) error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("cannot add lines to non-draft journal entry")
	}

	if !debit.IsZero() && !credit.IsZero() {
		return errors.New("a line cannot have both debit and credit amounts")
	}

	if debit.IsZero() && credit.IsZero() {
		return errors.New("a line must have either debit or credit amount")
	}

	je.lines = append(je.lines, JournalLine{
		accountID:   accountID,
		debit:       debit,
		credit:      credit,
		description: description,
	})

	return nil
}

func (je *JournalEntry) Post() error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("only draft entries can be posted")
	}

	if !je.IsBalanced() {
		return errors.New("journal entry is not balanced")
	}

	if len(je.lines) == 0 {
		return errors.New("journal entry must have at least one line")
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

func (je *JournalEntry) IsBalanced() bool {
	if len(je.lines) == 0 {
		return true
	}

	totalDebits := money.Zero(je.lines[0].debit.GetCurrency())
	totalCredits := money.Zero(je.lines[0].credit.GetCurrency())

	for _, line := range je.lines {
		var err error
		totalDebits, err = totalDebits.Add(line.debit)
		if err != nil {
			return false
		}
		totalCredits, err = totalCredits.Add(line.credit)
		if err != nil {
			return false
		}
	}

	return totalDebits.Equals(totalCredits)
}

func (je *JournalEntry) GetLineCount() int {
	return len(je.lines)
}

func (je *JournalEntry) GetTotalDebits() money.Money {
	if len(je.lines) == 0 {
		return money.Zero(string(CurrencyUSD))
	}
	total := money.Zero(je.lines[0].debit.GetCurrency())
	for _, line := range je.lines {
		var err error
		total, err = total.Add(line.debit)
		if err != nil {
			return money.Zero(je.lines[0].debit.GetCurrency())
		}
	}
	return total
}

func (je *JournalEntry) GetTotalCredits() money.Money {
	if len(je.lines) == 0 {
		return money.Zero(string(CurrencyUSD))
	}
	total := money.Zero(je.lines[0].credit.GetCurrency())
	for _, line := range je.lines {
		var err error
		total, err = total.Add(line.credit)
		if err != nil {
			return money.Zero(je.lines[0].credit.GetCurrency())
		}
	}
	return total
}

func (je *JournalEntry) SetPeriod(periodID AccountingPeriodID) error {
	if je.status != JournalEntryStatusDraft {
		return errors.New("cannot set period on non-draft entry")
	}
	je.periodID = periodID
	return nil
}

func (je *JournalEntry) GetID() JournalEntryID {
	return je.id
}

func (je *JournalEntry) GetStatus() JournalEntryStatus {
	return je.status
}

func (je *JournalEntry) GetLines() []JournalLine {
	return je.lines
}

func (je *JournalEntry) PullEvents() []DomainEvent {
	events := je.events
	je.events = make([]DomainEvent, 0)
	return events
}

func (jl *JournalLine) GetAccountID() AccountID {
	return jl.accountID
}

func (jl *JournalLine) GetDebit() money.Money {
	return jl.debit
}

func (jl *JournalLine) GetCredit() money.Money {
	return jl.credit
}

func (jl *JournalLine) GetDescription() string {
	return jl.description
}

func (jl *JournalLine) IsDebit() bool {
	return !jl.debit.IsZero()
}

func (jl *JournalLine) IsCredit() bool {
	return !jl.credit.IsZero()
}

func (je *JournalEntry) String() string {
	return fmt.Sprintf("JournalEntry{id=%s, status=%s, lines=%d, balanced=%v}",
		je.id, je.status, len(je.lines), je.IsBalanced())
}
