package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Supplier struct {
	id          PartyID
	partyType   PartyType
	name        string
	taxID       string
	contactInfo ContactInfo
	bankAccount BankAccount
	status      PartyStatus
	rating      SupplierRating
	contracts   []string
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

func NewSupplier(name, taxID string, contactInfo ContactInfo) (*Supplier, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now().UTC()
	s := &Supplier{
		id:          NewPartyID(),
		partyType:   PartyTypeSupplier,
		name:        name,
		taxID:       taxID,
		contactInfo: contactInfo,
		bankAccount: BankAccount{},
		status:      PartyStatusActive,
		rating:      SupplierRating{},
		contracts:   make([]string, 0),
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	s.events = append(s.events, SupplierCreated{
		PartyID:   s.id,
		Name:      s.name,
		Email:     s.contactInfo.Email,
		OcurredAt: now,
	})

	return s, nil
}

func (s *Supplier) GetID() PartyID              { return s.id }
func (s *Supplier) GetType() PartyType          { return s.partyType }
func (s *Supplier) GetName() string             { return s.name }
func (s *Supplier) GetContactInfo() ContactInfo { return s.contactInfo }
func (s *Supplier) GetTaxID() string            { return s.taxID }
func (s *Supplier) GetStatus() PartyStatus      { return s.status }
func (s *Supplier) GetBankAccount() BankAccount { return s.bankAccount }
func (s *Supplier) GetRating() SupplierRating   { return s.rating }
func (s *Supplier) GetContracts() []string      { return s.contracts }
func (s *Supplier) GetCreatedAt() time.Time     { return s.createdAt }
func (s *Supplier) GetUpdatedAt() time.Time     { return s.updatedAt }

func (s *Supplier) UpdateContactInfo(contactInfo ContactInfo) error {
	if s.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	s.contactInfo = contactInfo
	s.updatedAt = time.Now().UTC()
	return nil
}

func (s *Supplier) UpdateBankAccount(bankAccount BankAccount) error {
	if s.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	s.bankAccount = bankAccount
	s.updatedAt = time.Now().UTC()
	return nil
}

func (s *Supplier) UpdateRating(rating SupplierRating) {
	s.rating = rating
	s.updatedAt = time.Now().UTC()
}

func (s *Supplier) AssignContract(contractID string) error {
	if s.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	for _, c := range s.contracts {
		if c == contractID {
			return ErrDuplicateContract
		}
	}
	s.contracts = append(s.contracts, contractID)
	s.updatedAt = time.Now().UTC()
	s.events = append(s.events, ContractAssigned{
		PartyID:    s.id,
		ContractID: contractID,
		OcurredAt:  s.updatedAt,
	})
	return nil
}

func (s *Supplier) RemoveContract(contractID string) error {
	found := false
	for i, c := range s.contracts {
		if c == contractID {
			s.contracts = append(s.contracts[:i], s.contracts[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return ErrContractNotFound
	}
	s.updatedAt = time.Now().UTC()
	return nil
}

func (s *Supplier) Activate() error {
	if s.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	oldStatus := s.status
	s.status = PartyStatusActive
	s.updatedAt = time.Now().UTC()
	s.events = append(s.events, PartyStatusChanged{
		PartyID:   s.id,
		PartyType: s.partyType,
		OldStatus: oldStatus,
		NewStatus: s.status,
		OcurredAt: s.updatedAt,
	})
	return nil
}

func (s *Supplier) Deactivate() {
	oldStatus := s.status
	s.status = PartyStatusInactive
	s.updatedAt = time.Now().UTC()
	s.events = append(s.events, PartyStatusChanged{
		PartyID:   s.id,
		PartyType: s.partyType,
		OldStatus: oldStatus,
		NewStatus: s.status,
		OcurredAt: s.updatedAt,
	})
}

func (s *Supplier) Blacklist() {
	oldStatus := s.status
	s.status = PartyStatusBlacklisted
	s.updatedAt = time.Now().UTC()
	s.events = append(s.events, PartyStatusChanged{
		PartyID:   s.id,
		PartyType: s.partyType,
		OldStatus: oldStatus,
		NewStatus: s.status,
		OcurredAt: s.updatedAt,
	})
}

func (s *Supplier) PullEvents() []DomainEvent {
	events := s.events
	s.events = make([]DomainEvent, 0)
	return events
}

func (s *Supplier) RatingToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(s.rating)
	if err != nil {
		return nil, fmt.Errorf("marshal rating: %w", err)
	}
	return json.RawMessage(data), nil
}

func SupplierRatingFromJSON(data json.RawMessage) (SupplierRating, error) {
	var rating SupplierRating
	if len(data) == 0 {
		return SupplierRating{}, nil
	}
	if err := json.Unmarshal(data, &rating); err != nil {
		return SupplierRating{}, fmt.Errorf("unmarshal rating: %w", err)
	}
	return rating, nil
}

func ReconstituteSupplier(
	id PartyID,
	name, taxID string,
	contactInfo ContactInfo,
	bankAccount BankAccount,
	status PartyStatus,
	rating SupplierRating,
	contracts []string,
	createdAt, updatedAt time.Time,
) (*Supplier, error) {
	return &Supplier{
		id:          id,
		partyType:   PartyTypeSupplier,
		name:        name,
		taxID:       taxID,
		contactInfo: contactInfo,
		bankAccount: bankAccount,
		status:      status,
		rating:      rating,
		contracts:   contracts,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}, nil
}
