package catalog

import (
	"testing"
)

func TestNewItem(t *testing.T) {
	categoryID := NewCategoryID()

	tests := []struct {
		name        string
		categoryID  *CategoryID
		sku         string
		itemName    string
		description string
		basePrice   float64
		currency    string
		wantErr     bool
	}{
		{
			name:        "valid item with category",
			categoryID:  &categoryID,
			sku:         "SKU-123",
			itemName:    "Test Product",
			description: "Test Description",
			basePrice:   100.50,
			currency:    "UAH",
			wantErr:     false,
		},
		{
			name:        "valid item without category",
			categoryID:  nil,
			sku:         "SKU-456",
			itemName:    "Another Product",
			description: "Description",
			basePrice:   50.0,
			currency:    "USD",
			wantErr:     false,
		},
		{
			name:        "empty name",
			categoryID:  &categoryID,
			sku:         "SKU-789",
			itemName:    "",
			description: "Description",
			basePrice:   100.0,
			currency:    "UAH",
			wantErr:     true,
		},
		{
			name:        "negative price",
			categoryID:  &categoryID,
			sku:         "SKU-000",
			itemName:    "Bad Price Product",
			description: "Description",
			basePrice:   -10.0,
			currency:    "UAH",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItem(
				tt.categoryID,
				tt.sku,
				tt.itemName,
				tt.description,
				tt.basePrice,
				tt.currency,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if item == nil {
				t.Fatal("expected item, got nil")
			}

			if item.GetName() != tt.itemName {
				t.Errorf("GetName() = %v, want %v", item.GetName(), tt.itemName)
			}

			if item.GetSKU() != tt.sku {
				t.Errorf("GetSKU() = %v, want %v", item.GetSKU(), tt.sku)
			}

			if item.GetBasePrice() != tt.basePrice {
				t.Errorf("GetBasePrice() = %v, want %v", item.GetBasePrice(), tt.basePrice)
			}

			if item.GetStatus() != ItemStatusActive {
				t.Errorf("GetStatus() = %v, want %v", item.GetStatus(), ItemStatusActive)
			}

			if tt.categoryID == nil {
				if item.GetCategoryID() != nil {
					t.Error("expected category ID to be nil")
				}
			} else {
				if *item.GetCategoryID() != *tt.categoryID {
					t.Errorf("GetCategoryID() = %v, want %v", *item.GetCategoryID(), *tt.categoryID)
				}
			}
		})
	}
}

func TestItem_UpdatePrice(t *testing.T) {
	item, _ := NewItem(nil, "SKU-123", "Test Product", "Description", 100.0, "UAH")

	// Update price
	err := item.UpdatePrice(150.0)
	if err != nil {
		t.Errorf("UpdatePrice() error = %v", err)
	}

	if item.GetBasePrice() != 150.0 {
		t.Errorf("GetBasePrice() = %v, want 150.0", item.GetBasePrice())
	}

	// Cannot update to negative price
	err = item.UpdatePrice(-10.0)
	if err == nil {
		t.Error("expected error when updating to negative price")
	}
}

func TestItem_StatusTransitions(t *testing.T) {
	item, _ := NewItem(nil, "SKU-123", "Test Product", "Description", 100.0, "UAH")

	// Start as active
	if item.GetStatus() != ItemStatusActive {
		t.Errorf("initial status = %v, want %v", item.GetStatus(), ItemStatusActive)
	}

	// Deactivate
	item.Deactivate()
	if item.GetStatus() != ItemStatusInactive {
		t.Errorf("after deactivate, status = %v, want %v",
			item.GetStatus(), ItemStatusInactive)
	}

	// Activate again
	item.Activate()
	if item.GetStatus() != ItemStatusActive {
		t.Errorf("after activate, status = %v, want %v",
			item.GetStatus(), ItemStatusActive)
	}

	// Discontinue
	item.Discontinue()
	if item.GetStatus() != ItemStatusDiscontinued {
		t.Errorf("after discontinue, status = %v, want %v",
			item.GetStatus(), ItemStatusDiscontinued)
	}
}

func TestItem_Attributes(t *testing.T) {
	item, _ := NewItem(nil, "SKU-123", "Test Product", "Description", 100.0, "UAH")

	// Set attribute
	item.SetAttribute("color", "red")
	item.SetAttribute("size", "large")

	attrs := item.GetAttributes()
	if attrs["color"] != "red" {
		t.Errorf("color attribute = %v, want red", attrs["color"])
	}

	if attrs["size"] != "large" {
		t.Errorf("size attribute = %v, want large", attrs["size"])
	}

	// Remove attribute
	item.RemoveAttribute("color")

	attrs = item.GetAttributes()
	if _, exists := attrs["color"]; exists {
		t.Error("expected color attribute to be removed")
	}
}

func TestCategory_Hierarchy(t *testing.T) {
	parentCategory, _ := NewCategory("Electronics", "Electronic devices", nil)
	childCategory, err := NewCategory("Laptops", "Laptop computers", &parentCategory.id)

	if err != nil {
		t.Fatalf("NewCategory() error = %v", err)
	}

	// Verify hierarchy
	if childCategory.GetParentID() == nil {
		t.Fatal("expected parent ID to be set")
	}

	if *childCategory.GetParentID() != parentCategory.GetID() {
		t.Errorf("parent ID = %v, want %v",
			*childCategory.GetParentID(), parentCategory.GetID())
	}
}

func TestCategory_ActivateDeactivate(t *testing.T) {
	category, _ := NewCategory("Test Category", "Test Description", nil)

	// Start as active
	if !category.IsActive() {
		t.Error("category should be active by default")
	}

	// Deactivate
	category.Deactivate()
	if category.IsActive() {
		t.Error("category should be inactive after deactivate")
	}

	// Activate
	category.Activate()
	if !category.IsActive() {
		t.Error("category should be active after activate")
	}
}

func TestAttributes_JSON(t *testing.T) {
	attrs := Attributes{
		"color":  "red",
		"size":   "large",
		"weight": 1.5,
		"tags":   []string{"new", "sale"},
	}

	// Test ToJSON
	jsonData, err := attrs.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
	}

	// Test FromJSON
	parsedAttrs, err := AttributesFromJSON(jsonData)
	if err != nil {
		t.Errorf("AttributesFromJSON() error = %v", err)
	}

	if parsedAttrs["color"] != attrs["color"] {
		t.Errorf("color = %v, want %v", parsedAttrs["color"], attrs["color"])
	}

	// Test empty JSON
	emptyAttrs, err := AttributesFromJSON([]byte{})
	if err != nil {
		t.Errorf("AttributesFromJSON() error = %v", err)
	}
	if len(emptyAttrs) != 0 {
		t.Errorf("expected empty attributes, got %d items", len(emptyAttrs))
	}
}
