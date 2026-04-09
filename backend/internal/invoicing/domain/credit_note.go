package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type CreditNoteID string

func NewCreditNoteID() CreditNoteID {
	return CreditNoteID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id CreditNoteID) String() string {
	return string(id)
}

type CreditNoteStatus string

const (
	CreditNoteStatusDraft     CreditNoteStatus = "draft"
	CreditNoteStatusIssued    CreditNoteStatus = "issued"
	CreditNoteStatusApplied   CreditNoteStatus = "applied"
	CreditNoteStatusCancelled CreditNoteStatus = "cancelled"
)

func (s CreditNoteStatus) String() string {
	return string(s)
}

type CreditNoteReason string

const (
	CreditNoteReasonRefund     CreditNoteReason = "refund"
	CreditNoteReasonDiscount   CreditNoteReason = "discount"
	CreditNoteReasonCorrection CreditNoteReason = "correction"
	CreditNoteReasonReturn     CreditNoteReason = "return"
	CreditNoteReasonBadDebt    CreditNoteReason = "bad_debt"
)

func (r CreditNoteReason) String() string {
	return string(r)
}

type CreditNoteLine struct {
	description string
	quantity    float64
	unitPrice   money.Money
	total       money.Money
}

func NewCreditNoteLine(description string, quantity float64, unitPrice money.Money) CreditNoteLine {
	total, _ := unitPrice.Multiply(quantity)
	return CreditNoteLine{
		description: description,
		quantity:    quantity,
		unitPrice:   unitPrice,
		total:       total,
	}
}

func (l CreditNoteLine) GetDescription() string    { return l.description }
func (l CreditNoteLine) GetQuantity() float64      { return l.quantity }
func (l CreditNoteLine) GetUnitPrice() money.Money { return l.unitPrice }
func (l CreditNoteLine) GetTotal() money.Money     { return l.total }

type CreditNote struct {
	id               CreditNoteID
	creditNoteNumber string
	invoiceID        *InvoiceID
	customerID       string
	reason           CreditNoteReason
	description      string
	lines            []CreditNoteLine
	subtotal         money.Money
	taxAmount        money.Money
	total            money.Money
	currency         string
	appliedAmount    money.Money
	status           CreditNoteStatus
	issuedAt         *time.Time
	appliedAt        *time.Time
	cancelledAt      *time.Time
	createdAt        time.Time
	updatedAt        time.Time
	events           []DomainEvent
}

func NewCreditNote(creditNoteNumber string, customerID string, reason CreditNoteReason, description string, currency string) (*CreditNote, error) {
	if creditNoteNumber == "" {
		return nil, errors.New("credit note number is required")
	}
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}
	if currency == "" {
		return nil, errors.New("currency is required")
	}

	now := time.Now().UTC()
	zeroMoney, _ := money.New(0, currency)
	return &CreditNote{
		id:               NewCreditNoteID(),
		creditNoteNumber: creditNoteNumber,
		customerID:       customerID,
		reason:           reason,
		description:      description,
		currency:         currency,
		lines:            make([]CreditNoteLine, 0),
		subtotal:         zeroMoney,
		taxAmount:        zeroMoney,
		total:            zeroMoney,
		appliedAmount:    zeroMoney,
		status:           CreditNoteStatusDraft,
		createdAt:        now,
		updatedAt:        now,
		events:           make([]DomainEvent, 0),
	}, nil
}

func (cn *CreditNote) GetID() CreditNoteID           { return cn.id }
func (cn *CreditNote) GetCreditNoteNumber() string   { return cn.creditNoteNumber }
func (cn *CreditNote) GetInvoiceID() *InvoiceID      { return cn.invoiceID }
func (cn *CreditNote) GetCustomerID() string         { return cn.customerID }
func (cn *CreditNote) GetReason() CreditNoteReason   { return cn.reason }
func (cn *CreditNote) GetDescription() string        { return cn.description }
func (cn *CreditNote) GetLines() []CreditNoteLine    { return cn.lines }
func (cn *CreditNote) GetSubtotal() money.Money      { return cn.subtotal }
func (cn *CreditNote) GetTaxAmount() money.Money     { return cn.taxAmount }
func (cn *CreditNote) GetTotal() money.Money         { return cn.total }
func (cn *CreditNote) GetCurrency() string           { return cn.currency }
func (cn *CreditNote) GetAppliedAmount() money.Money { return cn.appliedAmount }
func (cn *CreditNote) GetStatus() CreditNoteStatus   { return cn.status }
func (cn *CreditNote) GetIssuedAt() *time.Time       { return cn.issuedAt }
func (cn *CreditNote) GetAppliedAt() *time.Time      { return cn.appliedAt }
func (cn *CreditNote) GetCancelledAt() *time.Time    { return cn.cancelledAt }
func (cn *CreditNote) GetCreatedAt() time.Time       { return cn.createdAt }
func (cn *CreditNote) GetUpdatedAt() time.Time       { return cn.updatedAt }

func (cn *CreditNote) LinkInvoice(invoiceID InvoiceID) error {
	if cn.status != CreditNoteStatusDraft {
		return errors.New("can only link invoice to draft credit note")
	}
	cn.invoiceID = &invoiceID
	cn.updatedAt = time.Now().UTC()
	return nil
}

