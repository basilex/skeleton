package domain

import (
	"testing"
	"time"
)

func TestNewLot(t *testing.T) {
	warehouseID := NewWarehouseID()
	manufacturingDate := time.Now().AddDate(-1, 0, 0)
	expiryDate := time.Now().AddDate(1, 0, 0)

	t.Run("creates lot with valid data", func(t *testing.T) {
		lot, err := NewLot("item-123", warehouseID, "LOT-001", 100, &manufacturingDate, &expiryDate)
		if err != nil {
			t.Fatalf("NewLot() error = %v", err)
		}

		if lot.GetLotNumber() != "LOT-001" {
			t.Errorf("lot number = %v, want LOT-001", lot.GetLotNumber())
		}

		if lot.GetQuantity() != 100 {
			t.Errorf("quantity = %v, want 100", lot.GetQuantity())
		}

		if lot.GetStatus() != LotStatusActive {
			t.Errorf("status = %v, want %v", lot.GetStatus(), LotStatusActive)
		}

		events := lot.PullEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("fails with empty item ID", func(t *testing.T) {
		_, err := NewLot("", warehouseID, "LOT-001", 100, nil, nil)
		if err == nil {
			t.Error("expected error for empty item ID")
		}
	})

	t.Run("fails with empty lot number", func(t *testing.T) {
		_, err := NewLot("item-123", warehouseID, "", 100, nil, nil)
		if err == nil {
			t.Error("expected error for empty lot number")
		}
	})

	t.Run("fails with non-positive quantity", func(t *testing.T) {
		_, err := NewLot("item-123", warehouseID, "LOT-001", 0, nil, nil)
		if err == nil {
			t.Error("expected error for zero quantity")
		}
	})
}

func TestLot_SerialNumbers(t *testing.T) {
	warehouseID := NewWarehouseID()
	lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, nil)

	t.Run("add serial number", func(t *testing.T) {
		err := lot.AddSerialNumber("SN-001")
		if err != nil {
			t.Errorf("AddSerialNumber() error = %v", err)
		}

		serials := lot.GetSerialNumbers()
		if len(serials) != 1 {
			t.Errorf("serial numbers count = %d, want 1", len(serials))
		}
	})

	t.Run("fails with duplicate serial number", func(t *testing.T) {
		err := lot.AddSerialNumber("SN-001")
		if err == nil {
			t.Error("expected error for duplicate serial number")
		}
	})

	t.Run("reserve serial number", func(t *testing.T) {
		err := lot.ReserveSerialNumber("SN-001", "user-123")
		if err != nil {
			t.Errorf("ReserveSerialNumber() error = %v", err)
		}

		serials := lot.GetSerialNumbers()
		if serials[0].GetStatus() != SerialNumberReserved {
			t.Errorf("status = %v, want %v", serials[0].GetStatus(), SerialNumberReserved)
		}
	})

	t.Run("fails to reserve already reserved serial", func(t *testing.T) {
		err := lot.ReserveSerialNumber("SN-001", "user-456")
		if err == nil {
			t.Error("expected error for already reserved serial")
		}
	})

	t.Run("release serial number", func(t *testing.T) {
		err := lot.ReleaseSerialNumber("SN-001")
		if err != nil {
			t.Errorf("ReleaseSerialNumber() error = %v", err)
		}

		serials := lot.GetSerialNumbers()
		if serials[0].GetStatus() != SerialNumberAvailable {
			t.Errorf("status = %v, want %v", serials[0].GetStatus(), SerialNumberAvailable)
		}
	})
}

func TestLot_AdjustQuantity(t *testing.T) {
	warehouseID := NewWarehouseID()
	lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, nil)

	t.Run("adjust quantity positively", func(t *testing.T) {
		err := lot.AdjustQuantity(50)
		if err != nil {
			t.Errorf("AdjustQuantity() error = %v", err)
		}

		if lot.GetQuantity() != 150 {
			t.Errorf("quantity = %v, want 150", lot.GetQuantity())
		}
	})

	t.Run("adjust quantity negatively", func(t *testing.T) {
		err := lot.AdjustQuantity(-50)
		if err != nil {
			t.Errorf("AdjustQuantity() error = %v", err)
		}

		if lot.GetQuantity() != 100 {
			t.Errorf("quantity = %v, want 100", lot.GetQuantity())
		}
	})

	t.Run("fails with negative result", func(t *testing.T) {
		err := lot.AdjustQuantity(-200)
		if err == nil {
			t.Error("expected error for negative quantity")
		}
	})

	t.Run("marks depleted when zero", func(t *testing.T) {
		lot2, _ := NewLot("item-456", warehouseID, "LOT-002", 10, nil, nil)
		_ = lot2.AdjustQuantity(-10)

		if lot2.GetStatus() != LotStatusDepleted {
			t.Errorf("status = %v, want %v", lot2.GetStatus(), LotStatusDepleted)
		}
	})
}

