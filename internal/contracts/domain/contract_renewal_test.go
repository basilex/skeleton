package domain

import (
	"testing"
	"time"
)

func TestContract_Renewal(t *testing.T) {
	startDate := time.Now().UTC()
	endDate := startDate.AddDate(1, 0, 0)
	paymentTerms := PaymentTerms{PaymentType: PaymentTypePrepaid, Currency: "USD"}
	deliveryTerms := DeliveryTerms{EstimatedDays: 30, DeliveryType: DeliveryTypeDelivery}

	contract, _ := NewContract(
		ContractTypeSupply,
		"party-123",
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		100000,
		"USD",
		"user-1",
	)

	t.Run("auto renewal disabled by default", func(t *testing.T) {
		if contract.GetAutoRenewal() {
			t.Error("auto renewal should be disabled by default")
		}
	})

	t.Run("enable auto renewal", func(t *testing.T) {
		contract.Activate(startDate)

		err := contract.EnableAutoRenewal(30, 3)
		if err != nil {
			t.Errorf("EnableAutoRenewal() error = %v", err)
		}

		if !contract.GetAutoRenewal() {
			t.Error("auto renewal should be enabled")
		}

		if contract.GetRenewalPeriodDays() != 30 {
			t.Errorf("renewal period = %d, want 30", contract.GetRenewalPeriodDays())
		}

		if contract.GetMaxRenewals() != 3 {
			t.Errorf("max renewals = %d, want 3", contract.GetMaxRenewals())
		}
	})

	t.Run("disable auto renewal", func(t *testing.T) {
		err := contract.DisableAutoRenewal()
		if err != nil {
			t.Errorf("DisableAutoRenewal() error = %v", err)
		}

		if contract.GetAutoRenewal() {
			t.Error("auto renewal should be disabled")
		}
	})

	t.Run("can renew check", func(t *testing.T) {
		contract2, _ := NewContract(
			ContractTypeSupply,
			"party-123",
			paymentTerms,
			deliveryTerms,
			startDate,
			endDate,
			100000,
			"USD",
			"user-1",
		)
		contract2.Activate(startDate)

		// Cannot renew without auto-renewal
		if contract2.CanRenew() {
			t.Error("should not be able to renew without auto-renewal")
		}

		// Enable auto-renewal with max
		contract2.EnableAutoRenewal(30, 2)

		// Renew twice
		newEndDate := endDate.AddDate(0, 0, 30)
		contract2.Renew(newEndDate)
		contract2.Renew(newEndDate.AddDate(0, 0, 30))

		// Should not be able to renew more
		if contract2.CanRenew() {
			t.Error("should not be able to renew after max renewals")
		}
	})

	t.Run("renew contract", func(t *testing.T) {
		contract3, _ := NewContract(
			ContractTypeSupply,
			"party-123",
			paymentTerms,
			deliveryTerms,
			startDate,
			endDate,
			100000,
			"USD",
			"user-1",
		)
		contract3.Activate(startDate)
		contract3.EnableAutoRenewal(30, 3)

		oldEndDate := contract3.GetValidityPeriod().EndDate
		newEndDate := oldEndDate.AddDate(0, 0, 30)

		err := contract3.Renew(newEndDate)
		if err != nil {
			t.Errorf("Renew() error = %v", err)
		}

		if contract3.GetRenewalCount() != 1 {
			t.Errorf("renewal count = %d, want 1", contract3.GetRenewalCount())
		}

		if contract3.GetRenewedAt() == nil {
			t.Error("renewedAt should be set")
		}

		if !contract3.GetValidityPeriod().EndDate.Equal(newEndDate) {
			t.Errorf("end date = %v, want %v", contract3.GetValidityPeriod().EndDate, newEndDate)
		}
	})

	t.Run("expire contract", func(t *testing.T) {
		contract4, _ := NewContract(
			ContractTypeSupply,
			"party-123",
			paymentTerms,
			deliveryTerms,
			startDate,
			endDate,
			100000,
			"USD",
			"user-1",
		)
		contract4.Activate(startDate)

		err := contract4.Expire()
		if err != nil {
			t.Errorf("Expire() error = %v", err)
		}

		if contract4.GetStatus() != ContractStatusExpired {
			t.Errorf("status = %v, want %v", contract4.GetStatus(), ContractStatusExpired)
		}
	})

	t.Run("days until expiry", func(t *testing.T) {
		contract5, _ := NewContract(
			ContractTypeSupply,
			"party-123",
			paymentTerms,
			deliveryTerms,
			startDate,
			startDate.AddDate(0, 0, 30),
			100000,
			"USD",
			"user-1",
		)
		contract5.Activate(startDate)

		days := contract5.DaysUntilExpiry()
		if days < 29 || days > 31 {
			t.Errorf("days until expiry = %d, expected around 30", days)
		}
	})
}

