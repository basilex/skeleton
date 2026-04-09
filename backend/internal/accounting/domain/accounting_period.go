package domain

import (
	"errors"
	"fmt"
	"time"
)

// PeriodStatus represents the status of an accounting period.
type PeriodStatus string

const (
	PeriodStatusOpen   PeriodStatus = "open"
	PeriodStatusClosed PeriodStatus = "closed"
	PeriodStatusLocked PeriodStatus = "locked"
)

// AccountingPeriod represents a fiscal period (month, quarter, year).
type AccountingPeriod struct {
	id        AccountingPeriodID
	name      string
	startDate time.Time
	endDate   time.Time
	status    PeriodStatus
	closedAt  *time.Time
	closedBy  string
}

// NewAccountingPeriod creates a new accounting period.
func NewAccountingPeriod(name string, startDate, endDate time.Time) (*AccountingPeriod, error) {
	if name == "" {
		return nil, errors.New("period name is required")
	}
	if startDate.After(endDate) {
		return nil, errors.New("start date must be before end date")
	}

	return &AccountingPeriod{
		id:        NewAccountingPeriodID(),
		name:      name,
		startDate: startDate,
		endDate:   endDate,
		status:    PeriodStatusOpen,
	}, nil
}

// Close closes the accounting period.
// Business rule: A closed period cannot be modified.
func (ap *AccountingPeriod) Close(closedBy string) error {
	if ap.status != PeriodStatusOpen {
		return errors.New("only open periods can be closed")
	}
	if closedBy == "" {
		return errors.New("closed by is required")
	}

	now := time.Now().UTC()
	ap.status = PeriodStatusClosed
	ap.closedAt = &now
	ap.closedBy = closedBy

	return nil
}

// Lock locks the accounting period.
// Business rule: A locked period cannot be reopened.
func (ap *AccountingPeriod) Lock() error {
	if ap.status != PeriodStatusClosed {
		return errors.New("only closed periods can be locked")
	}

	ap.status = PeriodStatusLocked
	return nil
}

// CanPostJournalEntry checks if a journal entry can be posted in this period.
func (ap *AccountingPeriod) CanPostJournalEntry(entryDate time.Time) bool {
	if ap.status == PeriodStatusLocked {
		return false
	}
	if ap.status == PeriodStatusClosed {
		return false
	}

	return (entryDate.Equal(ap.startDate) || entryDate.After(ap.startDate)) &&
		(entryDate.Equal(ap.endDate) || entryDate.Before(ap.endDate))
}

// GetID returns the period ID.
func (ap *AccountingPeriod) GetID() AccountingPeriodID {
	return ap.id
}

// GetName returns the period name.
func (ap *AccountingPeriod) GetName() string {
	return ap.name
}

// GetStatus returns the period status.
func (ap *AccountingPeriod) GetStatus() PeriodStatus {
	return ap.status
}

// GetDateRange returns the start and end dates.
func (ap *AccountingPeriod) GetDateRange() (time.Time, time.Time) {
	return ap.startDate, ap.endDate
}

// String returns a string representation.
func (ap *AccountingPeriod) String() string {
	return fmt.Sprintf("AccountingPeriod{id=%s, name=%s, status=%s}",
		ap.id, ap.name, ap.status)
}
