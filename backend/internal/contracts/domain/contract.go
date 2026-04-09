package domain

import (
	"fmt"
	"time"
)

type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

type Amendment struct {
	amendmentID string
	version     int
	description string
	changes     map[string]string
	amendedAt   time.Time
	amendedBy   string
}

func NewAmendment(amendmentID string, version int, description string, changes map[string]string, amendedBy string) Amendment {
	return Amendment{
		amendmentID: amendmentID,
		version:     version,
		description: description,
		changes:     changes,
		amendedAt:   time.Now().UTC(),
		amendedBy:   amendedBy,
	}
}

func (a Amendment) GetAmendmentID() string        { return a.amendmentID }
func (a Amendment) GetVersion() int               { return a.version }
func (a Amendment) GetDescription() string        { return a.description }
func (a Amendment) GetChanges() map[string]string { return a.changes }
func (a Amendment) GetAmendedAt() time.Time       { return a.amendedAt }
func (a Amendment) GetAmendedBy() string          { return a.amendedBy }

func NewDateRange(start, end time.Time) (DateRange, error) {
	if end.Before(start) {
		return DateRange{}, ErrInvalidDateRange
	}
	return DateRange{StartDate: start, EndDate: end}, nil
}

func (dr DateRange) IsValid() bool {
	return !dr.EndDate.Before(dr.StartDate)
}

func (dr DateRange) Contains(t time.Time) bool {
	return (t.Equal(dr.StartDate) || t.After(dr.StartDate)) &&
		(t.Equal(dr.EndDate) || t.Before(dr.EndDate))
}

func (dr DateRange) IsActive() bool {
	now := time.Now().UTC()
	return dr.Contains(now)
}

type Contract struct {
	id                ContractID
	contractType      ContractType
	status            ContractStatus
	partyID           string
	paymentTerms      PaymentTerms
	deliveryTerms     DeliveryTerms
	validityPeriod    DateRange
	documents         []string
	creditLimit       float64
	currency          string
	autoRenewal       bool
	renewalPeriodDays int
	renewalCount      int
	maxRenewals       int
	amendments        []Amendment
	version           int
	metadata          map[string]interface{}
	createdBy         string
	createdAt         time.Time
	updatedAt         time.Time
	signedAt          *time.Time
	terminatedAt      *time.Time
	renewedAt         *time.Time
	events            []DomainEvent
}

func NewContract(
	contractType ContractType,
	partyID string,
	paymentTerms PaymentTerms,
	deliveryTerms DeliveryTerms,
	startDate, endDate time.Time,
	creditLimit float64,
	currency string,
	createdBy string,
) (*Contract, error) {
	if partyID == "" {
		return nil, fmt.Errorf("party ID is required")
	}

	validityPeriod, err := NewDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if err := paymentTerms.Validate(); err != nil {
		return nil, err
	}

	if err := deliveryTerms.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	c := &Contract{
		id:                NewContractID(),
		contractType:      contractType,
		status:            ContractStatusDraft,
		partyID:           partyID,
		paymentTerms:      paymentTerms,
		deliveryTerms:     deliveryTerms,
		validityPeriod:    validityPeriod,
		documents:         make([]string, 0),
		creditLimit:       creditLimit,
		currency:          currency,
		autoRenewal:       false,
		renewalPeriodDays: 0,
		renewalCount:      0,
		maxRenewals:       0,
		amendments:        make([]Amendment, 0),
		version:           1,
		metadata:          make(map[string]interface{}),
		createdBy:         createdBy,
		createdAt:         now,
		updatedAt:         now,
		events:            make([]DomainEvent, 0),
	}

	c.events = append(c.events, ContractCreated{
		ContractID:   c.id,
		PartyID:      c.partyID,
		ContractType: c.contractType,
		occurredAt:   now,
	})

	return c, nil
}

