package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateCustomerHandler struct {
	customers domain.CustomerRepository
	bus       eventbus.Bus
}

func NewCreateCustomerHandler(
	customers domain.CustomerRepository,
	bus eventbus.Bus,
) *CreateCustomerHandler {
	return &CreateCustomerHandler{
		customers: customers,
		bus:       bus,
	}
}

type CreateCustomerCommand struct {
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

type CreateCustomerResult struct {
	CustomerID string
}

func (h *CreateCustomerHandler) Handle(ctx context.Context, cmd CreateCustomerCommand) (CreateCustomerResult, error) {
	address := domain.Address{
		Street:     cmd.Street,
		City:       cmd.City,
		Region:     cmd.Region,
		PostalCode: cmd.PostalCode,
		Country:    cmd.Country,
	}

	contactInfo, err := domain.NewContactInfo(cmd.Email, cmd.Phone, address)
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("validate contact info: %w", err)
	}

	customer, err := domain.NewCustomer(cmd.Name, cmd.TaxID, contactInfo)
	if err != nil {
		return CreateCustomerResult{}, fmt.Errorf("create customer: %w", err)
	}

	if err := h.customers.Save(ctx, customer); err != nil {
		return CreateCustomerResult{}, fmt.Errorf("save customer: %w", err)
	}

	events := customer.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateCustomerResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateCustomerResult{
		CustomerID: customer.GetID().String(),
	}, nil
}
