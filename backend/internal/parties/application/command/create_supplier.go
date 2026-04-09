package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateSupplierHandler struct {
	suppliers domain.SupplierRepository
	bus       eventbus.Bus
}

func NewCreateSupplierHandler(
	suppliers domain.SupplierRepository,
	bus eventbus.Bus,
) *CreateSupplierHandler {
	return &CreateSupplierHandler{
		suppliers: suppliers,
		bus:       bus,
	}
}

type CreateSupplierCommand struct {
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

type CreateSupplierResult struct {
	SupplierID string
}

func (h *CreateSupplierHandler) Handle(ctx context.Context, cmd CreateSupplierCommand) (CreateSupplierResult, error) {
	address := domain.Address{
		Street:     cmd.Street,
		City:       cmd.City,
		Region:     cmd.Region,
		PostalCode: cmd.PostalCode,
		Country:    cmd.Country,
	}

	contactInfo, err := domain.NewContactInfo(cmd.Email, cmd.Phone, address)
	if err != nil {
		return CreateSupplierResult{}, fmt.Errorf("validate contact info: %w", err)
	}

	supplier, err := domain.NewSupplier(cmd.Name, cmd.TaxID, contactInfo)
	if err != nil {
		return CreateSupplierResult{}, fmt.Errorf("create supplier: %w", err)
	}

	if err := h.suppliers.Save(ctx, supplier); err != nil {
		return CreateSupplierResult{}, fmt.Errorf("save supplier: %w", err)
	}

	events := supplier.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateSupplierResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateSupplierResult{
		SupplierID: supplier.GetID().String(),
	}, nil
}
