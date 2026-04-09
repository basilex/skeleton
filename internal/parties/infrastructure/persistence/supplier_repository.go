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

type SupplierRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewSupplierRepository(pool *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *SupplierRepository) Save(ctx context.Context, supplier *domain.Supplier) error {
	contactInfoJSON, err := supplier.GetContactInfo().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal contact info: %w", err)
	}

	bankAccountJSON, err := supplier.GetBankAccount().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal bank account: %w", err)
	}

	ratingJSON, err := supplier.RatingToJSON()
	if err != nil {
		return fmt.Errorf("marshal rating: %w", err)
	}

	query, args, err := r.psql.Insert("parties").
		Columns("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "rating", "contracts", "created_at", "updated_at").
		Values(
			supplier.GetID().String(),
			domain.PartyTypeSupplier.String(),
			supplier.GetName(),
			supplier.GetTaxID(),
			contactInfoJSON,
			bankAccountJSON,
			supplier.GetStatus().String(),
			ratingJSON,
			supplier.GetContracts(),
			supplier.GetCreatedAt(),
			supplier.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, contact_info = EXCLUDED.contact_info, bank_account = EXCLUDED.bank_account, status = EXCLUDED.status, rating = EXCLUDED.rating, contracts = EXCLUDED.contracts, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save supplier: %w", err)
	}
	return nil
}

func (r *SupplierRepository) FindByID(ctx context.Context, id domain.PartyID) (*domain.Supplier, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, rating, contracts, created_at, updated_at FROM parties WHERE id = $1 AND party_type = $2`,
		id, domain.PartyTypeSupplier.String())
	if err != nil {
		return nil, fmt.Errorf("find supplier by id: %w", err)
	}
	return dto.toSupplierDomain()
}

func (r *SupplierRepository) FindAll(ctx context.Context, filter domain.PartyFilter) (pagination.PageResult[*domain.Supplier], error) {
	q := r.psql.Select("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "rating", "contracts", "created_at", "updated_at").
		From("parties").
		Where(squirrel.Eq{"party_type": domain.PartyTypeSupplier.String()})

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
		return pagination.PageResult[*domain.Supplier]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []partyDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Supplier]{}, fmt.Errorf("select suppliers: %w", err)
	}

	suppliers := make([]*domain.Supplier, 0, len(dtos))
	for _, dto := range dtos {
		supplier, err := dto.toSupplierDomain()
		if err != nil {
			return pagination.PageResult[*domain.Supplier]{}, err
		}
		suppliers = append(suppliers, supplier)
	}

	return pagination.NewPageResult(suppliers, limit), nil
}

func (r *SupplierRepository) FindByEmail(ctx context.Context, email string) (*domain.Supplier, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, rating, contracts, created_at, updated_at FROM parties WHERE party_type = $1 AND contact_info->>'email' = $2`,
		domain.PartyTypeSupplier.String(), email)
	if err != nil {
		return nil, fmt.Errorf("find supplier by email: %w", err)
	}
	return dto.toSupplierDomain()
}

func (r *SupplierRepository) Delete(ctx context.Context, id domain.PartyID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM parties WHERE id = $1 AND party_type = $2`, id, domain.PartyTypeSupplier.String())
	if err != nil {
		return fmt.Errorf("delete supplier: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}
