package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *InvoiceRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
	var orderID *string
	if invoice.GetOrderID() != nil {
		orderID = invoice.GetOrderID()
	}

	var contractID *string
	if invoice.GetContractID() != nil {
		contractID = invoice.GetContractID()
	}

	var supplierID *string
	if invoice.GetSupplierID() != nil {
		supplierID = invoice.GetSupplierID()
	}

	var notes *string
	if invoice.GetNotes() != nil {
		notes = invoice.GetNotes()
	}

	query, args, err := r.psql.Insert("invoices").
		Columns("id", "invoice_number", "order_id", "contract_id", "customer_id", "supplier_id",
			"issue_date", "due_date", "status", "subtotal", "tax_amount", "discount", "total",
			"currency", "notes", "paid_amount", "created_at", "updated_at").
		Values(invoice.GetID().String(), invoice.GetInvoiceNumber(), orderID, contractID,
			invoice.GetCustomerID(), supplierID, invoice.GetIssueDate(), invoice.GetDueDate(),
			invoice.GetStatus().String(), invoice.GetSubtotal(), invoice.GetTaxAmount(),
			invoice.GetDiscount(), invoice.GetTotal(), invoice.GetCurrency(), notes,
			invoice.GetPaidAmount(), invoice.GetCreatedAt(), invoice.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, subtotal = EXCLUDED.subtotal, " +
			"tax_amount = EXCLUDED.tax_amount, discount = EXCLUDED.discount, total = EXCLUDED.total, " +
			"notes = EXCLUDED.notes, paid_amount = EXCLUDED.paid_amount, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	// Save invoice lines
	for _, line := range invoice.GetLines() {
		if err := r.saveLine(ctx, line); err != nil {
			return fmt.Errorf("save invoice line: %w", err)
		}
	}

	// Save payments
	for _, payment := range invoice.GetPayments() {
		if err := r.savePayment(ctx, payment); err != nil {
			return fmt.Errorf("save payment: %w", err)
		}
	}

	return nil
}

func (r *InvoiceRepository) saveLine(ctx context.Context, line *domain.InvoiceLine) error {
	query, args, err := r.psql.Insert("invoice_lines").
		Columns("id", "invoice_id", "description", "quantity", "unit_price", "unit", "discount", "total").
		Values(line.GetID().String(), line.GetInvoiceID().String(), line.GetDescription(),
			line.GetQuantity(), line.GetUnitPrice(), line.GetUnit(), line.GetDiscount(), line.GetTotal()).
		Suffix("ON CONFLICT(id) DO UPDATE SET description = EXCLUDED.description, " +
			"quantity = EXCLUDED.quantity, unit_price = EXCLUDED.unit_price, " +
			"discount = EXCLUDED.discount, total = EXCLUDED.total").
		ToSql()
	if err != nil {
		return fmt.Errorf("build line insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save line: %w", err)
	}

	return nil
}

func (r *InvoiceRepository) savePayment(ctx context.Context, payment *domain.Payment) error {
	query, args, err := r.psql.Insert("payments").
		Columns("id", "invoice_id", "amount", "currency", "method", "reference", "paid_at", "notes").
		Values(payment.GetID().String(), payment.GetInvoiceID().String(), payment.GetAmount(),
			payment.GetCurrency(), payment.GetMethod().String(), payment.GetReference(),
			payment.GetPaidAt(), payment.GetNotes()).
		Suffix("ON CONFLICT(id) DO UPDATE SET amount = EXCLUDED.amount, method = EXCLUDED.method, " +
			"reference = EXCLUDED.reference, notes = EXCLUDED.notes").
		ToSql()
	if err != nil {
		return fmt.Errorf("build payment insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save payment: %w", err)
	}

	return nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id domain.InvoiceID) (*domain.Invoice, error) {
	var dto invoiceDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, invoice_number, order_id, contract_id, customer_id, supplier_id, 
				issue_date, due_date, status, subtotal, tax_amount, discount, total, 
				currency, notes, paid_amount, created_at, updated_at 
		 FROM invoices WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find invoice by id: %w", err)
	}

	// Load invoice lines
	var lineDTOs []invoiceLineDTO
	err = pgxscan.Select(ctx, r.pool, &lineDTOs,
		`SELECT id, invoice_id, description, quantity, unit_price, unit, discount, total, created_at 
		 FROM invoice_lines WHERE invoice_id = $1 ORDER BY created_at`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find invoice lines: %w", err)
	}

	// Load payments
	var paymentDTOs []paymentDTO
	err = pgxscan.Select(ctx, r.pool, &paymentDTOs,
		`SELECT id, invoice_id, amount, currency, method, reference, paid_at, notes, created_at 
		 FROM payments WHERE invoice_id = $1 ORDER BY paid_at`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find payments: %w", err)
	}

	// Convert pointers for toDomain
	linePtrs := make([]*invoiceLineDTO, len(lineDTOs))
	for i := range lineDTOs {
		linePtrs[i] = &lineDTOs[i]
	}

	paymentPtrs := make([]*paymentDTO, len(paymentDTOs))
	for i := range paymentDTOs {
		paymentPtrs[i] = &paymentDTOs[i]
	}

	return dto.toDomain(linePtrs, paymentPtrs)
}

