package query

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/contracts/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type GetContractHandler struct {
	contracts domain.ContractRepository
}

func NewGetContractHandler(contracts domain.ContractRepository) *GetContractHandler {
	return &GetContractHandler{
		contracts: contracts,
	}
}

type GetContractQuery struct {
	ContractID string
}

type ContractDTO struct {
	ID             string           `json:"id"`
	ContractType   string           `json:"contract_type"`
	Status         string           `json:"status"`
	PartyID        string           `json:"party_id"`
	PaymentTerms   PaymentTermsDTO  `json:"payment_terms"`
	DeliveryTerms  DeliveryTermsDTO `json:"delivery_terms"`
	ValidityPeriod DateRangeDTO     `json:"validity_period"`
	Documents      []string         `json:"documents"`
	CreditLimit    float64          `json:"credit_limit"`
	Currency       string           `json:"currency"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	SignedAt       string           `json:"signed_at"`
	TerminatedAt   string           `json:"terminated_at"`
}

type PaymentTermsDTO struct {
	PaymentType  string  `json:"payment_type"`
	CreditDays   int     `json:"credit_days"`
	PenaltyRate  float64 `json:"penalty_rate"`
	DiscountRate float64 `json:"discount_rate"`
	Currency     string  `json:"currency"`
}

type DeliveryTermsDTO struct {
	DeliveryType     string  `json:"delivery_type"`
	EstimatedDays    int     `json:"estimated_days"`
	ShippingCost     float64 `json:"shipping_cost"`
	Insurance        bool    `json:"insurance"`
	ShippingCurrency string  `json:"shipping_currency"`
}

type DateRangeDTO struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (h *GetContractHandler) Handle(ctx context.Context, q GetContractQuery) (ContractDTO, error) {
	contractID, err := domain.ParseContractID(q.ContractID)
	if err != nil {
		return ContractDTO{}, fmt.Errorf("parse contract id: %w", err)
	}

	contract, err := h.contracts.FindByID(ctx, contractID)
	if err != nil {
		return ContractDTO{}, fmt.Errorf("find contract: %w", err)
	}

	return contractToDTO(contract), nil
}

type ListContractsHandler struct {
	contracts domain.ContractRepository
}

func NewListContractsHandler(contracts domain.ContractRepository) *ListContractsHandler {
	return &ListContractsHandler{
		contracts: contracts,
	}
}

type ListContractsQuery struct {
	PartyID      *string
	ContractType *string
	Status       *string
	ActiveOnly   bool
	Cursor       string
	Limit        int
}

func (h *ListContractsHandler) Handle(ctx context.Context, q ListContractsQuery) (pagination.PageResult[ContractDTO], error) {
	var contractType *domain.ContractType
	if q.ContractType != nil {
		ct, err := domain.ParseContractType(*q.ContractType)
		if err != nil {
			return pagination.PageResult[ContractDTO]{}, fmt.Errorf("parse contract type: %w", err)
		}
		contractType = &ct
	}

	var status *domain.ContractStatus
	if q.Status != nil {
		s, err := domain.ParseContractStatus(*q.Status)
		if err != nil {
			return pagination.PageResult[ContractDTO]{}, fmt.Errorf("parse status: %w", err)
		}
		status = &s
	}

	filter := domain.ContractFilter{
		PartyID:      q.PartyID,
		ContractType: contractType,
		Status:       status,
		ActiveOnly:   q.ActiveOnly,
		Cursor:       q.Cursor,
		Limit:        q.Limit,
	}

	page, err := h.contracts.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[ContractDTO]{}, fmt.Errorf("find all contracts: %w", err)
	}

	items := make([]ContractDTO, len(page.Items))
	for i, contract := range page.Items {
		items[i] = contractToDTO(contract)
	}

	return pagination.NewPageResultWithCursor(items, page.NextCursor, page.HasMore, page.Limit), nil
}

func contractToDTO(contract *domain.Contract) ContractDTO {
	dto := ContractDTO{
		ID:           contract.GetID().String(),
		ContractType: contract.GetType().String(),
		Status:       contract.GetStatus().String(),
		PartyID:      contract.GetPartyID(),
		PaymentTerms: PaymentTermsDTO{
			PaymentType:  contract.GetPaymentTerms().PaymentType.String(),
			CreditDays:   contract.GetPaymentTerms().CreditDays,
			PenaltyRate:  contract.GetPaymentTerms().PenaltyRate,
			DiscountRate: contract.GetPaymentTerms().DiscountRate,
			Currency:     contract.GetPaymentTerms().Currency,
		},
		DeliveryTerms: DeliveryTermsDTO{
			DeliveryType:     contract.GetDeliveryTerms().DeliveryType.String(),
			EstimatedDays:    contract.GetDeliveryTerms().EstimatedDays,
			ShippingCost:     contract.GetDeliveryTerms().ShippingCost,
			Insurance:        contract.GetDeliveryTerms().Insurance,
			ShippingCurrency: contract.GetDeliveryTerms().ShippingCurrency,
		},
		ValidityPeriod: DateRangeDTO{
			StartDate: contract.GetValidityPeriod().StartDate.Format("2006-01-02"),
			EndDate:   contract.GetValidityPeriod().EndDate.Format("2006-01-02"),
		},
		Documents:   contract.GetDocuments(),
		CreditLimit: contract.GetCreditLimit(),
		Currency:    contract.GetCurrency(),
		CreatedAt:   contract.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt:   contract.GetUpdatedAt().Format(time.RFC3339),
	}

	if contract.GetSignedAt() != nil {
		dto.SignedAt = contract.GetSignedAt().Format(time.RFC3339)
	}
	if contract.GetTerminatedAt() != nil {
		dto.TerminatedAt = contract.GetTerminatedAt().Format(time.RFC3339)
	}

	return dto
}
