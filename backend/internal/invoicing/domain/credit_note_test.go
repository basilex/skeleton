package domain

import (
	"testing"
	"time"

	moneypkg "github.com/basilex/skeleton/pkg/money"
)

func TestNewCreditNote(t *testing.T) {
	t.Run("creates credit note with valid data", func(t *testing.T) {
		cn, err := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "Product return", "USD")
		if err != nil {
			t.Fatalf("NewCreditNote() error = %v", err)
		}

		if cn.GetCreditNoteNumber() != "CN-001" {
			t.Errorf("number = %v, want CN-001", cn.GetCreditNoteNumber())
		}

		if cn.GetStatus() != CreditNoteStatusDraft {
			t.Errorf("status = %v, want %v", cn.GetStatus(), CreditNoteStatusDraft)
		}

		if !cn.GetTotal().IsZero() {
			t.Errorf("total = %v, want 0", cn.GetTotal())
		}
	})

	t.Run("fails with empty number", func(t *testing.T) {
		_, err := NewCreditNote("", "customer-123", CreditNoteReasonRefund, "desc", "USD")
		if err == nil {
			t.Error("expected error for empty number")
		}
	})

	t.Run("fails with empty customer ID", func(t *testing.T) {
		_, err := NewCreditNote("CN-001", "", CreditNoteReasonRefund, "desc", "USD")
		if err == nil {
			t.Error("expected error for empty customer ID")
		}
	})
}

func TestCreditNote_Lines(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")

	t.Run("add line", func(t *testing.T) {
		unitPrice, _ := moneypkg.NewFromFloat(50.0, "USD")
		err := cn.AddLine("Product A", 2, unitPrice)
		if err != nil {
			t.Errorf("AddLine() error = %v", err)
		}

		if len(cn.GetLines()) != 1 {
			t.Errorf("lines count = %d, want 1", len(cn.GetLines()))
		}

		expected, _ := moneypkg.NewFromFloat(100.0, "USD")
		if !cn.GetSubtotal().Equals(expected) {
			t.Errorf("subtotal = %v, want 100", cn.GetSubtotal())
		}
	})

	t.Run("add another line", func(t *testing.T) {
		unitPrice, _ := moneypkg.NewFromFloat(25.0, "USD")
		err := cn.AddLine("Product B", 1, unitPrice)
		if err != nil {
			t.Errorf("AddLine() error = %v", err)
		}

		expected, _ := moneypkg.NewFromFloat(125.0, "USD")
		if !cn.GetSubtotal().Equals(expected) {
			t.Errorf("subtotal = %v, want 125", cn.GetSubtotal())
		}
	})

	t.Run("set tax", func(t *testing.T) {
		taxAmount, _ := moneypkg.NewFromFloat(10.0, "USD")
		err := cn.SetTax(taxAmount)
		if err != nil {
			t.Errorf("SetTax() error = %v", err)
		}

		expectedTax, _ := moneypkg.NewFromFloat(10.0, "USD")
		if !cn.GetTaxAmount().Equals(expectedTax) {
			t.Errorf("tax amount = %v, want 10", cn.GetTaxAmount())
		}

		expectedTotal, _ := moneypkg.NewFromFloat(135.0, "USD")
		if !cn.GetTotal().Equals(expectedTotal) {
			t.Errorf("total = %v, want 135", cn.GetTotal())
		}
	})
}

func TestCreditNote_LinkInvoice(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")
	invoiceID := NewInvoiceID()

	err := cn.LinkInvoice(invoiceID)
	if err != nil {
		t.Errorf("LinkInvoice() error = %v", err)
	}

	if cn.GetInvoiceID() == nil {
		t.Error("expected invoice ID to be set")
	}

	if *cn.GetInvoiceID() != invoiceID {
		t.Errorf("invoice ID = %v, want %v", *cn.GetInvoiceID(), invoiceID)
	}
}

func TestCreditNote_Issue(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")

	t.Run("fails without lines", func(t *testing.T) {
		err := cn.Issue()
		if err == nil {
			t.Error("expected error for empty lines")
		}
	})

	t.Run("issues with lines", func(t *testing.T) {
		unitPrice, _ := moneypkg.NewFromFloat(50.0, "USD")
		cn.AddLine("Product A", 2, unitPrice)
		err := cn.Issue()
		if err != nil {
			t.Errorf("Issue() error = %v", err)
		}

		if cn.GetStatus() != CreditNoteStatusIssued {
			t.Errorf("status = %v, want %v", cn.GetStatus(), CreditNoteStatusIssued)
		}

		if cn.GetIssuedAt() == nil {
			t.Error("expected issuedAt to be set")
		}
	})

	t.Run("cannot add lines after issue", func(t *testing.T) {
		unitPrice, _ := moneypkg.NewFromFloat(25.0, "USD")
		err := cn.AddLine("Product B", 1, unitPrice)
		if err == nil {
			t.Error("expected error when adding line to issued credit note")
		}
	})
}

