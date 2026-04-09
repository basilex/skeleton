package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	catalog "github.com/basilex/skeleton/internal/catalog/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ItemRepository) Save(ctx context.Context, item *catalog.Item) error {
	dto := toItemDTO(item)
	var categoryID *string = dto.CategoryID

	query, args, err := r.psql.Insert("catalog_items").
		Columns("id", "category_id", "name", "description", "sku", "base_price", "currency", "status", "attributes", "created_at", "updated_at").
		Values(dto.ID, categoryID, dto.Name, dto.Description, dto.SKU, dto.BasePrice, dto.Currency, dto.Status, dto.Attributes, dto.CreatedAt, dto.UpdatedAt).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, base_price = EXCLUDED.base_price, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save item: %w", err)
	}
	return nil
}

func (r *ItemRepository) FindByID(ctx context.Context, id catalog.ItemID) (*catalog.Item, error) {
	var dto itemDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, category_id, name, description, sku, base_price, currency, status, attributes, created_at, updated_at FROM catalog_items WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find item by id: %w", err)
	}
	return dto.toDomain()
}

func (r *ItemRepository) FindBySKU(ctx context.Context, sku string) (*catalog.Item, error) {
	var dto itemDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, category_id, name, description, sku, base_price, currency, status, attributes, created_at, updated_at FROM catalog_items WHERE sku = $1`,
		sku)
	if err != nil {
		return nil, fmt.Errorf("find item by sku: %w", err)
	}
	return dto.toDomain()
}

func (r *ItemRepository) FindAll(ctx context.Context, filter catalog.ItemFilter) (pagination.PageResult[*catalog.Item], error) {
	q := r.psql.Select("id", "category_id", "name", "description", "sku", "base_price", "currency", "status", "attributes", "created_at", "updated_at").
		From("catalog_items")

	if filter.CategoryID != nil {
		q = q.Where(squirrel.Eq{"category_id": filter.CategoryID.String()})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
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
		return pagination.PageResult[*catalog.Item]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []itemDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*catalog.Item]{}, fmt.Errorf("select items: %w", err)
	}

	items := make([]*catalog.Item, 0, len(dtos))
	for _, dto := range dtos {
		item, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*catalog.Item]{}, err
		}
		items = append(items, item)
	}

	return pagination.NewPageResult(items, limit), nil
}

func (r *ItemRepository) FindByCategory(ctx context.Context, categoryID catalog.CategoryID, filter catalog.ItemFilter) (pagination.PageResult[*catalog.Item], error) {
	q := r.psql.Select("id", "category_id", "name", "description", "sku", "base_price", "currency", "status", "attributes", "created_at", "updated_at").
		From("catalog_items").
		Where(squirrel.Eq{"category_id": categoryID.String()})

	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
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
		return pagination.PageResult[*catalog.Item]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []itemDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*catalog.Item]{}, fmt.Errorf("select items by category: %w", err)
	}

	items := make([]*catalog.Item, 0, len(dtos))
	for _, dto := range dtos {
		item, err := dto.toDomain()
		if err != nil {
			return pagination.PageResult[*catalog.Item]{}, err
		}
		items = append(items, item)
	}

	return pagination.NewPageResult(items, limit), nil
}

func (r *ItemRepository) Delete(ctx context.Context, id catalog.ItemID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM catalog_items WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if result.RowsAffected() == 0 {
		return catalog.ErrItemNotFound
	}
	return nil
}
