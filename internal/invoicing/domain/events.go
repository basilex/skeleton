package domain

import (
	"time"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

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

// CreditNote events
type CreditNoteIssued struct {
	CreditNoteID     CreditNoteID
	CreditNoteNumber string
	CustomerID       string
	InvoiceID        *InvoiceID
	Total            float64
	Currency         string
	occurredAt       time.Time
}

func (e CreditNoteIssued) EventName() string {
	return "invoicing.credit_note_issued"
}

func (e CreditNoteIssued) OccurredAt() time.Time {
	return e.occurredAt
}

type CreditNoteFullyApplied struct {
	CreditNoteID     CreditNoteID
	CreditNoteNumber string
	CustomerID       string
	Total            float64
	occurredAt       time.Time
}

func (e CreditNoteFullyApplied) EventName() string {
	return "invoicing.credit_note_fully_applied"
}

func (e CreditNoteFullyApplied) OccurredAt() time.Time {
	return e.occurredAt
}

type CreditNoteCancelled struct {
	CreditNoteID     CreditNoteID
	CreditNoteNumber string
	Reason           string
	occurredAt       time.Time
}

func (e CreditNoteCancelled) EventName() string {
	return "invoicing.credit_note_cancelled"
}

func (e CreditNoteCancelled) OccurredAt() time.Time {
	return e.occurredAt
}

// Installment events
type InstallmentPaid struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	Amount        float64
	PaidAt        time.Time
	occurredAt    time.Time
}

func (e InstallmentPaid) EventName() string {
	return "invoicing.installment_paid"
}

func (e InstallmentPaid) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentPartiallyPaid struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	Amount        float64
	occurredAt    time.Time
}

func (e InstallmentPartiallyPaid) EventName() string {
	return "invoicing.installment_partially_paid"
}

func (e InstallmentPartiallyPaid) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentDue struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	DueDate       time.Time
	occurredAt    time.Time
}

func (e InstallmentDue) EventName() string {
	return "invoicing.installment_due"
}

func (e InstallmentDue) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentOverdue struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	DueDate       time.Time
	occurredAt    time.Time
}

func (e InstallmentOverdue) EventName() string {
	return "invoicing.installment_overdue"
}

func (e InstallmentOverdue) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentReminderSent struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	RemindedAt    time.Time
	occurredAt    time.Time
}

func (e InstallmentReminderSent) EventName() string {
	return "invoicing.installment_reminder_sent"
}

func (e InstallmentReminderSent) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentCancelled struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	Reason        string
	occurredAt    time.Time
}

func (e InstallmentCancelled) EventName() string {
	return "invoicing.installment_cancelled"
}

func (e InstallmentCancelled) OccurredAt() time.Time {
	return e.occurredAt
}

type InstallmentFailed struct {
	InstallmentID InstallmentID
	InvoiceID     InvoiceID
	Reason        string
	occurredAt    time.Time
}

func (e InstallmentFailed) EventName() string {
	return "invoicing.installment_failed"
}

func (e InstallmentFailed) OccurredAt() time.Time {
	return e.occurredAt
}
