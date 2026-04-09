package domain

import (
	"testing"
)

func TestNewWarehouse(t *testing.T) {
	tests := []struct {
		name          string
		warehouseName string
		code          string
		location      string
		wantErr       error
	}{
		{
			name:          "valid warehouse",
			warehouseName: "Main Warehouse",
			code:          "WH-001",
			location:      "New York",
			wantErr:       nil,
		},
		{
			name:          "empty name",
			warehouseName: "",
			code:          "WH-002",
			location:      "Boston",
			wantErr:       ErrWarehouseNameEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warehouse, err := NewWarehouse(tt.warehouseName, tt.code, tt.location)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewWarehouse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewWarehouse() unexpected error = %v", err)
				return
			}
			if warehouse.GetName() != tt.warehouseName {
				t.Errorf("Warehouse.Name = %v, want %v", warehouse.GetName(), tt.warehouseName)
			}
			if warehouse.GetCode() != tt.code {
				t.Errorf("Warehouse.Code = %v, want %v", warehouse.GetCode(), tt.code)
			}
			if warehouse.GetStatus() != WarehouseStatusActive {
				t.Errorf("Warehouse.Status = %v, want %v", warehouse.GetStatus(), WarehouseStatusActive)
			}
		})
	}
}

func TestWarehouse_Activate(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	warehouse.Deactivate()
	if err := warehouse.Activate(); err != nil {
		t.Errorf("Activate() error = %v", err)
	}
	if warehouse.GetStatus() != WarehouseStatusActive {
		t.Errorf("Warehouse.Status = %v, want %v", warehouse.GetStatus(), WarehouseStatusActive)
	}
}

func TestWarehouse_Deactivate(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	if err := warehouse.Deactivate(); err != nil {
		t.Errorf("Deactivate() error = %v", err)
	}
	if warehouse.GetStatus() != WarehouseStatusInactive {
		t.Errorf("Warehouse.Status = %v, want %v", warehouse.GetStatus(), WarehouseStatusInactive)
	}
}

func TestWarehouse_SetMaintenance(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	if err := warehouse.SetMaintenance(); err != nil {
		t.Errorf("SetMaintenance() error = %v", err)
	}
	if warehouse.GetStatus() != WarehouseStatusMaintenance {
		t.Errorf("Warehouse.Status = %v, want %v", warehouse.GetStatus(), WarehouseStatusMaintenance)
	}
}

func TestWarehouse_SetCapacity(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		capacity float64
		wantErr  bool
	}{
		{
			name:     "valid capacity",
			capacity: 1000.0,
			wantErr:  false,
		},
		{
			name:     "negative capacity",
			capacity: -100.0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := warehouse.SetCapacity(tt.capacity)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCapacity() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && warehouse.GetCapacity() != tt.capacity {
				t.Errorf("Warehouse.Capacity = %v, want %v", warehouse.GetCapacity(), tt.capacity)
			}
		})
	}
}

func TestWarehouse_IsActive(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	if !warehouse.IsActive() {
		t.Error("Warehouse should be active by default")
	}

	warehouse.Deactivate()
	if warehouse.IsActive() {
		t.Error("Warehouse should not be active after deactivation")
	}
}

func TestWarehouse_PullEvents(t *testing.T) {
	warehouse, err := NewWarehouse("Test Warehouse", "WH-001", "NY")
	if err != nil {
		t.Fatal(err)
	}

	events := warehouse.PullEvents()
	if len(events) != 1 {
		t.Errorf("PullEvents() returned %d events, want 1", len(events))
	}

	if len(warehouse.PullEvents()) != 0 {
		t.Error("PullEvents() should return empty slice on second call")
	}
}