func TestLot_Expiry(t *testing.T) {
	warehouseID := NewWarehouseID()

	t.Run("lot not expired", func(t *testing.T) {
		futureExpiry := time.Now().AddDate(0, 0, 30)
		lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, &futureExpiry)

		if lot.IsExpired() {
			t.Error("lot should not be expired")
		}
	})

	t.Run("lot expired", func(t *testing.T) {
		pastExpiry := time.Now().AddDate(0, 0, -1)
		lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, &pastExpiry)

		if !lot.IsExpired() {
			t.Error("lot should be expired")
		}
	})

	t.Run("expire within days", func(t *testing.T) {
		futureExpiry := time.Now().AddDate(0, 0, 15)
		lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, &futureExpiry)

		if !lot.IsExpiredWithin(30) {
			t.Error("lot should expire within 30 days")
		}

		if lot.IsExpiredWithin(10) {
			t.Error("lot should not expire within 10 days")
		}
	})

	t.Run("mark expired", func(t *testing.T) {
		lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, nil)
		lot.MarkExpired()

		if lot.GetStatus() != LotStatusExpired {
			t.Errorf("status = %v, want %v", lot.GetStatus(), LotStatusExpired)
		}
	})
}

func TestLot_Location(t *testing.T) {
	warehouseID := NewWarehouseID()
	lot, _ := NewLot("item-123", warehouseID, "LOT-001", 100, nil, nil)
	location, _ := NewStockLocation(warehouseID, "A", "01", "001")

	lot.SetLocation(location)

	if lot.GetLocation() == nil {
		t.Error("expected location to be set")
	}

	if lot.GetLocation().GetZone() != "A" {
		t.Errorf("zone = %v, want A", lot.GetLocation().GetZone())
	}
}

