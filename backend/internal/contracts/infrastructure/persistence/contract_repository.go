package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/contracts/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContractRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewContractRepository(pool *pgxpool.Pool) *ContractRepository {
	return &ContractRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ContractRepository) Save(ctx context.Context, contract *domain.Contract) error {
	paymentTermsJSON, err := contract.GetPaymentTerms().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal payment terms: %w", err)
	}

	deliveryTermsJSON, err := contract.GetDeliveryTerms().ToJSON()
	if err != nil {
		return fmt.Errorf("marshal delivery terms: %w", err)
	}

	validityPeriod := fmt.Sprintf("[%s,%s)",
		contract.GetValidityPeriod().StartDate.Format("2006-01-02"),
		contract.GetValidityPeriod().EndDate.Format("2006-01-02"),
	)

	metadataJSON, err := jsonMarshal(contract.GetMetadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query, args, err := r.psql.Insert("contracts").
		Columns("id", "contract_type", "status", "party_id", "payment_terms", "delivery_terms",
			"validity_period", "documents", "credit_limit", "currency", "metadata", "created_by",
			"created_at", "updated_at", "signed_at", "terminated_at").
		Values(
			contract.GetID().String(),
			contract.GetType().String(),
			contract.GetStatus().String(),
			contract.GetPartyID(),
			paymentTermsJSON,
			deliveryTermsJSON,
			validityPeriod,
			contract.GetDocuments(),
			contract.GetCreditLimit(),
			contract.GetCurrency(),
			metadataJSON,
			contract.GetCreatedBy(),
			contract.GetCreatedAt(),
			contract.GetUpdatedAt(),
			contract.GetSignedAt(),
			contract.GetTerminatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET " +
			"contract_type = EXCLUDED.contract_type, " +
			"status = EXCLUDED.status, " +
			"party_id = EXCLUDED.party_id, " +
			"payment_terms = EXCLUDED.payment_terms, " +
			"delivery_terms = EXCLUDED.delivery_terms, " +
			"validity_period = EXCLUDED.validity_period, " +
			"documents = EXCLUDED.documents, " +
			"credit_limit = EXCLUDED.credit_limit, " +
			"currency = EXCLUDED.currency, " +
			"metadata = EXCLUDED.metadata, " +
			"updated_at = EXCLUDED.updated_at, " +
			"signed_at = EXCLUDED.signed_at, " +
			"terminated_at = EXCLUDED.terminated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save contract: %w", err)
	}
	return nil
}

func (r *ContractRepository) FindByID(ctx context.Context, id domain.ContractID) (*domain.Contract, error) {
	var dto contractDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, contract_type, status, party_id, payment_terms, delivery_terms,
			validity_period::text, documents, credit_limit, currency, metadata,
			created_by, created_at, updated_at, signed_at, terminated_at
		 FROM contracts WHERE id = $1`,
		id)
	if err != nil {
		return nil, fmt.Errorf("find contract by id: %w", err)
	}
	return dto.toDomain()
}

func (r *ContractRepository) FindByPartyID(ctx context.Context, partyID string, filter domain.ContractFilter) (pagination.PageResult[*domain.Contract], error) {
	q := r.psql.Select("id", "contract_type", "status", "party_id", "payment_terms", "delivery_terms",
		"validity_period::text", "documents", "credit_limit", "currency", "metadata",
		"created_by", "created_at", "updated_at", "signed_at", "terminated_at").
		From("contracts").
		Where(squirrel.Eq{"party_id": partyID})

	if filter.ContractType != nil {
		q = q.Where(squirrel.Eq{"contract_type": filter.ContractType.String()})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.ActiveOnly {
		q = q.Where("status = 'active' AND validity_period @> CURRENT_DATE")
	}

	q = q.OrderBy("created_at DESC")

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}
	if filter.Cursor != "" {
		q = q.Where(squirrel.Lt{"created_at": filter.Cursor})
	}
	q = q.Limit(uint64(limit + 1))

	query, args, err := q.ToSql()
	if err != nil {
		return pagination.PageResult[*domain.Contract]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []contractDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Contract]{}, fmt.Errorf("select contracts: %w", err)
	}

	contracts := make([]*domain.Contract, 0, len(dtos))
	for _, dto := range dtos {
		contract, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Contract]{}, err
		}
		contracts = append(contracts, contract)
	}

	return pagination.NewPageResult(contracts, limit), nil
}

func (r *ContractRepository) FindAll(ctx context.Context, filter domain.ContractFilter) (pagination.PageResult[*domain.Contract], error) {
	q := r.psql.Select("id", "contract_type", "status", "party_id", "payment_terms", "delivery_terms",
		"validity_period::text", "documents", "credit_limit", "currency", "metadata",
		"created_by", "created_at", "updated_at", "signed_at", "terminated_at").
		From("contracts")

	if filter.PartyID != nil {
		q = q.Where(squirrel.Eq{"party_id": *filter.PartyID})
	}
	if filter.ContractType != nil {
		q = q.Where(squirrel.Eq{"contract_type": filter.ContractType.String()})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.ActiveOnly {
		q = q.Where("status = 'active' AND validity_period @> CURRENT_DATE")
	}

	q = q.OrderBy("created_at DESC")

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}
	if filter.Cursor != "" {
		q = q.Where(squirrel.Lt{"created_at": filter.Cursor})
	}
	q = q.Limit(uint64(limit + 1))

	query, args, err := q.ToSql()
	if err != nil {
		return pagination.PageResult[*domain.Contract]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []contractDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Contract]{}, fmt.Errorf("select contracts: %w", err)
	}

	contracts := make([]*domain.Contract, 0, len(dtos))
	for _, dto := range dtos {
		contract, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Contract]{}, err
		}
		contracts = append(contracts, contract)
	}

	return pagination.NewPageResult(contracts, limit), nil
}

func (r *ContractRepository) Delete(ctx context.Context, id domain.ContractID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrContractNotFound
	}
	return nil
}

func jsonMarshal(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}