func TestCreditNote_Apply(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")
	unitPrice, _ := moneypkg.NewFromFloat(50.0, "USD")
	cn.AddLine("Product A", 2, unitPrice)
	taxAmount, _ := moneypkg.NewFromFloat(10.0, "USD")
	cn.SetTax(taxAmount)
	cn.Issue()

	t.Run("apply partial amount", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(50.0, "USD")
		err := cn.Apply(amount)
		if err != nil {
			t.Errorf("Apply() error = %v", err)
		}

		expectedApplied, _ := moneypkg.NewFromFloat(50.0, "USD")
		if !cn.GetAppliedAmount().Equals(expectedApplied) {
			t.Errorf("applied amount = %v, want 50", cn.GetAppliedAmount())
		}

		expectedRemaining, _ := moneypkg.NewFromFloat(60.0, "USD")
		if !cn.GetRemainingAmount().Equals(expectedRemaining) {
			t.Errorf("remaining amount = %v, want 60", cn.GetRemainingAmount())
		}
	})

	t.Run("apply full remaining", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(60.0, "USD")
		err := cn.Apply(amount)
		if err != nil {
			t.Errorf("Apply() error = %v", err)
		}

		if cn.GetStatus() != CreditNoteStatusApplied {
			t.Errorf("status = %v, want %v", cn.GetStatus(), CreditNoteStatusApplied)
		}

		if !cn.IsFullyApplied() {
			t.Error("expected credit note to be fully applied")
		}
	})

	t.Run("cannot apply more than total", func(t *testing.T) {
		cn2, _ := NewCreditNote("CN-002", "customer-123", CreditNoteReasonRefund, "desc", "USD")
		unitPrice, _ := moneypkg.NewFromFloat(100.0, "USD")
		cn2.AddLine("Product A", 1, unitPrice)
		cn2.Issue()

		amount, _ := moneypkg.NewFromFloat(150.0, "USD")
		err := cn2.Apply(amount)
		if err == nil {
			t.Error("expected error when applying more than total")
		}
	})
}

func TestCreditNote_Cancel(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")
	unitPrice, _ := moneypkg.NewFromFloat(100.0, "USD")
	cn.AddLine("Product A", 1, unitPrice)
	cn.Issue()

	err := cn.Cancel("Customer requested cancellation")
	if err != nil {
		t.Errorf("Cancel() error = %v", err)
	}

	if cn.GetStatus() != CreditNoteStatusCancelled {
		t.Errorf("status = %v, want %v", cn.GetStatus(), CreditNoteStatusCancelled)
	}

	if cn.GetCancelledAt() == nil {
		t.Error("expected cancelledAt to be set")
	}
}

