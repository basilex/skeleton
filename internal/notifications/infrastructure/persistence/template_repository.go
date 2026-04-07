package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/jmoiron/sqlx"
)

type TemplateRepository struct {
	db *sqlx.DB
}

func NewTemplateRepository(db *sqlx.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

type templateRow struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Channel   string         `db:"channel"`
	Subject   sql.NullString `db:"subject"`
	Body      string         `db:"body"`
	HTMLBody  sql.NullString `db:"html_body"`
	Variables sql.NullString `db:"variables"`
	IsActive  int            `db:"is_active"`
	CreatedAt string         `db:"created_at"`
	UpdatedAt string         `db:"updated_at"`
}

func (r *TemplateRepository) Create(ctx context.Context, template *domain.NotificationTemplate) error {
	variablesJSON, err := json.Marshal(template.Variables())
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	query := `
		INSERT INTO notification_templates (
			id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		) VALUES (
			:id, :name, :channel, :subject, :body, :html_body, :variables, :is_active, :created_at, :updated_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":         template.ID().String(),
		"name":       template.Name(),
		"channel":    template.Channel().String(),
		"subject":    template.Subject(),
		"body":       template.Body(),
		"html_body":  nullString(template.HTMLBody()),
		"variables":  string(variablesJSON),
		"is_active":  boolToInt(template.IsActive()),
		"created_at": template.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		"updated_at": template.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})

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

	query := `
		UPDATE notification_templates SET
			name = :name,
			channel = :channel,
			subject = :subject,
			body = :body,
			html_body = :html_body,
			variables = :variables,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id
	`

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":         template.ID().String(),
		"name":       template.Name(),
		"channel":    template.Channel().String(),
		"subject":    template.Subject(),
		"body":       template.Body(),
		"html_body":  nullString(template.HTMLBody()),
		"variables":  string(variablesJSON),
		"is_active":  boolToInt(template.IsActive()),
		"updated_at": template.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	})

	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	return nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id domain.TemplateID) (*domain.NotificationTemplate, error) {
	var row templateRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE id = ?
	`, id.String())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template by id: %w", err)
	}

	return r.scanTemplate(row)
}

func (r *TemplateRepository) GetByName(ctx context.Context, name string) (*domain.NotificationTemplate, error) {
	var row templateRow
	err := r.db.GetContext(ctx, &row, `
		SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
		FROM notification_templates
		WHERE name = ?
	`, name)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template by name: %w", err)
	}

	return r.scanTemplate(row)
}

func (r *TemplateRepository) List(ctx context.Context, channel *domain.Channel) ([]*domain.NotificationTemplate, error) {
	var rows []templateRow
	var err error

	if channel != nil {
		err = r.db.SelectContext(ctx, &rows, `
			SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
			FROM notification_templates
			WHERE channel = ?
			ORDER BY name ASC
		`, channel.String())
	} else {
		err = r.db.SelectContext(ctx, &rows, `
			SELECT id, name, channel, subject, body, html_body, variables, is_active, created_at, updated_at
			FROM notification_templates
			ORDER BY name ASC
		`)
	}

	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	templates := make([]*domain.NotificationTemplate, len(rows))
	for i, row := range rows {
		t, err := r.scanTemplate(row)
		if err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		templates[i] = t
	}

	return templates, nil
}

func (r *TemplateRepository) Delete(ctx context.Context, id domain.TemplateID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_templates WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func (r *TemplateRepository) scanTemplate(row templateRow) (*domain.NotificationTemplate, error) {
	channel, err := domain.ParseChannel(row.Channel)
	if err != nil {
		return nil, fmt.Errorf("parse channel: %w", err)
	}

	var variables []string
	if row.Variables.Valid && row.Variables.String != "" {
		if err := json.Unmarshal([]byte(row.Variables.String), &variables); err != nil {
			return nil, fmt.Errorf("unmarshal variables: %w", err)
		}
	}

	opts := []domain.TemplateOption{}
	if row.HTMLBody.Valid && row.HTMLBody.String != "" {
		opts = append(opts, domain.WithHTMLBody(row.HTMLBody.String))
	}

	template, err := domain.NewNotificationTemplate(
		row.Name,
		channel,
		row.Subject.String,
		row.Body,
		variables,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}

	if row.IsActive == 0 {
		_ = template.Deactivate()
	}

	return template, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
