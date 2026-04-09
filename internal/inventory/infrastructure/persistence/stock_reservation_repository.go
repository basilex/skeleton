package persistence

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockReservationRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewStockReservationRepository(pool *pgxpool.Pool) *StockReservationRepository {
	return &StockReservationRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *StockReservationRepository) Save(ctx context.Context, reservation *domain.StockReservation) error {
	query, args, err := r.psql.Insert("stock_reservations").
		Columns("id", "stock_id", "order_id", "quantity", "status", "reserved_at",
			"expires_at", "fulfilled_at", "cancelled_at", "created_at", "updated_at").
		Values(reservation.GetID().String(), reservation.GetStockID().String(), reservation.GetOrderID(),
			reservation.GetQuantity(), reservation.GetStatus().String(), reservation.GetReservedAt(),
			reservation.GetExpiresAt(), reservation.GetFulfilledAt(), reservation.GetCancelledAt(),
			reservation.GetCreatedAt(), reservation.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, " +
			"fulfilled_at = EXCLUDED.fulfilled_at, cancelled_at = EXCLUDED.cancelled_at, " +
			"updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save stock reservation: %w", err)
	}

	return nil
}

func (r *StockReservationRepository) FindByID(ctx context.Context, id domain.StockReservationID) (*domain.StockReservation, error) {
	var dto stockReservationDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, stock_id, order_id, quantity, status, reserved_at, 
				expires_at, fulfilled_at, cancelled_at, created_at, updated_at 
		 FROM stock_reservations WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find stock reservation by id: %w", err)
	}

	return dto.toDomain()
}

func (r *StockReservationRepository) FindByOrder(ctx context.Context, orderID string) ([]*domain.StockReservation, error) {
	var dtos []stockReservationDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, stock_id, order_id, quantity, status, reserved_at, 
				expires_at, fulfilled_at, cancelled_at, created_at, updated_at 
		 FROM stock_reservations WHERE order_id = $1 ORDER BY reserved_at DESC`, orderID)
	if err != nil {
		return nil, fmt.Errorf("find stock reservations by order: %w", err)
	}

	reservations := make([]*domain.StockReservation, 0, len(dtos))
	for _, dto := range dtos {
		reservation, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

func (r *StockReservationRepository) FindActiveByStock(ctx context.Context, stockID domain.StockID) ([]*domain.StockReservation, error) {
	var dtos []stockReservationDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, stock_id, order_id, quantity, status, reserved_at, 
				expires_at, fulfilled_at, cancelled_at, created_at, updated_at 
		 FROM stock_reservations 
		 WHERE stock_id = $1 AND status = $2 
		 ORDER BY reserved_at DESC`, stockID.String(), domain.ReservationStatusActive.String())
	if err != nil {
		return nil, fmt.Errorf("find active stock reservations: %w", err)
	}

	reservations := make([]*domain.StockReservation, 0, len(dtos))
	for _, dto := range dtos {
		reservation, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

func (r *StockReservationRepository) Delete(ctx context.Context, id domain.StockReservationID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM stock_reservations WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete stock reservation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrReservationNotFound
	}

	return nil
}
