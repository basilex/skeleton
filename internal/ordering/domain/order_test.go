package domain

import (
	"testing"

	moneypkg "github.com/basilex/skeleton/pkg/money"
)

func TestOrder_Create(t *testing.T) {
	customerID := "customer-123"
	supplierID := "supplier-456"
	contractID := "contract-789"

	order, err := NewOrder(
		"ORD-001",
		customerID,
		supplierID,
		contractID,
		"UAH",
		"user-123",
	)

	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	if order == nil {
		t.Fatal("expected order, got nil")
	}

	// Verify initial state
	if order.GetStatus() != OrderStatusDraft {
		t.Errorf("GetStatus() = %v, want %v", order.GetStatus(), OrderStatusDraft)
	}

	if order.GetCustomerID() != customerID {
		t.Errorf("GetCustomerID() = %v, want %v", order.GetCustomerID(), customerID)
	}

	if order.GetSupplierID() != supplierID {
		t.Errorf("GetSupplierID() = %v, want %v", order.GetSupplierID(), supplierID)
	}

	if order.GetContractID() != contractID {
		t.Errorf("GetContractID() = %v, want %v", order.GetContractID(), contractID)
	}

	if order.GetCurrency() != "UAH" {
		t.Errorf("GetCurrency() = %v, want UAH", order.GetCurrency())
	}

	if len(order.GetLines()) != 0 {
		t.Errorf("expected 0 lines, got %d", len(order.GetLines()))
	}

	// Verify event
	events := order.PullEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if _, ok := events[0].(OrderCreated); !ok {
		t.Errorf("expected OrderCreated event, got %T", events[0])
	}
}

func TestOrder_AddLine(t *testing.T) {
	order, _ := NewOrder("ORD-001", "customer-123", "supplier-456", "", "UAH", "user-123")

	// Create order line
	unitPrice, _ := moneypkg.NewFromFloat(100.0, "UAH")
	discount := moneypkg.Zero("UAH")

	line, err := NewOrderLine(
		order.GetID(),
		"item-123",
		"Test Item",
		10.0, // quantity
		"piece",
		unitPrice,
		discount,
	)

	if err != nil {
		t.Fatalf("NewOrderLine() error = %v", err)
	}

	// Add line to order
	err = order.AddLine(line)
	if err != nil {
		t.Errorf("AddLine() error = %v", err)
	}

	if len(order.GetLines()) != 1 {
		t.Errorf("expected 1 line, got %d", len(order.GetLines()))
	}

	expectedSubtotal, _ := moneypkg.NewFromFloat(1000.0, "UAH") // 10 * 100
	if !order.GetSubtotal().Equals(expectedSubtotal) {
		t.Errorf("GetSubtotal() = %v, want 1000", order.GetSubtotal())
	}

	expectedTotal, _ := moneypkg.NewFromFloat(1000.0, "UAH")
	if !order.GetTotal().Equals(expectedTotal) {
		t.Errorf("GetTotal() = %v, want 1000", order.GetTotal())
	}
}

func TestOrder_RemoveLine(t *testing.T) {
	order, _ := NewOrder("ORD-001", "customer-123", "supplier-456", "", "UAH", "user-123")

	unitPrice1, _ := moneypkg.NewFromFloat(100.0, "UAH")
	discount1 := moneypkg.Zero("UAH")
	line1, _ := NewOrderLine(order.GetID(), "item-1", "Item 1", 5, "piece", unitPrice1, discount1)

	unitPrice2, _ := moneypkg.NewFromFloat(50.0, "UAH")
	discount2 := moneypkg.Zero("UAH")
	line2, _ := NewOrderLine(order.GetID(), "item-2", "Item 2", 10, "piece", unitPrice2, discount2)

	order.AddLine(line1)
	order.AddLine(line2)

	if len(order.GetLines()) != 2 {
		t.Errorf("expected 2 lines, got %d", len(order.GetLines()))
	}

	// Remove first line
	err := order.RemoveLine(line1.GetID())
	if err != nil {
		t.Errorf("RemoveLine() error = %v", err)
	}

	if len(order.GetLines()) != 1 {
		t.Errorf("expected 1 line after removal, got %d", len(order.GetLines()))
	}

	// Cannot remove non-existent line
	err = order.RemoveLine(line1.GetID())
	if err == nil {
		t.Error("expected error when removing non-existent line")
	}
}

