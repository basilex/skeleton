package domain

import (
	"testing"
)

func TestNewStock(t *testing.T) {
	warehouseID := NewWarehouseID()

	tests := []struct {
		name        string
		itemID      string
		warehouseID WarehouseID
		wantErr     bool
	}{
		{
			name:        "valid stock",
			itemID:      "item-001",
			warehouseID: warehouseID,
			wantErr:     false,
		},
		{
			name:        "empty item ID",
			itemID:      "",
			warehouseID: warehouseID,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stock, err := NewStock(tt.itemID, tt.warehouseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if stock.GetItemID() != tt.itemID {
					t.Errorf("Stock.ItemID = %v, want %v", stock.GetItemID(), tt.itemID)
				}
				if stock.GetQuantity() != 0 {
					t.Errorf("Stock.Quantity = %v, want 0", stock.GetQuantity())
				}
				if stock.GetAvailableQty() != 0 {
					t.Errorf("Stock.AvailableQty = %v, want 0", stock.GetAvailableQty())
				}
			}
		})
	}
}

func TestStock_AdjustQuantity(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	movementID := NewStockMovementID()
	stock.AdjustQuantity(100, movementID)

	if stock.GetQuantity() != 100 {
		t.Errorf("Stock.Quantity = %v, want 100", stock.GetQuantity())
	}
	if stock.GetAvailableQty() != 100 {
		t.Errorf("Stock.AvailableQty = %v, want 100", stock.GetAvailableQty())
	}

	stock.AdjustQuantity(-30, movementID)
	if stock.GetQuantity() != 70 {
		t.Errorf("Stock.Quantity = %v, want 70", stock.GetQuantity())
	}
	if stock.GetAvailableQty() != 70 {
		t.Errorf("Stock.AvailableQty = %v, want 70", stock.GetAvailableQty())
	}
}

func TestStock_Reserve(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	movementID := NewStockMovementID()
	stock.AdjustQuantity(100, movementID)

	reservationID := NewStockReservationID()
	if err := stock.Reserve(30, reservationID); err != nil {
		t.Errorf("Reserve() error = %v", err)
	}

	if stock.GetReservedQty() != 30 {
		t.Errorf("Stock.ReservedQty = %v, want 30", stock.GetReservedQty())
	}
	if stock.GetAvailableQty() != 70 {
		t.Errorf("Stock.AvailableQty = %v, want 70", stock.GetAvailableQty())
	}

	reservationID2 := NewStockReservationID()
	if err := stock.Reserve(80, reservationID2); err == nil {
		t.Error("Reserve() should return error when insufficient stock")
	}
}

func TestStock_ReleaseReservation(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	movementID := NewStockMovementID()
	stock.AdjustQuantity(100, movementID)
	reservationID := NewStockReservationID()
	stock.Reserve(30, reservationID)

	stock.ReleaseReservation(10)

	if stock.GetReservedQty() != 20 {
		t.Errorf("Stock.ReservedQty = %v, want 20", stock.GetReservedQty())
	}
	if stock.GetAvailableQty() != 80 {
		t.Errorf("Stock.AvailableQty = %v, want 80", stock.GetAvailableQty())
	}
}

func TestStock_FulfillReservation(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	movementID := NewStockMovementID()
	stock.AdjustQuantity(100, movementID)
	reservationID := NewStockReservationID()
	stock.Reserve(30, reservationID)

	stock.FulfillReservation(10)

	if stock.GetQuantity() != 90 {
		t.Errorf("Stock.Quantity = %v, want 90", stock.GetQuantity())
	}
	if stock.GetReservedQty() != 20 {
		t.Errorf("Stock.ReservedQty = %v, want 20", stock.GetReservedQty())
	}
	if stock.GetAvailableQty() != 70 {
		t.Errorf("Stock.AvailableQty = %v, want 70", stock.GetAvailableQty())
	}
}

func TestStock_SetReorderPoint(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	if err := stock.SetReorderPoint(10); err != nil {
		t.Errorf("SetReorderPoint() error = %v", err)
	}
	if stock.GetReorderPoint() != 10 {
		t.Errorf("Stock.ReorderPoint = %v, want 10", stock.GetReorderPoint())
	}

	if err := stock.SetReorderPoint(-5); err == nil {
		t.Error("SetReorderPoint() should return error for negative value")
	}
}

func TestStock_NeedsReorder(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	stock.SetReorderPoint(10)

	movementID := NewStockMovementID()
	stock.AdjustQuantity(5, movementID)

	if !stock.NeedsReorder() {
		t.Error("Stock should need reorder when available qty (5) <= reorder point (10)")
	}

	stock.AdjustQuantity(10, movementID)

	if stock.NeedsReorder() {
		t.Error("Stock should not need reorder when available qty (15) > reorder point (10)")
	}
}

func TestStock_IsAvailable(t *testing.T) {
	warehouseID := NewWarehouseID()
	stock, err := NewStock("item-001", warehouseID)
	if err != nil {
		t.Fatal(err)
	}

	movementID := NewStockMovementID()
	stock.AdjustQuantity(100, movementID)

	if !stock.IsAvailable(50) {
		t.Error("Stock should be available for 50 units")
	}

	if stock.IsAvailable(150) {
		t.Error("Stock should not be available for 150 units")
	}
}