func TestStockTake(t *testing.T) {
	warehouseID := NewWarehouseID()

	t.Run("creates stock take", func(t *testing.T) {
		stockTake, err := NewStockTake(warehouseID, "ST-2024-001", "user-123")
		if err != nil {
			t.Fatalf("NewStockTake() error = %v", err)
		}

		if stockTake.GetReference() != "ST-2024-001" {
			t.Errorf("reference = %v, want ST-2024-001", stockTake.GetReference())
		}

		if stockTake.GetStatus() != StockTakeStatusPending {
			t.Errorf("status = %v, want %v", stockTake.GetStatus(), StockTakeStatusPending)
		}

		events := stockTake.PullEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("fails with empty reference", func(t *testing.T) {
		_, err := NewStockTake(warehouseID, "", "user-123")
		if err == nil {
			t.Error("expected error for empty reference")
		}
	})

	t.Run("add items", func(t *testing.T) {
		stockTake, _ := NewStockTake(warehouseID, "ST-001", "user-123")

		err := stockTake.AddItem("item-1", 100)
		if err != nil {
			t.Errorf("AddItem() error = %v", err)
		}

		if stockTake.GetTotalItems() != 1 {
			t.Errorf("total items = %d, want 1", stockTake.GetTotalItems())
		}
	})

	t.Run("count items", func(t *testing.T) {
		stockTake, _ := NewStockTake(warehouseID, "ST-001", "user-123")
		stockTake.AddItem("item-1", 100)

		err := stockTake.Start("user-456")
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}

		err = stockTake.CountItem("item-1", 98, "counter-1")
		if err != nil {
			t.Errorf("CountItem() error = %v", err)
		}

		lines := stockTake.GetLines()
		line := lines["item-1"]
		if line.GetVariance() != -2 {
			t.Errorf("variance = %v, want -2", line.GetVariance())
		}

		if stockTake.GetCountedItems() != 1 {
			t.Errorf("counted items = %d, want 1", stockTake.GetCountedItems())
		}
	})

	t.Run("complete stock take", func(t *testing.T) {
		stockTake, _ := NewStockTake(warehouseID, "ST-001", "user-123")
		stockTake.AddItem("item-1", 100)
		stockTake.Start("user-456")
		stockTake.CountItem("item-1", 100, "counter-1")

		err := stockTake.Complete("user-789")
		if err != nil {
			t.Errorf("Complete() error = %v", err)
		}

		if stockTake.GetStatus() != StockTakeStatusCompleted {
			t.Errorf("status = %v, want %v", stockTake.GetStatus(), StockTakeStatusCompleted)
		}
	})

	t.Run("complete with uncounted items fails", func(t *testing.T) {
		stockTake, _ := NewStockTake(warehouseID, "ST-001", "user-123")
		stockTake.AddItem("item-1", 100)
		stockTake.AddItem("item-2", 50)
		stockTake.Start("user-456")
		stockTake.CountItem("item-1", 100, "counter-1")
		// item-2 not counted

		err := stockTake.Complete("user-789")
		if err == nil {
			t.Error("expected error for uncounted items")
		}
	})

	t.Run("variance detection", func(t *testing.T) {
		stockTake, _ := NewStockTake(warehouseID, "ST-001", "user-123")
		stockTake.AddItem("item-1", 100)
		stockTake.AddItem("item-2", 50)
		stockTake.Start("user-456")
		stockTake.CountItem("item-1", 98, "counter-1")
		stockTake.CountItem("item-2", 50, "counter-1")

		if !stockTake.HasVariance() {
			t.Error("expected stock take to have variance")
		}

		if stockTake.GetVarianceCount() != 1 {
			t.Errorf("variance count = %d, want 1", stockTake.GetVarianceCount())
		}

		variances := stockTake.GetVariances()
		if variances["item-1"] != -2 {
			t.Errorf("item-1 variance = %v, want -2", variances["item-1"])
		}
	})
}

func TestStockLocation(t *testing.T) {
	warehouseID := NewWarehouseID()

	t.Run("creates location", func(t *testing.T) {
		location, err := NewStockLocation(warehouseID, "A", "01", "001")
		if err != nil {
			t.Fatalf("NewStockLocation() error = %v", err)
		}

		if location.GetZone() != "A" {
			t.Errorf("zone = %v, want A", location.GetZone())
		}
		if location.GetAisle() != "01" {
			t.Errorf("aisle = %v, want 01", location.GetAisle())
		}
	})

	t.Run("fails with empty zone", func(t *testing.T) {
		_, err := NewStockLocation(warehouseID, "", "01", "001")
		if err == nil {
			t.Error("expected error for empty zone")
		}
	})

	t.Run("string representation", func(t *testing.T) {
		location, _ := NewStockLocation(warehouseID, "A", "01", "001")

		shortCode := location.ShortCode()
		if shortCode != "A/01/001" {
			t.Errorf("short code = %v, want A/01/001", shortCode)
		}
	})

	t.Run("equals comparison", func(t *testing.T) {
		loc1, _ := NewStockLocation(warehouseID, "A", "01", "001")
		loc2, _ := NewStockLocation(warehouseID, "A", "01", "001")
		loc3, _ := NewStockLocation(warehouseID, "B", "02", "002")

		if !loc1.Equals(loc2) {
			t.Error("locations should be equal")
		}
		if loc1.Equals(loc3) {
			t.Error("locations should not be equal")
		}
	})
}
