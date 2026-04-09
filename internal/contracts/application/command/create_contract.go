package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/contracts/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateContractHandler struct {
	contracts domain.ContractRepository
	bus       eventbus.Bus
}

func NewCreateContractHandler(
	contracts domain.ContractRepository,
	bus eventbus.Bus,
) *CreateContractHandler {
	return &CreateContractHandler{
		contracts: contracts,
		bus:       bus,
	}
}

type CreateContractCommand struct {
	ContractType  string
	PartyID       string
	PaymentType   string
	CreditDays    int
	Currency      string
	DeliveryType  string
	EstimatedDays int
	StartDate     string
	EndDate       string
	CreditLimit   float64
	CreatedBy     string
}

type CreateContractResult struct {
	ContractID string
}

func (h *CreateContractHandler) Handle(ctx context.Context, cmd CreateContractCommand) (CreateContractResult, error) {
	contractType, err := domain.ParseContractType(cmd.ContractType)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("parse contract type: %w", err)
	}

	paymentType, err := domain.ParsePaymentType(cmd.PaymentType)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("parse payment type: %w", err)
	}

	paymentTerms, err := domain.NewPaymentTerms(paymentType, cmd.CreditDays, cmd.Currency)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("create payment terms: %w", err)
	}

	deliveryType, err := domain.ParseDeliveryType(cmd.DeliveryType)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("parse delivery type: %w", err)
	}

	deliveryTerms := domain.NewDeliveryTerms(deliveryType, cmd.EstimatedDays)

	startDate, err := time.Parse("2006-01-02", cmd.StartDate)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("parse start date: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", cmd.EndDate)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("parse end date: %w", err)
	}

	contract, err := domain.NewContract(
		contractType,
		cmd.PartyID,
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		cmd.CreditLimit,
		cmd.Currency,
		cmd.CreatedBy,
	)
	if err != nil {
		return CreateContractResult{}, fmt.Errorf("create contract: %w", err)
	}

	if err := h.contracts.Save(ctx, contract); err != nil {
		return CreateContractResult{}, fmt.Errorf("save contract: %w", err)
	}

	events := contract.PullEvents()
	for _, e := range events {
		if err := h.bus.Publish(ctx, e); err != nil {
			return CreateContractResult{}, fmt.Errorf("publish event: %w", err)
		}
	}

	return CreateContractResult{
		ContractID: contract.GetID().String(),
	}, nil
}
