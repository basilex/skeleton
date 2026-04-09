package persistence

import (
	"encoding/json"
	"time"

	"github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/money"
)

type orderDTO struct {
	ID          string     `db:"id"`
	OrderNumber string     `db:"order_number"`
	CustomerID  string     `db:"customer_id"`
	SupplierID  string     `db:"supplier_id"`
	ContractID  *string    `db:"contract_id"`
	Subtotal    int64      `db:"subtotal"`
	TaxAmount   int64      `db:"tax_amount"`
	Discount    int64      `db:"discount"`
	Total       int64      `db:"total"`
	Currency    string     `db:"currency"`
	Status      string     `db:"status"`
	OrderDate   time.Time  `db:"order_date"`
	DueDate     *time.Time `db:"due_date"`
	CompletedAt *time.Time `db:"completed_at"`
	CancelledAt *time.Time `db:"cancelled_at"`
	Notes       *string    `db:"notes"`
	CreatedBy   *string    `db:"created_by"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type orderLineDTO struct {
	ID        string          `db:"id"`
	OrderID   string          `db:"order_id"`
	ItemID    string          `db:"item_id"`
	ItemName  string          `db:"item_name"`
	Quantity  float64         `db:"quantity"`
	Unit      string          `db:"unit"`
	UnitPrice int64           `db:"unit_price"`
	Discount  int64           `db:"discount"`
	Total     int64           `db:"total"`
	Metadata  json.RawMessage `db:"metadata"`
	CreatedAt time.Time       `db:"created_at"`
}

func (dto *orderDTO) toDomain(lines []orderLineDTO) (*domain.Order, error) {
	orderID, err := domain.ParseOrderID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := domain.ParseOrderStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	var contractID string
	if dto.ContractID != nil {
		contractID = *dto.ContractID
	}

	subtotal, _ := money.New(dto.Subtotal, dto.Currency)
	taxAmount, _ := money.New(dto.TaxAmount, dto.Currency)
	discount, _ := money.New(dto.Discount, dto.Currency)
	total, _ := money.New(dto.Total, dto.Currency)

	orderLines := make([]*domain.OrderLine, 0, len(lines))
	for _, line := range lines {
		orderLineID, err := domain.ParseOrderLineID(line.ID)
		if err != nil {
			return nil, err
		}

		lineUnitPrice, _ := money.New(line.UnitPrice, dto.Currency)
		lineDiscount, _ := money.New(line.Discount, dto.Currency)
		lineTotal, _ := money.New(line.Total, dto.Currency)

		orderLine, err := domain.ReconstituteOrderLine(
			orderLineID,
			orderID,
			line.ItemID,
			line.ItemName,
			line.Quantity,
			line.Unit,
			lineUnitPrice,
			lineDiscount,
			lineTotal,
			line.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orderLines = append(orderLines, orderLine)
	}

	var notes string
	if dto.Notes != nil {
		notes = *dto.Notes
	}

	var createdBy string
	if dto.CreatedBy != nil {
		createdBy = *dto.CreatedBy
	}

	return domain.ReconstituteOrder(
		orderID,
		dto.OrderNumber,
		status,
		dto.CustomerID,
		dto.SupplierID,
		contractID,
		subtotal,
		taxAmount,
		discount,
		total,
		dto.Currency,
		orderLines,
		dto.OrderDate,
		dto.DueDate,
		dto.CompletedAt,
		dto.CancelledAt,
		notes,
		make(map[string]interface{}),
		createdBy,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}