func TestOrder_Confirm(t *testing.T) {
	order, _ := NewOrder("ORD-001", "customer-123", "supplier-456", "", "UAH", "user-123")

	// Clear events from order creation
	_ = order.PullEvents()

	// Cannot confirm order without lines
	err := order.Confirm()
	if err == nil {
		t.Error("expected error when confirming order without lines")
	}

	// Add a line
	unitPrice, _ := moneypkg.NewFromFloat(100.0, "UAH")
	discount := moneypkg.Zero("UAH")
	line, _ := NewOrderLine(order.GetID(), "item-1", "Item 1", 1, "piece", unitPrice, discount)
	order.AddLine(line)

	// Now can confirm
	err = order.Confirm()
	if err != nil {
		t.Errorf("Confirm() error = %v", err)
	}

	if order.GetStatus() != OrderStatusConfirmed {
		t.Errorf("GetStatus() = %v, want %v", order.GetStatus(), OrderStatusConfirmed)
	}

	// Verify events
	events := order.PullEvents()
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	statusChange, ok := events[0].(OrderStatusChanged)
	if !ok {
		t.Errorf("expected OrderStatusChanged event, got %T", events[0])
	}

	if statusChange.OldStatus != OrderStatusDraft {
		t.Errorf("OldStatus = %v, want %v", statusChange.OldStatus, OrderStatusDraft)
	}

	if statusChange.NewStatus != OrderStatusConfirmed {
		t.Errorf("NewStatus = %v, want %v", statusChange.NewStatus, OrderStatusConfirmed)
	}
}

func TestOrder_Complete(t *testing.T) {
	order, _ := NewOrder("ORD-001", "customer-123", "supplier-456", "", "UAH", "user-123")
	unitPrice, _ := moneypkg.NewFromFloat(100.0, "UAH")
	discount := moneypkg.Zero("UAH")
	line, _ := NewOrderLine(order.GetID(), "item-1", "Item 1", 1, "piece", unitPrice, discount)
	order.AddLine(line)

	// Cannot complete draft order
	err := order.Complete()
	if err == nil {
		t.Error("expected error when completing draft order")
	}

	// Confirm first
	order.Confirm()

	// Now can complete
	err = order.Complete()
	if err != nil {
		t.Errorf("Complete() error = %v", err)
	}

	if order.GetStatus() != OrderStatusCompleted {
		t.Errorf("GetStatus() = %v, want %v", order.GetStatus(), OrderStatusCompleted)
	}

	if order.GetCompletedAt() == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestOrder_Cancel(t *testing.T) {
	order, _ := NewOrder("ORD-001", "customer-123", "supplier-456", "", "UAH", "user-123")

	// Can cancel draft order
	err := order.Cancel("customer cancellation")
	if err != nil {
		t.Errorf("Cancel() error = %v", err)
	}

	if order.GetStatus() != OrderStatusCancelled {
		t.Errorf("GetStatus() = %v, want %v", order.GetStatus(), OrderStatusCancelled)
	}

	if order.GetCancelledAt() == nil {
		t.Error("expected cancelled_at to be set")
	}

	if order.GetNotes() != "customer cancellation" {
		t.Errorf("GetNotes() = %v, want 'customer cancellation'", order.GetNotes())
	}

	// Cannot cancel completed order
	order = nil
	order, _ = NewOrder("ORD-002", "customer-123", "supplier-456", "", "UAH", "user-123")
	unitPrice, _ := moneypkg.NewFromFloat(100.0, "UAH")
	discount := moneypkg.Zero("UAH")
	line, _ := NewOrderLine(order.GetID(), "item-1", "Item 1", 1, "piece", unitPrice, discount)
	order.AddLine(line)
	order.Confirm()
	order.Complete()

	err = order.Cancel("too late")
	if err == nil {
		t.Error("expected error when cancelling completed order")
	}
}

func TestOrderLine_Quantity(t *testing.T) {
	orderID := NewOrderID()

	tests := []struct {
		name      string
		quantity  float64
		unitPrice float64
		discount  float64
		wantTotal float64
		wantErr   bool
	}{
		{
			name:      "valid line",
			quantity:  10,
			unitPrice: 100,
			discount:  50,
			wantTotal: 950, // (10 * 100) - 50
			wantErr:   false,
		},
		{
			name:      "zero quantity",
			quantity:  0,
			unitPrice: 100,
			discount:  0,
			wantTotal: 0,
			wantErr:   true,
		},
		{
			name:      "negative quantity",
			quantity:  -5,
			unitPrice: 100,
			discount:  0,
			wantTotal: 0,
			wantErr:   true,
		},
		{
			name:      "negative unit price",
			quantity:  10,
			unitPrice: -100,
			discount:  0,
			wantTotal: 0,
			wantErr:   true,
		},
		{
			name:      "discount greater than total",
			quantity:  1,
			unitPrice: 100,
			discount:  200,
			wantTotal: 0, // Floor at 0
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitPrice, err := moneypkg.NewFromFloat(tt.unitPrice, "USD")
			if err != nil {
				// If NewFromFloat returns error (e.g., negative price), that's expected for negative prices
				if tt.wantErr {
					return // Test passed - expected error
				}
				t.Errorf("NewFromFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			discount, _ := moneypkg.NewFromFloat(tt.discount, "USD")

			line, err := NewOrderLine(orderID, "item-123", "Test Item", tt.quantity, "piece", unitPrice, discount)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewOrderLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			expectedTotal, _ := moneypkg.NewFromFloat(tt.wantTotal, "USD")
			if !line.GetTotal().Equals(expectedTotal) {
				t.Errorf("GetTotal() = %v, want %v", line.GetTotal(), tt.wantTotal)
			}
		})
	}
}