func TestContract_Amendments(t *testing.T) {
	startDate := time.Now().UTC()
	endDate := startDate.AddDate(1, 0, 0)
	paymentTerms := PaymentTerms{PaymentType: PaymentTypePrepaid, Currency: "USD"}
	deliveryTerms := DeliveryTerms{EstimatedDays: 30, DeliveryType: DeliveryTypeDelivery}

	contract, _ := NewContract(
		ContractTypeSupply,
		"party-123",
		paymentTerms,
		deliveryTerms,
		startDate,
		endDate,
		100000,
		"USD",
		"user-1",
	)

	t.Run("create amendment", func(t *testing.T) {
		contract.Activate(startDate)

		changes := map[string]string{
			"credit_limit":  "increased from 100000 to 150000",
			"payment_terms": "changed to net45",
		}

		err := contract.CreateAmendment("AMD-001", "Credit limit increase", changes, "user-2")
		if err != nil {
			t.Errorf("CreateAmendment() error = %v", err)
		}

		if contract.GetVersion() != 2 {
			t.Errorf("version = %d, want 2", contract.GetVersion())
		}

		amendments := contract.GetAmendments()
		if len(amendments) != 1 {
			t.Errorf("amendments count = %d, want 1", len(amendments))
		}
	})

	t.Run("get amendment by version", func(t *testing.T) {
		amd := contract.GetAmendment(2)
		if amd == nil {
			t.Fatal("amendment not found")
		}

		if amd.GetDescription() != "Credit limit increase" {
			t.Errorf("description = %v, want 'Credit limit increase'", amd.GetDescription())
		}
	})

	t.Run("get latest amendment", func(t *testing.T) {
		// Add another amendment
		changes := map[string]string{
			"delivery_terms": "updated delivery schedule",
		}
		contract.CreateAmendment("AMD-002", "Delivery terms update", changes, "user-3")

		latest := contract.GetLatestAmendment()
		if latest == nil {
			t.Fatal("latest amendment not found")
		}

		if latest.GetVersion() != 3 {
			t.Errorf("version = %d, want 3", latest.GetVersion())
		}
	})

	t.Run("cannot amend terminated contract", func(t *testing.T) {
		contract2, _ := NewContract(
			ContractTypeSupply,
			"party-123",
			paymentTerms,
			deliveryTerms,
			startDate,
			endDate,
			100000,
			"USD",
			"user-1",
		)

		contract2.Activate(startDate)
		err := contract2.Terminate("testing")
		if err != nil {
			t.Fatalf("Terminate() error = %v", err)
		}

		err = contract2.CreateAmendment("AMD-003", "test", nil, "user-1")
		if err == nil {
			t.Error("expected error when amending terminated contract")
		}
	})
}

func TestDateRange_Contains(t *testing.T) {
	start := time.Now().UTC()
	end := start.AddDate(0, 1, 0)

	dr, _ := NewDateRange(start, end)

	t.Run("contains start date", func(t *testing.T) {
		if !dr.Contains(start) {
			t.Error("should contain start date")
		}
	})

	t.Run("contains date in range", func(t *testing.T) {
		middle := start.AddDate(0, 0, 15)
		if !dr.Contains(middle) {
			t.Error("should contain date in range")
		}
	})

	t.Run("does not contain date before", func(t *testing.T) {
		before := start.AddDate(0, 0, -1)
		if dr.Contains(before) {
			t.Error("should not contain date before range")
		}
	})

	t.Run("does not contain date after", func(t *testing.T) {
		after := end.AddDate(0, 0, 1)
		if dr.Contains(after) {
			t.Error("should not contain date after range")
		}
	})
}
