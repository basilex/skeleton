package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *TransactionRepository) Save(ctx context.Context, transaction *domain.Transaction) error {
	query, args, err := r.psql.Insert("transactions").
		Columns("id", "from_account", "to_account", "amount", "currency", "reference_type", "description", "occurred_at", "posted_at", "posted_by").
		Values(
			transaction.GetID().String(),
			transaction.GetFromAccount().String(),
			transaction.GetToAccount().String(),
			transaction.GetAmount().Amount,
			transaction.GetCurrency().String(),
			transaction.GetReference(),
			transaction.GetDescription(),
			transaction.GetOccurredAt(),
			transaction.GetPostedAt(),
			transaction.GetPostedBy(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET from_account = EXCLUDED.from_account, to_account = EXCLUDED.to_account, amount = EXCLUDED.amount, currency = EXCLUDED.currency, reference_type = EXCLUDED.reference_type, description = EXCLUDED.description, occurred_at = EXCLUDED.occurred_at, posted_at = EXCLUDED.posted_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) FindByID(ctx context.Context, id domain.TransactionID) (*domain.Transaction, error) {
	var dto transactionDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, from_account, to_account, amount, currency, reference_type, description, occurred_at, posted_at, posted_by FROM transactions WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find transaction by id: %w", err)
	}
	return dto.toDomain()
}

func (r *TransactionRepository) FindByAccount(ctx context.Context, accountID domain.AccountID, limit int) ([]*domain.Transaction, error) {
	if limit <= 0 {
		limit = 50
	}

	query, args, err := r.psql.Select("id", "from_account", "to_account", "amount", "currency", "reference_type", "description", "occurred_at", "posted_at", "posted_by").
		From("transactions").
		Where(squirrel.Or{
			squirrel.Eq{"from_account": accountID.String()},
			squirrel.Eq{"to_account": accountID.String()},
		}).
		OrderBy("occurred_at DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dtos []transactionDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return nil, fmt.Errorf("select transactions: %w", err)
	}

	transactions := make([]*domain.Transaction, 0, len(dtos))
	for _, dto := range dtos {
		transaction, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *TransactionRepository) FindByReference(ctx context.Context, reference string) ([]*domain.Transaction, error) {
	query, args, err := r.psql.Select("id", "from_account", "to_account", "amount", "currency", "reference_type", "description", "occurred_at", "posted_at", "posted_by").
		From("transactions").
		Where(squirrel.Eq{"reference_type": reference}).
		OrderBy("occurred_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dtos []transactionDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return nil, fmt.Errorf("select transactions: %w", err)
	}

	transactions := make([]*domain.Transaction, 0, len(dtos))
	for _, dto := range dtos {
		transaction, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
