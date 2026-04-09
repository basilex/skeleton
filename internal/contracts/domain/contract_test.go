package domain

import (
	"testing"
	"time"
)

func TestNewContract(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0) // 1 year

	paymentTerms, _ := NewPaymentTerms(PaymentTypePrepaid, 0, "UAH")
	deliveryTerms := NewDeliveryTerms(DeliveryTypeDelivery, 5)

	tests := []struct {
		name          string
		contractType  ContractType
		partyID       string
		paymentTerms  PaymentTerms
		deliveryTerms DeliveryTerms
		startDate     time.Time
		endDate       time.Time
		creditLimit   float64
		currency      string
		wantErr       bool
	}{
		{
			name:          "valid supply contract",
			contractType:  ContractTypeSupply,
			partyID:       "party-123",
			paymentTerms:  paymentTerms,
			deliveryTerms: deliveryTerms,
			startDate:     startDate,
			endDate:       endDate,
			creditLimit:   100000,
			currency:      "UAH",
			wantErr:       false,
		},
		{
			name:          "empty party ID",
			contractType:  ContractTypeSupply,
			partyID:       "",
			paymentTerms:  paymentTerms,
			deliveryTerms: deliveryTerms,
			startDate:     startDate,
			endDate:       endDate,
			creditLimit:   100000,
			currency:      "UAH",
			wantErr:       true,
		},
		{
			name:          "invalid date range",
			contractType:  ContractTypeSupply,
			partyID:       "party-123",
			paymentTerms:  paymentTerms,
			deliveryTerms: deliveryTerms,
			startDate:     endDate,
			endDate:       startDate,
			creditLimit:   100000,
			currency:      "UAH",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, err := NewContract(
				tt.contractType,
				tt.partyID,
				tt.paymentTerms,
				tt.deliveryTerms,
				tt.startDate,
				tt.endDate,
				tt.creditLimit,
				tt.currency,
				"user-123",
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewContract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if contract == nil {
				t.Fatal("expected contract, got nil")
			}

			if contract.GetType() != tt.contractType {
				t.Errorf("GetType() = %v, want %v", contract.GetType(), tt.contractType)
			}

			if contract.GetStatus() != ContractStatusDraft {
				t.Errorf("GetStatus() = %v, want %v", contract.GetStatus(), ContractStatusDraft)
			}

			if contract.GetPartyID() != tt.partyID {
				t.Errorf("GetPartyID() = %v, want %v", contract.GetPartyID(), tt.partyID)
			}

			events := contract.PullEvents()
			if len(events) != 1 {
				t.Errorf("expected 1 event, got %d", len(events))
			}

			if _, ok := events[0].(ContractCreated); !ok {
				t.Errorf("expected ContractCreated event, got %T", events[0])
			}
		})
	}
}

func TestContract_Activate(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)
	paymentTerms, _ := NewPaymentTerms(PaymentTypePrepaid, 0, "UAH")
	deliveryTerms := NewDeliveryTerms(DeliveryTypeDelivery, 5)

	contract, _ := NewContract(
		ContractTypeSupply,
		"party-123",
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		100000,
		"UAH",
		"user-123",
	)

	// Clear the ContractCreated event from creation
	_ = contract.PullEvents()

	// Cannot activate a draft contract without signing
	signedAt := time.Now()
	err := contract.Activate(signedAt)
	if err != nil {
		t.Errorf("Activate() error = %v", err)
	}

	if contract.GetStatus() != ContractStatusActive {
		t.Errorf("GetStatus() = %v, want %v", contract.GetStatus(), ContractStatusActive)
	}

	if contract.GetSignedAt() == nil {
		t.Error("expected signed_at to be set")
	}

	events := contract.PullEvents()
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	if _, ok := events[0].(ContractActivated); !ok {
		t.Errorf("expected ContractActivated event, got %T", events[0])
	}
}

