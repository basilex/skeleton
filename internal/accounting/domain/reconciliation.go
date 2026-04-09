package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type Reconciliation struct {
	id               ReconciliationID
	accountID        AccountID
	statementDate    time.Time
	statementBalance money.Money
	bookBalance      money.Money
	difference       money.Money
	status           ReconciliationStatus
	reconciledAt     *time.Time
	reconciledBy     string
	items            []ReconciliationItem
}

type ReconciliationItem struct {
	journalEntryID JournalEntryID
	transactionID  TransactionID
	amount         money.Money
	cleared        bool
	clearedAt      *time.Time
	reference      string
}

type ReconciliationStatus string

const (
	ReconciliationStatusInProgress  ReconciliationStatus = "in_progress"
	ReconciliationStatusCompleted   ReconciliationStatus = "completed"
	ReconciliationStatusDiscrepancy ReconciliationStatus = "discrepancy"
)

func NewReconciliation(accountID AccountID, statementDate time.Time, statementBalance, bookBalance money.Money) (*Reconciliation, error) {
	if statementBalance.GetCurrency() != bookBalance.GetCurrency() {
		return nil, errors.New("currencies must match")
	}

	difference, err := statementBalance.Subtract(bookBalance)
	if err != nil {
		return nil, err
	}

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

func (r *Reconciliation) AddItem(journalEntryID JournalEntryID, transactionID TransactionID, amount money.Money, reference string) error {
	if r.status != ReconciliationStatusInProgress {
		return errors.New("cannot add items to non-in-progress reconciliation")
	}

	r.items = append(r.items, ReconciliationItem{
		journalEntryID: journalEntryID,
		transactionID:  transactionID,
		amount:         amount,
		reference:      reference,
	})

	return nil
}

func (r *Reconciliation) MarkItemCleared(transactionID TransactionID) error {
	for i := range r.items {
		if r.items[i].transactionID == transactionID {
			now := time.Now().UTC()
			r.items[i].cleared = true
			r.items[i].clearedAt = &now
			return nil
		}
	}
	return fmt.Errorf("transaction %s not found in reconciliation", transactionID)
}

func (r *Reconciliation) Complete(reconciledBy string) error {
	if r.status != ReconciliationStatusInProgress {
		return errors.New("can only complete in-progress reconciliation")
	}

	now := time.Now().UTC()
	r.status = ReconciliationStatusCompleted
	r.reconciledAt = &now
	r.reconciledBy = reconciledBy

	if !r.difference.IsZero() {
		r.status = ReconciliationStatusDiscrepancy
	}

	return nil
}

func (r *Reconciliation) GetID() ReconciliationID {
	return r.id
}

func (r *Reconciliation) GetAccountID() AccountID {
	return r.accountID
}

func (r *Reconciliation) GetStatementDate() time.Time {
	return r.statementDate
}

func (r *Reconciliation) GetStatementBalance() money.Money {
	return r.statementBalance
}

func (r *Reconciliation) GetBookBalance() money.Money {
	return r.bookBalance
}

func (r *Reconciliation) GetDifference() money.Money {
	return r.difference
}

func (r *Reconciliation) GetStatus() ReconciliationStatus {
	return r.status
}

func (r *Reconciliation) GetReconciledAt() *time.Time {
	return r.reconciledAt
}

func (r *Reconciliation) GetReconciledBy() string {
	return r.reconciledBy
}

func (r *Reconciliation) GetItems() []ReconciliationItem {
	return r.items
}

func (r *Reconciliation) IsCleared() bool {
	for _, item := range r.items {
		if !item.cleared {
			return false
		}
	}
	return true
}

func (r *Reconciliation) HasDiscrepancy() bool {
	return r.status == ReconciliationStatusDiscrepancy
}

func (ri *ReconciliationItem) GetJournalEntryID() JournalEntryID {
	return ri.journalEntryID
}

func (ri *ReconciliationItem) GetTransactionID() TransactionID {
	return ri.transactionID
}

func (ri *ReconciliationItem) GetAmount() money.Money {
	return ri.amount
}

func (ri *ReconciliationItem) IsCleared() bool {
	return ri.cleared
}

func (ri *ReconciliationItem) GetClearedAt() *time.Time {
	return ri.clearedAt
}

func (ri *ReconciliationItem) GetReference() string {
	return ri.reference
}
