package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/documents/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *TemplateRepository) Save(ctx context.Context, template *domain.Template) error {
	variables := template.GetVariables()
	if variables == nil {
		variables = []string{}
	}

	query, args, err := r.psql.Insert("document_templates").
		Columns("id", "name", "document_type", "content", "variables", "created_at", "updated_at").
		Values(template.GetID().String(), template.GetName(), template.GetDocumentType().String(),
			template.GetContent(), variables, template.GetCreatedAt(), template.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name, content = EXCLUDED.content, " +
			"variables = EXCLUDED.variables, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save template: %w", err)
	}

	return nil
}

func (r *TemplateRepository) FindByID(ctx context.Context, id domain.TemplateID) (*domain.Template, error) {
	var dto templateDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, name, document_type, content, variables, created_at, updated_at 
		 FROM document_templates WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find template by id: %w", err)
	}

	return dto.toDomain(), nil
}

func (r *TemplateRepository) FindByDocumentType(ctx context.Context, documentType domain.DocumentType) ([]*domain.Template, error) {
	var dtos []templateDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, name, document_type, content, variables, created_at, updated_at 
		 FROM document_templates WHERE document_type = $1 ORDER BY created_at DESC`, documentType.String())
	if err != nil {
		return nil, fmt.Errorf("find templates by type: %w", err)
	}

	templates := make([]*domain.Template, 0, len(dtos))
	for _, dto := range dtos {
		templates = append(templates, dto.toDomain())
	}

	return templates, nil
}

func (r *TemplateRepository) FindAll(ctx context.Context) ([]*domain.Template, error) {
	var dtos []templateDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, name, document_type, content, variables, created_at, updated_at 
		 FROM document_templates ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("find all templates: %w", err)
	}

	templates := make([]*domain.Template, 0, len(dtos))
	for _, dto := range dtos {
		templates = append(templates, dto.toDomain())
	}

	return templates, nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id domain.TemplateID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM document_templates WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}
