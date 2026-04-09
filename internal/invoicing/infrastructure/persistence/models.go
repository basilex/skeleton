package persistence

import (
	"time"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/money"
)

type invoiceDTO struct {
	ID            string    `db:"id"`
	InvoiceNumber string    `db:"invoice_number"`
	OrderID       *string   `db:"order_id"`
	ContractID    *string   `db:"contract_id"`
	CustomerID    string    `db:"customer_id"`
	SupplierID    *string   `db:"supplier_id"`
	IssueDate     time.Time `db:"issue_date"`
	DueDate       time.Time `db:"due_date"`
	Status        string    `db:"status"`
	Subtotal      int64     `db:"subtotal"`
	TaxRate       float64   `db:"tax_rate"`
	TaxAmount     int64     `db:"tax_amount"`
	Discount      int64     `db:"discount"`
	Total         int64     `db:"total"`
	Currency      string    `db:"currency"`
	Notes         *string   `db:"notes"`
	PaidAmount    int64     `db:"paid_amount"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (dto *invoiceDTO) toDomain(lines []*invoiceLineDTO, payments []*paymentDTO) (*domain.Invoice, error) {
	invoiceID, err := domain.ParseInvoiceID(dto.ID)
	if err != nil {
		return nil, err
	}

	status := domain.InvoiceStatus(dto.Status)
	domainLines := make([]*domain.InvoiceLine, 0, len(lines))
	for _, line := range lines {
		domainLines = append(domainLines, line.toDomain())
	}

	domainPayments := make([]*domain.Payment, 0, len(payments))
	for _, payment := range payments {
		domainPayments = append(domainPayments, payment.toDomain())
	}

	subtotal, _ := money.New(dto.Subtotal, dto.Currency)
	taxAmount, _ := money.New(dto.TaxAmount, dto.Currency)
	discount, _ := money.New(dto.Discount, dto.Currency)
	total, _ := money.New(dto.Total, dto.Currency)
	paidAmount, _ := money.New(dto.PaidAmount, dto.Currency)

	return domain.RestoreInvoice(
		invoiceID,
		dto.InvoiceNumber,
		dto.OrderID,
		dto.ContractID,
		dto.CustomerID,
		dto.SupplierID,
		dto.IssueDate,
		dto.DueDate,
		status,
		domainLines,
		subtotal,
		dto.TaxRate,
		taxAmount,
		discount,
		total,
		dto.Currency,
		dto.Notes,
		paidAmount,
		domainPayments,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}

type invoiceLineDTO struct {
	ID          string    `db:"id"`
	InvoiceID   string    `db:"invoice_id"`
	Description string    `db:"description"`
	Quantity    float64   `db:"quantity"`
	UnitPrice   int64     `db:"unit_price"`
	Unit        string    `db:"unit"`
	Discount    int64     `db:"discount"`
	Total       int64     `db:"total"`
	CreatedAt   time.Time `db:"created_at"`
}

func (dto *invoiceLineDTO) toDomain() *domain.InvoiceLine {
	lineID, _ := domain.ParseInvoiceLineID(dto.ID)
	invoiceID, _ := domain.ParseInvoiceID(dto.InvoiceID)

	unitPrice, _ := money.New(dto.UnitPrice, "USD")
	discount, _ := money.New(dto.Discount, "USD")
	total, _ := money.New(dto.Total, "USD")

	return domain.RestoreInvoiceLine(
		lineID,
		invoiceID,
		dto.Description,
		dto.Quantity,
		unitPrice,
		dto.Unit,
		discount,
		total,
	)
}

type paymentDTO struct {
	ID        string    `db:"id"`
	InvoiceID string    `db:"invoice_id"`
	Amount    int64     `db:"amount"`
	Currency  string    `db:"currency"`
	Method    string    `db:"method"`
	Reference string    `db:"reference"`
	PaidAt    time.Time `db:"paid_at"`
	Notes     string    `db:"notes"`
	CreatedAt time.Time `db:"created_at"`
}

func (dto *paymentDTO) toDomain() *domain.Payment {
	paymentID, _ := domain.ParsePaymentID(dto.ID)
	invoiceID, _ := domain.ParseInvoiceID(dto.InvoiceID)

	amount, _ := money.New(dto.Amount, dto.Currency)

	return domain.RestorePayment(
		paymentID,
		invoiceID,
		amount,
		dto.Currency,
		domain.PaymentMethod(dto.Method),
		dto.Reference,
		dto.PaidAt,
		dto.Notes,
	)
}
