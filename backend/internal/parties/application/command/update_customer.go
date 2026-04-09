package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/parties/domain"
)

type UpdateCustomerHandler struct {
	customers domain.CustomerRepository
}

func NewUpdateCustomerHandler(customers domain.CustomerRepository) *UpdateCustomerHandler {
	return &UpdateCustomerHandler{
		customers: customers,
	}
}

type UpdateCustomerCommand struct {
	CustomerID  string
	Name        string
	TaxID       string
	Email       string
	Phone       string
	Street      string
	City        string
	Region      string
	PostalCode  string
	Country     string
	Website     string
	SocialMedia map[string]string
}

func (h *UpdateCustomerHandler) Handle(ctx context.Context, cmd UpdateCustomerCommand) error {
	partyID, err := domain.ParsePartyID(cmd.CustomerID)
	if err != nil {
		return fmt.Errorf("parse customer id: %w", err)
	}

	customer, err := h.customers.FindByID(ctx, partyID)
	if err != nil {
		return fmt.Errorf("find customer: %w", err)
	}

	address := domain.Address{
		Street:     cmd.Street,
		City:       cmd.City,
		Region:     cmd.Region,
		PostalCode: cmd.PostalCode,
		Country:    cmd.Country,
	}

	contactInfo, err := domain.NewContactInfo(cmd.Email, cmd.Phone, address)
	if err != nil {
		return fmt.Errorf("validate contact info: %w", err)
	}

	if err := customer.UpdateContactInfo(contactInfo); err != nil {
		return fmt.Errorf("update contact info: %w", err)
	}

	if err := h.customers.Save(ctx, customer); err != nil {
		return fmt.Errorf("save customer: %w", err)
	}

	return nil
}
