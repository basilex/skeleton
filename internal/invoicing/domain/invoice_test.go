package domain

import (
	"testing"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

func TestNewInvoice(t *testing.T) {
	tests := []struct {
		name          string
		invoiceNumber string
		customerID    string
		currency      string
		dueDate       time.Time
		wantErr       error
	}{
		{
			name:          "valid invoice",
			invoiceNumber: "INV-2024-001",
			customerID:    "0193a7b2-1234-5678-9abc-def012345678",
			currency:      "USD",
			dueDate:       time.Now().Add(30 * 24 * time.Hour),
			wantErr:       nil,
		},
		{
			name:          "empty invoice number",
			invoiceNumber: "",
			customerID:    "0193a7b2-1234-5678-9abc-def012345678",
			currency:      "USD",
			dueDate:       time.Now().Add(30 * 24 * time.Hour),
			wantErr:       ErrEmptyInvoiceNumber,
		},
		{
			name:          "empty customer ID",
			invoiceNumber: "INV-2024-001",
			customerID:    "",
			currency:      "USD",
			dueDate:       time.Now().Add(30 * 24 * time.Hour),
			wantErr:       ErrEmptyCustomerID,
		},
		{
			name:          "past due date",
			invoiceNumber: "INV-2024-001",
			customerID:    "0193a7b2-1234-5678-9abc-def012345678",
			currency:      "USD",
			dueDate:       time.Now().Add(-24 * time.Hour),
			wantErr:       ErrInvalidDueDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice, err := NewInvoice(tt.invoiceNumber, tt.customerID, tt.currency, tt.dueDate)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("NewInvoice() unexpected error: %v", err)
				} else if err != tt.wantErr {
					t.Errorf("NewInvoice() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErr != nil {
				t.Errorf("NewInvoice() expected error %v, got nil", tt.wantErr)
				return
			}
			if invoice.GetInvoiceNumber() != tt.invoiceNumber {
				t.Errorf("InvoiceNumber = %v, want %v", invoice.GetInvoiceNumber(), tt.invoiceNumber)
			}
			if invoice.GetCustomerID() != tt.customerID {
				t.Errorf("CustomerID = %v, want %v", invoice.GetCustomerID(), tt.customerID)
			}
			if invoice.GetStatus() != InvoiceStatusDraft {
				t.Errorf("Status = %v, want %v", invoice.GetStatus(), InvoiceStatusDraft)
			}
		})
	}
}

func TestInvoice_AddLine(t *testing.T) {
	invoice, err := NewInvoice("INV-001", "customer-123", "USD", time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("NewInvoice() error: %v", err)
	}

	unitPrice, _ := money.NewFromFloat(100.0, "USD")
	discount, _ := money.NewFromFloat(0, "USD")

	err = invoice.AddLine("Service A", 10, unitPrice, "hours", discount)
	if err != nil {
		t.Errorf("AddLine() error: %v", err)
	}

	if len(invoice.GetLines()) != 1 {
		t.Errorf("Lines count = %d, want 1", len(invoice.GetLines()))
	}

	expectedSubtotal, _ := money.NewFromFloat(1000.0, "USD")
	if !invoice.GetSubtotal().Equals(expectedSubtotal) {
		t.Errorf("Subtotal = %v, want %v", invoice.GetSubtotal(), expectedSubtotal)
	}
}

func TestInvoice_Send(t *testing.T) {
	invoice, err := NewInvoice("INV-001", "customer-123", "USD", time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("NewInvoice() error: %v", err)
	}

	err = invoice.Send()
	if err == nil {
		t.Errorf("Send() should fail without lines")
	}

	unitPrice, _ := money.NewFromFloat(100.0, "USD")
	discount, _ := money.NewFromFloat(0, "USD")
	invoice.AddLine("Service", 1, unitPrice, "hours", discount)
	err = invoice.Send()
	if err != nil {
		t.Errorf("Send() error: %v", err)
	}

	if invoice.GetStatus() != InvoiceStatusSent {
		t.Errorf("Status = %v, want %v", invoice.GetStatus(), InvoiceStatusSent)
	}

	err = invoice.Send()
	if err == nil {
		t.Errorf("Send() should fail for sent invoice")
	}
}

func TestInvoice_RecordPayment(t *testing.T) {
	invoice, _ := NewInvoice("INV-001", "customer-123", "USD", time.Now().Add(30*24*time.Hour))
	unitPrice, _ := money.NewFromFloat(100.0, "USD")
	discount, _ := money.NewFromFloat(0, "USD")
	invoice.AddLine("Service", 1, unitPrice, "hours", discount)
	invoice.Send()

	amount, _ := money.NewFromFloat(100.0, "USD")
	_, err := invoice.RecordPayment(amount, PaymentMethodBankTransfer, "REF-001")
	if err != nil {
		t.Errorf("RecordPayment() error: %v", err)
	}

	if invoice.GetStatus() != InvoiceStatusPaid {
		t.Errorf("Status = %v, want %v", invoice.GetStatus(), InvoiceStatusPaid)
	}

	expectedPaid, _ := money.NewFromFloat(100.0, "USD")
	if !invoice.GetPaidAmount().Equals(expectedPaid) {
		t.Errorf("PaidAmount = %v, want %v", invoice.GetPaidAmount(), expectedPaid)
	}
}

func TestInvoice_Cancel(t *testing.T) {
	invoice, _ := NewInvoice("INV-001", "customer-123", "USD", time.Now().Add(30*24*time.Hour))

	err := invoice.Cancel("Customer request")
	if err != nil {
		t.Errorf("Cancel() error: %v", err)
	}

	if invoice.GetStatus() != InvoiceStatusCancelled {
		t.Errorf("Status = %v, want %v", invoice.GetStatus(), InvoiceStatusCancelled)
	}

	err = invoice.Cancel("Again")
	if err != ErrInvoiceAlreadyCancelled {
		t.Errorf("Cancel() should fail for cancelled invoice")
	}
}

func TestInvoice_StatusTransitions(t *testing.T) {
	tests := []struct {
		name string
		from InvoiceStatus
		to   InvoiceStatus
		want bool
	}{
		{"Draft to Sent", InvoiceStatusDraft, InvoiceStatusSent, true},
		{"Draft to Cancelled", InvoiceStatusDraft, InvoiceStatusCancelled, true},
		{"Draft to Paid", InvoiceStatusDraft, InvoiceStatusPaid, false},
		{"Sent to Viewed", InvoiceStatusSent, InvoiceStatusViewed, true},
		{"Sent to Paid", InvoiceStatusSent, InvoiceStatusPaid, true},
		{"Paid to Draft", InvoiceStatusPaid, InvoiceStatusDraft, false},
		{"Cancelled to Draft", InvoiceStatusCancelled, InvoiceStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Errorf("CanTransitionTo(%v -> %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