func (c *Contract) GetID() ContractID                   { return c.id }
func (c *Contract) GetType() ContractType               { return c.contractType }
func (c *Contract) GetStatus() ContractStatus           { return c.status }
func (c *Contract) GetPartyID() string                  { return c.partyID }
func (c *Contract) GetPaymentTerms() PaymentTerms       { return c.paymentTerms }
func (c *Contract) GetDeliveryTerms() DeliveryTerms     { return c.deliveryTerms }
func (c *Contract) GetValidityPeriod() DateRange        { return c.validityPeriod }
func (c *Contract) GetDocuments() []string              { return c.documents }
func (c *Contract) GetCreditLimit() float64             { return c.creditLimit }
func (c *Contract) GetCurrency() string                 { return c.currency }
func (c *Contract) GetAutoRenewal() bool                { return c.autoRenewal }
func (c *Contract) GetRenewalPeriodDays() int           { return c.renewalPeriodDays }
func (c *Contract) GetRenewalCount() int                { return c.renewalCount }
func (c *Contract) GetMaxRenewals() int                 { return c.maxRenewals }
func (c *Contract) GetAmendments() []Amendment          { return c.amendments }
func (c *Contract) GetVersion() int                     { return c.version }
func (c *Contract) GetCreatedBy() string                { return c.createdBy }
func (c *Contract) GetCreatedAt() time.Time             { return c.createdAt }
func (c *Contract) GetUpdatedAt() time.Time             { return c.updatedAt }
func (c *Contract) GetSignedAt() *time.Time             { return c.signedAt }
func (c *Contract) GetTerminatedAt() *time.Time         { return c.terminatedAt }
func (c *Contract) GetRenewedAt() *time.Time            { return c.renewedAt }
func (c *Contract) GetMetadata() map[string]interface{} { return c.metadata }

func (c *Contract) IsActive() bool {
	return c.status == ContractStatusActive && c.validityPeriod.IsActive()
}

func (c *Contract) CanRenew() bool {
	if !c.autoRenewal {
		return false
	}
	if c.maxRenewals > 0 && c.renewalCount >= c.maxRenewals {
		return false
	}
	if c.status != ContractStatusActive {
		return false
	}
	return true
}

func (c *Contract) EnableAutoRenewal(renewalPeriodDays int, maxRenewals int) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}
	if renewalPeriodDays <= 0 {
		return fmt.Errorf("renewal period must be positive")
	}

	c.autoRenewal = true
	c.renewalPeriodDays = renewalPeriodDays
	c.maxRenewals = maxRenewals
	c.updatedAt = time.Now().UTC()

	c.events = append(c.events, ContractAutoRenewalEnabled{
		ContractID:        c.id,
		RenewalPeriodDays: renewalPeriodDays,
		MaxRenewals:       maxRenewals,
		occurredAt:        c.updatedAt,
	})

	return nil
}

func (c *Contract) DisableAutoRenewal() error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	c.autoRenewal = false
	c.updatedAt = time.Now().UTC()

	c.events = append(c.events, ContractAutoRenewalDisabled{
		ContractID: c.id,
		occurredAt: c.updatedAt,
	})

	return nil
}

func (c *Contract) Renew(newEndDate time.Time) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}
	if c.status == ContractStatusDraft {
		return fmt.Errorf("cannot renew draft contract")
	}
	if newEndDate.Before(c.validityPeriod.EndDate) {
		return fmt.Errorf("new end date must be after current end date")
	}

	now := time.Now().UTC()
	oldEndDate := c.validityPeriod.EndDate
	c.validityPeriod.EndDate = newEndDate
	c.renewalCount++
	c.renewedAt = &now
	c.updatedAt = now

	c.events = append(c.events, ContractRenewed{
		ContractID:   c.id,
		OldEndDate:   oldEndDate,
		NewEndDate:   newEndDate,
		RenewalCount: c.renewalCount,
		occurredAt:   now,
	})

	return nil
}

func (c *Contract) Expire() error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	now := time.Now().UTC()
	c.status = ContractStatusExpired
	c.updatedAt = now

	c.events = append(c.events, ContractExpired{
		ContractID: c.id,
		occurredAt: now,
	})

	return nil
}

func (c *Contract) IsExpired() bool {
	return c.status == ContractStatusExpired ||
		(c.status == ContractStatusActive && time.Now().UTC().After(c.validityPeriod.EndDate))
}

func (c *Contract) DaysUntilExpiry() int {
	if c.status != ContractStatusActive {
		return 0
	}
	now := time.Now().UTC()
	if now.After(c.validityPeriod.EndDate) {
		return 0
	}
	return int(c.validityPeriod.EndDate.Sub(now).Hours() / 24)
}