func TestInstallment(t *testing.T) {
	invoiceID := NewInvoiceID()
	dueDate := time.Now().AddDate(0, 1, 0)

	t.Run("creates installment with valid data", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, err := NewInstallment(invoiceID, 1, amount, dueDate)
		if err != nil {
			t.Fatalf("NewInstallment() error = %v", err)
		}

		expectedAmount, _ := moneypkg.NewFromFloat(100.0, "USD")
		if !inst.GetTotalAmount().Equals(expectedAmount) {
			t.Errorf("total = %v, want 100.0", inst.GetTotalAmount())
		}

		if inst.GetStatus() != InstallmentStatusPending {
			t.Errorf("status = %v, want %v", inst.GetStatus(), InstallmentStatusPending)
		}
	})

	t.Run("fails with non-positive amount", func(t *testing.T) {
		zeroAmount := moneypkg.Zero("USD")
		_, err := NewInstallment(invoiceID, 1, zeroAmount, dueDate)
		if err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("make payment", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, _ := NewInstallment(invoiceID, 1, amount, dueDate)

		paymentAmount, _ := moneypkg.NewFromFloat(50.0, "USD")
		err := inst.MakePayment(paymentAmount)
		if err != nil {
			t.Errorf("MakePayment() error = %v", err)
		}

		expectedPaid, _ := moneypkg.NewFromFloat(50.0, "USD")
		if !inst.GetPaidAmount().Equals(expectedPaid) {
			t.Errorf("paid amount = %v, want 50.0", inst.GetPaidAmount())
		}

		expectedRemaining, _ := moneypkg.NewFromFloat(50.0, "USD")
		if !inst.GetRemainingAmount().Equals(expectedRemaining) {
			t.Errorf("remaining amount = %v, want 50.0", inst.GetRemainingAmount())
		}
	})

	t.Run("full payment marks as paid", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, _ := NewInstallment(invoiceID, 1, amount, dueDate)

		err := inst.MakePayment(amount)
		if err != nil {
			t.Errorf("MakePayment() error = %v", err)
		}

		if !inst.IsPaid() {
			t.Error("expected installment to be paid")
		}

		if inst.GetPaidAt() == nil {
			t.Error("expected paidAt to be set")
		}
	})

	t.Run("status transitions", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, _ := NewInstallment(invoiceID, 1, amount, dueDate)

		err := inst.MarkAsDue()
		if err != nil {
			t.Errorf("MarkAsDue() error = %v", err)
		}

		if inst.GetStatus() != InstallmentStatusDue {
			t.Errorf("status = %v, want %v", inst.GetStatus(), InstallmentStatusDue)
		}

		err = inst.MarkAsOverdue()
		if err != nil {
			t.Errorf("MarkAsOverdue() error = %v", err)
		}

		if !inst.IsOverdue() {
			t.Error("expected installment to be overdue")
		}
	})

	t.Run("send reminder", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, _ := NewInstallment(invoiceID, 1, amount, dueDate)

		err := inst.SendReminder()
		if err != nil {
			t.Errorf("SendReminder() error = %v", err)
		}

		if inst.GetRemindedAt() == nil {
			t.Error("expected remindedAt to be set")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		amount, _ := moneypkg.NewFromFloat(100.0, "USD")
		inst, _ := NewInstallment(invoiceID, 1, amount, dueDate)

		err := inst.Cancel("Customer bankruptcy")
		if err != nil {
			t.Errorf("Cancel() error = %v", err)
		}

		if inst.GetStatus() != InstallmentStatusCancelled {
			t.Errorf("status = %v, want %v", inst.GetStatus(), InstallmentStatusCancelled)
		}
	})
}

func TestInstallmentPlan(t *testing.T) {
	invoiceID := NewInvoiceID()

	t.Run("create plan", func(t *testing.T) {
		totalAmount, _ := moneypkg.NewFromFloat(1000.0, "USD")
		plan := NewInstallmentPlan(invoiceID, totalAmount, "USD")
		expectedTotal, _ := moneypkg.NewFromFloat(1000.0, "USD")
		if !plan.GetTotalAmount().Equals(expectedTotal) {
			t.Errorf("total = %v, want 1000.0", plan.GetTotalAmount())
		}
	})

	t.Run("add installments", func(t *testing.T) {
		totalAmount, _ := moneypkg.NewFromFloat(1000.0, "USD")
		plan := NewInstallmentPlan(invoiceID, totalAmount, "USD")

		instAmount1, _ := moneypkg.NewFromFloat(500.0, "USD")
		err := plan.AddInstallment(instAmount1, time.Now().AddDate(0, 1, 0))
		if err != nil {
			t.Errorf("AddInstallment() error = %v", err)
		}

		instAmount2, _ := moneypkg.NewFromFloat(500.0, "USD")
		err = plan.AddInstallment(instAmount2, time.Now().AddDate(0, 2, 0))
		if err != nil {
			t.Errorf("AddInstallment() error = %v", err)
		}

		if len(plan.GetInstallments()) != 2 {
			t.Errorf("installments count = %d, want 2", len(plan.GetInstallments()))
		}
	})

	t.Run("fails when exceeds total", func(t *testing.T) {
		totalAmount, _ := moneypkg.NewFromFloat(1000.0, "USD")
		plan := NewInstallmentPlan(invoiceID, totalAmount, "USD")

		instAmount1, _ := moneypkg.NewFromFloat(500.0, "USD")
		_ = plan.AddInstallment(instAmount1, time.Now().AddDate(0, 1, 0))

		instAmount2, _ := moneypkg.NewFromFloat(600.0, "USD")
		_ = plan.AddInstallment(instAmount2, time.Now().AddDate(0, 2, 0))
		// Total so far: 1100, which exceeds 1000
		// So the second add should fail

		// Try to add another that would exceed
		totalAmount2, _ := moneypkg.NewFromFloat(500.0, "USD")
		plan2 := NewInstallmentPlan(invoiceID, totalAmount2, "USD")

		instAmount3, _ := moneypkg.NewFromFloat(400.0, "USD")
		_ = plan2.AddInstallment(instAmount3, time.Now().AddDate(0, 1, 0))

		instAmount4, _ := moneypkg.NewFromFloat(200.0, "USD")
		err := plan2.AddInstallment(instAmount4, time.Now().AddDate(0, 2, 0))
		if err == nil {
			t.Error("expected error when exceeding total")
		}
	})
}
