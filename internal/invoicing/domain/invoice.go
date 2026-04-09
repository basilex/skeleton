package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Invoice struct {
	id            InvoiceID
	invoiceNumber string
	orderID       *string
	contractID    *string
	customerID    string
	supplierID    *string
	issueDate     time.Time
	dueDate       time.Time
	status        InvoiceStatus
	lines         []*InvoiceLine
	subtotal      float64
	taxRate       float64
	taxAmount     float64
	discount      float64
	total         float64
	currency      string
	notes         *string
	paidAmount    float64
	payments      []*Payment
	createdAt     time.Time
	updatedAt     time.Time
	events        []eventbus.Event
}

func NewInvoice(
	invoiceNumber string,
	customerID string,
	currency string,
	dueDate time.Time,
) (*Invoice, error) {
	if invoiceNumber == "" {
		return nil, ErrEmptyInvoiceNumber
	}
	if customerID == "" {
		return nil, ErrEmptyCustomerID
	}
	if currency == "" {
		return nil, errors.New("currency cannot be empty")
	}
	if dueDate.Before(time.Now()) {
		return nil, ErrInvalidDueDate
	}

	now := time.Now()
	return &Invoice{
		id:            NewInvoiceID(),
		invoiceNumber: invoiceNumber,
		customerID:    customerID,
		status:        InvoiceStatusDraft,
		issueDate:     now,
		dueDate:       dueDate,
		currency:      currency,
		lines:         make([]*InvoiceLine, 0),
		payments:      make([]*Payment, 0),
		createdAt:     now,
		updatedAt:     now,
		events:        make([]eventbus.Event, 0),
	}, nil
}

func RestoreInvoice(
	id InvoiceID,
	invoiceNumber string,
	orderID *string,
	contractID *string,
	customerID string,
	supplierID *string,
	issueDate time.Time,
	dueDate time.Time,
	status InvoiceStatus,
	lines []*InvoiceLine,
	subtotal float64,
	taxRate float64,
	taxAmount float64,
	discount float64,
	total float64,
	currency string,
	notes *string,
	paidAmount float64,
	payments []*Payment,
	createdAt time.Time,
	updatedAt time.Time,
) *Invoice {
	return &Invoice{
		id:            id,
		invoiceNumber: invoiceNumber,
		orderID:       orderID,
		contractID:    contractID,
		customerID:    customerID,
		supplierID:    supplierID,
		issueDate:     issueDate,
		dueDate:       dueDate,
		status:        status,
		lines:         lines,
		subtotal:      subtotal,
		taxRate:       taxRate,
		taxAmount:     taxAmount,
		discount:      discount,
		total:         total,
		currency:      currency,
		notes:         notes,
		paidAmount:    paidAmount,
		payments:      payments,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		events:        make([]eventbus.Event, 0),
	}
}

func (i *Invoice) AddLine(
	description string,
	quantity float64,
	unitPrice float64,
	unit string,
	discount float64,
) error {
	if i.status != InvoiceStatusDraft {
		return errors.New("cannot add lines to non-draft invoice")
	}

	line, err := NewInvoiceLine(i.id, description, quantity, unitPrice, unit, discount)
	if err != nil {
		return err
	}

	i.lines = append(i.lines, line)
	i.recalculateTotals()
	i.updatedAt = time.Now()
	return nil
}

func (i *Invoice) RemoveLine(lineID InvoiceLineID) error {
	if i.status != InvoiceStatusDraft {
		return errors.New("cannot remove lines from non-draft invoice")
	}

	for idx, line := range i.lines {
		if line.GetID() == lineID {
			i.lines = append(i.lines[:idx], i.lines[idx+1:]...)
			i.recalculateTotals()
			i.updatedAt = time.Now()
			return nil
		}
	}

	return ErrInvoiceLineNotFound
}

