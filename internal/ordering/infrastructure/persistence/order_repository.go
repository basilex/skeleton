package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	var contractID *string
	if order.GetContractID() != "" {
		cid := order.GetContractID()
		contractID = &cid
	}

	var notes *string
	if order.GetNotes() != "" {
		n := order.GetNotes()
		notes = &n
	}

	var createdBy *string
	if order.GetCreatedBy() != "" {
		cb := order.GetCreatedBy()
		createdBy = &cb
	}

	query, args, err := r.psql.Insert("orders").
		Columns("id", "order_number", "customer_id", "supplier_id", "contract_id", "subtotal", "tax_amount", "discount", "total", "currency", "status", "order_date", "due_date", "notes", "created_by", "created_at", "updated_at").
		Values(
			order.GetID().String(),
			order.GetOrderNumber(),
			order.GetCustomerID(),
			order.GetSupplierID(),
			contractID,
			order.GetSubtotal(),
			order.GetTaxAmount(),
			order.GetDiscount(),
			order.GetTotal(),
			order.GetCurrency(),
			order.GetStatus().String(),
			order.GetOrderDate(),
			order.GetDueDate(),
			notes,
			createdBy,
			order.GetCreatedAt(),
			order.GetUpdatedAt(),
		).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, subtotal = EXCLUDED.subtotal, total = EXCLUDED.total, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save order: %w", err)
	}

	// Delete existing lines
	_, err = r.pool.Exec(ctx, "DELETE FROM order_lines WHERE order_id = $1", order.GetID().String())
	if err != nil {
		return fmt.Errorf("delete old lines: %w", err)
	}

	// Insert new lines
	for _, line := range order.GetLines() {
		lineQuery, lineArgs, err := r.psql.Insert("order_lines").
			Columns("id", "order_id", "item_id", "item_name", "quantity", "unit", "unit_price", "discount", "total", "created_at").
			Values(
				line.GetID().String(),
				order.GetID().String(),
				line.GetItemID(),
				line.GetItemName(),
				line.GetQuantity(),
				line.GetUnit(),
				line.GetUnitPrice(),
				line.GetDiscount(),
				line.GetTotal(),
				line.GetCreatedAt(),
			).
			ToSql()
		if err != nil {
			return fmt.Errorf("build insert line query: %w", err)
		}

		_, err = r.pool.Exec(ctx, lineQuery, lineArgs...)
		if err != nil {
			return fmt.Errorf("save order line: %w", err)
		}
	}

	return nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	var orderDTO orderDTO
	err := pgxscan.Get(ctx, r.pool, &orderDTO,
		`SELECT id, order_number, customer_id, supplier_id, contract_id, subtotal, tax_amount, discount, total, currency, status, order_date, due_date, completed_at, cancelled_at, notes, created_by, created_at, updated_at FROM orders WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find order by id: %w", err)
	}

	var lineDTOs []orderLineDTO
	err = pgxscan.Select(ctx, r.pool, &lineDTOs,
		`SELECT id, order_id, item_id, item_name, quantity, unit, unit_price, discount, total, created_at FROM order_lines WHERE order_id = $1 ORDER BY created_at`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("find order lines: %w", err)
	}

	return orderDTO.toDomain(lineDTOs)
}

func (r *OrderRepository) FindByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	var orderDTO orderDTO
	err := pgxscan.Get(ctx, r.pool, &orderDTO,
		`SELECT id, order_number, customer_id, supplier_id, contract_id, subtotal, tax_amount, discount, total, currency, status, order_date, due_date, completed_at, cancelled_at, notes, created_by, created_at, updated_at FROM orders WHERE order_number = $1`,
		orderNumber)
	if err != nil {
		return nil, fmt.Errorf("find order by number: %w", err)
	}

	var lineDTOs []orderLineDTO
	err = pgxscan.Select(ctx, r.pool, &lineDTOs,
		`SELECT id, order_id, item_id, item_name, quantity, unit, unit_price, discount, total, created_at FROM order_lines WHERE order_id = $1 ORDER BY created_at`,
		orderDTO.ID)
	if err != nil {
		return nil, fmt.Errorf("find order lines: %w", err)
	}

	return orderDTO.toDomain(lineDTOs)
}

func (r *OrderRepository) FindByCustomerID(ctx context.Context, customerID string, filter domain.OrderFilter) (pagination.PageResult[*domain.Order], error) {
	filter.CustomerID = &customerID
	return r.FindAll(ctx, filter)
}

func (r *OrderRepository) FindBySupplierID(ctx context.Context, supplierID string, filter domain.OrderFilter) (pagination.PageResult[*domain.Order], error) {
	filter.SupplierID = &supplierID
	return r.FindAll(ctx, filter)
}

func (r *OrderRepository) FindAll(ctx context.Context, filter domain.OrderFilter) (pagination.PageResult[*domain.Order], error) {
	q := r.psql.Select("id", "order_number", "customer_id", "supplier_id", "contract_id", "subtotal", "tax_amount", "discount", "total", "currency", "status", "order_date", "due_date", "completed_at", "cancelled_at", "notes", "created_by", "created_at", "updated_at").
		From("orders")

	if filter.CustomerID != nil {
		q = q.Where(squirrel.Eq{"customer_id": *filter.CustomerID})
	}
	if filter.SupplierID != nil {
		q = q.Where(squirrel.Eq{"supplier_id": *filter.SupplierID})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.StartDate != nil {
		startTime, _ := time.Parse("2006-01-02", *filter.StartDate)
		q = q.Where(squirrel.GtOrEq{"order_date": startTime})
	}
	if filter.EndDate != nil {
		endTime, _ := time.Parse("2006-01-02", *filter.EndDate)
		q = q.Where(squirrel.LtOrEq{"order_date": endTime})
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
		return pagination.PageResult[*domain.Order]{}, fmt.Errorf("build query: %w", err)
	}

	var orderDTOs []orderDTO
	if err := pgxscan.Select(ctx, r.pool, &orderDTOs, query, args...); err != nil {
		return pagination.PageResult[*domain.Order]{}, fmt.Errorf("select orders: %w", err)
	}

	orders := make([]*domain.Order, 0, len(orderDTOs))
	for _, dto := range orderDTOs {
		order, err := r.FindByID(ctx, domain.MustParseOrderID(dto.ID))
		if err != nil {
			return pagination.PageResult[*domain.Order]{}, err
		}
		orders = append(orders, order)
	}

	return pagination.NewPageResult(orders, limit), nil
}

func (r *OrderRepository) Delete(ctx context.Context, id domain.OrderID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
