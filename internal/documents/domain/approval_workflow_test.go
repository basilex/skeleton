package domain

import (
	"testing"
)

func TestNewApprovalWorkflow(t *testing.T) {
	docID := NewDocumentID()
	docType := DocumentTypeContract

	t.Run("creates workflow with valid data", func(t *testing.T) {
		workflow, err := NewApprovalWorkflow(docID, docType, "user-123")
		if err != nil {
			t.Fatalf("NewApprovalWorkflow() error = %v", err)
		}

		if workflow.GetDocumentID() != docID {
			t.Errorf("document ID = %v, want %v", workflow.GetDocumentID(), docID)
		}

		if workflow.GetStatus() != ApprovalStatusPending {
			t.Errorf("status = %v, want %v", workflow.GetStatus(), ApprovalStatusPending)
		}

		if workflow.GetRequestedBy() != "user-123" {
			t.Errorf("requested by = %v, want user-123", workflow.GetRequestedBy())
		}
	})

	t.Run("fails with empty document ID", func(t *testing.T) {
		_, err := NewApprovalWorkflow(DocumentID{}, docType, "user-123")
		if err == nil {
			t.Error("expected error for empty document ID")
		}
	})

	t.Run("fails with empty requested by", func(t *testing.T) {
		_, err := NewApprovalWorkflow(docID, docType, "")
		if err == nil {
			t.Error("expected error for empty requested by")
		}
	})
}

func TestApprovalWorkflow_AddStep(t *testing.T) {
	docID := NewDocumentID()
	workflow, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")

	t.Run("adds step successfully", func(t *testing.T) {
		err := workflow.AddStep("approver-1", "manager")
		if err != nil {
			t.Errorf("AddStep() error = %v", err)
		}

		steps := workflow.GetSteps()
		if len(steps) != 1 {
			t.Errorf("steps count = %d, want 1", len(steps))
		}

		if steps[0].GetApproverID() != "approver-1" {
			t.Errorf("approver ID = %v, want approver-1", steps[0].GetApproverID())
		}
	})

	t.Run("fails when workflow not pending", func(t *testing.T) {
		workflow2, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
		workflow2.AddStep("approver-1", "manager")
		workflow2.Cancel("reason")

		err := workflow2.AddStep("approver-2", "director")
		if err == nil {
			t.Error("expected error when adding step to cancelled workflow")
		}
	})
}

func TestApprovalWorkflow_Approve(t *testing.T) {
	docID := NewDocumentID()
	workflow, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
	workflow.AddStep("approver-1", "manager")
	workflow.AddStep("approver-2", "director")

	t.Run("approves step successfully", func(t *testing.T) {
		err := workflow.Approve("approver-1", "approved by manager")
		if err != nil {
			t.Errorf("Approve() error = %v", err)
		}

		current, total := workflow.GetProgress()
		if current != 1 {
			t.Errorf("progress = %d/%d, want 1/%d", current, total, total)
		}

		step := workflow.GetCurrentStep()
		if step == nil {
			t.Fatal("expected current step to exist")
		}
		if step.GetApproverID() != "approver-2" {
			t.Errorf("current step approver = %v, want approver-2", step.GetApproverID())
		}
	})

	t.Run("fails with wrong approver", func(t *testing.T) {
		workflow2, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
		workflow2.AddStep("approver-1", "manager")

		err := workflow2.Approve("wrong-approver", "comment")
		if err == nil {
			t.Error("expected error for wrong approver")
		}
	})

	t.Run("completes workflow when all steps approved", func(t *testing.T) {
		workflow3, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
		workflow3.AddStep("approver-1", "manager")

		err := workflow3.Approve("approver-1", "approved")
		if err != nil {
			t.Errorf("Approve() error = %v", err)
		}

		if workflow3.GetStatus() != ApprovalStatusApproved {
			t.Errorf("status = %v, want %v", workflow3.GetStatus(), ApprovalStatusApproved)
		}

		if !workflow3.IsComplete() {
			t.Error("expected workflow to be complete")
		}

		events := workflow3.PullEvents()
		if len(events) != 1 {
			t.Errorf("events count = %d, want 1", len(events))
		}
	})
}

