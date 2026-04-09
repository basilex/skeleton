package domain

import (
	"testing"
)

func TestNewCustomer(t *testing.T) {
	tests := []struct {
		name         string
		customerName string
		taxID        string
		email        string
		phone        string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid customer",
			customerName: "ACME Corp",
			taxID:        "12345678",
			email:        "info@acme.com",
			phone:        "+1234567890",
			wantErr:      false,
		},
		{
			name:         "empty name",
			customerName: "",
			taxID:        "12345678",
			email:        "info@acme.com",
			phone:        "+1234567890",
			wantErr:      true,
			errMsg:       "name is required",
		},
		{
			name:         "invalid email format",
			customerName: "ACME Corp",
			taxID:        "12345678",
			email:        "invalid-email-at-domain",
			phone:        "+1234567890",
			wantErr:      true,
			errMsg:       "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := Address{
				City:    "Kyiv",
				Country: "Ukraine",
			}
			contactInfo, err := NewContactInfo(tt.email, tt.phone, address)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("NewContactInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			customer, err := NewCustomer(tt.customerName, tt.taxID, contactInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustomer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errMsg)
				}
				return
			}

			if customer == nil {
				t.Fatal("expected customer, got nil")
			}

			if customer.GetName() != tt.customerName {
				t.Errorf("GetName() = %v, want %v", customer.GetName(), tt.customerName)
			}

			if customer.GetTaxID() != tt.taxID {
				t.Errorf("GetTaxID() = %v, want %v", customer.GetTaxID(), tt.taxID)
			}

			if customer.GetStatus() != PartyStatusActive {
				t.Errorf("GetStatus() = %v, want %v", customer.GetStatus(), PartyStatusActive)
			}

			events := customer.PullEvents()
			if len(events) != 1 {
				t.Errorf("expected 1 event, got %d", len(events))
			}

			if _, ok := events[0].(CustomerCreated); !ok {
				t.Errorf("expected CustomerCreated event, got %T", events[0])
			}
		})
	}
}

func TestCustomer_AddPurchase(t *testing.T) {
	address := Address{City: "Kyiv", Country: "Ukraine"}
	contactInfo, _ := NewContactInfo("test@test.com", "+1234567890", address)
	customer, err := NewCustomer("Test Customer", "12345678", contactInfo)
	if err != nil {
		t.Fatalf("NewCustomer() error = %v", err)
	}

	initialPurchases := customer.GetTotalPurchases()
	customer.AddPurchase(1000.50)

	if customer.GetTotalPurchases() != initialPurchases+1000.50 {
		t.Errorf("GetTotalPurchases() = %v, want %v",
			customer.GetTotalPurchases(), initialPurchases+1000.50)
	}

	// Test loyalty level upgrade (thresholds: 20000=Silver, 50000=Gold, 100000=Platinum)
	customer.AddPurchase(20000) // Total > 20000, should upgrade to Silver
	if customer.GetLoyaltyLevel() != LoyaltyLevelSilver {
		t.Errorf("GetLoyaltyLevel() = %v, want %v (total: %v)",
			customer.GetLoyaltyLevel(), LoyaltyLevelSilver, customer.GetTotalPurchases())
	}

	customer.AddPurchase(30000) // Total > 50000, should upgrade to Gold
	if customer.GetLoyaltyLevel() != LoyaltyLevelGold {
		t.Errorf("GetLoyaltyLevel() = %v, want %v (total: %v)",
			customer.GetLoyaltyLevel(), LoyaltyLevelGold, customer.GetTotalPurchases())
	}
}

