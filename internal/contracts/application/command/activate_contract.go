package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/contracts/domain"
)

type ActivateContractHandler struct {
	contracts domain.ContractRepository
}

func NewActivateContractHandler(contracts domain.ContractRepository) *ActivateContractHandler {
	return &ActivateContractHandler{
		contracts: contracts,
	}
}

type ActivateContractCommand struct {
	ContractID string
	SignedAt   string
}

func (h *ActivateContractHandler) Handle(ctx context.Context, cmd ActivateContractCommand) error {
	contractID, err := domain.ParseContractID(cmd.ContractID)
	if err != nil {
		return fmt.Errorf("parse contract id: %w", err)
	}

	contract, err := h.contracts.FindByID(ctx, contractID)
	if err != nil {
		return fmt.Errorf("find contract: %w", err)
	}

	signedAt, err := time.Parse(time.RFC3339, cmd.SignedAt)
	if err != nil {
		return fmt.Errorf("parse signed at: %w", err)
	}

	if err := contract.Activate(signedAt); err != nil {
		return fmt.Errorf("activate contract: %w", err)
	}

	if err := h.contracts.Save(ctx, contract); err != nil {
		return fmt.Errorf("save contract: %w", err)
	}

	return nil
}

type TerminateContractHandler struct {
	contracts domain.ContractRepository
}

func NewTerminateContractHandler(contracts domain.ContractRepository) *TerminateContractHandler {
	return &TerminateContractHandler{
		contracts: contracts,
	}
}

type TerminateContractCommand struct {
	ContractID string
	Reason     string
}

func (h *TerminateContractHandler) Handle(ctx context.Context, cmd TerminateContractCommand) error {
	contractID, err := domain.ParseContractID(cmd.ContractID)
	if err != nil {
		return fmt.Errorf("parse contract id: %w", err)
	}

	contract, err := h.contracts.FindByID(ctx, contractID)
	if err != nil {
		return fmt.Errorf("find contract: %w", err)
	}

	if err := contract.Terminate(cmd.Reason); err != nil {
		return fmt.Errorf("terminate contract: %w", err)
	}

	if err := h.contracts.Save(ctx, contract); err != nil {
		return fmt.Errorf("save contract: %w", err)
	}

	return nil
}