func (c *Contract) CreateAmendment(amendmentID string, description string, changes map[string]string, amendedBy string) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	amendment := NewAmendment(amendmentID, c.version+1, description, changes, amendedBy)
	c.amendments = append(c.amendments, amendment)
	c.version++
	c.updatedAt = time.Now().UTC()

	c.events = append(c.events, ContractAmended{
		ContractID:  c.id,
		AmendmentID: amendmentID,
		Version:     c.version,
		Description: description,
		AmendedBy:   amendedBy,
		occurredAt:  c.updatedAt,
	})

	return nil
}

func (c *Contract) GetAmendment(version int) *Amendment {
	for i := range c.amendments {
		if c.amendments[i].GetVersion() == version {
			return &c.amendments[i]
		}
	}
	return nil
}

func (c *Contract) GetLatestAmendment() *Amendment {
	if len(c.amendments) == 0 {
		return nil
	}
	return &c.amendments[len(c.amendments)-1]
}

func (c *Contract) Activate(signedAt time.Time) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}
	if c.status == ContractStatusExpired {
		return ErrContractExpired
	}

	c.status = ContractStatusActive
	c.signedAt = &signedAt
	c.updatedAt = time.Now().UTC()

	c.events = append(c.events, ContractActivated{
		ContractID:   c.id,
		ContractType: c.contractType,
		occurredAt:   c.updatedAt,
	})

	return nil
}

func (c *Contract) Terminate(reason string) error {
	if c.status == ContractStatusDraft || c.status == ContractStatusTerminated {
		return fmt.Errorf("cannot terminate contract in %s status", c.status)
	}

	now := time.Now().UTC()
	c.status = ContractStatusTerminated
	c.terminatedAt = &now
	c.updatedAt = now

	c.events = append(c.events, ContractTerminated{
		ContractID: c.id,
		Reason:     reason,
		occurredAt: now,
	})

	return nil
}

func (c *Contract) AddDocument(documentID string) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	for _, docID := range c.documents {
		if docID == documentID {
			return fmt.Errorf("document already attached")
		}
	}

	c.documents = append(c.documents, documentID)
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Contract) RemoveDocument(documentID string) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	for i, docID := range c.documents {
		if docID == documentID {
			c.documents = append(c.documents[:i], c.documents[i+1:]...)
			c.updatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("document not found")
}

func (c *Contract) UpdatePaymentTerms(terms PaymentTerms) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	if err := terms.Validate(); err != nil {
		return err
	}

	c.paymentTerms = terms
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Contract) UpdateDeliveryTerms(terms DeliveryTerms) error {
	if c.status == ContractStatusTerminated {
		return ErrContractTerminated
	}

	if err := terms.Validate(); err != nil {
		return err
	}

	c.deliveryTerms = terms
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Contract) PullEvents() []DomainEvent {
	events := c.events
	c.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteContract(
	id ContractID,
	contractType ContractType,
	status ContractStatus,
	partyID string,
	paymentTerms PaymentTerms,
	deliveryTerms DeliveryTerms,
	validityPeriod DateRange,
	documents []string,
	creditLimit float64,
	currency string,
	autoRenewal bool,
	renewalPeriodDays int,
	renewalCount int,
	maxRenewals int,
	amendments []Amendment,
	version int,
	metadata map[string]interface{},
	createdBy string,
	createdAt, updatedAt time.Time,
	signedAt, terminatedAt *time.Time,
	renewedAt *time.Time,
) (*Contract, error) {
	return &Contract{
		id:                id,
		contractType:      contractType,
		status:            status,
		partyID:           partyID,
		paymentTerms:      paymentTerms,
		deliveryTerms:     deliveryTerms,
		validityPeriod:    validityPeriod,
		documents:         documents,
		creditLimit:       creditLimit,
		currency:          currency,
		autoRenewal:       autoRenewal,
		renewalPeriodDays: renewalPeriodDays,
		renewalCount:      renewalCount,
		maxRenewals:       maxRenewals,
		amendments:        amendments,
		version:           version,
		metadata:          metadata,
		createdBy:         createdBy,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		signedAt:          signedAt,
		terminatedAt:      terminatedAt,
		renewedAt:         renewedAt,
		events:            make([]DomainEvent, 0),
	}, nil
}
