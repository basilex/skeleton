package domain

import (
	"errors"
	"fmt"
	"time"
)

type InstallmentID string

func NewInstallmentID() InstallmentID {
	return InstallmentID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id InstallmentID) String() string {
	return string(id)
}

type InstallmentStatus string

const (
	InstallmentStatusPending   InstallmentStatus = "pending"
	InstallmentStatusDue       InstallmentStatus = "due"
	InstallmentStatusPaid      InstallmentStatus = "paid"
	InstallmentStatusOverdue   InstallmentStatus = "overdue"
	InstallmentStatusCancelled InstallmentStatus = "cancelled"
	InstallmentStatusFailed    InstallmentStatus = "failed"
)

func (s InstallmentStatus) String() string {
	return string(s)
}

type Installment struct {
	id                InstallmentID
	invoiceID         InvoiceID
	installmentNumber int
	totalAmount       float64
	paidAmount        float64
	dueDate           time.Time
	paidAt            *time.Time
	remindedAt        *time.Time
	status            InstallmentStatus
	failedReason      string
	createdAt         time.Time
	updatedAt         time.Time
	events            []DomainEvent
}

func NewInstallment(invoiceID InvoiceID, installmentNumber int, totalAmount float64, dueDate time.Time) (*Installment, error) {
	if totalAmount <= 0 {
		return nil, errors.New("installment amount must be positive")
	}
	if dueDate.Before(time.Now()) {
		return nil, errors.New("due date must be in the future")
	}

	now := time.Now().UTC()
	return &Installment{
		id:                NewInstallmentID(),
		invoiceID:         invoiceID,
		installmentNumber: installmentNumber,
		totalAmount:       totalAmount,
		paidAmount:        0,
		dueDate:           dueDate,
		status:            InstallmentStatusPending,
		createdAt:         now,
		updatedAt:         now,
		events:            make([]DomainEvent, 0),
	}, nil
}

func (i *Installment) GetID() InstallmentID         { return i.id }
func (i *Installment) GetInvoiceID() InvoiceID      { return i.invoiceID }
func (i *Installment) GetInstallmentNumber() int    { return i.installmentNumber }
func (i *Installment) GetTotalAmount() float64      { return i.totalAmount }
func (i *Installment) GetPaidAmount() float64       { return i.paidAmount }
func (i *Installment) GetDueDate() time.Time        { return i.dueDate }
func (i *Installment) GetPaidAt() *time.Time        { return i.paidAt }
func (i *Installment) GetRemindedAt() *time.Time    { return i.remindedAt }
func (i *Installment) GetStatus() InstallmentStatus { return i.status }
func (i *Installment) GetFailedReason() string      { return i.failedReason }
func (i *Installment) GetCreatedAt() time.Time      { return i.createdAt }
func (i *Installment) GetUpdatedAt() time.Time      { return i.updatedAt }

func (i *Installment) GetRemainingAmount() float64 {
	return i.totalAmount - i.paidAmount
}

func (i *Installment) IsPaid() bool {
	return i.status == InstallmentStatusPaid
}

func (i *Installment) IsOverdue() bool {
	return i.status == InstallmentStatusOverdue
}

func (i *Installment) MakePayment(amount float64) error {
	if i.status == InstallmentStatusPaid {
		return errors.New("installment already paid")
	}
	if i.status == InstallmentStatusCancelled {
		return errors.New("installment is cancelled")
	}
	if i.status == InstallmentStatusFailed {
		return errors.New("installment has failed")
	}

	newPaidAmount := i.paidAmount + amount
	if newPaidAmount > i.totalAmount {
		return fmt.Errorf("payment amount %.2f exceeds remaining amount %.2f", amount, i.GetRemainingAmount())
	}

	i.paidAmount = newPaidAmount
	i.updatedAt = time.Now().UTC()

	if i.paidAmount >= i.totalAmount {
		now := time.Now().UTC()
		i.status = InstallmentStatusPaid
		i.paidAt = &now

		i.events = append(i.events, InstallmentPaid{
			InstallmentID: i.id,
			InvoiceID:     i.invoiceID,
			Amount:        amount,
			PaidAt:        now,
			occurredAt:    now,
		})
	} else {
		i.events = append(i.events, InstallmentPartiallyPaid{
			InstallmentID: i.id,
			InvoiceID:     i.invoiceID,
			Amount:        amount,
			occurredAt:    time.Now().UTC(),
		})
	}

	return nil
}

func (i *Installment) MarkAsDue() error {
	if i.status != InstallmentStatusPending {
		return fmt.Errorf("cannot mark installment in %s status as due", i.status)
	}

	i.status = InstallmentStatusDue
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, InstallmentDue{
		InstallmentID: i.id,
		InvoiceID:     i.invoiceID,
		DueDate:       i.dueDate,
		occurredAt:    time.Now().UTC(),
	})

	return nil
}

func (i *Installment) MarkAsOverdue() error {
	if i.status != InstallmentStatusPending && i.status != InstallmentStatusDue {
		return fmt.Errorf("cannot mark installment in %s status as overdue", i.status)
	}
	if i.IsPaid() {
		return errors.New("installment is already paid")
	}

	i.status = InstallmentStatusOverdue
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, InstallmentOverdue{
		InstallmentID: i.id,
		InvoiceID:     i.invoiceID,
		DueDate:       i.dueDate,
		occurredAt:    time.Now().UTC(),
	})

	return nil
}

