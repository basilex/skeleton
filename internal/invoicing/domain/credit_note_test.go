package domain

import (
	"testing"
	"time"
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

		if cn.GetTotal() != 0 {
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
		err := cn.AddLine("Product A", 2, 50.0)
		if err != nil {
			t.Errorf("AddLine() error = %v", err)
		}

		if len(cn.GetLines()) != 1 {
			t.Errorf("lines count = %d, want 1", len(cn.GetLines()))
		}

		if cn.GetSubtotal() != 100 {
			t.Errorf("subtotal = %v, want 100", cn.GetSubtotal())
		}
	})

	t.Run("add another line", func(t *testing.T) {
		err := cn.AddLine("Product B", 1, 25.0)
		if err != nil {
			t.Errorf("AddLine() error = %v", err)
		}

		if cn.GetSubtotal() != 125 {
			t.Errorf("subtotal = %v, want 125", cn.GetSubtotal())
		}
	})

	t.Run("set tax", func(t *testing.T) {
		err := cn.SetTax(10.0)
		if err != nil {
			t.Errorf("SetTax() error = %v", err)
		}

		if cn.GetTaxAmount() != 10 {
			t.Errorf("tax amount = %v, want 10", cn.GetTaxAmount())
		}

		if cn.GetTotal() != 135 {
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
		cn.AddLine("Product A", 2, 50.0)
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
		err := cn.AddLine("Product B", 1, 25.0)
		if err == nil {
			t.Error("expected error when adding line to issued credit note")
		}
	})
}

func TestCreditNote_Apply(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")
	cn.AddLine("Product A", 2, 50.0)
	cn.SetTax(10.0)
	cn.Issue()

	t.Run("apply partial amount", func(t *testing.T) {
		err := cn.Apply(50.0)
		if err != nil {
			t.Errorf("Apply() error = %v", err)
		}

		if cn.GetAppliedAmount() != 50 {
			t.Errorf("applied amount = %v, want 50", cn.GetAppliedAmount())
		}

		if cn.GetRemainingAmount() != 60 {
			t.Errorf("remaining amount = %v, want 60", cn.GetRemainingAmount())
		}
	})

	t.Run("apply full remaining", func(t *testing.T) {
		err := cn.Apply(60.0)
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
		cn2.AddLine("Product A", 1, 100.0)
		cn2.Issue()

		err := cn2.Apply(150.0)
		if err == nil {
			t.Error("expected error when applying more than total")
		}
	})
}

func TestCreditNote_Cancel(t *testing.T) {
	cn, _ := NewCreditNote("CN-001", "customer-123", CreditNoteReasonRefund, "desc", "USD")
	cn.AddLine("Product A", 1, 100.0)
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
		inst, err := NewInstallment(invoiceID, 1, 100.0, dueDate)
		if err != nil {
			t.Fatalf("NewInstallment() error = %v", err)
		}

		if inst.GetTotalAmount() != 100.0 {
			t.Errorf("total = %v, want 100.0", inst.GetTotalAmount())
		}

		if inst.GetStatus() != InstallmentStatusPending {
			t.Errorf("status = %v, want %v", inst.GetStatus(), InstallmentStatusPending)
		}
	})

	t.Run("fails with non-positive amount", func(t *testing.T) {
		_, err := NewInstallment(invoiceID, 1, 0, dueDate)
		if err == nil {
			t.Error("expected error for zero amount")
		}
	})

	t.Run("make payment", func(t *testing.T) {
		inst, _ := NewInstallment(invoiceID, 1, 100.0, dueDate)

		err := inst.MakePayment(50.0)
		if err != nil {
			t.Errorf("MakePayment() error = %v", err)
		}

		if inst.GetPaidAmount() != 50.0 {
			t.Errorf("paid amount = %v, want 50.0", inst.GetPaidAmount())
		}

		if inst.GetRemainingAmount() != 50.0 {
			t.Errorf("remaining amount = %v, want 50.0", inst.GetRemainingAmount())
		}
	})

	t.Run("full payment marks as paid", func(t *testing.T) {
		inst, _ := NewInstallment(invoiceID, 1, 100.0, dueDate)

		err := inst.MakePayment(100.0)
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
		inst, _ := NewInstallment(invoiceID, 1, 100.0, dueDate)

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
		inst, _ := NewInstallment(invoiceID, 1, 100.0, dueDate)

		err := inst.SendReminder()
		if err != nil {
			t.Errorf("SendReminder() error = %v", err)
		}

		if inst.GetRemindedAt() == nil {
			t.Error("expected remindedAt to be set")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		inst, _ := NewInstallment(invoiceID, 1, 100.0, dueDate)

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
		plan := NewInstallmentPlan(invoiceID, 1000.0, "USD")
		if plan.GetTotalAmount() != 1000.0 {
			t.Errorf("total = %v, want 1000.0", plan.GetTotalAmount())
		}
	})

	t.Run("add installments", func(t *testing.T) {
		plan := NewInstallmentPlan(invoiceID, 1000.0, "USD")

		err := plan.AddInstallment(500.0, time.Now().AddDate(0, 1, 0))
		if err != nil {
			t.Errorf("AddInstallment() error = %v", err)
		}

		err = plan.AddInstallment(500.0, time.Now().AddDate(0, 2, 0))
		if err != nil {
			t.Errorf("AddInstallment() error = %v", err)
		}

		if len(plan.GetInstallments()) != 2 {
			t.Errorf("installments count = %d, want 2", len(plan.GetInstallments()))
		}
	})

	t.Run("fails when exceeds total", func(t *testing.T) {
		plan := NewInstallmentPlan(invoiceID, 1000.0, "USD")

		_ = plan.AddInstallment(500.0, time.Now().AddDate(0, 1, 0))
		_ = plan.AddInstallment(600.0, time.Now().AddDate(0, 2, 0))
		// Total so far: 1100, which exceeds 1000
		// So the second add should fail

		// Try to add another that would exceed
		plan2 := NewInstallmentPlan(invoiceID, 500.0, "USD")
		_ = plan2.AddInstallment(400.0, time.Now().AddDate(0, 1, 0))

		err := plan2.AddInstallment(200.0, time.Now().AddDate(0, 2, 0))
		if err == nil {
			t.Error("expected error when exceeding total")
		}
	})
}
