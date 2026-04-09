package domain

import (
	"errors"
	"fmt"
	"time"
)

type ApprovalWorkflowID string

func NewApprovalWorkflowID() ApprovalWorkflowID {
	return ApprovalWorkflowID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id ApprovalWorkflowID) String() string {
	return string(id)
}

type ApprovalStatus string

const (
	ApprovalStatusPending   ApprovalStatus = "pending"
	ApprovalStatusApproved  ApprovalStatus = "approved"
	ApprovalStatusRejected  ApprovalStatus = "rejected"
	ApprovalStatusCancelled ApprovalStatus = "cancelled"
)

func (s ApprovalStatus) String() string {
	return string(s)
}

func (s ApprovalStatus) IsFinal() bool {
	return s == ApprovalStatusApproved || s == ApprovalStatusRejected || s == ApprovalStatusCancelled
}

type ApprovalStepStatus string

const (
	ApprovalStepStatusPending   ApprovalStepStatus = "pending"
	ApprovalStepStatusCompleted ApprovalStepStatus = "completed"
	ApprovalStepStatusRejected  ApprovalStepStatus = "rejected"
	ApprovalStepStatusSkipped   ApprovalStepStatus = "skipped"
)

func (s ApprovalStepStatus) String() string {
	return string(s)
}

type ApprovalStep struct {
	stepNumber int
	approverID string
	role       string
	status     ApprovalStepStatus
	approvedAt *time.Time
	comment    string
}

func (s *ApprovalStep) GetStepNumber() int {
	return s.stepNumber
}

func (s *ApprovalStep) GetApproverID() string {
	return s.approverID
}

func (s *ApprovalStep) GetRole() string {
	return s.role
}

func (s *ApprovalStep) GetStatus() ApprovalStepStatus {
	return s.status
}

func (s *ApprovalStep) GetApprovedAt() *time.Time {
	return s.approvedAt
}

func (s *ApprovalStep) GetComment() string {
	return s.comment
}

type ApprovalWorkflow struct {
	id           ApprovalWorkflowID
	documentID   DocumentID
	documentType DocumentType
	steps        []ApprovalStep
	currentStep  int
	status       ApprovalStatus
	requestedBy  string
	requestedAt  time.Time
	completedAt  *time.Time
	events       []DomainEvent
}

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

func NewApprovalWorkflow(documentID DocumentID, documentType DocumentType, requestedBy string) (*ApprovalWorkflow, error) {
	if documentID.IsZero() {
		return nil, errors.New("document ID is required")
	}
	if !documentType.IsValid() {
		return nil, errors.New("invalid document type")
	}
	if requestedBy == "" {
		return nil, errors.New("requested by is required")
	}

	now := time.Now().UTC()
	return &ApprovalWorkflow{
		id:           NewApprovalWorkflowID(),
		documentID:   documentID,
		documentType: documentType,
		steps:        make([]ApprovalStep, 0),
		currentStep:  0,
		status:       ApprovalStatusPending,
		requestedBy:  requestedBy,
		requestedAt:  now,
		events:       make([]DomainEvent, 0),
	}, nil
}

func (w *ApprovalWorkflow) AddStep(approverID string, role string) error {
	if w.status != ApprovalStatusPending {
		return errors.New("cannot add steps to non-pending workflow")
	}

	step := ApprovalStep{
		stepNumber: len(w.steps) + 1,
		approverID: approverID,
		role:       role,
		status:     ApprovalStepStatusPending,
	}

	w.steps = append(w.steps, step)
	return nil
}

func (w *ApprovalWorkflow) Approve(approverID string, comment string) error {
	if w.status.IsFinal() {
		return fmt.Errorf("workflow is already %s", w.status)
	}

	if len(w.steps) == 0 {
		return errors.New("no approval steps defined")
	}

	if w.currentStep >= len(w.steps) {
		return errors.New("all steps already completed")
	}

	step := &w.steps[w.currentStep]
	if step.approverID != approverID {
		return errors.New("only assigned approver can approve")
	}

	now := time.Now().UTC()
	step.status = ApprovalStepStatusCompleted
	step.approvedAt = &now
	step.comment = comment

	w.currentStep++

	if w.currentStep >= len(w.steps) {
		w.status = ApprovalStatusApproved
		w.completedAt = &now

		w.events = append(w.events, ApprovalCompleted{
			WorkflowID:   w.id,
			DocumentID:   w.documentID,
			DocumentType: w.documentType,
			Status:       ApprovalStatusApproved,
			completedAt:  now,
		})
	}

	return nil
}

func (w *ApprovalWorkflow) Reject(approverID string, reason string) error {
	if w.status.IsFinal() {
		return fmt.Errorf("workflow is already %s", w.status)
	}

	if len(w.steps) == 0 {
		return errors.New("no approval steps defined")
	}

	if w.currentStep >= len(w.steps) {
		return errors.New("all steps already completed")
	}

	step := &w.steps[w.currentStep]
	if step.approverID != approverID {
		return errors.New("only assigned approver can reject")
	}

	now := time.Now().UTC()
	step.status = ApprovalStepStatusRejected
	step.approvedAt = &now
	step.comment = reason

	w.status = ApprovalStatusRejected
	w.completedAt = &now

	for i := w.currentStep + 1; i < len(w.steps); i++ {
		w.steps[i].status = ApprovalStepStatusSkipped
	}

	w.events = append(w.events, ApprovalCompleted{
		WorkflowID:   w.id,
		DocumentID:   w.documentID,
		DocumentType: w.documentType,
		Status:       ApprovalStatusRejected,
		completedAt:  now,
	})

	return nil
}

func (w *ApprovalWorkflow) Cancel(reason string) error {
	if w.status.IsFinal() {
		return fmt.Errorf("cannot cancel %s workflow", w.status)
	}

	now := time.Now().UTC()
	w.status = ApprovalStatusCancelled
	w.completedAt = &now

	for i := w.currentStep; i < len(w.steps); i++ {
		w.steps[i].status = ApprovalStepStatusSkipped
	}

	return nil
}

func (w *ApprovalWorkflow) GetCurrentStep() *ApprovalStep {
	if w.currentStep >= len(w.steps) {
		return nil
	}
	return &w.steps[w.currentStep]
}

func (w *ApprovalWorkflow) GetProgress() (int, int) {
	return w.currentStep, len(w.steps)
}

func (w *ApprovalWorkflow) IsComplete() bool {
	return w.status.IsFinal()
}

func (w *ApprovalWorkflow) GetID() ApprovalWorkflowID {
	return w.id
}

func (w *ApprovalWorkflow) GetDocumentID() DocumentID {
	return w.documentID
}

func (w *ApprovalWorkflow) GetDocumentType() DocumentType {
	return w.documentType
}

func (w *ApprovalWorkflow) GetSteps() []ApprovalStep {
	return w.steps
}

func (w *ApprovalWorkflow) GetStatus() ApprovalStatus {
	return w.status
}

func (w *ApprovalWorkflow) GetRequestedBy() string {
	return w.requestedBy
}

func (w *ApprovalWorkflow) GetRequestedAt() time.Time {
	return w.requestedAt
}

func (w *ApprovalWorkflow) GetCompletedAt() *time.Time {
	return w.completedAt
}

func (w *ApprovalWorkflow) PullEvents() []DomainEvent {
	events := w.events
	w.events = make([]DomainEvent, 0)
	return events
}

func (w *ApprovalWorkflow) String() string {
	return fmt.Sprintf("ApprovalWorkflow{id=%s, documentID=%s, status=%s, progress=%d/%d}",
		w.id, w.documentID, w.status, w.currentStep, len(w.steps))
}