func TestCustomer_ActivateDeactivate(t *testing.T) {
	address := Address{City: "Kyiv", Country: "Ukraine"}
	contactInfo, _ := NewContactInfo("test@test.com", "+1234567890", address)
	customer, _ := NewCustomer("Test Customer", "12345678", contactInfo)

	// Start as active
	if customer.GetStatus() != PartyStatusActive {
		t.Errorf("initial status = %v, want %v", customer.GetStatus(), PartyStatusActive)
	}

	// Deactivate
	customer.Deactivate()
	if customer.GetStatus() != PartyStatusInactive {
		t.Errorf("after deactivate, status = %v, want %v",
			customer.GetStatus(), PartyStatusInactive)
	}

	// Activate
	err := customer.Activate()
	if err != nil {
		t.Errorf("Activate() error = %v", err)
	}
	if customer.GetStatus() != PartyStatusActive {
		t.Errorf("after activate, status = %v, want %v",
			customer.GetStatus(), PartyStatusActive)
	}

	// Blacklist
	customer.Blacklist()
	if customer.GetStatus() != PartyStatusBlacklisted {
		t.Errorf("after blacklist, status = %v, want %v",
			customer.GetStatus(), PartyStatusBlacklisted)
	}

	// Cannot activate blacklisted customer
	err = customer.Activate()
	if err == nil {
		t.Error("expected error when activating blacklisted customer")
	}
	if err != ErrPartyBlacklisted {
		t.Errorf("expected ErrPartyBlacklisted, got %v", err)
	}
}

func TestSupplier_AssignContract(t *testing.T) {
	address := Address{City: "Kyiv", Country: "Ukraine"}
	contactInfo, _ := NewContactInfo("test@test.com", "+1234567890", address)
	supplier, _ := NewSupplier("Test Supplier", "12345678", contactInfo)

	contractID := "contract-123"

	// Assign contract
	err := supplier.AssignContract(contractID)
	if err != nil {
		t.Errorf("AssignContract() error = %v", err)
	}

	contracts := supplier.GetContracts()
	if len(contracts) != 1 {
		t.Errorf("expected 1 contract, got %d", len(contracts))
	}

	if contracts[0] != contractID {
		t.Errorf("contract = %v, want %v", contracts[0], contractID)
	}

	// Cannot assign duplicate contract
	err = supplier.AssignContract(contractID)
	if err == nil {
		t.Error("expected error when assigning duplicate contract")
	}

	// Blacklist supplier
	supplier.Blacklist()
	err = supplier.AssignContract("contract-456")
	if err == nil {
		t.Error("expected error when assigning contract to blacklisted supplier")
	}
}

func TestContactInfo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		phone   string
		wantErr bool
	}{
		{
			name:    "valid contact info",
			email:   "test@example.com",
			phone:   "+1234567890",
			wantErr: false,
		},
		{
			name:    "invalid email",
			email:   "invalid-email",
			phone:   "+1234567890",
			wantErr: true,
		},
		{
			name:    "invalid phone",
			email:   "test@example.com",
			phone:   "123",
			wantErr: true,
		},
		{
			name:    "empty contact info",
			email:   "",
			phone:   "",
			wantErr: false, // Empty is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := Address{City: "Kyiv", Country: "Ukraine"}
			contactInfo, err := NewContactInfo(tt.email, tt.phone, address)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContactInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				err := contactInfo.Validate()
				if err != nil {
					t.Errorf("Validate() error = %v", err)
				}
			}
		})
	}
}

func TestBankAccount_Validation(t *testing.T) {
	tests := []struct {
		name          string
		bankName      string
		accountName   string
		accountNumber string
		currency      string
		wantErr       bool
	}{
		{
			name:          "valid bank account",
			bankName:      "Test Bank",
			accountName:   "John Doe",
			accountNumber: "UA1234567890",
			currency:      "UAH",
			wantErr:       false,
		},
		{
			name:          "missing bank name",
			bankName:      "",
			accountName:   "John Doe",
			accountNumber: "UA1234567890",
			currency:      "UAH",
			wantErr:       true,
		},
		{
			name:          "invalid currency",
			bankName:      "Test Bank",
			accountName:   "John Doe",
			accountNumber: "UA1234567890",
			currency:      "INVALID",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBankAccount(tt.bankName, tt.accountName, tt.accountNumber, tt.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBankAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
