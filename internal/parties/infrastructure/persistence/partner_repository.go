package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/parties/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartnerRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewPartnerRepository(pool *pgxpool.Pool) *PartnerRepository {
	return &PartnerRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *PartnerRepository) Save(ctx context.Context, partner *domain.Partner) error {
	contactInfoJSON, err := partner.GetContactInfo().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal contact info: %w", err)
	}

	bankAccountJSON, err := partner.GetBankAccount().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal bank account: %w", err)
	}

	query, args, err := r.psql.Insert("parties").
		Columns("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "created_at", "updated_at").
		Values(
			partner.GetID().String(),
			domain.PartyTypePartner.String(),
			partner.GetName(),
			partner.GetTaxID(),
			contactInfoJSON,
			bankAccountJSON,
			partner.GetStatus().String(),
			partner.GetCreatedAt(),
			partner.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, contact_info = EXCLUDED.contact_info, bank_account = EXCLUDED.bank_account, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save partner: %w", err)
	}
	return nil
}

func (r *PartnerRepository) FindByID(ctx context.Context, id domain.PartyID) (*domain.Partner, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, created_at, updated_at FROM parties WHERE id = $1 AND party_type = $2`,
		id, domain.PartyTypePartner.String())
	if err != nil {
		return nil, fmt.Errorf("find partner by id: %w", err)
	}
	return dto.toPartnerDomain()
}

func (r *PartnerRepository) FindAll(ctx context.Context, filter domain.PartyFilter) (pagination.PageResult[*domain.Partner], error) {
	q := r.psql.Select("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "created_at", "updated_at").
		From("parties").
		Where(squirrel.Eq{"party_type": domain.PartyTypePartner.String()})

	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"name": "%" + filter.Search + "%"})
	}
	if filter.TaxID != "" {
		q = q.Where(squirrel.Eq{"tax_id": filter.TaxID})
	}
	if filter.Cursor != "" {
		q = q.Where(squirrel.Lt{"id": filter.Cursor})
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	q = q.OrderBy("id DESC").Limit(uint64(limit + 1))

	query, args, err := q.ToSql()
	if err != nil {
		return pagination.PageResult[*domain.Partner]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []partyDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Partner]{}, fmt.Errorf("select partners: %w", err)
	}

	partners := make([]*domain.Partner, 0, len(dtos))
	for _, dto := range dtos {
		partner, err := dto.toPartnerDomain()
		if err != nil {
			return pagination.PageResult[*domain.Partner]{}, err
		}
		partners = append(partners, partner)
	}

	return pagination.NewPageResult(partners, limit), nil
}

func (r *PartnerRepository) FindByEmail(ctx context.Context, email string) (*domain.Partner, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, created_at, updated_at FROM parties WHERE party_type = $1 AND contact_info->>'email' = $2`,
		domain.PartyTypePartner.String(), email)
	if err != nil {
		return nil, fmt.Errorf("find partner by email: %w", err)
	}
	return dto.toPartnerDomain()
}

func (r *PartnerRepository) Delete(ctx context.Context, id domain.PartyID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM parties WHERE id = $1 AND party_type = $2`, id, domain.PartyTypePartner.String())
	if err != nil {
		return fmt.Errorf("delete partner: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPartnerNotFound
	}
	return nil
}
