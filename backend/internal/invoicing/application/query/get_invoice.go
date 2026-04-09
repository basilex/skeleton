package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/invoicing/domain"
)

type GetInvoiceHandler struct {
	invoices domain.InvoiceRepository
}

func NewGetInvoiceHandler(invoices domain.InvoiceRepository) *GetInvoiceHandler {
	return &GetInvoiceHandler{
		invoices: invoices,
	}
}

type GetInvoiceQuery struct {
	InvoiceID string
}

type InvoiceDTO struct {
	ID            string           `json:"id"`
	InvoiceNumber string           `json:"invoice_number"`
	OrderID       *string          `json:"order_id"`
	ContractID    *string          `json:"contract_id"`
	CustomerID    string           `json:"customer_id"`
	SupplierID    *string          `json:"supplier_id"`
	IssueDate     string           `json:"issue_date"`
	DueDate       string           `json:"due_date"`
	Status        string           `json:"status"`
	Lines         []InvoiceLineDTO `json:"lines"`
	Subtotal      float64          `json:"subtotal"`
	TaxAmount     float64          `json:"tax_amount"`
	Discount      float64          `json:"discount"`
	Total         float64          `json:"total"`
	Currency      string           `json:"currency"`
	Notes         *string          `json:"notes"`
	PaidAmount    float64          `json:"paid_amount"`
	Payments      []PaymentDTO     `json:"payments"`
	CreatedAt     string           `json:"created_at"`
	UpdatedAt     string           `json:"updated_at"`
}

type InvoiceLineDTO struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Unit        string  `json:"unit"`
	Discount    float64 `json:"discount"`
	Total       float64 `json:"total"`
}

type PaymentDTO struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method"`
	Reference string  `json:"reference"`
	PaidAt    string  `json:"paid_at"`
	Notes     string  `json:"notes"`
}

func (h *GetInvoiceHandler) Handle(ctx context.Context, query GetInvoiceQuery) (*InvoiceDTO, error) {
	invoiceID, err := domain.ParseInvoiceID(query.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("parse invoice ID: %w", err)
	}

	invoice, err := h.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("find invoice: %w", err)
	}

	return toDTO(invoice), nil
}

func toDTO(invoice *domain.Invoice) *InvoiceDTO {
	lines := make([]InvoiceLineDTO, 0, len(invoice.GetLines()))
	for _, line := range invoice.GetLines() {
		lines = append(lines, InvoiceLineDTO{
			ID:          line.GetID().String(),
			Description: line.GetDescription(),
			Quantity:    line.GetQuantity(),
			UnitPrice:   line.GetUnitPrice().ToFloat64(),
			Unit:        line.GetUnit(),
			Discount:    line.GetDiscount().ToFloat64(),
			Total:       line.GetTotal().ToFloat64(),
		})
	}

	payments := make([]PaymentDTO, 0, len(invoice.GetPayments()))
	for _, payment := range invoice.GetPayments() {
		payments = append(payments, PaymentDTO{
			ID:        payment.GetID().String(),
			Amount:    payment.GetAmount().ToFloat64(),
			Currency:  payment.GetCurrency(),
			Method:    payment.GetMethod().String(),
			Reference: payment.GetReference(),
			PaidAt:    payment.GetPaidAt().Format("2006-01-02T15:04:05Z07:00"),
			Notes:     payment.GetNotes(),
		})
	}

	return &InvoiceDTO{
		ID:            invoice.GetID().String(),
		InvoiceNumber: invoice.GetInvoiceNumber(),
		OrderID:       invoice.GetOrderID(),
		ContractID:    invoice.GetContractID(),
		CustomerID:    invoice.GetCustomerID(),
		SupplierID:    invoice.GetSupplierID(),
		IssueDate:     invoice.GetIssueDate().Format("2006-01-02"),
		DueDate:       invoice.GetDueDate().Format("2006-01-02"),
		Status:        invoice.GetStatus().String(),
		Lines:         lines,
		Subtotal:      invoice.GetSubtotal().ToFloat64(),
		TaxAmount:     invoice.GetTaxAmount().ToFloat64(),
		Discount:      invoice.GetDiscount().ToFloat64(),
		Total:         invoice.GetTotal().ToFloat64(),
		Currency:      invoice.GetCurrency(),
		Notes:         invoice.GetNotes(),
		PaidAmount:    invoice.GetPaidAmount().ToFloat64(),
		Payments:      payments,
		CreatedAt:     invoice.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     invoice.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
