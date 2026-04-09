package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type GetSupplierHandler struct {
	suppliers domain.SupplierRepository
}

func NewGetSupplierHandler(suppliers domain.SupplierRepository) *GetSupplierHandler {
	return &GetSupplierHandler{
		suppliers: suppliers,
	}
}

type GetSupplierQuery struct {
	SupplierID string
}

type SupplierDTO struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	TaxID            string            `json:"tax_id"`
	Email            string            `json:"email"`
	Phone            string            `json:"phone"`
	Address          AddressDTO        `json:"address"`
	Website          string            `json:"website"`
	SocialMedia      map[string]string `json:"social_media"`
	BankAccount      BankAccountDTO    `json:"bank_account"`
	Status           string            `json:"status"`
	Rating           SupplierRatingDTO `json:"rating"`
	PerformanceLevel string            `json:"performance_level"`
	Contracts        []string          `json:"contracts"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

type SupplierRatingDTO struct {
	QualityScore       float64 `json:"quality_score"`
	DeliveryScore      float64 `json:"delivery_score"`
	CommunicationScore float64 `json:"communication_score"`
	OverallScore       float64 `json:"overall_score"`
	RatingCount        int     `json:"rating_count"`
}

func (h *GetSupplierHandler) Handle(ctx context.Context, q GetSupplierQuery) (SupplierDTO, error) {
	partyID, err := domain.ParsePartyID(q.SupplierID)
	if err != nil {
		return SupplierDTO{}, fmt.Errorf("parse supplier id: %w", err)
	}

	supplier, err := h.suppliers.FindByID(ctx, partyID)
	if err != nil {
		return SupplierDTO{}, fmt.Errorf("find supplier: %w", err)
	}

	contactInfo := supplier.GetContactInfo()
	bankAccount := supplier.GetBankAccount()
	rating := supplier.GetRating()

	return SupplierDTO{
		ID:    supplier.GetID().String(),
		Name:  supplier.GetName(),
		TaxID: supplier.GetTaxID(),
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
		Status: supplier.GetStatus().String(),
		Rating: SupplierRatingDTO{
			QualityScore:       rating.QualityScore,
			DeliveryScore:      rating.DeliveryScore,
			CommunicationScore: rating.CommunicationScore,
			OverallScore:       rating.OverallScore,
			RatingCount:        rating.RatingCount,
		},
		PerformanceLevel: supplier.GetPerformanceLevel().String(),
		Contracts:        supplier.GetContracts(),
		CreatedAt:        supplier.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        supplier.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}

type ListSuppliersHandler struct {
	suppliers domain.SupplierRepository
}

func NewListSuppliersHandler(suppliers domain.SupplierRepository) *ListSuppliersHandler {
	return &ListSuppliersHandler{
		suppliers: suppliers,
	}
}

type ListSuppliersQuery struct {
	Status *string
	Search string
	TaxID  string
	Cursor string
	Limit  int
}

func (h *ListSuppliersHandler) Handle(ctx context.Context, q ListSuppliersQuery) (pagination.PageResult[SupplierDTO], error) {
	var status *domain.PartyStatus
	if q.Status != nil {
		s, err := domain.ParsePartyStatus(*q.Status)
		if err != nil {
			return pagination.PageResult[SupplierDTO]{}, fmt.Errorf("parse status: %w", err)
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

	page, err := h.suppliers.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[SupplierDTO]{}, fmt.Errorf("find all suppliers: %w", err)
	}

	items := make([]SupplierDTO, len(page.Items))
	for i, s := range page.Items {
		contactInfo := s.GetContactInfo()
		bankAccount := s.GetBankAccount()
		rating := s.GetRating()

		items[i] = SupplierDTO{
			ID:    s.GetID().String(),
			Name:  s.GetName(),
			TaxID: s.GetTaxID(),
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
			Status: s.GetStatus().String(),
			Rating: SupplierRatingDTO{
				QualityScore:       rating.QualityScore,
				DeliveryScore:      rating.DeliveryScore,
				CommunicationScore: rating.CommunicationScore,
				OverallScore:       rating.OverallScore,
				RatingCount:        rating.RatingCount,
			},
			PerformanceLevel: s.GetPerformanceLevel().String(),
			Contracts:        s.GetContracts(),
			CreatedAt:        s.GetCreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        s.GetUpdatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	return pagination.NewPageResultWithCursor(items, page.NextCursor, page.HasMore, page.Limit), nil
}
