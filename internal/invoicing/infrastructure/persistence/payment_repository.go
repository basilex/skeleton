package persistence

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	query := `INSERT INTO payments (id, invoice_id, amount, currency, method, reference, paid_at, notes)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			  ON CONFLICT(id) DO UPDATE SET amount = EXCLUDED.amount, method = EXCLUDED.method, 
			  reference = EXCLUDED.reference, notes = EXCLUDED.notes`

	_, err := r.pool.Exec(ctx, query,
		payment.GetID().String(),
		payment.GetInvoiceID().String(),
		payment.GetAmount(),
		payment.GetCurrency(),
		payment.GetMethod().String(),
		payment.GetReference(),
		payment.GetPaidAt(),
		payment.GetNotes(),
	)

	if err != nil {
		return fmt.Errorf("save payment: %w", err)
	}

	return nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id domain.PaymentID) (*domain.Payment, error) {
	var dto paymentDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, invoice_id, amount, currency, method, reference, paid_at, notes, created_at 
		 FROM payments WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find payment by id: %w", err)
	}

	return dto.toDomain(), nil
}

func (r *PaymentRepository) FindByInvoiceID(ctx context.Context, invoiceID domain.InvoiceID) ([]*domain.Payment, error) {
	var dtos []paymentDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, invoice_id, amount, currency, method, reference, paid_at, notes, created_at 
		 FROM payments WHERE invoice_id = $1 ORDER BY paid_at`, invoiceID.String())
	if err != nil {
		return nil, fmt.Errorf("find payments by invoice id: %w", err)
	}

	payments := make([]*domain.Payment, 0, len(dtos))
	for _, dto := range dtos {
		payments = append(payments, dto.toDomain())
	}

	return payments, nil
}

func (r *PaymentRepository) Delete(ctx context.Context, id domain.PaymentID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM payments WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrPaymentNotFound
	}

	return nil
}
