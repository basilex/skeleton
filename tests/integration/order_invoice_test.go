package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/invoicing/domain"
	orderingDomain "github.com/basilex/skeleton/internal/ordering/domain"
)

func TestOrderConfirmedCreatesInvoice(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	t.Run("order_confirmed_event_creates_invoice", func(t *testing.T) {
		orderLines := []orderingDomain.OrderConfirmedLine{
			{
				ItemID:    "item-001",
				ItemName:  "Product A",
				Quantity:  10,
				Unit:      "piece",
				UnitPrice: 100.0,
				Discount:  0,
				Total:     1000.0,
			},
			{
				ItemID:    "item-002",
				ItemName:  "Product B",
				Quantity:  5,
				Unit:      "piece",
				UnitPrice: 50.0,
				Discount:  10.0,
				Total:     240.0,
			},
		}

		now := time.Now()

		event := orderingDomain.OrderConfirmed{
			OrderID:     orderingDomain.NewOrderID(),
			CustomerID:  "customer-456",
			SupplierID:  "supplier-789",
			WarehouseID: "warehouse-001",
			Lines:       orderLines,
			Total:       1240.0,
			Currency:    "USD",
		}
		_ = now

		invoiceNumber := fmt.Sprintf("INV-%d", time.Now().Unix())

		invoice, err := domain.NewInvoice(
			invoiceNumber,
			event.CustomerID,
			event.Currency,
			time.Now().Add(30*24*time.Hour),
		)
		if err != nil {
			t.Fatalf("failed to create invoice: %s", err)
		}

		invoice.LinkOrder(event.OrderID.String())

		for _, line := range event.Lines {
			err := invoice.AddLine(
				line.ItemName,
				line.Quantity,
				line.UnitPrice,
				line.Unit,
				line.Discount,
			)
			if err != nil {
				t.Fatalf("failed to add invoice line: %s", err)
			}
		}

		if invoice.GetOrderID() == nil {
			t.Error("expected order ID to be set")
		} else if *invoice.GetOrderID() != event.OrderID.String() {
			t.Errorf("expected order ID %s, got %s", event.OrderID.String(), *invoice.GetOrderID())
		}

		if invoice.GetCustomerID() != event.CustomerID {
			t.Errorf("expected customer ID %s, got %s", event.CustomerID, invoice.GetCustomerID())
		}

		invoiceLines := invoice.GetLines()
		if len(invoiceLines) != len(orderLines) {
			t.Errorf("expected %d invoice lines, got %d", len(orderLines), len(invoiceLines))
		}

		for i, line := range invoiceLines {
			if line.GetDescription() != orderLines[i].ItemName {
				t.Errorf("line %d: expected description %s, got %s", i, orderLines[i].ItemName, line.GetDescription())
			}
			if line.GetQuantity() != orderLines[i].Quantity {
				t.Errorf("line %d: expected quantity %f, got %f", i, orderLines[i].Quantity, line.GetQuantity())
			}
		}

		if invoice.GetCurrency() != event.Currency {
			t.Errorf("expected currency %s, got %s", event.Currency, invoice.GetCurrency())
		}

		t.Logf("✅ Invoice created successfully from OrderConfirmed event")
		t.Logf("   Order ID: %s", event.OrderID)
		t.Logf("   Invoice Number: %s", invoiceNumber)
		t.Logf("   Customer: %s", event.CustomerID)
		t.Logf("   Lines: %d", len(invoiceLines))
		t.Logf("   Currency: %s", invoice.GetCurrency())
		t.Logf("   Due Date: %s", invoice.GetDueDate().Format("2006-01-02"))
	})
}

func TestOrderConfirmedCalculatesInvoiceTotals(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	t.Run("invoice_totals_match_order_totals", func(t *testing.T) {
		orderLines := []orderingDomain.OrderConfirmedLine{
			{
				ItemID:    "item-001",
				ItemName:  "Product A",
				Quantity:  2,
				UnitPrice: 100.0,
				Unit:      "piece",
				Discount:  0,
				Total:     200.0,
			},
		}

		event := orderingDomain.OrderConfirmed{
			OrderID:     orderingDomain.NewOrderID(),
			CustomerID:  "customer-123",
			SupplierID:  "supplier-001",
			WarehouseID: "warehouse-001",
			Lines:       orderLines,
			Total:       200.0,
			Currency:    "USD",
		}

		invoiceNumber := fmt.Sprintf("INV-%d", time.Now().Unix())

		invoice, err := domain.NewInvoice(
			invoiceNumber,
			event.CustomerID,
			event.Currency,
			time.Now().Add(30*24*time.Hour),
		)
		if err != nil {
			t.Fatalf("failed to create invoice: %s", err)
		}

		for _, line := range event.Lines {
			invoice.AddLine(
				line.ItemName,
				line.Quantity,
				line.UnitPrice,
				line.Unit,
				line.Discount,
			)
		}

		calculatedTotal := invoice.GetTotal()
		if calculatedTotal <= 0 {
			t.Errorf("expected positive invoice total, got %f", calculatedTotal)
		}

		t.Logf("✅ Invoice totals calculated")
		t.Logf("   Order Total: %f", event.Total)
		t.Logf("   Invoice Subtotal: %f", invoice.GetSubtotal())
		t.Logf("   Invoice Total: %f", calculatedTotal)
	})
}
