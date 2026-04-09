package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
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
	totalAmount       money.Money
	paidAmount        money.Money
	dueDate           time.Time
	paidAt            *time.Time
	remindedAt        *time.Time
	status            InstallmentStatus
	failedReason      string
	createdAt         time.Time
	updatedAt         time.Time
	events            []DomainEvent
}

func NewInstallment(invoiceID InvoiceID, installmentNumber int, totalAmount money.Money, dueDate time.Time) (*Installment, error) {
	if totalAmount.IsNegative() || totalAmount.IsZero() {
		return nil, errors.New("installment amount must be positive")
	}
	if dueDate.Before(time.Now()) {
		return nil, errors.New("due date must be in the future")
	}

	now := time.Now().UTC()
	zeroMoney, _ := money.New(0, totalAmount.GetCurrency())
	return &Installment{
		id:                NewInstallmentID(),
		invoiceID:         invoiceID,
		installmentNumber: installmentNumber,
		totalAmount:       totalAmount,
		paidAmount:        zeroMoney,
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
func (i *Installment) GetTotalAmount() money.Money  { return i.totalAmount }
func (i *Installment) GetPaidAmount() money.Money   { return i.paidAmount }
func (i *Installment) GetDueDate() time.Time        { return i.dueDate }
func (i *Installment) GetPaidAt() *time.Time        { return i.paidAt }
func (i *Installment) GetRemindedAt() *time.Time    { return i.remindedAt }
func (i *Installment) GetStatus() InstallmentStatus { return i.status }
func (i *Installment) GetFailedReason() string      { return i.failedReason }
func (i *Installment) GetCreatedAt() time.Time      { return i.createdAt }
func (i *Installment) GetUpdatedAt() time.Time      { return i.updatedAt }

func (i *Installment) GetRemainingAmount() money.Money {
	remaining, _ := i.totalAmount.Subtract(i.paidAmount)
	return remaining
}

func (i *Installment) IsPaid() bool {
	return i.status == InstallmentStatusPaid
}

func (i *Installment) IsOverdue() bool {
	return i.status == InstallmentStatusOverdue
}

func (i *Installment) MakePayment(amount money.Money) error {
	if i.status == InstallmentStatusPaid {
		return errors.New("installment already paid")
	}
	if i.status == InstallmentStatusCancelled {
		return errors.New("installment is cancelled")
	}
	if i.status == InstallmentStatusFailed {
		return errors.New("installment has failed")
	}

	newPaidAmount, err := i.paidAmount.Add(amount)
	if err != nil {
		return err
	}

	if newPaidAmount.GreaterThan(i.totalAmount) {
		return fmt.Errorf("payment amount %s exceeds remaining amount %s", amount.String(), i.GetRemainingAmount().String())
	}

	i.paidAmount = newPaidAmount
	i.updatedAt = time.Now().UTC()

	if i.paidAmount.GreaterThanOrEqual(i.totalAmount) {
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
	return fmt.Sprintf("Installment{id=%s, invoice=%s, number=%d, amount=%s, status=%s}",
		i.id, i.invoiceID, i.installmentNumber, i.totalAmount.String(), i.status)
}

func ReconstituteInstallment(
	id InstallmentID,
	invoiceID InvoiceID,
	installmentNumber int,
	totalAmount money.Money,
	paidAmount money.Money,
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
	totalAmount  money.Money
	currency     string
	createdAt    time.Time
}

func NewInstallmentPlan(invoiceID InvoiceID, totalAmount money.Money, currency string) *InstallmentPlan {
	return &InstallmentPlan{
		invoiceID:    invoiceID,
		installments: make([]*Installment, 0),
		totalAmount:  totalAmount,
		currency:     currency,
		createdAt:    time.Now().UTC(),
	}
}

func (p *InstallmentPlan) AddInstallment(amount money.Money, dueDate time.Time) error {
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("installment amount must be positive")
	}

	totalPlanned, _ := money.New(0, p.currency)
	for _, inst := range p.installments {
		totalPlanned, _ = totalPlanned.Add(inst.GetTotalAmount())
	}

	newTotal, err := totalPlanned.Add(amount)
	if err != nil {
		return err
	}

	if newTotal.GreaterThan(p.totalAmount) {
		return fmt.Errorf("installment total %s would exceed invoice total %s", newTotal.String(), p.totalAmount.String())
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
func (p *InstallmentPlan) GetTotalAmount() money.Money     { return p.totalAmount }
func (p *InstallmentPlan) GetCurrency() string             { return p.currency }
func (p *InstallmentPlan) GetCreatedAt() time.Time         { return p.createdAt }

func (p *InstallmentPlan) GetTotalPaid() money.Money {
	total, _ := money.New(0, p.currency)
	for _, inst := range p.installments {
		total, _ = total.Add(inst.GetPaidAmount())
	}
	return total
}

func (p *InstallmentPlan) GetTotalRemaining() money.Money {
	remaining, _ := p.totalAmount.Subtract(p.GetTotalPaid())
	return remaining
}

func (p *InstallmentPlan) IsFullyPaid() bool {
	return p.GetTotalPaid().GreaterThanOrEqual(p.totalAmount)
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
