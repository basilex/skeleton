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

type StockRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewStockRepository(pool *pgxpool.Pool) *StockRepository {
	return &StockRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *StockRepository) Save(ctx context.Context, stock *domain.Stock) error {
	var lastMovementID *string
	if stock.GetLastMovementID().String() != "" {
		id := stock.GetLastMovementID().String()
		lastMovementID = &id
	}

	query, args, err := r.psql.Insert("stock").
		Columns("id", "item_id", "warehouse_id", "quantity", "reserved_qty", "available_qty",
			"reorder_point", "reorder_quantity", "last_movement_id", "created_at", "updated_at").
		Values(stock.GetID().String(), stock.GetItemID(), stock.GetWarehouseID().String(),
			stock.GetQuantity(), stock.GetReservedQty(), stock.GetAvailableQty(),
			stock.GetReorderPoint(), stock.GetReorderQuantity(), lastMovementID,
			stock.GetCreatedAt(), stock.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET quantity = EXCLUDED.quantity, " +
			"reserved_qty = EXCLUDED.reserved_qty, available_qty = EXCLUDED.available_qty, " +
			"reorder_point = EXCLUDED.reorder_point, reorder_quantity = EXCLUDED.reorder_quantity, " +
			"last_movement_id = EXCLUDED.last_movement_id, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save stock: %w", err)
	}

	return nil
}

func (r *StockRepository) FindByID(ctx context.Context, id domain.StockID) (*domain.Stock, error) {
	var dto stockDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, item_id, warehouse_id, quantity, reserved_qty, available_qty, 
				reorder_point, reorder_quantity, last_movement_id, created_at, updated_at 
		 FROM stock WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find stock by id: %w", err)
	}

	return dto.toDomain()
}

func (r *StockRepository) FindByItemAndWarehouse(ctx context.Context, itemID string, warehouseID domain.WarehouseID) (*domain.Stock, error) {
	var dto stockDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, item_id, warehouse_id, quantity, reserved_qty, available_qty, 
				reorder_point, reorder_quantity, last_movement_id, created_at, updated_at 
		 FROM stock WHERE item_id = $1 AND warehouse_id = $2`, itemID, warehouseID.String())
	if err != nil {
		return nil, fmt.Errorf("find stock by item and warehouse: %w", err)
	}

	return dto.toDomain()
}

func (r *StockRepository) FindByWarehouse(ctx context.Context, warehouseID domain.WarehouseID, filter domain.StockFilter) (pagination.PageResult[*domain.Stock], error) {
	q := r.psql.Select("id", "item_id", "warehouse_id", "quantity", "reserved_qty", "available_qty",
		"reorder_point", "reorder_quantity", "last_movement_id", "created_at", "updated_at").
		From("stock").
		Where(squirrel.Eq{"warehouse_id": warehouseID.String()})

	if filter.ItemID != nil {
		q = q.Where(squirrel.Eq{"item_id": *filter.ItemID})
	}
	if filter.Available != nil {
		if *filter.Available {
			q = q.Where(squirrel.Gt{"available_qty": 0})
		} else {
			q = q.Where(squirrel.Eq{"available_qty": 0})
		}
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
		return pagination.PageResult[*domain.Stock]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []stockDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Stock]{}, fmt.Errorf("select stock: %w", err)
	}

	stocks := make([]*domain.Stock, 0, len(dtos))
	for _, dto := range dtos {
		stock, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Stock]{}, err
		}
		stocks = append(stocks, stock)
	}

	return pagination.NewPageResult(stocks, limit), nil
}

func (r *StockRepository) FindAll(ctx context.Context, filter domain.StockFilter) (pagination.PageResult[*domain.Stock], error) {
	q := r.psql.Select("id", "item_id", "warehouse_id", "quantity", "reserved_qty", "available_qty",
		"reorder_point", "reorder_quantity", "last_movement_id", "created_at", "updated_at").
		From("stock")

	if filter.ItemID != nil {
		q = q.Where(squirrel.Eq{"item_id": *filter.ItemID})
	}
	if filter.WarehouseID != nil {
		q = q.Where(squirrel.Eq{"warehouse_id": filter.WarehouseID.String()})
	}
	if filter.Available != nil {
		if *filter.Available {
			q = q.Where(squirrel.Gt{"available_qty": 0})
		} else {
			q = q.Where(squirrel.Eq{"available_qty": 0})
		}
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
		return pagination.PageResult[*domain.Stock]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []stockDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Stock]{}, fmt.Errorf("select stock: %w", err)
	}

	stocks := make([]*domain.Stock, 0, len(dtos))
	for _, dto := range dtos {
		stock, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*domain.Stock]{}, err
		}
		stocks = append(stocks, stock)
	}

	return pagination.NewPageResult(stocks, limit), nil
}

func (r *StockRepository) Delete(ctx context.Context, id domain.StockID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM stock WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete stock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrStockNotFound
	}

	return nil
}
