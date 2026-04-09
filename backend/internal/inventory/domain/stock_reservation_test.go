package domain

import (
	"testing"
	"time"
)

func TestNewStockReservation(t *testing.T) {
	stockID := NewStockID()
	expiresAt := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name      string
		stockID   StockID
		orderID   string
		quantity  float64
		expiresAt *time.Time
		wantErr   bool
	}{
		{
			name:      "valid reservation",
			stockID:   stockID,
			orderID:   "order-001",
			quantity:  10,
			expiresAt: &expiresAt,
			wantErr:   false,
		},
		{
			name:      "empty order ID",
			stockID:   stockID,
			orderID:   "",
			quantity:  10,
			expiresAt: &expiresAt,
			wantErr:   true,
		},
		{
			name:      "negative quantity",
			stockID:   stockID,
			orderID:   "order-001",
			quantity:  -5,
			expiresAt: &expiresAt,
			wantErr:   true,
		},
		{
			name:      "zero quantity",
			stockID:   stockID,
			orderID:   "order-001",
			quantity:  0,
			expiresAt: &expiresAt,
			wantErr:   true,
		},
		{
			name:      "no expiration",
			stockID:   stockID,
			orderID:   "order-001",
			quantity:  10,
			expiresAt: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation, err := NewStockReservation(tt.stockID, tt.orderID, tt.quantity, tt.expiresAt)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStockReservation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if reservation.GetStockID() != tt.stockID {
					t.Errorf("Reservation.StockID = %v, want %v", reservation.GetStockID(), tt.stockID)
				}
				if reservation.GetOrderID() != tt.orderID {
					t.Errorf("Reservation.OrderID = %v, want %v", reservation.GetOrderID(), tt.orderID)
				}
				if reservation.GetQuantity() != tt.quantity {
					t.Errorf("Reservation.Quantity = %v, want %v", reservation.GetQuantity(), tt.quantity)
				}
				if reservation.GetStatus() != ReservationStatusActive {
					t.Errorf("Reservation.Status = %v, want %v", reservation.GetStatus(), ReservationStatusActive)
				}
			}
		})
	}
}

func TestStockReservation_Fulfill(t *testing.T) {
	stockID := NewStockID()
	reservation, err := NewStockReservation(stockID, "order-001", 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := reservation.Fulfill(); err != nil {
		t.Errorf("Fulfill() error = %v", err)
	}

	if reservation.GetStatus() != ReservationStatusFulfilled {
		t.Errorf("Reservation.Status = %v, want %v", reservation.GetStatus(), ReservationStatusFulfilled)
	}

	if reservation.GetFulfilledAt() == nil {
		t.Error("FulfilledAt should not be nil after fulfillment")
	}

	if err := reservation.Fulfill(); err == nil {
		t.Error("Fulfill() should return error when already fulfilled")
	}
}

func TestStockReservation_Cancel(t *testing.T) {
	stockID := NewStockID()
	reservation, err := NewStockReservation(stockID, "order-001", 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := reservation.Cancel(); err != nil {
		t.Errorf("Cancel() error = %v", err)
	}

	if reservation.GetStatus() != ReservationStatusCancelled {
		t.Errorf("Reservation.Status = %v, want %v", reservation.GetStatus(), ReservationStatusCancelled)
	}

	if reservation.GetCancelledAt() == nil {
		t.Error("CancelledAt should not be nil after cancellation")
	}

	if err := reservation.Cancel(); err == nil {
		t.Error("Cancel() should return error when already cancelled")
	}
}

func TestStockReservation_IsExpired(t *testing.T) {
	stockID := NewStockID()

	pastTime := time.Now().Add(-24 * time.Hour)
	reservation1, err := NewStockReservation(stockID, "order-001", 10, &pastTime)
	if err != nil {
		t.Fatal(err)
	}

	if !reservation1.IsExpired() {
		t.Error("Reservation should be expired when expiresAt is in the past")
	}

	futureTime := time.Now().Add(24 * time.Hour)
	reservation2, err := NewStockReservation(stockID, "order-002", 10, &futureTime)
	if err != nil {
		t.Fatal(err)
	}

	if reservation2.IsExpired() {
		t.Error("Reservation should not be expired when expiresAt is in the future")
	}

	reservation3, err := NewStockReservation(stockID, "order-003", 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	if reservation3.IsExpired() {
		t.Error("Reservation without expiration should not be expired")
	}
}

func TestStockReservation_IsActive(t *testing.T) {
	stockID := NewStockID()
	pastTime := time.Now().Add(-24 * time.Hour)
	futureTime := time.Now().Add(24 * time.Hour)

	reservation1, err := NewStockReservation(stockID, "order-001", 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reservation1.IsActive() {
		t.Error("New reservation should be active")
	}

	reservation2, err := NewStockReservation(stockID, "order-002", 10, &pastTime)
	if err != nil {
		t.Fatal(err)
	}
	if reservation2.IsActive() {
		t.Error("Reservation with past expiration should not be active")
	}

	reservation3, err := NewStockReservation(stockID, "order-003", 10, &futureTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reservation3.IsActive() {
		t.Error("Reservation with future expiration should be active")
	}

	reservation3.Fulfill()
	if reservation3.IsActive() {
		t.Error("Fulfilled reservation should not be active")
	}
}
