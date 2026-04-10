package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type userDTO struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	query, args, err := r.psql.Insert("users").
		Columns("id", "email", "password_hash", "is_active", "created_at", "updated_at").
		Values(user.ID(), user.Email().String(), user.PasswordHash(), user.IsActive(), user.CreatedAt(), user.UpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET email = EXCLUDED.email, password_hash = EXCLUDED.password_hash, is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	var dto userDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, email, password_hash, is_active, created_at, updated_at FROM users WHERE id = $1`,
		id)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	roleIDs, err := r.findRoleIDs(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf("find user roles: %w", err)
	}

	return r.dtoToDomain(dto, roleIDs)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var dto userDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, email, password_hash, is_active, created_at, updated_at FROM users WHERE email = $1`,
		email.String())
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	roleIDs, err := r.findRoleIDs(ctx, dto.ID)
	if err != nil {
		return nil, fmt.Errorf("find user roles: %w", err)
	}

	return r.dtoToDomain(dto, roleIDs)
}

func (r *UserRepository) FindAll(ctx context.Context, filter domain.UserFilter) (pagination.PageResult[*domain.User], error) {
	q := r.psql.Select("id", "email", "password_hash", "is_active", "created_at", "updated_at").
		From("users")

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"email": "%" + filter.Search + "%"})
	}
	if filter.IsActive != nil {
		q = q.Where(squirrel.Eq{"is_active": *filter.IsActive})
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
		return pagination.PageResult[*domain.User]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []userDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.User]{}, fmt.Errorf("select users: %w", err)
	}

	users := make([]*domain.User, 0, len(dtos))
	for _, dto := range dtos {
		roleIDs, err := r.findRoleIDs(ctx, dto.ID)
		if err != nil {
			return pagination.PageResult[*domain.User]{}, fmt.Errorf("find user roles: %w", err)
		}
		user, err := r.dtoToDomain(dto, roleIDs)
		if err != nil {
			return pagination.PageResult[*domain.User]{}, err
		}
		users = append(users, user)
	}

	return pagination.NewPageResult(users, limit), nil
}

func (r *UserRepository) Delete(ctx context.Context, id domain.UserID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) findRoleIDs(ctx context.Context, userID string) ([]domain.RoleID, error) {
	rows, err := r.pool.Query(ctx, `SELECT role_id FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	var roleIDs []domain.RoleID
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("scan role id: %w", err)
		}
		parsed, err := domain.ParseRoleID(roleID)
		if err != nil {
			continue
		}
		roleIDs = append(roleIDs, parsed)
	}
	return roleIDs, nil
}

func (r *UserRepository) dtoToDomain(dto userDTO, roleIDs []domain.RoleID) (*domain.User, error) {
	userID, err := domain.ParseUserID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	userEmail, err := domain.NewEmail(dto.Email)
	if err != nil {
		return nil, fmt.Errorf("parse email: %w", err)
	}
	if roleIDs == nil {
		roleIDs = []domain.RoleID{}
	}
	return domain.ReconstituteUser(userID, userEmail, domain.PasswordHash(dto.PasswordHash), roleIDs, dto.IsActive, dto.CreatedAt, dto.UpdatedAt)
}