func (r *InvoiceRepository) FindByInvoiceNumber(ctx context.Context, invoiceNumber string) (*domain.Invoice, error) {
	var dto invoiceDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, invoice_number, order_id, contract_id, customer_id, supplier_id, 
				issue_date, due_date, status, subtotal, tax_amount, discount, total, 
				currency, notes, paid_amount, created_at, updated_at 
		 FROM invoices WHERE invoice_number = $1`, invoiceNumber)
	if err != nil {
		return nil, fmt.Errorf("find invoice by number: %w", err)
	}

	invoiceID, _ := domain.ParseInvoiceID(dto.ID)
	return r.FindByID(ctx, invoiceID)
}

func (r *InvoiceRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Invoice, error) {
	var dto invoiceDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, invoice_number, order_id, contract_id, customer_id, supplier_id, 
				issue_date, due_date, status, subtotal, tax_amount, discount, total, 
				currency, notes, paid_amount, created_at, updated_at 
		 FROM invoices WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("find invoice by order id: %w", err)
	}

	invoiceID, _ := domain.ParseInvoiceID(dto.ID)
	return r.FindByID(ctx, invoiceID)
}

func (r *InvoiceRepository) FindByCustomerID(ctx context.Context, customerID string, filter domain.InvoiceFilter) (pagination.PageResult[*domain.Invoice], error) {
	q := r.psql.Select("id", "invoice_number", "order_id", "contract_id", "customer_id", "supplier_id",
		"issue_date", "due_date", "status", "subtotal", "tax_amount", "discount", "total",
		"currency", "notes", "paid_amount", "created_at", "updated_at").
		From("invoices").
		Where(squirrel.Eq{"customer_id": customerID})

	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.StartDate != nil {
		q = q.Where(squirrel.GtOrEq{"issue_date": *filter.StartDate})
	}
	if filter.EndDate != nil {
		q = q.Where(squirrel.LtOrEq{"issue_date": *filter.EndDate})
	}
	if filter.Overdue != nil && *filter.Overdue {
		q = q.Where(squirrel.Lt{"due_date": time.Now()})
		q = q.Where(squirrel.NotEq{"status": domain.InvoiceStatusPaid})
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
		return pagination.PageResult[*domain.Invoice]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []invoiceDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Invoice]{}, fmt.Errorf("select invoices: %w", err)
	}

	invoices := make([]*domain.Invoice, 0, len(dtos))
	for _, dto := range dtos {
		invoiceID, _ := domain.ParseInvoiceID(dto.ID)
		invoice, err := r.FindByID(ctx, invoiceID)
		if err != nil {
			return pagination.PageResult[*domain.Invoice]{}, err
		}
		invoices = append(invoices, invoice)
	}

	return pagination.NewPageResult(invoices, limit), nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context, filter domain.InvoiceFilter) (pagination.PageResult[*domain.Invoice], error) {
	q := r.psql.Select("id", "invoice_number", "order_id", "contract_id", "customer_id", "supplier_id",
		"issue_date", "due_date", "status", "subtotal", "tax_amount", "discount", "total",
		"currency", "notes", "paid_amount", "created_at", "updated_at").
		From("invoices")

	if filter.CustomerID != nil {
		q = q.Where(squirrel.Eq{"customer_id": *filter.CustomerID})
	}
	if filter.SupplierID != nil {
		q = q.Where(squirrel.Eq{"supplier_id": *filter.SupplierID})
	}
	if filter.OrderID != nil {
		q = q.Where(squirrel.Eq{"order_id": *filter.OrderID})
	}
	if filter.ContractID != nil {
		q = q.Where(squirrel.Eq{"contract_id": *filter.ContractID})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.StartDate != nil {
		q = q.Where(squirrel.GtOrEq{"issue_date": *filter.StartDate})
	}
	if filter.EndDate != nil {
		q = q.Where(squirrel.LtOrEq{"issue_date": *filter.EndDate})
	}
	if filter.Overdue != nil && *filter.Overdue {
		q = q.Where(squirrel.Lt{"due_date": time.Now()})
		q = q.Where(squirrel.NotEq{"status": domain.InvoiceStatusPaid})
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
		return pagination.PageResult[*domain.Invoice]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []invoiceDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Invoice]{}, fmt.Errorf("select invoices: %w", err)
	}

	invoices := make([]*domain.Invoice, 0, len(dtos))
	for _, dto := range dtos {
		invoiceID, _ := domain.ParseInvoiceID(dto.ID)
		invoice, err := r.FindByID(ctx, invoiceID)
		if err != nil {
			return pagination.PageResult[*domain.Invoice]{}, err
		}
		invoices = append(invoices, invoice)
	}

	return pagination.NewPageResult(invoices, limit), nil
}

func (r *InvoiceRepository) Delete(ctx context.Context, id domain.InvoiceID) error {
	// Delete payments first
	_, err := r.pool.Exec(ctx, `DELETE FROM payments WHERE invoice_id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete payments: %w", err)
	}

	// Delete invoice lines
	_, err = r.pool.Exec(ctx, `DELETE FROM invoice_lines WHERE invoice_id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete invoice lines: %w", err)
	}

	// Delete invoice
	result, err := r.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrInvoiceNotFound
	}

	return nil
}
