package domain

import (
	"testing"
)

func TestNewStockMovement(t *testing.T) {
	tests := []struct {
		name         string
		movementType MovementType
		itemID       string
		quantity     float64
		wantErr      bool
	}{
		{
			name:         "valid movement",
			movementType: MovementTypeReceipt,
			itemID:       "item-001",
			quantity:     100,
			wantErr:      false,
		},
		{
			name:         "negative quantity",
			movementType: MovementTypeReceipt,
			itemID:       "item-001",
			quantity:     -10,
			wantErr:      true,
		},
		{
			name:         "zero quantity",
			movementType: MovementTypeReceipt,
			itemID:       "item-001",
			quantity:     0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movement, err := NewStockMovement(tt.movementType, tt.itemID, tt.quantity)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStockMovement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if movement.GetItemID() != tt.itemID {
					t.Errorf("Movement.ItemID = %v, want %v", movement.GetItemID(), tt.itemID)
				}
				if movement.GetMovementType() != tt.movementType {
					t.Errorf("Movement.Type = %v, want %v", movement.GetMovementType(), tt.movementType)
				}
			}
		})
	}
}

func TestNewReceipt(t *testing.T) {
	warehouseID := NewWarehouseID()
	movement, err := NewReceipt("item-001", warehouseID, 100, "receipt-001")
	if err != nil {
		t.Fatal(err)
	}

	if movement.GetMovementType() != MovementTypeReceipt {
		t.Errorf("Movement.Type = %v, want %v", movement.GetMovementType(), MovementTypeReceipt)
	}
	if movement.GetToWarehouse() != warehouseID {
		t.Errorf("Movement.ToWarehouse = %v, want %v", movement.GetToWarehouse(), warehouseID)
	}
	if movement.GetReferenceID() != "receipt-001" {
		t.Errorf("Movement.ReferenceID = %v, want receipt-001", movement.GetReferenceID())
	}
	if movement.GetReferenceType() != "receipt" {
		t.Errorf("Movement.ReferenceType = %v, want receipt", movement.GetReferenceType())
	}
	if !movement.IsInbound() {
		t.Error("Receipt should be inbound")
	}
	if movement.IsOutbound() {
		t.Error("Receipt should not be outbound")
	}
}

func TestNewIssue(t *testing.T) {
	warehouseID := NewWarehouseID()
	movement, err := NewIssue("item-001", warehouseID, 50, "order-001")
	if err != nil {
		t.Fatal(err)
	}

	if movement.GetMovementType() != MovementTypeIssue {
		t.Errorf("Movement.Type = %v, want %v", movement.GetMovementType(), MovementTypeIssue)
	}
	if movement.GetFromWarehouse() != warehouseID {
		t.Errorf("Movement.FromWarehouse = %v, want %v", movement.GetFromWarehouse(), warehouseID)
	}
	if movement.GetReferenceID() != "order-001" {
		t.Errorf("Movement.ReferenceID = %v, want order-001", movement.GetReferenceID())
	}
	if movement.GetReferenceType() != "order" {
		t.Errorf("Movement.ReferenceType = %v, want order", movement.GetReferenceType())
	}
	if movement.IsInbound() {
		t.Error("Issue should not be inbound")
	}
	if !movement.IsOutbound() {
		t.Error("Issue should be outbound")
	}
}

func TestNewTransfer(t *testing.T) {
	fromWarehouseID := NewWarehouseID()
	toWarehouseID := NewWarehouseID()

	movement, err := NewTransfer("item-001", fromWarehouseID, toWarehouseID, 30)
	if err != nil {
		t.Fatal(err)
	}

	if movement.GetMovementType() != MovementTypeTransfer {
		t.Errorf("Movement.Type = %v, want %v", movement.GetMovementType(), MovementTypeTransfer)
	}
	if movement.GetFromWarehouse() != fromWarehouseID {
		t.Errorf("Movement.FromWarehouse = %v, want %v", movement.GetFromWarehouse(), fromWarehouseID)
	}
	if movement.GetToWarehouse() != toWarehouseID {
		t.Errorf("Movement.ToWarehouse = %v, want %v", movement.GetToWarehouse(), toWarehouseID)
	}

	if _, err := NewTransfer("item-001", fromWarehouseID, fromWarehouseID, 30); err == nil {
		t.Error("NewTransfer should return error when from and to warehouses are the same")
	}
}

func TestNewAdjustment(t *testing.T) {
	warehouseID := NewWarehouseID()
	movement, err := NewAdjustment("item-001", warehouseID, 50, "inventory correction")
	if err != nil {
		t.Fatal(err)
	}

	if movement.GetMovementType() != MovementTypeAdjustment {
		t.Errorf("Movement.Type = %v, want %v", movement.GetMovementType(), MovementTypeAdjustment)
	}
	if movement.GetFromWarehouse() != warehouseID {
		t.Errorf("Movement.FromWarehouse = %v, want %v", movement.GetFromWarehouse(), warehouseID)
	}
	if movement.GetToWarehouse() != warehouseID {
		t.Errorf("Movement.ToWarehouse = %v, want %v", movement.GetToWarehouse(), warehouseID)
	}
	if movement.GetNotes() != "inventory correction" {
		t.Errorf("Movement.Notes = %v, want inventory correction", movement.GetNotes())
	}
	if movement.GetReferenceType() != "adjustment" {
		t.Errorf("Movement.ReferenceType = %v, want adjustment", movement.GetReferenceType())
	}
}
