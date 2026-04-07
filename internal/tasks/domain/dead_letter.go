// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// DeadLetterID is a unique identifier for a dead letter task.
type DeadLetterID string

// NewDeadLetterID generates a new unique DeadLetterID using UUID v7.
func NewDeadLetterID() DeadLetterID {
	return DeadLetterID(uuid.NewV7().String())
}

// ParseDeadLetterID validates and converts a string to DeadLetterID.
func ParseDeadLetterID(s string) (DeadLetterID, error) {
	if s == "" {
		return "", nil
	}
	return DeadLetterID(s), nil
}

// String returns the string representation of DeadLetterID.
func (id DeadLetterID) String() string {
	return string(id)
}

// DeadLetterAction represents the action to take on a dead letter task.
type DeadLetterAction string

// Dead letter action constants.
const (
	DeadLetterActionRetry       DeadLetterAction = "retry"
	DeadLetterActionIgnore      DeadLetterAction = "ignore"
	DeadLetterActionInvestigate DeadLetterAction = "investigate"
)

// String returns the string representation of DeadLetterAction.
func (a DeadLetterAction) String() string {
	return string(a)
}

// DeadLetterTask represents a task that has exceeded its maximum retry attempts.
// These tasks are stored for manual review and potential reprocessing.
type DeadLetterTask struct {
	id           DeadLetterID
	originalTask *Task
	failedAt     time.Time
	reason       string
	reviewed     bool
	reviewedAt   *time.Time
	reviewedBy   *string
	action       DeadLetterAction
	createdAt    time.Time
}

// NewDeadLetterTask creates a new dead letter task from a failed task.
func NewDeadLetterTask(task *Task, reason string) *DeadLetterTask {
	return &DeadLetterTask{
		id:           NewDeadLetterID(),
		originalTask: task,
		failedAt:     time.Now(),
		reason:       reason,
		reviewed:     false,
		createdAt:    time.Now(),
	}
}

// ID returns the dead letter task's unique identifier.
func (d *DeadLetterTask) ID() DeadLetterID {
	return d.id
}

// OriginalTask returns the original task that failed.
func (d *DeadLetterTask) OriginalTask() *Task {
	return d.originalTask
}

// FailedAt returns when the original task failed.
func (d *DeadLetterTask) FailedAt() time.Time {
	return d.failedAt
}

// Reason returns the failure reason.
func (d *DeadLetterTask) Reason() string {
	return d.reason
}

// IsReviewed returns whether the dead letter task has been reviewed.
func (d *DeadLetterTask) IsReviewed() bool {
	return d.reviewed
}

// ReviewedAt returns when the dead letter task was reviewed.
func (d *DeadLetterTask) ReviewedAt() *time.Time {
	return d.reviewedAt
}

// ReviewedBy returns who reviewed the dead letter task.
func (d *DeadLetterTask) ReviewedBy() *string {
	return d.reviewedBy
}

// Action returns the action taken on the dead letter task.
func (d *DeadLetterTask) Action() DeadLetterAction {
	return d.action
}

// CreatedAt returns when the dead letter task was created.
func (d *DeadLetterTask) CreatedAt() time.Time {
	return d.createdAt
}

// MarkReviewed marks the dead letter task as reviewed with the specified action.
func (d *DeadLetterTask) MarkReviewed(action DeadLetterAction, reviewedBy *string) {
	d.reviewed = true
	now := time.Now()
	d.reviewedAt = &now
	d.reviewedBy = reviewedBy
	d.action = action
}

// CanRetry returns whether the dead letter task can be retried.
func (d *DeadLetterTask) CanRetry() bool {
	return d.originalTask != nil && d.originalTask.CanRetry()
}
