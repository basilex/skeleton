package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type GetCustomerHandler struct {
	customers domain.CustomerRepository
}

func NewGetCustomerHandler(customers domain.CustomerRepository) *GetCustomerHandler {
	return &GetCustomerHandler{
		customers: customers,
	}
}

type GetCustomerQuery struct {
	CustomerID string
}

type CustomerDTO struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	TaxID          string            `json:"tax_id"`
	Email          string            `json:"email"`
	Phone          string            `json:"phone"`
	Address        AddressDTO        `json:"address"`
	Website        string            `json:"website"`
	SocialMedia    map[string]string `json:"social_media"`
	BankAccount    BankAccountDTO    `json:"bank_account"`
	Status         string            `json:"status"`
	LoyaltyLevel   string            `json:"loyalty_level"`
	TotalPurchases float64           `json:"total_purchases"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type AddressDTO struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type BankAccountDTO struct {
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	SWIFTCode     string `json:"swift_code"`
	IBAN          string `json:"iban"`
	Currency      string `json:"currency"`
}

func (h *GetCustomerHandler) Handle(ctx context.Context, q GetCustomerQuery) (CustomerDTO, error) {
	partyID, err := domain.ParsePartyID(q.CustomerID)
	if err != nil {
		return CustomerDTO{}, fmt.Errorf("parse customer id: %w", err)
	}

	customer, err := h.customers.FindByID(ctx, partyID)
	if err != nil {
		return CustomerDTO{}, fmt.Errorf("find customer: %w", err)
	}

	contactInfo := customer.GetContactInfo()
	bankAccount := customer.GetBankAccount()

	return CustomerDTO{
		ID:    customer.GetID().String(),
		Name:  customer.GetName(),
		TaxID: customer.GetTaxID(),
		Email: contactInfo.Email,
		Phone: contactInfo.Phone,
		Address: AddressDTO{
			Street:     contactInfo.Address.Street,
			City:       contactInfo.Address.City,
			Region:     contactInfo.Address.Region,
			PostalCode: contactInfo.Address.PostalCode,
			Country:    contactInfo.Address.Country,
		},
		Website:     contactInfo.Website,
		SocialMedia: contactInfo.SocialMedia,
		BankAccount: BankAccountDTO{
			BankName:      bankAccount.BankName,
			AccountName:   bankAccount.AccountName,
			AccountNumber: bankAccount.AccountNumber,
			SWIFTCode:     bankAccount.SWIFTCode,
			IBAN:          bankAccount.IBAN,
			Currency:      bankAccount.Currency,
		},
		Status:         customer.GetStatus().String(),
		LoyaltyLevel:   customer.GetLoyaltyLevel().String(),
		TotalPurchases: customer.GetTotalPurchases().ToFloat64(),
		CreatedAt:      customer.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      customer.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type ListCustomersHandler struct {
	customers domain.CustomerRepository
}

func NewListCustomersHandler(customers domain.CustomerRepository) *ListCustomersHandler {
	return &ListCustomersHandler{
		customers: customers,
	}
}

type ListCustomersQuery struct {
	Status *string
	Search string
	TaxID  string
	Cursor string
	Limit  int
}

func (h *ListCustomersHandler) Handle(ctx context.Context, q ListCustomersQuery) (pagination.PageResult[CustomerDTO], error) {
	var status *domain.PartyStatus
	if q.Status != nil {
		s, err := domain.ParsePartyStatus(*q.Status)
		if err != nil {
			return pagination.PageResult[CustomerDTO]{}, fmt.Errorf("parse status: %w", err)
		}
		status = &s
	}

	filter := domain.PartyFilter{
		Status: status,
		Search: q.Search,
		TaxID:  q.TaxID,
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}

	page, err := h.customers.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[CustomerDTO]{}, fmt.Errorf("find all customers: %w", err)
	}

	items := make([]CustomerDTO, len(page.Items))
	for i, c := range page.Items {
		contactInfo := c.GetContactInfo()
		bankAccount := c.GetBankAccount()

		items[i] = CustomerDTO{
			ID:    c.GetID().String(),
			Name:  c.GetName(),
			TaxID: c.GetTaxID(),
			Email: contactInfo.Email,
			Phone: contactInfo.Phone,
			Address: AddressDTO{
				Street:     contactInfo.Address.Street,
				City:       contactInfo.Address.City,
				Region:     contactInfo.Address.Region,
				PostalCode: contactInfo.Address.PostalCode,
				Country:    contactInfo.Address.Country,
			},
			Website:     contactInfo.Website,
			SocialMedia: contactInfo.SocialMedia,
			BankAccount: BankAccountDTO{
				BankName:      bankAccount.BankName,
				AccountName:   bankAccount.AccountName,
				AccountNumber: bankAccount.AccountNumber,
				SWIFTCode:     bankAccount.SWIFTCode,
				IBAN:          bankAccount.IBAN,
				Currency:      bankAccount.Currency,
			},
			Status:         c.GetStatus().String(),
			LoyaltyLevel:   c.GetLoyaltyLevel().String(),
			TotalPurchases: c.GetTotalPurchases().ToFloat64(),
			CreatedAt:      c.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      c.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	return pagination.NewPageResultWithCursor(items, page.NextCursor, page.HasMore, page.Limit), nil
}
