package domain

import (
	"errors"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type Payment struct {
	id        PaymentID
	invoiceID InvoiceID
	amount    money.Money
	currency  string
	method    PaymentMethod
	reference string
	paidAt    time.Time
	notes     string
}

func NewPayment(
	invoiceID InvoiceID,
	amount money.Money,
	currency string,
	method PaymentMethod,
	reference string,
) (*Payment, error) {
	if invoiceID.IsZero() {
		return nil, errors.New("invoice ID cannot be empty")
	}
	if amount.IsNegative() || amount.IsZero() {
		return nil, ErrInvalidAmount
	}
	if currency == "" {
		return nil, errors.New("currency cannot be empty")
	}

	return &Payment{
		id:        NewPaymentID(),
		invoiceID: invoiceID,
		amount:    amount,
		currency:  currency,
		method:    method,
		reference: reference,
		paidAt:    time.Now(),
	}, nil
}

func RestorePayment(
	id PaymentID,
	invoiceID InvoiceID,
	amount money.Money,
	currency string,
	method PaymentMethod,
	reference string,
	paidAt time.Time,
	notes string,
) *Payment {
	return &Payment{
		id:        id,
		invoiceID: invoiceID,
		amount:    amount,
		currency:  currency,
		method:    method,
		reference: reference,
		paidAt:    paidAt,
		notes:     notes,
	}
}

func (p *Payment) GetID() PaymentID {
	return p.id
}

func (p *Payment) GetInvoiceID() InvoiceID {
	return p.invoiceID
}

func (p *Payment) GetAmount() money.Money {
	return p.amount
}

func (p *Payment) GetCurrency() string {
	return p.currency
}

func (p *Payment) GetMethod() PaymentMethod {
	return p.method
}

func (p *Payment) GetReference() string {
	return p.reference
}

func (p *Payment) GetPaidAt() time.Time {
	return p.paidAt
}

func (p *Payment) GetNotes() string {
	return p.notes
}

func (p *Payment) AddNotes(notes string) {
	p.notes = notes
}