func (cn *CreditNote) AddLine(description string, quantity float64, unitPrice money.Money) error {
	if cn.status != CreditNoteStatusDraft {
		return errors.New("can only add lines to draft credit note")
	}

	line := NewCreditNoteLine(description, quantity, unitPrice)
	cn.lines = append(cn.lines, line)
	cn.recalculateTotals()
	cn.updatedAt = time.Now().UTC()
	return nil
}

func (cn *CreditNote) SetTax(taxAmount money.Money) error {
	if cn.status != CreditNoteStatusDraft {
		return errors.New("can only set tax on draft credit note")
	}
	cn.taxAmount = taxAmount
	cn.recalculateTotals()
	cn.updatedAt = time.Now().UTC()
	return nil
}

func (cn *CreditNote) recalculateTotals() {
	zeroMoney, _ := money.New(0, cn.currency)
	subtotal := zeroMoney

	for _, line := range cn.lines {
		subtotal, _ = subtotal.Add(line.GetTotal())
	}

	cn.subtotal = subtotal
	total, _ := subtotal.Add(cn.taxAmount)
	cn.total = total
}

func (cn *CreditNote) Issue() error {
	if cn.status != CreditNoteStatusDraft {
		return fmt.Errorf("cannot issue credit note in %s status", cn.status)
	}
	if len(cn.lines) == 0 {
		return errors.New("credit note must have at least one line")
	}

	now := time.Now().UTC()
	cn.status = CreditNoteStatusIssued
	cn.issuedAt = &now
	cn.updatedAt = now

	cn.events = append(cn.events, CreditNoteIssued{
		CreditNoteID:     cn.id,
		CreditNoteNumber: cn.creditNoteNumber,
		CustomerID:       cn.customerID,
		InvoiceID:        cn.invoiceID,
		Total:            cn.total,
		Currency:         cn.currency,
		occurredAt:       now,
	})

	return nil
}

func (cn *CreditNote) Apply(amount money.Money) error {
	if cn.status != CreditNoteStatusIssued {
		return fmt.Errorf("cannot apply credit note in %s status", cn.status)
	}
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("amount must be positive")
	}

	newApplied, err := cn.appliedAmount.Add(amount)
	if err != nil {
		return err
	}

	if newApplied.GreaterThan(cn.total) {
		return fmt.Errorf("applied amount %s exceeds credit note total %s", newApplied.String(), cn.total.String())
	}

	cn.appliedAmount = newApplied
	cn.updatedAt = time.Now().UTC()

	if cn.appliedAmount.GreaterThanOrEqual(cn.total) {
		now := time.Now().UTC()
		cn.status = CreditNoteStatusApplied
		cn.appliedAt = &now

		cn.events = append(cn.events, CreditNoteFullyApplied{
			CreditNoteID:     cn.id,
			CreditNoteNumber: cn.creditNoteNumber,
			CustomerID:       cn.customerID,
			Total:            cn.total,
			occurredAt:       now,
		})
	}

	return nil
}

func (cn *CreditNote) Cancel(reason string) error {
	if cn.status == CreditNoteStatusApplied {
		return errors.New("cannot cancel applied credit note")
	}
	if cn.status == CreditNoteStatusCancelled {
		return errors.New("credit note already cancelled")
	}

	now := time.Now().UTC()
	cn.status = CreditNoteStatusCancelled
	cn.cancelledAt = &now
	cn.updatedAt = now

	cn.events = append(cn.events, CreditNoteCancelled{
		CreditNoteID:     cn.id,
		CreditNoteNumber: cn.creditNoteNumber,
		Reason:           reason,
		occurredAt:       now,
	})

	return nil
}

func (cn *CreditNote) GetRemainingAmount() money.Money {
	remaining, _ := cn.total.Subtract(cn.appliedAmount)
	return remaining
}

func (cn *CreditNote) IsFullyApplied() bool {
	return cn.appliedAmount.GreaterThanOrEqual(cn.total)
}

func (cn *CreditNote) PullEvents() []DomainEvent {
	events := cn.events
	cn.events = make([]DomainEvent, 0)
	return events
}

func (cn *CreditNote) String() string {
	return fmt.Sprintf("CreditNote{id=%s, number=%s, customer=%s, total=%s, status=%s}",
		cn.id, cn.creditNoteNumber, cn.customerID, cn.total.String(), cn.status)
}

func ReconstituteCreditNote(
	id CreditNoteID,
	creditNoteNumber string,
	invoiceID *InvoiceID,
	customerID string,
	reason CreditNoteReason,
	description string,
	lines []CreditNoteLine,
	subtotal money.Money,
	taxAmount money.Money,
	total money.Money,
	currency string,
	appliedAmount money.Money,
	status CreditNoteStatus,
	issuedAt *time.Time,
	appliedAt *time.Time,
	cancelledAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *CreditNote {
	return &CreditNote{
		id:               id,
		creditNoteNumber: creditNoteNumber,
		invoiceID:        invoiceID,
		customerID:       customerID,
		reason:           reason,
		description:      description,
		lines:            lines,
		subtotal:         subtotal,
		taxAmount:        taxAmount,
		total:            total,
		currency:         currency,
		appliedAmount:    appliedAmount,
		status:           status,
		issuedAt:         issuedAt,
		appliedAt:        appliedAt,
		cancelledAt:      cancelledAt,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		events:           make([]DomainEvent, 0),
	}
}
