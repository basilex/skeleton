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

type CustomerRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *CustomerRepository) Save(ctx context.Context, customer *domain.Customer) error {
	contactInfoJSON, err := customer.GetContactInfo().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal contact info: %w", err)
	}

	bankAccountJSON, err := customer.GetBankAccount().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal bank account: %w", err)
	}

	query, args, err := r.psql.Insert("parties").
		Columns("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "loyalty_level", "total_purchases", "credit_limit", "current_credit", "created_at", "updated_at").
		Values(
			customer.GetID().String(),
			domain.PartyTypeCustomer.String(),
			customer.GetName(),
			customer.GetTaxID(),
			contactInfoJSON,
			bankAccountJSON,
			customer.GetStatus().String(),
			customer.GetLoyaltyLevel().String(),
			customer.GetTotalPurchases().GetAmount(),
			customer.GetCreditLimit().GetAmount(),
			customer.GetCurrentCredit().GetAmount(),
			customer.GetCreatedAt(),
			customer.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, contact_info = EXCLUDED.contact_info, bank_account = EXCLUDED.bank_account, status = EXCLUDED.status, loyalty_level = EXCLUDED.loyalty_level, total_purchases = EXCLUDED.total_purchases, credit_limit = EXCLUDED.credit_limit, current_credit = EXCLUDED.current_credit, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) FindByID(ctx context.Context, id domain.PartyID) (*domain.Customer, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, loyalty_level, total_purchases, created_at, updated_at FROM parties WHERE id = $1 AND party_type = $2`,
		id, domain.PartyTypeCustomer.String())
	if err != nil {
		return nil, fmt.Errorf("find customer by id: %w", err)
	}
	return dto.toCustomerDomain()
}

func (r *CustomerRepository) FindAll(ctx context.Context, filter domain.PartyFilter) (pagination.PageResult[*domain.Customer], error) {
	q := r.psql.Select("id", "party_type", "name", "tax_id", "contact_info", "bank_account", "status", "loyalty_level", "total_purchases", "created_at", "updated_at").
		From("parties").
		Where(squirrel.Eq{"party_type": domain.PartyTypeCustomer.String()})

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
		return pagination.PageResult[*domain.Customer]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []partyDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Customer]{}, fmt.Errorf("select customers: %w", err)
	}

	customers := make([]*domain.Customer, 0, len(dtos))
	for _, dto := range dtos {
		customer, err := dto.toCustomerDomain()
		if err != nil {
			return pagination.PageResult[*domain.Customer]{}, err
		}
		customers = append(customers, customer)
	}

	return pagination.NewPageResult(customers, limit), nil
}

func (r *CustomerRepository) FindByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, contact_info, bank_account, status, loyalty_level, total_purchases, created_at, updated_at FROM parties WHERE party_type = $1 AND contact_info->>'email' = $2`,
		domain.PartyTypeCustomer.String(), email)
	if err != nil {
		return nil, fmt.Errorf("find customer by email: %w", err)
	}
	return dto.toCustomerDomain()
}

func (r *CustomerRepository) Delete(ctx context.Context, id domain.PartyID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM parties WHERE id = $1 AND party_type = $2`, id, domain.PartyTypeCustomer.String())
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCustomerNotFound
	}
	return nil
}