func TestContract_Terminate(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)
	paymentTerms, _ := NewPaymentTerms(PaymentTypePrepaid, 0, "UAH")
	deliveryTerms := NewDeliveryTerms(DeliveryTypeDelivery, 5)

	contract, _ := NewContract(
		ContractTypeSupply,
		"party-123",
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		100000,
		"UAH",
		"user-123",
	)

	// Cannot terminate a draft contract
	err := contract.Terminate("test termination")
	if err == nil {
		t.Error("expected error when terminating draft contract")
	}

	// Activate the contract first
	contract.Activate(time.Now())

	// Now terminate
	err = contract.Terminate("test termination")
	if err != nil {
		t.Errorf("Terminate() error = %v", err)
	}

	if contract.GetStatus() != ContractStatusTerminated {
		t.Errorf("GetStatus() = %v, want %v",
			contract.GetStatus(), ContractStatusTerminated)
	}

	if contract.GetTerminatedAt() == nil {
		t.Error("expected terminated_at to be set")
	}
}

func TestContract_AddRemoveDocument(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)
	paymentTerms, _ := NewPaymentTerms(PaymentTypePrepaid, 0, "UAH")
	deliveryTerms := NewDeliveryTerms(DeliveryTypeDelivery, 5)

	contract, _ := NewContract(
		ContractTypeSupply,
		"party-123",
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		100000,
		"UAH",
		"user-123",
	)

	docID1 := "document-123"
	docID2 := "document-456"

	// Add documents
	err := contract.AddDocument(docID1)
	if err != nil {
		t.Errorf("AddDocument() error = %v", err)
	}

	err = contract.AddDocument(docID2)
	if err != nil {
		t.Errorf("AddDocument() error = %v", err)
	}

	docs := contract.GetDocuments()
	if len(docs) != 2 {
		t.Errorf("expected 2 documents, got %d", len(docs))
	}

	// Cannot add duplicate document
	err = contract.AddDocument(docID1)
	if err == nil {
		t.Error("expected error when adding duplicate document")
	}

	// Remove document
	err = contract.RemoveDocument(docID1)
	if err != nil {
		t.Errorf("RemoveDocument() error = %v", err)
	}

	docs = contract.GetDocuments()
	if len(docs) != 1 {
		t.Errorf("expected 1 document, got %d", len(docs))
	}

	// Cannot remove non-existent document
	err = contract.RemoveDocument("non-existent")
	if err == nil {
		t.Error("expected error when removing non-existent document")
	}

	// Terminated contract cannot add/remove documents
	contract.Activate(time.Now())
	contract.Terminate("test")

	err = contract.AddDocument("document-789")
	if err == nil {
		t.Error("expected error when adding document to terminated contract")
	}
}

func TestPaymentTerms_Validation(t *testing.T) {
	tests := []struct {
		name        string
		paymentType PaymentType
		creditDays  int
		currency    string
		wantErr     bool
	}{
		{
			name:        "valid prepaid",
			paymentType: PaymentTypePrepaid,
			creditDays:  0,
			currency:    "UAH",
			wantErr:     false,
		},
		{
			name:        "valid credit",
			paymentType: PaymentTypeCredit,
			creditDays:  30,
			currency:    "UAH",
			wantErr:     false,
		},
		{
			name:        "invalid credit without days",
			paymentType: PaymentTypeCredit,
			creditDays:  0,
			currency:    "UAH",
			wantErr:     true,
		},
		{
			name:        "invalid currency",
			paymentType: PaymentTypePrepaid,
			creditDays:  0,
			currency:    "TOOLONG",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPaymentTerms(tt.paymentType, tt.creditDays, tt.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaymentTerms() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeliveryTerms_Validation(t *testing.T) {
	tests := []struct {
		name          string
		deliveryType  DeliveryType
		estimatedDays int
		wantErr       bool
	}{
		{
			name:          "valid delivery",
			deliveryType:  DeliveryTypeDelivery,
			estimatedDays: 5,
			wantErr:       false,
		},
		{
			name:          "negative days",
			deliveryType:  DeliveryTypeDelivery,
			estimatedDays: -1,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deliveryTerms := NewDeliveryTerms(tt.deliveryType, tt.estimatedDays)
			err := deliveryTerms.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateRange_Validation(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)

	// Valid range
	_, err := NewDateRange(startDate, endDate)
	if err != nil {
		t.Errorf("NewDateRange() error = %v for valid range", err)
	}

	// Invalid range (end before start)
	_, err = NewDateRange(endDate, startDate)
	if err == nil {
		t.Error("expected error for invalid range")
	}
	if err != ErrInvalidDateRange {
		t.Errorf("expected ErrInvalidDateRange, got %v", err)
	}
}
