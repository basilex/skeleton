package persistence

import (
	"encoding/json"
	"time"

	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/money"
)

type partyDTO struct {
	ID               string          `db:"id"`
	PartyType        string          `db:"party_type"`
	Name             string          `db:"name"`
	TaxID            string          `db:"tax_id"`
	ContactInfo      json.RawMessage `db:"contact_info"`
	BankAccount      json.RawMessage `db:"bank_account"`
	Status           string          `db:"status"`
	LoyaltyLevel     string          `db:"loyalty_level"`
	TotalPurchases   int64           `db:"total_purchases"`
	CreditLimit      int64           `db:"credit_limit"`
	CurrentCredit    int64           `db:"current_credit"`
	Rating           json.RawMessage `db:"rating"`
	PerformanceLevel string          `db:"performance_level"`
	Contracts        []string        `db:"contracts"`
	Position         string          `db:"position"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}

func (dto *partyDTO) toCustomerDomain() (*domain.Customer, error) {
	contactInfo, err := domain.ContactInfoFromJSON(dto.ContactInfo)
	if err != nil {
		return nil, err
	}

	bankAccount, err := domain.BankAccountFromJSON(dto.BankAccount)
	if err != nil {
		return nil, err
	}

	partyID, err := domain.ParsePartyID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParsePartyStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	loyaltyLevel := domain.LoyaltyLevel(dto.LoyaltyLevel)
	if loyaltyLevel == "" {
		loyaltyLevel = domain.LoyaltyLevelBronze
	}

	totalPurchases, _ := money.New(dto.TotalPurchases, "USD")
	creditLimit, _ := money.New(dto.CreditLimit, "USD")
	currentCredit, _ := money.New(dto.CurrentCredit, "USD")

	return domain.ReconstituteCustomer(
		partyID,
		dto.Name,
		dto.TaxID,
		contactInfo,
		bankAccount,
		status,
		loyaltyLevel,
		totalPurchases,
		creditLimit,
		currentCredit,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func (dto *partyDTO) toSupplierDomain() (*domain.Supplier, error) {
	contactInfo, err := domain.ContactInfoFromJSON(dto.ContactInfo)
	if err != nil {
		return nil, err
	}

	bankAccount, err := domain.BankAccountFromJSON(dto.BankAccount)
	if err != nil {
		return nil, err
	}

	partyID, err := domain.ParsePartyID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParsePartyStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	rating, err := domain.SupplierRatingFromJSON(dto.Rating)
	if err != nil {
		return nil, err
	}

	performanceLevel := domain.PerformanceLevel(dto.PerformanceLevel)
	if performanceLevel == "" {
		performanceLevel = domain.PerformanceLevelAverage
	}

	return domain.ReconstituteSupplier(
		partyID,
		dto.Name,
		dto.TaxID,
		contactInfo,
		bankAccount,
		status,
		rating,
		performanceLevel,
		dto.Contracts,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func (dto *partyDTO) toPartnerDomain() (*domain.Partner, error) {
	contactInfo, err := domain.ContactInfoFromJSON(dto.ContactInfo)
	if err != nil {
		return nil, err
	}

	bankAccount, err := domain.BankAccountFromJSON(dto.BankAccount)
	if err != nil {
		return nil, err
	}

	partyID, err := domain.ParsePartyID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParsePartyStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	return domain.ReconstitutePartner(
		partyID,
		dto.Name,
		dto.TaxID,
		contactInfo,
		bankAccount,
		status,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func (dto *partyDTO) toEmployeeDomain() (*domain.Employee, error) {
	contactInfo, err := domain.ContactInfoFromJSON(dto.ContactInfo)
	if err != nil {
		return nil, err
	}

	bankAccount, err := domain.BankAccountFromJSON(dto.BankAccount)
	if err != nil {
		return nil, err
	}

	partyID, err := domain.ParsePartyID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParsePartyStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	return domain.ReconstituteEmployee(
		partyID,
		dto.Name,
		dto.TaxID,
		dto.Position,
		contactInfo,
		bankAccount,
		status,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}