func (i *Invoice) Send() error {
	if i.status != InvoiceStatusDraft {
		return fmt.Errorf("%w: can only send draft invoices", ErrInvalidInvoiceStatus)
	}

	if len(i.lines) == 0 {
		return errors.New("cannot send invoice without lines")
	}

	i.status = InvoiceStatusSent
	i.updatedAt = time.Now()
	i.events = append(i.events, InvoiceSent{
		InvoiceID:     i.id,
		InvoiceNumber: i.invoiceNumber,
		CustomerID:    i.customerID,
		SentAt:        time.Now().Format(time.RFC3339),
		occurredAt:    time.Now(),
	})
	return nil
}

func (i *Invoice) MarkAsViewed() error {
	if i.status != InvoiceStatusSent {
		return fmt.Errorf("%w: can only mark sent invoices as viewed", ErrInvalidInvoiceStatus)
	}

	i.status = InvoiceStatusViewed
	i.updatedAt = time.Now()
	return nil
}

func (i *Invoice) RecordPayment(
	amount float64,
	method PaymentMethod,
	reference string,
) (*Payment, error) {
	if i.status == InvoiceStatusDraft {
		return nil, errors.New("cannot record payment for draft invoice")
	}
	if i.status == InvoiceStatusCancelled {
		return nil, ErrInvoiceAlreadyCancelled
	}
	if i.status == InvoiceStatusPaid {
		return nil, ErrInvoiceAlreadyPaid
	}

	if i.paidAmount+amount > i.total {
		return nil, ErrPaymentExceedsAmount
	}

	payment, err := NewPayment(i.id, amount, i.currency, method, reference)
	if err != nil {
		return nil, err
	}

	i.payments = append(i.payments, payment)
	i.paidAmount += amount
	i.updatedAt = time.Now()

	if i.paidAmount >= i.total {
		i.status = InvoiceStatusPaid
		i.events = append(i.events, InvoicePaid{
			InvoiceID:     i.id,
			InvoiceNumber: i.invoiceNumber,
			CustomerID:    i.customerID,
			Total:         i.total,
			PaidAmount:    i.paidAmount,
			PaidAt:        time.Now().Format(time.RFC3339),
			occurredAt:    time.Now(),
		})
	}

	i.events = append(i.events, InvoicePaymentRecorded{
		InvoiceID:  i.id,
		PaymentID:  payment.id,
		Amount:     amount,
		Method:     method,
		occurredAt: time.Now(),
	})

	return payment, nil
}

func (i *Invoice) MarkAsOverdue() error {
	if i.status != InvoiceStatusSent && i.status != InvoiceStatusViewed {
		return fmt.Errorf("%w: can only mark sent/viewed invoices as overdue", ErrInvalidInvoiceStatus)
	}

	if time.Now().Before(i.dueDate) {
		return errors.New("invoice is not past due date yet")
	}

	i.status = InvoiceStatusOverdue
	i.updatedAt = time.Now()
	i.events = append(i.events, InvoiceOverdue{
		InvoiceID:     i.id,
		InvoiceNumber: i.invoiceNumber,
		CustomerID:    i.customerID,
		OverdueAt:     time.Now().Format(time.RFC3339),
		occurredAt:    time.Now(),
	})
	return nil
}

func (i *Invoice) Cancel(reason string) error {
	if i.status == InvoiceStatusPaid {
		return ErrInvoiceAlreadyPaid
	}
	if i.status == InvoiceStatusCancelled {
		return ErrInvoiceAlreadyCancelled
	}

	i.status = InvoiceStatusCancelled
	if i.notes == nil {
		notes := fmt.Sprintf("Cancelled: %s", reason)
		i.notes = &notes
	} else {
		notes := fmt.Sprintf("%s\nCancelled: %s", *i.notes, reason)
		i.notes = &notes
	}
	i.updatedAt = time.Now()
	i.events = append(i.events, InvoiceCancelled{
		InvoiceID:     i.id,
		InvoiceNumber: i.invoiceNumber,
		CustomerID:    i.customerID,
		Reason:        reason,
		CancelledAt:   time.Now().Format(time.RFC3339),
		occurredAt:    time.Now(),
	})
	return nil
}

