package domain

import (
	"time"
)

type InvoiceCreated struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	Total         float64
	Currency      string
	occurredAt    time.Time
}

func (e InvoiceCreated) EventName() string {
	return "invoicing.invoice_created"
}

func (e InvoiceCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoiceSent struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	SentAt        string
	occurredAt    time.Time
}

func (e InvoiceSent) EventName() string {
	return "invoicing.invoice_sent"
}

func (e InvoiceSent) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoiceViewed struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	ViewedAt      time.Time
	occurredAt    time.Time
}

func (e InvoiceViewed) EventName() string {
	return "invoicing.invoice_viewed"
}

func (e InvoiceViewed) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoicePaid struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	Total         float64
	PaidAmount    float64
	PaidAt        string
	occurredAt    time.Time
}

func (e InvoicePaid) EventName() string {
	return "invoicing.invoice_paid"
}

func (e InvoicePaid) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoiceOverdue struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	OverdueAt     string
	occurredAt    time.Time
}

func (e InvoiceOverdue) EventName() string {
	return "invoicing.invoice_overdue"
}

func (e InvoiceOverdue) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoiceCancelled struct {
	InvoiceID     InvoiceID
	InvoiceNumber string
	CustomerID    string
	Reason        string
	CancelledAt   string
	occurredAt    time.Time
}

func (e InvoiceCancelled) EventName() string {
	return "invoicing.invoice_cancelled"
}

func (e InvoiceCancelled) OccurredAt() time.Time {
	return e.occurredAt
}

type InvoicePaymentRecorded struct {
	InvoiceID  InvoiceID
	PaymentID  PaymentID
	Amount     float64
	Method     PaymentMethod
	occurredAt time.Time
}

func (e InvoicePaymentRecorded) EventName() string {
	return "invoicing.payment_recorded"
}

func (e InvoicePaymentRecorded) OccurredAt() time.Time {
	return e.occurredAt
}
