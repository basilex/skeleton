package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepository struct {
	pool *pgxpool.Pool
}

func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

func (r *TemplateRepository) Create(ctx context.Context, template *domain.NotificationTemplate) error {
	variablesJSON, err := json.Marshal(template.Variables())
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	var htmlBody *string
	if template.HTMLBody() != "" {
		htmlBody = new(string)
		*htmlBody = template.HTMLBody()
	}

	query := `
		INSERT INTO notification_templates (
			id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.pool.Exec(ctx, query,
		template.ID().String(),
		template.Name(),
		template.Channel().String(),
		template.Subject(),
		template.Body(),
		htmlBody,
		string(variablesJSON),
		template.IsActive(),
		template.CreatedAt(),
		template.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("create template: %w", err)
	}

	return nil
}

func (r *TemplateRepository) Update(ctx context.Context, template *domain.NotificationTemplate) error {
	variablesJSON, err := json.Marshal(template.Variables())
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	var htmlBody *string
	if template.HTMLBody() != "" {
		htmlBody = new(string)
		*htmlBody = template.HTMLBody()
	}

	query := `
		UPDATE notification_templates SET
			name = $1,
			channel = $2,
			subject = $3,
			body = $4,
			html_body = $5,
			variables = $6,
			is_active = $7,
			updated_at = $8
		WHERE id = $9
	`

	_, err = r.pool.Exec(ctx, query,
		template.Name(),
		template.Channel().String(),
		template.Subject(),
		template.Body(),
		htmlBody,
		string(variablesJSON),
		template.IsActive(),
		template.UpdatedAt(),
		template.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	return nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id domain.TemplateID) (*domain.NotificationTemplate, error) {
	query := `
		SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id.String())

	return r.scanTemplate(row)
}

func (r *TemplateRepository) GetByName(ctx context.Context, name string) (*domain.NotificationTemplate, error) {
	query := `
		SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE name = $1
	`

	row := r.pool.QueryRow(ctx, query, name)

	return r.scanTemplate(row)
}

func (r *TemplateRepository) List(ctx context.Context, channel *domain.Channel) ([]*domain.NotificationTemplate, error) {
	var query string
	var args []interface{}

	if channel != nil {
		query = `
			SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
			FROM notification_templates
			WHERE channel = $1
			ORDER BY name ASC
		`
		args = []interface{}{channel.String()}
	} else {
		query = `
			SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
			FROM notification_templates
			ORDER BY name ASC
		`
		args = []interface{}{}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*domain.NotificationTemplate, 0)
	for rows.Next() {
		t, err := r.scanTemplateFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return templates, nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id domain.TemplateID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_templates WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) scanTemplate(row pgx.Row) (*domain.NotificationTemplate, error) {
	var id, name, channel, body string
	var subject, htmlBody *string
	var variablesBytes []byte
	var isActive bool
	var createdAt, updatedAt string

	err := row.Scan(
		&id, &name, &channel, &subject, &body, &htmlBody, &variablesBytes, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("scan template: %w", err)
	}

	channelEnum, err := domain.ParseChannel(channel)
	if err != nil {
		return nil, fmt.Errorf("parse channel: %w", err)
	}

	var variables []string
	if len(variablesBytes) > 0 {
		if err := json.Unmarshal(variablesBytes, &variables); err != nil {
			return nil, fmt.Errorf("unmarshal variables: %w", err)
		}
	}

	opts := []domain.TemplateOption{}
	if htmlBody != nil && *htmlBody != "" {
		opts = append(opts, domain.WithHTMLBody(*htmlBody))
	}

	var subjectStr string
	if subject != nil {
		subjectStr = *subject
	}

	template, err := domain.NewNotificationTemplate(
		name,
		channelEnum,
		subjectStr,
		body,
		variables,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}

	if !isActive {
		_ = template.Deactivate()
	}

	return template, nil
}

func (r *TemplateRepository) scanTemplateFromRows(rows pgx.Rows) (*domain.NotificationTemplate, error) {
	var id, name, channel, body string
	var subject, htmlBody *string
	var variablesBytes []byte
	var isActive bool
	var createdAt, updatedAt string

	err := rows.Scan(
		&id, &name, &channel, &subject, &body, &htmlBody, &variablesBytes, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan template: %w", err)
	}

	channelEnum, err := domain.ParseChannel(channel)
	if err != nil {
		return nil, fmt.Errorf("parse channel: %w", err)
	}

	var variables []string
	if len(variablesBytes) > 0 {
		if err := json.Unmarshal(variablesBytes, &variables); err != nil {
			return nil, fmt.Errorf("unmarshal variables: %w", err)
		}
	}

	opts := []domain.TemplateOption{}
	if htmlBody != nil && *htmlBody != "" {
		opts = append(opts, domain.WithHTMLBody(*htmlBody))
	}

	var subjectStr string
	if subject != nil {
		subjectStr = *subject
	}

	template, err := domain.NewNotificationTemplate(
		name,
		channelEnum,
		subjectStr,
		body,
		variables,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}

	if !isActive {
		_ = template.Deactivate()
	}

	return template, nil
}
