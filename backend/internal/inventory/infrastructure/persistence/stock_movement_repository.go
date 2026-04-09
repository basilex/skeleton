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

type StockMovementRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewStockMovementRepository(pool *pgxpool.Pool) *StockMovementRepository {
	return &StockMovementRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *StockMovementRepository) Save(ctx context.Context, movement *domain.StockMovement) error {
	var fromWarehouse, toWarehouse *string
	if movement.GetFromWarehouse().String() != "" {
		id := movement.GetFromWarehouse().String()
		fromWarehouse = &id
	}
	if movement.GetToWarehouse().String() != "" {
		id := movement.GetToWarehouse().String()
		toWarehouse = &id
	}

	query, args, err := r.psql.Insert("stock_movements").
		Columns("id", "movement_type", "item_id", "from_warehouse", "to_warehouse",
			"quantity", "reference_id", "reference_type", "notes", "occurred_at", "created_at").
		Values(movement.GetID().String(), movement.GetMovementType().String(), movement.GetItemID(),
			fromWarehouse, toWarehouse, movement.GetQuantity(), movement.GetReferenceID(),
			movement.GetReferenceType(), movement.GetNotes(), movement.GetOccurredAt(), movement.GetCreatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET from_warehouse = EXCLUDED.from_warehouse, " +
			"to_warehouse = EXCLUDED.to_warehouse, quantity = EXCLUDED.quantity, " +
			"reference_id = EXCLUDED.reference_id, reference_type = EXCLUDED.reference_type, " +
			"notes = EXCLUDED.notes").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save stock movement: %w", err)
	}

	return nil
}

func (r *StockMovementRepository) FindByID(ctx context.Context, id domain.StockMovementID) (*domain.StockMovement, error) {
	var dto stockMovementDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, movement_type, item_id, from_warehouse, to_warehouse, 
				quantity, reference_id, reference_type, notes, occurred_at, created_at 
		 FROM stock_movements WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find stock movement by id: %w", err)
	}

	return dto.toDomain()
}

func (r *StockMovementRepository) FindByItem(ctx context.Context, itemID string, filter domain.StockMovementFilter) (pagination.PageResult[*domain.StockMovement], error) {
	q := r.psql.Select("id", "movement_type", "item_id", "from_warehouse", "to_warehouse",
		"quantity", "reference_id", "reference_type", "notes", "occurred_at", "created_at").
		From("stock_movements").
		Where(squirrel.Eq{"item_id": itemID})

	if filter.MovementType != nil {
		q = q.Where(squirrel.Eq{"movement_type": filter.MovementType.String()})
	}
	if filter.ReferenceType != nil {
		q = q.Where(squirrel.Eq{"reference_type": *filter.ReferenceType})
	}
	if filter.StartDate != nil {
		q = q.Where(squirrel.GtOrEq{"occurred_at": *filter.StartDate})
	}
	if filter.EndDate != nil {
		q = q.Where(squirrel.LtOrEq{"occurred_at": *filter.EndDate})
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
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []stockMovementDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("select stock movements: %w", err)
	}

	movements := make([]*domain.StockMovement, 0, len(dtos))
	for _, dto := range dtos {
		movement, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.StockMovement]{}, err
		}
		movements = append(movements, movement)
	}

	return pagination.NewPageResult(movements, limit), nil
}

func (r *StockMovementRepository) FindByWarehouse(ctx context.Context, warehouseID domain.WarehouseID, filter domain.StockMovementFilter) (pagination.PageResult[*domain.StockMovement], error) {
	warehouseStr := warehouseID.String()
	q := r.psql.Select("id", "movement_type", "item_id", "from_warehouse", "to_warehouse",
		"quantity", "reference_id", "reference_type", "notes", "occurred_at", "created_at").
		From("stock_movements").
		Where("(from_warehouse = ? OR to_warehouse = ?)", warehouseStr, warehouseStr)

	if filter.ItemID != nil {
		q = q.Where(squirrel.Eq{"item_id": *filter.ItemID})
	}
	if filter.MovementType != nil {
		q = q.Where(squirrel.Eq{"movement_type": filter.MovementType.String()})
	}
	if filter.ReferenceType != nil {
		q = q.Where(squirrel.Eq{"reference_type": *filter.ReferenceType})
	}
	if filter.StartDate != nil {
		q = q.Where(squirrel.GtOrEq{"occurred_at": *filter.StartDate})
	}
	if filter.EndDate != nil {
		q = q.Where(squirrel.LtOrEq{"occurred_at": *filter.EndDate})
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
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []stockMovementDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("select stock movements: %w", err)
	}

	movements := make([]*domain.StockMovement, 0, len(dtos))
	for _, dto := range dtos {
		movement, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.StockMovement]{}, err
		}
		movements = append(movements, movement)
	}

	return pagination.NewPageResult(movements, limit), nil
}

func (r *StockMovementRepository) FindAll(ctx context.Context, filter domain.StockMovementFilter) (pagination.PageResult[*domain.StockMovement], error) {
	q := r.psql.Select("id", "movement_type", "item_id", "from_warehouse", "to_warehouse",
		"quantity", "reference_id", "reference_type", "notes", "occurred_at", "created_at").
		From("stock_movements")

	if filter.ItemID != nil {
		q = q.Where(squirrel.Eq{"item_id": *filter.ItemID})
	}
	if filter.WarehouseID != nil {
		warehouseStr := filter.WarehouseID.String()
		q = q.Where("(from_warehouse = ? OR to_warehouse = ?)", warehouseStr, warehouseStr)
	}
	if filter.MovementType != nil {
		q = q.Where(squirrel.Eq{"movement_type": filter.MovementType.String()})
	}
	if filter.ReferenceType != nil {
		q = q.Where(squirrel.Eq{"reference_type": *filter.ReferenceType})
	}
	if filter.StartDate != nil {
		q = q.Where(squirrel.GtOrEq{"occurred_at": *filter.StartDate})
	}
	if filter.EndDate != nil {
		q = q.Where(squirrel.LtOrEq{"occurred_at": *filter.EndDate})
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
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []stockMovementDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.StockMovement]{}, fmt.Errorf("select stock movements: %w", err)
	}

	movements := make([]*domain.StockMovement, 0, len(dtos))
	for _, dto := range dtos {
		movement, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.StockMovement]{}, err
		}
		movements = append(movements, movement)
	}

	return pagination.NewPageResult(movements, limit), nil
}
