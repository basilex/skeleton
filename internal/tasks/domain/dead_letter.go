package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

type DeadLetterID string

func NewDeadLetterID() DeadLetterID {
	return DeadLetterID(uuid.NewV7().String())
}

func ParseDeadLetterID(s string) (DeadLetterID, error) {
	if s == "" {
		return "", nil
	}
	return DeadLetterID(s), nil
}

func (id DeadLetterID) String() string {
	return string(id)
}

type DeadLetterAction string

const (
	DeadLetterActionRetry       DeadLetterAction = "retry"
	DeadLetterActionIgnore      DeadLetterAction = "ignore"
	DeadLetterActionInvestigate DeadLetterAction = "investigate"
)

func (a DeadLetterAction) String() string {
	return string(a)
}

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

func (d *DeadLetterTask) ID() DeadLetterID {
	return d.id
}

func (d *DeadLetterTask) OriginalTask() *Task {
	return d.originalTask
}

func (d *DeadLetterTask) FailedAt() time.Time {
	return d.failedAt
}

func (d *DeadLetterTask) Reason() string {
	return d.reason
}

func (d *DeadLetterTask) IsReviewed() bool {
	return d.reviewed
}

func (d *DeadLetterTask) ReviewedAt() *time.Time {
	return d.reviewedAt
}

func (d *DeadLetterTask) ReviewedBy() *string {
	return d.reviewedBy
}

func (d *DeadLetterTask) Action() DeadLetterAction {
	return d.action
}

func (d *DeadLetterTask) CreatedAt() time.Time {
	return d.createdAt
}

func (d *DeadLetterTask) MarkReviewed(action DeadLetterAction, reviewedBy *string) {
	d.reviewed = true
	now := time.Now()
	d.reviewedAt = &now
	d.reviewedBy = reviewedBy
	d.action = action
}

func (d *DeadLetterTask) CanRetry() bool {
	return d.originalTask != nil && d.originalTask.CanRetry()
}
