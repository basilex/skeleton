package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/uuid"
)

type ScheduleID string

func NewScheduleID() ScheduleID {
	return ScheduleID(uuid.NewV7().String())
}

func ParseScheduleID(s string) (ScheduleID, error) {
	if s == "" {
		return "", fmt.Errorf("schedule ID cannot be empty")
	}
	return ScheduleID(s), nil
}

func (id ScheduleID) String() string {
	return string(id)
}

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

type ScheduleOption func(*TaskSchedule)

func WithTimezone(timezone string) ScheduleOption {
	return func(s *TaskSchedule) {
		s.timezone = timezone
	}
}

func (s *TaskSchedule) ID() ScheduleID {
	return s.id
}

func (s *TaskSchedule) Name() string {
	return s.name
}

func (s *TaskSchedule) TaskType() TaskType {
	return s.taskType
}

func (s *TaskSchedule) Payload() TaskPayload {
	return s.payload
}

func (s *TaskSchedule) Cron() string {
	return s.cron
}

func (s *TaskSchedule) Timezone() string {
	return s.timezone
}

func (s *TaskSchedule) LastRunAt() *time.Time {
	return s.lastRunAt
}

func (s *TaskSchedule) NextRunAt() *time.Time {
	return s.nextRunAt
}

func (s *TaskSchedule) IsActive() bool {
	return s.isActive
}

func (s *TaskSchedule) CreatedAt() time.Time {
	return s.createdAt
}

func (s *TaskSchedule) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *TaskSchedule) Activate() {
	s.isActive = true
	s.updatedAt = time.Now()
}

func (s *TaskSchedule) Deactivate() {
	s.isActive = false
	s.updatedAt = time.Now()
}

func (s *TaskSchedule) UpdateNextRun(nextRun time.Time) {
	s.nextRunAt = &nextRun
	s.updatedAt = time.Now()
}

func (s *TaskSchedule) MarkRun(runTime time.Time) {
	s.lastRunAt = &runTime
	s.calculateNextRun()
	s.updatedAt = time.Now()
}

func (s *TaskSchedule) ShouldRun(now time.Time) bool {
	if !s.isActive {
		return false
	}
	if s.nextRunAt == nil {
		return false
	}
	return s.nextRunAt.Before(now) || s.nextRunAt.Equal(now)
}

func (s *TaskSchedule) calculateNextRun() {
	next := calculateNextCronRun(s.cron, s.timezone, s.lastRunAt)
	s.nextRunAt = &next
}

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
