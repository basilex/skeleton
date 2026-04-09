package domain

import (
	"fmt"
	"time"
)

type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

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
	id             ContractID
	contractType   ContractType
	status         ContractStatus
	partyID        string
	paymentTerms   PaymentTerms
	deliveryTerms  DeliveryTerms
	validityPeriod DateRange
	documents      []string
	creditLimit    float64
	currency       string
	metadata       map[string]interface{}
	createdBy      string
	createdAt      time.Time
	updatedAt      time.Time
	signedAt       *time.Time
	terminatedAt   *time.Time
	events         []DomainEvent
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
		id:             NewContractID(),
		contractType:   contractType,
		status:         ContractStatusDraft,
		partyID:        partyID,
		paymentTerms:   paymentTerms,
		deliveryTerms:  deliveryTerms,
		validityPeriod: validityPeriod,
		documents:      make([]string, 0),
		creditLimit:    creditLimit,
		currency:       currency,
		metadata:       make(map[string]interface{}),
		createdBy:      createdBy,
		createdAt:      now,
		updatedAt:      now,
		events:         make([]DomainEvent, 0),
	}

	c.events = append(c.events, ContractCreated{
		ContractID:   c.id,
		PartyID:      c.partyID,
		ContractType: c.contractType,
		OcurredAt:    now,
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
func (c *Contract) GetCreatedBy() string                { return c.createdBy }
func (c *Contract) GetCreatedAt() time.Time             { return c.createdAt }
func (c *Contract) GetUpdatedAt() time.Time             { return c.updatedAt }
func (c *Contract) GetSignedAt() *time.Time             { return c.signedAt }
func (c *Contract) GetTerminatedAt() *time.Time         { return c.terminatedAt }
func (c *Contract) GetMetadata() map[string]interface{} { return c.metadata }

func (c *Contract) IsActive() bool {
	return c.status == ContractStatusActive && c.validityPeriod.IsActive()
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
		OcurredAt:    c.updatedAt,
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
		OcurredAt:  now,
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
	metadata map[string]interface{},
	createdBy string,
	createdAt, updatedAt time.Time,
	signedAt, terminatedAt *time.Time,
) (*Contract, error) {
	return &Contract{
		id:             id,
		contractType:   contractType,
		status:         status,
		partyID:        partyID,
		paymentTerms:   paymentTerms,
		deliveryTerms:  deliveryTerms,
		validityPeriod: validityPeriod,
		documents:      documents,
		creditLimit:    creditLimit,
		currency:       currency,
		metadata:       metadata,
		createdBy:      createdBy,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		signedAt:       signedAt,
		terminatedAt:   terminatedAt,
		events:         make([]DomainEvent, 0),
	}, nil
}