func (i *Invoice) LinkOrder(orderID string) {
	i.orderID = &orderID
	i.updatedAt = time.Now()
}

func (i *Invoice) LinkContract(contractID string) {
	i.contractID = &contractID
	i.updatedAt = time.Now()
}

func (i *Invoice) SetSupplier(supplierID string) {
	i.supplierID = &supplierID
	i.updatedAt = time.Now()
}

func (i *Invoice) SetNotes(notes string) {
	i.notes = &notes
	i.updatedAt = time.Now()
}

func (i *Invoice) recalculateTotals() {
	i.subtotal = 0
	for _, line := range i.lines {
		i.subtotal += line.GetTotal()
	}

	// Calculate tax based on rate and net amount
	netAmount := i.subtotal - i.discount
	if i.taxRate > 0 {
		i.taxAmount = netAmount * (i.taxRate / 100)
	}

	i.total = i.subtotal + i.taxAmount - i.discount
}

func (i *Invoice) SetTaxRate(taxRate float64) error {
	if i.status != InvoiceStatusDraft {
		return errors.New("cannot set tax rate on non-draft invoice")
	}
	if taxRate < 0 {
		return errors.New("tax rate cannot be negative")
	}
	if taxRate > 100 {
		return errors.New("tax rate cannot exceed 100%")
	}

	i.taxRate = taxRate
	i.recalculateTotals()
	i.updatedAt = time.Now()
	return nil
}

func (i *Invoice) SetDiscount(discount float64) error {
	if i.status != InvoiceStatusDraft {
		return errors.New("cannot set discount on non-draft invoice")
	}
	if discount < 0 {
		return errors.New("discount cannot be negative")
	}
	if discount > i.subtotal {
		return errors.New("discount cannot exceed subtotal")
	}

	i.discount = discount
	i.recalculateTotals()
	i.updatedAt = time.Now()
	return nil
}

func (i *Invoice) GetID() InvoiceID {
	return i.id
}

func (i *Invoice) GetInvoiceNumber() string {
	return i.invoiceNumber
}

func (i *Invoice) GetOrderID() *string {
	return i.orderID
}

func (i *Invoice) GetContractID() *string {
	return i.contractID
}

func (i *Invoice) GetCustomerID() string {
	return i.customerID
}

func (i *Invoice) GetSupplierID() *string {
	return i.supplierID
}

func (i *Invoice) GetIssueDate() time.Time {
	return i.issueDate
}

func (i *Invoice) GetDueDate() time.Time {
	return i.dueDate
}

func (i *Invoice) GetStatus() InvoiceStatus {
	return i.status
}

func (i *Invoice) GetLines() []*InvoiceLine {
	return i.lines
}

func (i *Invoice) GetSubtotal() float64 {
	return i.subtotal
}

func (i *Invoice) GetTaxRate() float64 {
	return i.taxRate
}

func (i *Invoice) GetTaxAmount() float64 {
	return i.taxAmount
}

func (i *Invoice) GetNetAmount() float64 {
	return i.subtotal - i.discount
}

func (i *Invoice) GetGrossAmount() float64 {
	return i.subtotal + i.taxAmount - i.discount
}

func (i *Invoice) GetDiscount() float64 {
	return i.discount
}

func (i *Invoice) GetTotal() float64 {
	return i.total
}

func (i *Invoice) GetCurrency() string {
	return i.currency
}

func (i *Invoice) GetNotes() *string {
	return i.notes
}

func (i *Invoice) GetPaidAmount() float64 {
	return i.paidAmount
}

func (i *Invoice) GetPayments() []*Payment {
	return i.payments
}

func (i *Invoice) GetCreatedAt() time.Time {
	return i.createdAt
}

func (i *Invoice) GetUpdatedAt() time.Time {
	return i.updatedAt
}

func (i *Invoice) PullEvents() []eventbus.Event {
	events := i.events
	i.events = make([]eventbus.Event, 0)
	return events
}
