package persistence

import (
	"encoding/json"
	"time"

	"github.com/basilex/skeleton/internal/contracts/domain"
)

type contractDTO struct {
	ID             string          `db:"id"`
	ContractType   string          `db:"contract_type"`
	Status         string          `db:"status"`
	PartyID        string          `db:"party_id"`
	PaymentTerms   json.RawMessage `db:"payment_terms"`
	DeliveryTerms  json.RawMessage `db:"delivery_terms"`
	ValidityPeriod string          `db:"validity_period"`
	Documents      []string        `db:"documents"`
	CreditLimit    float64         `db:"credit_limit"`
	Currency       string          `db:"currency"`
	Metadata       json.RawMessage `db:"metadata"`
	CreatedBy      string          `db:"created_by"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
	SignedAt       *time.Time      `db:"signed_at"`
	TerminatedAt   *time.Time      `db:"terminated_at"`
}

func (dto *contractDTO) toDomain() (*domain.Contract, error) {
	contractID, err := domain.ParseContractID(dto.ID)
	if err != nil {
		return nil, err
	}

	contractType, err := domain.ParseContractType(dto.ContractType)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParseContractStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	paymentTerms, err := domain.PaymentTermsFromJSON(dto.PaymentTerms)
	if err != nil {
		return nil, err
	}

	deliveryTerms, err := domain.DeliveryTermsFromJSON(dto.DeliveryTerms)
	if err != nil {
		return nil, err
	}

	// Parse validity period from PostgreSQL DATERANGE format
	// Format: [2024-01-01,2024-12-31)
	var startDate, endDate time.Time
	if len(dto.ValidityPeriod) > 2 {
		// Remove brackets and parse
		period := dto.ValidityPeriod[1 : len(dto.ValidityPeriod)-1]
		parts := splitDateRange(period)
		if len(parts) == 2 {
			startDate, err = time.Parse("2006-01-02", parts[0])
			if err != nil {
				startDate = time.Now()
			}
			endDate, err = time.Parse("2006-01-02", parts[1])
			if err != nil {
				endDate = time.Now().AddDate(1, 0, 0)
			}
		}
	}

	validityPeriod, err := domain.NewDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if len(dto.Metadata) > 0 {
		if err := json.Unmarshal(dto.Metadata, &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	}

	return domain.ReconstituteContract(
		contractID,
		contractType,
		status,
		dto.PartyID,
		paymentTerms,
		deliveryTerms,
		validityPeriod,
		dto.Documents,
		dto.CreditLimit,
		dto.Currency,
		false,
		0,
		0,
		0,
		make([]domain.Amendment, 0),
		1,
		metadata,
		dto.CreatedBy,
		dto.CreatedAt,
		dto.UpdatedAt,
		dto.SignedAt,
		dto.TerminatedAt,
		nil,
	)
}

func splitDateRange(s string) []string {
	// Split by comma, handling PostgreSQL daterange format
	for i, c := range s {
		if c == ',' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