func (i *Installment) SendReminder() error {
	if i.status != InstallmentStatusPending && i.status != InstallmentStatusDue && i.status != InstallmentStatusOverdue {
		return fmt.Errorf("cannot send reminder for installment in %s status", i.status)
	}

	now := time.Now().UTC()
	i.remindedAt = &now
	i.updatedAt = now

	i.events = append(i.events, InstallmentReminderSent{
		InstallmentID: i.id,
		InvoiceID:     i.invoiceID,
		RemindedAt:    now,
		occurredAt:    now,
	})

	return nil
}

func (i *Installment) Cancel(reason string) error {
	if i.status == InstallmentStatusPaid {
		return errors.New("cannot cancel paid installment")
	}

	i.status = InstallmentStatusCancelled
	i.failedReason = reason
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, InstallmentCancelled{
		InstallmentID: i.id,
		InvoiceID:     i.invoiceID,
		Reason:        reason,
		occurredAt:    time.Now().UTC(),
	})

	return nil
}

func (i *Installment) MarkAsFailed(reason string) error {
	if i.status == InstallmentStatusPaid {
		return errors.New("cannot mark paid installment as failed")
	}

	i.status = InstallmentStatusFailed
	i.failedReason = reason
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, InstallmentFailed{
		InstallmentID: i.id,
		InvoiceID:     i.invoiceID,
		Reason:        reason,
		occurredAt:    time.Now().UTC(),
	})

	return nil
}

func (i *Installment) PullEvents() []DomainEvent {
	events := i.events
	i.events = make([]DomainEvent, 0)
	return events
}

func (i *Installment) String() string {
	return fmt.Sprintf("Installment{id=%s, invoice=%s, number=%d, amount=%.2f, status=%s}",
		i.id, i.invoiceID, i.installmentNumber, i.totalAmount, i.status)
}

func ReconstituteInstallment(
	id InstallmentID,
	invoiceID InvoiceID,
	installmentNumber int,
	totalAmount float64,
	paidAmount float64,
	dueDate time.Time,
	paidAt *time.Time,
	remindedAt *time.Time,
	status InstallmentStatus,
	failedReason string,
	createdAt time.Time,
	updatedAt time.Time,
) *Installment {
	return &Installment{
		id:                id,
		invoiceID:         invoiceID,
		installmentNumber: installmentNumber,
		totalAmount:       totalAmount,
		paidAmount:        paidAmount,
		dueDate:           dueDate,
		paidAt:            paidAt,
		remindedAt:        remindedAt,
		status:            status,
		failedReason:      failedReason,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		events:            make([]DomainEvent, 0),
	}
}

type InstallmentPlan struct {
	invoiceID    InvoiceID
	installments []*Installment
	totalAmount  float64
	currency     string
	createdAt    time.Time
}

func NewInstallmentPlan(invoiceID InvoiceID, totalAmount float64, currency string) *InstallmentPlan {
	return &InstallmentPlan{
		invoiceID:    invoiceID,
		installments: make([]*Installment, 0),
		totalAmount:  totalAmount,
		currency:     currency,
		createdAt:    time.Now().UTC(),
	}
}

func (p *InstallmentPlan) AddInstallment(amount float64, dueDate time.Time) error {
	if amount <= 0 {
		return errors.New("installment amount must be positive")
	}

	totalPlanned := float64(0)
	for _, inst := range p.installments {
		totalPlanned += inst.GetTotalAmount()
	}

	if totalPlanned+amount > p.totalAmount {
		return fmt.Errorf("installment total %.2f would exceed invoice total %.2f", totalPlanned+amount, p.totalAmount)
	}

	inst, err := NewInstallment(p.invoiceID, len(p.installments)+1, amount, dueDate)
	if err != nil {
		return err
	}

	p.installments = append(p.installments, inst)
	return nil
}

func (p *InstallmentPlan) GetInvoiceID() InvoiceID         { return p.invoiceID }
func (p *InstallmentPlan) GetInstallments() []*Installment { return p.installments }
func (p *InstallmentPlan) GetTotalAmount() float64         { return p.totalAmount }
func (p *InstallmentPlan) GetCurrency() string             { return p.currency }
func (p *InstallmentPlan) GetCreatedAt() time.Time         { return p.createdAt }

func (p *InstallmentPlan) GetTotalPaid() float64 {
	total := float64(0)
	for _, inst := range p.installments {
		total += inst.GetPaidAmount()
	}
	return total
}

func (p *InstallmentPlan) GetTotalRemaining() float64 {
	return p.totalAmount - p.GetTotalPaid()
}

func (p *InstallmentPlan) IsFullyPaid() bool {
	return p.GetTotalPaid() >= p.totalAmount
}

func (p *InstallmentPlan) GetNextPendingInstallment() *Installment {
	for _, inst := range p.installments {
		if inst.GetStatus() == InstallmentStatusPending || inst.GetStatus() == InstallmentStatusDue {
			return inst
		}
	}
	return nil
}

func (p *InstallmentPlan) GetOverdueInstallments() []*Installment {
	var overdue []*Installment
	for _, inst := range p.installments {
		if inst.GetStatus() == InstallmentStatusOverdue {
			overdue = append(overdue, inst)
		}
	}
	return overdue
}
