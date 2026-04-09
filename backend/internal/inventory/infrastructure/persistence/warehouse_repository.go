package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WarehouseRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewWarehouseRepository(pool *pgxpool.Pool) *WarehouseRepository {
	return &WarehouseRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *WarehouseRepository) Save(ctx context.Context, warehouse *domain.Warehouse) error {
	query, args, err := r.psql.Insert("warehouses").
		Columns("id", "name", "code", "location", "capacity", "status", "metadata", "created_at", "updated_at").
		Values(warehouse.GetID().String(), warehouse.GetName(), warehouse.GetCode(),
			warehouse.GetLocation(), warehouse.GetCapacity(), warehouse.GetStatus().String(),
			warehouse.GetMetadata(), warehouse.GetCreatedAt(), warehouse.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, code = EXCLUDED.code, " +
			"location = EXCLUDED.location, capacity = EXCLUDED.capacity, status = EXCLUDED.status, " +
			"metadata = EXCLUDED.metadata, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save warehouse: %w", err)
	}

	return nil
}

func (r *WarehouseRepository) FindByID(ctx context.Context, id domain.WarehouseID) (*domain.Warehouse, error) {
	var dto warehouseDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, name, code, location, capacity, status, metadata, created_at, updated_at 
		 FROM warehouses WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find warehouse by id: %w", err)
	}

	return dto.toDomain()
}

func (r *WarehouseRepository) FindByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	var dto warehouseDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, name, code, location, capacity, status, metadata, created_at, updated_at 
		 FROM warehouses WHERE code = $1`, code)
	if err != nil {
		return nil, fmt.Errorf("find warehouse by code: %w", err)
	}

	return dto.toDomain()
}

func (r *WarehouseRepository) FindAll(ctx context.Context, filter domain.WarehouseFilter) (pagination.PageResult[*domain.Warehouse], error) {
	q := r.psql.Select("id", "name", "code", "location", "capacity", "status", "metadata", "created_at", "updated_at").
		From("warehouses")

	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.Code != nil {
		q = q.Where(squirrel.Eq{"code": *filter.Code})
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
		return pagination.PageResult[*domain.Warehouse]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []warehouseDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Warehouse]{}, fmt.Errorf("select warehouses: %w", err)
	}

	warehouses := make([]*domain.Warehouse, 0, len(dtos))
	for _, dto := range dtos {
		warehouse, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Warehouse]{}, err
		}
		warehouses = append(warehouses, warehouse)
	}

	return pagination.NewPageResult(warehouses, limit), nil
}

func (r *WarehouseRepository) Delete(ctx context.Context, id domain.WarehouseID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM warehouses WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete warehouse: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrWarehouseNotFound
	}

	return nil
}