func TestApprovalWorkflow_Reject(t *testing.T) {
	docID := NewDocumentID()
	workflow, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
	workflow.AddStep("approver-1", "manager")
	workflow.AddStep("approver-2", "director")

	t.Run("rejects workflow", func(t *testing.T) {
		err := workflow.Reject("approver-1", "rejected by manager")
		if err != nil {
			t.Errorf("Reject() error = %v", err)
		}

		if workflow.GetStatus() != ApprovalStatusRejected {
			t.Errorf("status = %v, want %v", workflow.GetStatus(), ApprovalStatusRejected)
		}

		steps := workflow.GetSteps()
		if steps[0].GetStatus() != ApprovalStepStatusRejected {
			t.Errorf("first step status = %v, want %v", steps[0].GetStatus(), ApprovalStepStatusRejected)
		}
		if steps[1].GetStatus() != ApprovalStepStatusSkipped {
			t.Errorf("second step status = %v, want %v", steps[1].GetStatus(), ApprovalStepStatusSkipped)
		}
	})
}

func TestApprovalWorkflow_Cancel(t *testing.T) {
	docID := NewDocumentID()
	workflow, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
	workflow.AddStep("approver-1", "manager")

	t.Run("cancels pending workflow", func(t *testing.T) {
		err := workflow.Cancel("cancelled by user")
		if err != nil {
			t.Errorf("Cancel() error = %v", err)
		}

		if workflow.GetStatus() != ApprovalStatusCancelled {
			t.Errorf("status = %v, want %v", workflow.GetStatus(), ApprovalStatusCancelled)
		}

		steps := workflow.GetSteps()
		if steps[0].GetStatus() != ApprovalStepStatusSkipped {
			t.Errorf("step status = %v, want %v", steps[0].GetStatus(), ApprovalStepStatusSkipped)
		}
	})

	t.Run("fails when already completed", func(t *testing.T) {
		workflow2, _ := NewApprovalWorkflow(docID, DocumentTypeContract, "user-123")
		workflow2.AddStep("approver-1", "manager")
		workflow2.Approve("approver-1", "approved")

		err := workflow2.Cancel("reason")
		if err == nil {
			t.Error("expected error when cancelling completed workflow")
		}
	})
}

func TestDocumentVersion(t *testing.T) {
	t.Run("creates version with valid data", func(t *testing.T) {
		version, err := NewDocumentVersion(1, ChangeTypeCreate, "user-1", "initial version", "checksum123", "file-1")
		if err != nil {
			t.Fatalf("NewDocumentVersion() error = %v", err)
		}

		if version.GetVersion() != 1 {
			t.Errorf("version = %d, want 1", version.GetVersion())
		}

		if version.GetChangeType() != ChangeTypeCreate {
			t.Errorf("change type = %v, want %v", version.GetChangeType(), ChangeTypeCreate)
		}
	})

	t.Run("fails with invalid version number", func(t *testing.T) {
		_, err := NewDocumentVersion(0, ChangeTypeCreate, "user-1", "desc", "checksum", "file")
		if err == nil {
			t.Error("expected error for zero version number")
		}
	})

	t.Run("fails with empty changed by", func(t *testing.T) {
		_, err := NewDocumentVersion(1, ChangeTypeCreate, "", "desc", "checksum", "file")
		if err == nil {
			t.Error("expected error for empty changed by")
		}
	})
}

func TestVersionNumber(t *testing.T) {
	t.Run("creates valid version number", func(t *testing.T) {
		v, err := NewVersionNumber(5)
		if err != nil {
			t.Fatalf("NewVersionNumber() error = %v", err)
		}

		if v.Int() != 5 {
			t.Errorf("version = %d, want 5", v.Int())
		}

		if v.String() != "v5" {
			t.Errorf("version string = %s, want v5", v.String())
		}
	})

	t.Run("fails with negative number", func(t *testing.T) {
		_, err := NewVersionNumber(0)
		if err == nil {
			t.Error("expected error for zero version number")
		}
	})

	t.Run("next version", func(t *testing.T) {
		v, _ := NewVersionNumber(3)
		next := v.Next()

		if next.Int() != 4 {
			t.Errorf("next version = %d, want 4", next.Int())
		}
	})
}
