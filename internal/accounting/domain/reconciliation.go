package domain

import (
	"errors"
	"fmt"
	"time"
)

// Reconciliation represents a bank statement reconciliation.
// It matches accounting records with external statements (bank, credit card, etc.)
type Reconciliation struct {
	id               ReconciliationID
	accountID        AccountID
	statementDate    time.Time
	statementBalance Money
	bookBalance      Money
	difference       Money
	status           ReconciliationStatus
	reconciledAt     *time.Time
	reconciledBy     string
	items            []ReconciliationItem
}

// ReconciliationItem represents a single reconciled transaction.
type ReconciliationItem struct {
	journalEntryID JournalEntryID
	transactionID  TransactionID
	amount         Money
	cleared        bool
	clearedAt      *time.Time
	reference      string
}

// ReconciliationStatus represents the status of a reconciliation.
type ReconciliationStatus string

const (
	ReconciliationStatusInProgress  ReconciliationStatus = "in_progress"
	ReconciliationStatusCompleted   ReconciliationStatus = "completed"
	ReconciliationStatusDiscrepancy ReconciliationStatus = "discrepancy"
)

// NewReconciliation creates a new reconciliation.
func NewReconciliation(accountID AccountID, statementDate time.Time, statementBalance, bookBalance Money) (*Reconciliation, error) {
	if statementBalance.Currency != bookBalance.Currency {
		return nil, errors.New("currencies must match")
	}

	difference := statementBalance.Subtract(bookBalance)

	return &Reconciliation{
		id:               NewReconciliationID(),
		accountID:        accountID,
		statementDate:    statementDate,
		statementBalance: statementBalance,
		bookBalance:      bookBalance,
		difference:       difference,
		status:           ReconciliationStatusInProgress,
		items:            make([]ReconciliationItem, 0),
	}, nil
}

// AddItem adds a reconciled item to the reconciliation.
func (r *Reconciliation) AddItem(journalEntryID JournalEntryID, amount Money, reference string) error {
	if r.status != ReconciliationStatusInProgress {
		return errors.New("cannot add items to completed reconciliation")
	}

	item := ReconciliationItem{
		journalEntryID: journalEntryID,
		amount:         amount,
		reference:      reference,
		cleared:        false,
	}

	r.items = append(r.items, item)
	return nil
}

// MarkCleared marks an item as cleared.
func (r *Reconciliation) MarkCleared(journalEntryID JournalEntryID) error {
	for i, item := range r.items {
		if item.journalEntryID == journalEntryID {
			now := time.Now().UTC()
			r.items[i].cleared = true
			r.items[i].clearedAt = &now
			return nil
		}
	}
	return errors.New("item not found")
}

// Complete completes the reconciliation if all items are cleared and balanced.
func (r *Reconciliation) Complete(reconciledBy string) error {
	if r.status != ReconciliationStatusInProgress {
		return errors.New("reconciliation is not in progress")
	}

	if !r.AllCleared() {
		return errors.New("not all items are cleared")
	}

	if !r.IsBalanced() {
		r.status = ReconciliationStatusDiscrepancy
		return errors.New("reconciliation has discrepancy")
	}

	now := time.Now().UTC()
	r.status = ReconciliationStatusCompleted
	r.reconciledAt = &now
	r.reconciledBy = reconciledBy

	return nil
}

// IsBalanced checks if the reconciliation balances (statement == book).
func (r *Reconciliation) IsBalanced() bool {
	return r.difference.IsZero()
}

// AllCleared checks if all items are cleared.
func (r *Reconciliation) AllCleared() bool {
	for _, item := range r.items {
		if !item.cleared {
			return false
		}
	}
	return true
}

// GetDifference returns the difference between statement and book balance.
func (r *Reconciliation) GetDifference() Money {
	return r.difference
}

// GetID returns the reconciliation ID.
func (r *Reconciliation) GetID() ReconciliationID {
	return r.id
}

// GetStatus returns the current status.
func (r *Reconciliation) GetStatus() ReconciliationStatus {
	return r.status
}

// GetItems returns all reconciliation items.
func (r *Reconciliation) GetItems() []ReconciliationItem {
	return r.items
}

// String returns a string representation.
func (r *Reconciliation) String() string {
	return fmt.Sprintf("Reconciliation{id=%s, status=%s, balanced=%v}",
		r.id, r.status, r.IsBalanced())
}
