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

type EmployeeRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewEmployeeRepository(pool *pgxpool.Pool) *EmployeeRepository {
	return &EmployeeRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *EmployeeRepository) Save(ctx context.Context, employee *domain.Employee) error {
	contactInfoJSON, err := employee.GetContactInfo().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal contact info: %w", err)
	}

	bankAccountJSON, err := employee.GetBankAccount().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal bank account: %w", err)
	}

	query, args, err := r.psql.Insert("parties").
		Columns("id", "party_type", "name", "tax_id", "position", "contact_info", "bank_account", "status", "created_at", "updated_at").
		Values(
			employee.GetID().String(),
			domain.PartyTypeEmployee.String(),
			employee.GetName(),
			employee.GetTaxID(),
			employee.GetPosition(),
			contactInfoJSON,
			bankAccountJSON,
			employee.GetStatus().String(),
			employee.GetCreatedAt(),
			employee.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, tax_id = EXCLUDED.tax_id, position = EXCLUDED.position, contact_info = EXCLUDED.contact_info, bank_account = EXCLUDED.bank_account, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save employee: %w", err)
	}
	return nil
}

func (r *EmployeeRepository) FindByID(ctx context.Context, id domain.PartyID) (*domain.Employee, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, position, contact_info, bank_account, status, created_at, updated_at FROM parties WHERE id = $1 AND party_type = $2`,
		id, domain.PartyTypeEmployee.String())
	if err != nil {
		return nil, fmt.Errorf("find employee by id: %w", err)
	}
	return dto.toEmployeeDomain()
}

func (r *EmployeeRepository) FindAll(ctx context.Context, filter domain.PartyFilter) (pagination.PageResult[*domain.Employee], error) {
	q := r.psql.Select("id", "party_type", "name", "tax_id", "position", "contact_info", "bank_account", "status", "created_at", "updated_at").
		From("parties").
		Where(squirrel.Eq{"party_type": domain.PartyTypeEmployee.String()})

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
		return pagination.PageResult[*domain.Employee]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []partyDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Employee]{}, fmt.Errorf("select employees: %w", err)
	}

	employees := make([]*domain.Employee, 0, len(dtos))
	for _, dto := range dtos {
		employee, err := dto.toEmployeeDomain()
		if err != nil {
			return pagination.PageResult[*domain.Employee]{}, err
		}
		employees = append(employees, employee)
	}

	return pagination.NewPageResult(employees, limit), nil
}

func (r *EmployeeRepository) FindByEmail(ctx context.Context, email string) (*domain.Employee, error) {
	var dto partyDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, party_type, name, tax_id, position, contact_info, bank_account, status, created_at, updated_at FROM parties WHERE party_type = $1 AND contact_info->>'email' = $2`,
		domain.PartyTypeEmployee.String(), email)
	if err != nil {
		return nil, fmt.Errorf("find employee by email: %w", err)
	}
	return dto.toEmployeeDomain()
}

func (r *EmployeeRepository) Delete(ctx context.Context, id domain.PartyID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM parties WHERE id = $1 AND party_type = $2`, id, domain.PartyTypeEmployee.String())
	if err != nil {
		return fmt.Errorf("delete employee: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}
