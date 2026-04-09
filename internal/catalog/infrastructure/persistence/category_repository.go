package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	catalog "github.com/basilex/skeleton/internal/catalog/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *CategoryRepository) Save(ctx context.Context, category *catalog.Category) error {
	dto := toCategoryDTO(category)

	query, args, err := r.psql.Insert("catalog_categories").
		Columns("id", "name", "description", "path", "is_active", "created_at", "updated_at").
		Values(dto.ID, dto.Name, dto.Description, dto.Path, dto.IsActive, dto.CreatedAt, dto.UpdatedAt).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) FindByID(ctx context.Context, id catalog.CategoryID) (*catalog.Category, error) {
	var dto categoryDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, name, description, path, is_active, created_at, updated_at FROM catalog_categories WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find category by id: %w", err)
	}
	return dto.toDomain()
}

func (r *CategoryRepository) FindByPath(ctx context.Context, path string) (*catalog.Category, error) {
	var dto categoryDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, name, description, path, is_active, created_at, updated_at FROM catalog_categories WHERE path = $1`,
		path)
	if err != nil {
		return nil, fmt.Errorf("find category by path: %w", err)
	}
	return dto.toDomain()
}

func (r *CategoryRepository) FindAll(ctx context.Context) ([]*catalog.Category, error) {
	query, args, err := r.psql.Select("id", "name", "description", "path", "is_active", "created_at", "updated_at").
		From("catalog_categories").
		OrderBy("path").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dtos []categoryDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}

	categories := make([]*catalog.Category, 0, len(dtos))
	for _, dto := range dtos {
		category, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) FindChildren(ctx context.Context, parentPath string) ([]*catalog.Category, error) {
	query, args, err := r.psql.Select("id", "name", "description", "path", "is_active", "created_at", "updated_at").
		From("catalog_categories").
		Where(squirrel.Like{"path": parentPath + "%"}).
		OrderBy("path").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dtos []categoryDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}

	categories := make([]*catalog.Category, 0, len(dtos))
	for _, dto := range dtos {
		category, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
