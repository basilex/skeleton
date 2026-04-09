package persistence

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

type sessionDTO struct {
	ID            string     `db:"id"`
	UserID        string     `db:"user_id"`
	Status        string     `db:"status"`
	DeviceType    *string    `db:"device_type"`
	OS            *string    `db:"os"`
	Browser       *string    `db:"browser"`
	DeviceName    *string    `db:"device_name"`
	UserAgent     *string    `db:"user_agent"`
	IPAddress     *net.IP    `db:"ip_address"`
	ExpiresAt     time.Time  `db:"expires_at"`
	LastActivity  time.Time  `db:"last_activity"`
	CreatedAt     time.Time  `db:"created_at"`
	RevokedAt     *time.Time `db:"revoked_at"`
	RevokedReason *string    `db:"revoked_reason"`
}

func (r *SessionRepository) Save(ctx context.Context, session *domain.Session) error {
	var ipAddress *net.IP
	if ip := session.IPAddress(); ip != nil {
		ipAddress = &ip
	}

	var deviceType, os, browser, deviceName, userAgent *string
	if d := session.Device(); d.UserAgent() != "" {
		ua := d.UserAgent()
		userAgent = &ua
	}
	if d := session.Device(); d.DeviceType() != "" {
		dt := d.DeviceType()
		deviceType = &dt
	}
	if d := session.Device(); d.OS() != "" {
		osVal := d.OS()
		os = &osVal
	}
	if d := session.Device(); d.Browser() != "" {
		b := d.Browser()
		browser = &b
	}
	if d := session.Device(); d.DeviceName() != "" {
		dn := d.DeviceName()
		deviceName = &dn
	}

	query, args, err := r.psql.Insert("sessions").
		Columns("id", "user_id", "status", "device_type", "os", "browser", "device_name",
			"user_agent", "ip_address", "expires_at", "last_activity", "created_at",
			"revoked_at", "revoked_reason").
		Values(session.ID(), session.UserID(), session.Status(),
			deviceType, os, browser, deviceName, userAgent,
			ipAddress, session.ExpiresAt(), session.LastActivity(), session.CreatedAt(),
			session.RevokedAt(), session.RevokedReason()).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, last_activity = EXCLUDED.last_activity, revoked_at = EXCLUDED.revoked_at, revoked_reason = EXCLUDED.revoked_reason").
		ToSql()
	if err != nil {
		return fmt.Errorf("build session insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	var dto sessionDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, user_id, status, device_type, os, browser, device_name, user_agent,
		 ip_address, expires_at, last_activity, created_at, revoked_at, revoked_reason
		 FROM sessions WHERE id = $1`,
		id)
	if err != nil {
		return nil, fmt.Errorf("find session by id: %w", err)
	}
	return r.dtoToDomain(dto)
}

func (r *SessionRepository) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	var dtos []sessionDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, user_id, status, device_type, os, browser, device_name, user_agent,
		 ip_address, expires_at, last_activity, created_at, revoked_at, revoked_reason
		 FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("find sessions by user id: %w", err)
	}
	return r.dtosToDomains(dtos)
}

func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	var dtos []sessionDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, user_id, status, device_type, os, browser, device_name, user_agent,
		 ip_address, expires_at, last_activity, created_at, revoked_at, revoked_reason
		 FROM sessions WHERE user_id = $1 AND status = 'active' AND expires_at > NOW() ORDER BY last_activity DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("find active sessions: %w", err)
	}
	return r.dtosToDomains(dtos)
}

func (r *SessionRepository) Delete(ctx context.Context, id domain.SessionID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM sessions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID domain.UserID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return result.RowsAffected(), nil
}

func (r *SessionRepository) dtoToDomain(dto sessionDTO) (*domain.Session, error) {
	userID, err := domain.ParseUserID(dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	sessionID, err := domain.ParseSessionID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("parse session id: %w", err)
	}

	status, err := domain.ParseSessionStatus(dto.Status)
	if err != nil {
		return nil, fmt.Errorf("parse session status: %w", err)
	}

	var device domain.DeviceInfo
	if dto.DeviceType != nil || dto.OS != nil || dto.Browser != nil || dto.DeviceName != nil || dto.UserAgent != nil {
		device = domain.NewDeviceInfo(
			ptrToString(dto.UserAgent),
			ptrToString(dto.DeviceType),
			ptrToString(dto.OS),
			ptrToString(dto.Browser),
			ptrToString(dto.DeviceName),
		)
	}

	var ipAddress net.IP
	if dto.IPAddress != nil {
		ipAddress = *dto.IPAddress
	}

	return domain.ReconstituteSession(
		sessionID,
		userID,
		status,
		device,
		ipAddress,
		dto.ExpiresAt,
		dto.LastActivity,
		dto.CreatedAt,
		dto.RevokedAt,
		ptrToString(dto.RevokedReason),
	), nil
}

func (r *SessionRepository) dtosToDomains(dtos []sessionDTO) ([]*domain.Session, error) {
	sessions := make([]*domain.Session, 0, len(dtos))
	for _, dto := range dtos {
		session, err := r.dtoToDomain(dto)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
