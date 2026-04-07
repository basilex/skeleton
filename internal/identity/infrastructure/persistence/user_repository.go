package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/jmoiron/sqlx"
)

// UserRepository implements the user repository interface using SQL database storage.
// It handles persistence of user entities with support for filtering and pagination.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository with the provided database connection.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :is_active, :created_at, :updated_at)
		ON CONFLICT(id) DO UPDATE SET
			email = excluded.email,
			password_hash = excluded.password_hash,
			is_active = excluded.is_active,
			updated_at = excluded.updated_at
	`
	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":            string(user.ID()),
		"email":         user.Email().String(),
		"password_hash": user.PasswordHash(),
		"is_active":     user.IsActive(),
		"created_at":    user.CreatedAt().Format(time.RFC3339),
		"updated_at":    user.UpdatedAt().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their unique identifier.
// Returns domain.ErrUserNotFound if no matching user exists.
func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var row struct {
		ID           string `db:"id"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		IsActive     bool   `db:"is_active"`
		CreatedAt    string `db:"created_at"`
		UpdatedAt    string `db:"updated_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT id, email, password_hash, is_active, created_at, updated_at FROM users WHERE id = ?`, string(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return r.scanUser(row)
}

// FindByEmail retrieves a user by their email address.
// Returns domain.ErrUserNotFound if no matching user exists.
func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var row struct {
		ID           string `db:"id"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		IsActive     bool   `db:"is_active"`
		CreatedAt    string `db:"created_at"`
		UpdatedAt    string `db:"updated_at"`
	}
	err := r.db.GetContext(ctx, &row, `SELECT id, email, password_hash, is_active, created_at, updated_at FROM users WHERE email = ?`, email.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return r.scanUser(row)
}

// FindAll retrieves users with optional filtering and pagination.
// Supports filtering by search term (email), active status, and cursor-based pagination.
func (r *UserRepository) FindAll(ctx context.Context, filter domain.UserFilter) (pagination.PageResult[*domain.User], error) {
	query := `SELECT id, email, password_hash, is_active, created_at, updated_at FROM users`
	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	if filter.Search != "" {
		conditions = append(conditions, "email LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}
	if filter.Cursor != "" {
		conditions = append(conditions, "id < ?")
		args = append(args, filter.Cursor)
	}

	if len(conditions) > 0 {
		where := " WHERE " + joinConditions(conditions)
		query += where
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit+1)

	var rows []struct {
		ID           string `db:"id"`
		Email        string `db:"email"`
		PasswordHash string `db:"password_hash"`
		IsActive     bool   `db:"is_active"`
		CreatedAt    string `db:"created_at"`
		UpdatedAt    string `db:"updated_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return pagination.PageResult[*domain.User]{}, fmt.Errorf("select users: %w", err)
	}

	users := make([]*domain.User, len(rows))
	for i, row := range rows {
		u, err := r.scanUser(row)
		if err != nil {
			return pagination.PageResult[*domain.User]{}, fmt.Errorf("scan user: %w", err)
		}
		users[i] = u
	}

	return pagination.NewPageResult(users, limit), nil
}

// Delete removes a user from the database by ID.
// Returns domain.ErrUserNotFound if no matching user exists.
func (r *UserRepository) Delete(ctx context.Context, id domain.UserID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, string(id))
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// scanUser converts a database row into a domain User entity.
func (r *UserRepository) scanUser(row struct {
	ID           string `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	IsActive     bool   `db:"is_active"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}) (*domain.User, error) {
	userID, err := domain.ParseUserID(row.ID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	email, err := domain.NewEmail(row.Email)
	if err != nil {
		return nil, fmt.Errorf("parse email: %w", err)
	}
	passwordHash := domain.PasswordHash(row.PasswordHash)
	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, row.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	user, err := domain.ReconstituteUser(userID, email, passwordHash, []domain.RoleID{}, row.IsActive, createdAt, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("reconstitute user: %w", err)
	}
	return user, nil
}

// joinConditions concatenates multiple SQL conditions with AND operators.
func joinConditions(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}
