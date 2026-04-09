package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/basilex/skeleton/pkg/transaction"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// exec executes a query using transaction from context if available, otherwise uses pool
func (r *AccountRepository) exec(ctx context.Context, query string, args ...interface{}) error {
	if tx, ok := transaction.FromContext(ctx); ok {
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// queryRow queries a single row using transaction from context if available, otherwise uses pool
func (r *AccountRepository) queryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	if tx, ok := transaction.FromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return r.pool.QueryRow(ctx, query, args...)
}

// query queries multiple rows using transaction from context if available, otherwise uses pool
func (r *AccountRepository) query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if tx, ok := transaction.FromContext(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}

func (r *AccountRepository) Save(ctx context.Context, account *domain.Account) error {
	var parentID *string
	if account.GetParentID() != nil {
		pid := account.GetParentID().String()
		parentID = &pid
	}

	query, args, err := r.psql.Insert("accounts").
		Columns("id", "code", "name", "account_type", "currency", "balance", "parent_id", "is_active", "created_at", "updated_at").
		Values(
			account.GetID().String(),
			account.GetCode(),
			account.GetName(),
			account.GetType().String(),
			account.GetCurrency().String(),
			account.GetBalance().Amount,
			parentID,
			account.IsActive(),
			account.GetCreatedAt(),
			account.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET code = EXCLUDED.code, name = EXCLUDED.name, account_type = EXCLUDED.account_type, currency = EXCLUDED.currency, balance = EXCLUDED.balance, parent_id = EXCLUDED.parent_id, is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	if err := r.exec(ctx, query, args...); err != nil {
		return fmt.Errorf("save account: %w", err)
	}
	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	var dto accountDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, code, name, account_type, currency, balance, parent_id, is_active, created_at, updated_at FROM accounts WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find account by id: %w", err)
	}
	return dto.toDomain()
}

func (r *AccountRepository) FindByCode(ctx context.Context, code string) (*domain.Account, error) {
	var dto accountDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, code, name, account_type, currency, balance, parent_id, is_active, created_at, updated_at FROM accounts WHERE code = $1`,
		code)
	if err != nil {
		return nil, fmt.Errorf("find account by code: %w", err)
	}
	return dto.toDomain()
}

func (r *AccountRepository) FindAll(ctx context.Context, filter domain.AccountFilter) (pagination.PageResult[*domain.Account], error) {
	q := r.psql.Select("id", "code", "name", "account_type", "currency", "balance", "parent_id", "is_active", "created_at", "updated_at").
		From("accounts")

	if filter.AccountType != nil {
		q = q.Where(squirrel.Eq{"account_type": filter.AccountType.String()})
	}
	if filter.IsActive != nil {
		q = q.Where(squirrel.Eq{"is_active": *filter.IsActive})
	}
	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"name": "%" + filter.Search + "%"})
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
		return pagination.PageResult[*domain.Account]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []accountDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Account]{}, fmt.Errorf("select accounts: %w", err)
	}

	accounts := make([]*domain.Account, 0, len(dtos))
	for _, dto := range dtos {
		account, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Account]{}, err
		}
		accounts = append(accounts, account)
	}

	return pagination.NewPageResult(accounts, limit), nil
}

func (r *AccountRepository) Delete(ctx context.Context, id domain.AccountID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}
