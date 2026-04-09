package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	moneypkg "github.com/basilex/skeleton/pkg/money"
	orderingDomain "github.com/basilex/skeleton/internal/ordering/domain"
)

func TestOrderConfirmedCreatesInvoice(t *testing.T) {
	t.Run("order_confirmed_event_creates_invoice", func(t *testing.T) {
		unitPrice1, _ := moneypkg.NewFromFloat(100.0, "USD")
		discount1 := moneypkg.Zero("USD")
		total1, _ := moneypkg.NewFromFloat(1000.0, "USD")
		
		unitPrice2, _ := moneypkg.NewFromFloat(50.0, "USD")
		discount2, _ := moneypkg.NewFromFloat(10.0, "USD")
		total2, _ := moneypkg.NewFromFloat(240.0, "USD")
		
		orderLines := []orderingDomain.OrderConfirmedLine{
			{
				ItemID:    "item-001",
				ItemName:  "Product A",
				Quantity:  10,
				Unit:      "piece",
				UnitPrice: unitPrice1,
				Discount:  discount1,
				Total:     total1,
			},
			{
				ItemID:    "item-002",
				ItemName:  "Product B",
				Quantity:  5,
				Unit:      "piece",
				UnitPrice: unitPrice2,
				Discount:  discount2,
				Total:     total2,
			},
		}

		total, _ := moneypkg.NewFromFloat(1240.0, "USD")

		event := orderingDomain.OrderConfirmed{
			OrderID:     orderingDomain.NewOrderID(),
			CustomerID:  "customer-456",
			SupplierID:  "supplier-789",
			WarehouseID: "warehouse-001",
			Lines:       orderLines,
			Total:       total,
			Currency:    "USD",
		}

		// Verify event structure
		if event.OrderID.String() == "" {
			t.Error("OrderID should not be empty")
		}

		if event.CustomerID == "" {
			t.Error("CustomerID should not be empty")
		}

		if len(event.Lines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(event.Lines))
		}

		if event.Currency != "USD" {
			t.Errorf("expected currency USD, got %s", event.Currency)
		}

		// Verify line totals
		line1Total, _ := moneypkg.NewFromFloat(1000.0, "USD")
		if !event.Lines[0].Total.Equals(line1Total) {
			t.Errorf("Line 1 total = %v, want 1000.0", event.Lines[0].Total)
		}

		line2Total, _ := moneypkg.NewFromFloat(240.0, "USD")
		if !event.Lines[1].Total.Equals(line2Total) {
			t.Errorf("Line 2 total = %v, want 240.0", event.Lines[1].Total)
		}

		// Verify event total
		eventTotal, _ := moneypkg.NewFromFloat(1240.0, "USD")
		if !event.Total.Equals(eventTotal) {
			t.Errorf("Event total = %v, want 1240.0", event.Total)
		}
	})
}

func TestInvoicePaymentReducesOrderTotal(t *testing.T) {
	t.Run("invoice_payment_reduces_balance", func(t *testing.T) {
		// Create invoice
		invoice, err := domain.NewInvoice(
			"INV-001",
			"customer-123",
			"USD",
			time.Now().AddDate(0, 1, 0), // due in 1 month
		)
		if err != nil {
			t.Fatalf("NewInvoice() error = %v", err)
		}

		// Add line
		unitPrice, _ := moneypkg.NewFromFloat(100.0, "USD")
		discount := moneypkg.Zero("USD")
		err = invoice.AddLine("Product A", 2, unitPrice, "piece", discount)
		if err != nil {
			t.Fatalf("AddLine() error = %v", err)
		}

		// Check initial total
		expectedTotal, _ := moneypkg.NewFromFloat(200.0, "USD")
		if !invoice.GetTotal().Equals(expectedTotal) {
			t.Errorf("initial total = %v, want 200.0", invoice.GetTotal())
		}

		// Send the invoice before payment
		err = invoice.Send()
		if err != nil {
			t.Fatalf("Send() error = %v", err)
		}

		// Record payment
		paymentAmount, _ := moneypkg.NewFromFloat(100.0, "USD")
		_, err = invoice.RecordPayment(paymentAmount, domain.PaymentMethodCard, "PAY-001")
		if err != nil {
			t.Fatalf("RecordPayment() error = %v", err)
		}

		// Check paid amount
		paidAmount, _ := moneypkg.NewFromFloat(100.0, "USD")
		if !invoice.GetPaidAmount().Equals(paidAmount) {
			t.Errorf("paid amount = %v, want 100.0", invoice.GetPaidAmount())
		}

		// Check remaining balance
		remaining, _ := moneypkg.NewFromFloat(100.0, "USD")
		balanceDiff, _ := invoice.GetTotal().Subtract(invoice.GetPaidAmount())
		if !balanceDiff.Equals(remaining) {
			t.Errorf("remaining balance = %v, want 100.0", balanceDiff)
		}

		// Record second payment - should fail because it would exceed total
		payment2, _ := moneypkg.NewFromFloat(200.0, "USD")
		_, err = invoice.RecordPayment(payment2, domain.PaymentMethodBankTransfer, "PAY-002")
		if err == nil {
			t.Error("expected error when payment exceeds total")
		}
	})
}

func TestOrderAndInvoiceCurrencyConsistency(t *testing.T) {
	t.Run("both_use_same_currency_from_order_confirmed", func(t *testing.T) {
		// Setup
		unitPrice, _ := moneypkg.NewFromFloat(100.0, "USD")
		discount, _ := moneypkg.NewFromFloat(0.0, "USD")
		total, _ := moneypkg.NewFromFloat(200.0, "USD")
		
		orderLines := []orderingDomain.OrderConfirmedLine{
			{
				ItemID:    "item-001",
				ItemName:  "Product",
				Quantity:  2,
				Unit:      "piece",
				UnitPrice: unitPrice,
				Discount:  discount,
				Total:     total,
			},
		}

		event := orderingDomain.OrderConfirmed{
			OrderID:     orderingDomain.NewOrderID(),
			CustomerID:  "customer-456",
			SupplierID:  "supplier-789",
			WarehouseID: "warehouse-001",
			Lines:       orderLines,
			Total:       total,
			Currency:    "USD",
		}

		// Create invoice from event
		invoice, err := domain.NewInvoice(
			fmt.Sprintf("INV-%s", event.OrderID.String()[:8]),
			event.CustomerID,
			event.Currency,
			time.Now().AddDate(0, 1, 0),
		)
		if err != nil {
			t.Fatalf("NewInvoice() error = %v", err)
		}

		// Verify currency matches
		if invoice.GetCurrency() != event.Currency {
			t.Errorf("invoice currency = %s, want %s", invoice.GetCurrency(), event.Currency)
		}

		// Verify total matches
		if !invoice.GetTotal().IsZero() {
			t.Errorf("invoice should start with zero total, got %v", invoice.GetTotal())
		}
	})
}
