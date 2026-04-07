// Package domain provides domain entities and value objects for the tasks module.
// This package contains the core business logic types for task management,
// including scheduled tasks, dead letter tasks, and domain events.
package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

// ScheduleID is a unique identifier for a task schedule.
type ScheduleID string

// NewScheduleID generates a new unique ScheduleID using UUID v7.
func NewScheduleID() ScheduleID {
	return ScheduleID(uuid.NewV7().String())
}

// ParseScheduleID validates and converts a string to ScheduleID.
func ParseScheduleID(s string) (ScheduleID, error) {
	if s == "" {
		return "", fmt.Errorf("schedule ID cannot be empty")
	}
	return ScheduleID(s), nil
}

// String returns the string representation of ScheduleID.
func (id ScheduleID) String() string {
	return string(id)
}

// TaskSchedule represents a scheduled task that runs on a cron expression.
type TaskSchedule struct {
	id        ScheduleID
	name      string
	taskType  TaskType
	payload   TaskPayload
	cron      string
	timezone  string
	lastRunAt *time.Time
	nextRunAt *time.Time
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
}

// NewTaskSchedule creates a new scheduled task with the provided details.
// The schedule calculates its next run time based on the cron expression.
func NewTaskSchedule(
	name string,
	taskType TaskType,
	cron string,
	payload TaskPayload,
	opts ...ScheduleOption,
) (*TaskSchedule, error) {
	if name == "" {
		return nil, fmt.Errorf("schedule name cannot be empty")
	}
	if taskType == "" {
		return nil, fmt.Errorf("task type cannot be empty")
	}
	if cron == "" {
		return nil, fmt.Errorf("cron expression cannot be empty")
	}

	now := time.Now()
	schedule := &TaskSchedule{
		id:        NewScheduleID(),
		name:      name,
		taskType:  taskType,
		cron:      cron,
		payload:   payload,
		timezone:  "UTC",
		isActive:  true,
		createdAt: now,
		updatedAt: now,
	}

	for _, opt := range opts {
		opt(schedule)
	}

	schedule.calculateNextRun()

	return schedule, nil
}

// ScheduleOption is a functional option for configuring a TaskSchedule.
type ScheduleOption func(*TaskSchedule)

// WithTimezone sets the timezone for the schedule.
func WithTimezone(timezone string) ScheduleOption {
	return func(s *TaskSchedule) {
		s.timezone = timezone
	}
}

// ID returns the schedule's unique identifier.
func (s *TaskSchedule) ID() ScheduleID {
	return s.id
}

// Name returns the schedule's name.
func (s *TaskSchedule) Name() string {
	return s.name
}

// TaskType returns the type of task to be executed.
func (s *TaskSchedule) TaskType() TaskType {
	return s.taskType
}

// Payload returns the task payload.
func (s *TaskSchedule) Payload() TaskPayload {
	return s.payload
}

// Cron returns the cron expression for the schedule.
func (s *TaskSchedule) Cron() string {
	return s.cron
}

// Timezone returns the timezone for the schedule.
func (s *TaskSchedule) Timezone() string {
	return s.timezone
}

// LastRunAt returns when the schedule was last executed.
func (s *TaskSchedule) LastRunAt() *time.Time {
	return s.lastRunAt
}

// NextRunAt returns when the schedule will next execute.
func (s *TaskSchedule) NextRunAt() *time.Time {
	return s.nextRunAt
}

// IsActive returns whether the schedule is active.
func (s *TaskSchedule) IsActive() bool {
	return s.isActive
}

// CreatedAt returns when the schedule was created.
func (s *TaskSchedule) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns when the schedule was last updated.
func (s *TaskSchedule) UpdatedAt() time.Time {
	return s.updatedAt
}

// Activate enables the schedule for execution.
func (s *TaskSchedule) Activate() {
	s.isActive = true
	s.updatedAt = time.Now()
}

// Deactivate disables the schedule from execution.
func (s *TaskSchedule) Deactivate() {
	s.isActive = false
	s.updatedAt = time.Now()
}

// UpdateNextRun sets the next run time for the schedule.
func (s *TaskSchedule) UpdateNextRun(nextRun time.Time) {
	s.nextRunAt = &nextRun
	s.updatedAt = time.Now()
}

// MarkRun records that the schedule has been executed and calculates the next run time.
func (s *TaskSchedule) MarkRun(runTime time.Time) {
	s.lastRunAt = &runTime
	s.calculateNextRun()
	s.updatedAt = time.Now()
}

// ShouldRun returns whether the schedule should be executed at the given time.
func (s *TaskSchedule) ShouldRun(now time.Time) bool {
	if !s.isActive {
		return false
	}
	if s.nextRunAt == nil {
		return false
	}
	return s.nextRunAt.Before(now) || s.nextRunAt.Equal(now)
}

// calculateNextRun calculates the next execution time based on the cron expression.
func (s *TaskSchedule) calculateNextRun() {
	next := calculateNextCronRun(s.cron, s.timezone, s.lastRunAt)
	s.nextRunAt = &next
}

// calculateNextCronRun calculates the next run time for a cron expression.
func calculateNextCronRun(cronExpr string, timezone string, lastRun *time.Time) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	if lastRun != nil {
		now = lastRun.In(loc)
	}

	minute := now.Minute()
	hour := now.Hour()
	day := now.Day()

	switch cronExpr {
	case "@every_minute":
		next := now.Add(1 * time.Minute)
		return next
	case "@hourly":
		next := now.Add(1 * time.Hour).Truncate(time.Hour)
		return next
	case "@daily":
		next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
		return next
	case "@weekly":
		daysUntilNextWeek := 7 - int(now.Weekday())
		next := now.Add(time.Duration(daysUntilNextWeek) * 24 * time.Hour).Truncate(24 * time.Hour)
		return next
	default:
		minute++
		if minute >= 60 {
			minute = 0
			hour++
			if hour >= 24 {
				hour = 0
				day++
			}
		}
		return time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, loc)
	}
}
