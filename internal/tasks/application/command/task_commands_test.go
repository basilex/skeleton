package command

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/stretchr/testify/require"
)

type mockTaskRepo struct {
	t       *testing.T
	saved   *domain.Task
	tasks   map[domain.TaskID]*domain.Task
	created bool
}

func newMockTaskRepo(t *testing.T) *mockTaskRepo {
	return &mockTaskRepo{
		t:     t,
		tasks: make(map[domain.TaskID]*domain.Task),
	}
}

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	m.saved = task
	m.tasks[task.ID()] = task
	m.created = true
	return nil
}

func (m *mockTaskRepo) Update(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID()] = task
	return nil
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, domain.ErrTaskNotFound
}

func (m *mockTaskRepo) GetPendingTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for _, task := range m.tasks {
		if task.Status() == domain.TaskStatusPending {
			tasks = append(tasks, task)
			if len(tasks) >= limit {
				break
			}
		}
	}
	return tasks, nil
}

func (m *mockTaskRepo) GetTasksByStatus(ctx context.Context, status domain.TaskStatus, limit int) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for _, task := range m.tasks {
		if task.Status() == status {
			tasks = append(tasks, task)
			if len(tasks) >= limit {
				break
			}
		}
	}
	return tasks, nil
}

func (m *mockTaskRepo) GetTasksByType(ctx context.Context, taskType domain.TaskType, limit int) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for _, task := range m.tasks {
		if task.Type() == taskType {
			tasks = append(tasks, task)
			if len(tasks) >= limit {
				break
			}
		}
	}
	return tasks, nil
}

func (m *mockTaskRepo) GetScheduledTasks(ctx context.Context, before time.Time, limit int) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
}

func (m *mockTaskRepo) GetActiveTasks(ctx context.Context) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
}

func (m *mockTaskRepo) GetStalledTasks(ctx context.Context, olderThan time.Duration) ([]*domain.Task, error) {
	return []*domain.Task{}, nil
}

func (m *mockTaskRepo) Delete(ctx context.Context, id domain.TaskID) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) DeleteCompletedTasks(ctx context.Context, olderThan time.Duration) (int64, error) {
	return 0, nil
}

type mockScheduleRepo struct {
	schedules map[domain.ScheduleID]*domain.TaskSchedule
}

func newMockScheduleRepo() *mockScheduleRepo {
	return &mockScheduleRepo{
		schedules: make(map[domain.ScheduleID]*domain.TaskSchedule),
	}
}

func (m *mockScheduleRepo) Create(ctx context.Context, schedule *domain.TaskSchedule) error {
	m.schedules[schedule.ID()] = schedule
	return nil
}

func (m *mockScheduleRepo) Update(ctx context.Context, schedule *domain.TaskSchedule) error {
	m.schedules[schedule.ID()] = schedule
	return nil
}

func (m *mockScheduleRepo) GetByID(ctx context.Context, id domain.ScheduleID) (*domain.TaskSchedule, error) {
	if schedule, ok := m.schedules[id]; ok {
		return schedule, nil
	}
	return nil, domain.ErrScheduleNotFound
}

func (m *mockScheduleRepo) GetByName(ctx context.Context, name string) (*domain.TaskSchedule, error) {
	for _, schedule := range m.schedules {
		if schedule.Name() == name {
			return schedule, nil
		}
	}
	return nil, domain.ErrScheduleNotFound
}

func (m *mockScheduleRepo) GetActiveSchedules(ctx context.Context) ([]*domain.TaskSchedule, error) {
	var schedules []*domain.TaskSchedule
	for _, schedule := range m.schedules {
		if schedule.IsActive() {
			schedules = append(schedules, schedule)
		}
	}
	return schedules, nil
}

func (m *mockScheduleRepo) List(ctx context.Context) ([]*domain.TaskSchedule, error) {
	var schedules []*domain.TaskSchedule
	for _, schedule := range m.schedules {
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (m *mockScheduleRepo) Delete(ctx context.Context, id domain.ScheduleID) error {
	delete(m.schedules, id)
	return nil
}

type mockDeadLetterRepo struct {
	tasks map[domain.DeadLetterID]*domain.DeadLetterTask
}

func newMockDeadLetterRepo() *mockDeadLetterRepo {
	return &mockDeadLetterRepo{
		tasks: make(map[domain.DeadLetterID]*domain.DeadLetterTask),
	}
}

func (m *mockDeadLetterRepo) Create(ctx context.Context, task *domain.DeadLetterTask) error {
	m.tasks[task.ID()] = task
	return nil
}

func (m *mockDeadLetterRepo) GetByID(ctx context.Context, id domain.DeadLetterID) (*domain.DeadLetterTask, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, domain.ErrDeadLetterNotFound
}

func (m *mockDeadLetterRepo) List(ctx context.Context, limit int, offset int) ([]*domain.DeadLetterTask, error) {
	var tasks []*domain.DeadLetterTask
	for _, task := range m.tasks {
		tasks = append(tasks, task)
		if len(tasks) >= limit {
			break
		}
	}
	return tasks, nil
}

func (m *mockDeadLetterRepo) MarkReviewed(ctx context.Context, id domain.DeadLetterID, action domain.DeadLetterAction, reviewedBy *string) error {
	return nil
}

func (m *mockDeadLetterRepo) Delete(ctx context.Context, id domain.DeadLetterID) error {
	delete(m.tasks, id)
	return nil
}

func TestCreateTaskHandler_Handle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		bus := memory.New()
		handler := NewCreateTaskHandler(taskRepo, bus)

		payload := domain.TaskPayload{
			"email":   "test@example.com",
			"subject": "Test",
		}

		taskID, err := handler.Handle(context.Background(), CreateTaskCommand{
			TaskType: domain.TaskTypeSendEmail,
			Payload:  payload,
			Priority: domain.TaskPriorityNormal,
		})

		require.NoError(t, err)
		require.NotEmpty(t, taskID)
		require.True(t, taskRepo.created)
		require.NotNil(t, taskRepo.saved)
		require.Equal(t, domain.TaskTypeSendEmail, taskRepo.saved.Type())
		require.Equal(t, domain.TaskStatusPending, taskRepo.saved.Status())
		require.Equal(t, payload, taskRepo.saved.Payload())
	})

	t.Run("with priority", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		bus := memory.New()
		handler := NewCreateTaskHandler(taskRepo, bus)

		taskID, err := handler.Handle(context.Background(), CreateTaskCommand{
			TaskType: domain.TaskTypeSendSMS,
			Payload:  domain.TaskPayload{"phone": "+1234567890"},
			Priority: domain.TaskPriorityHigh,
		})

		require.NoError(t, err)
		require.NotEmpty(t, taskID)
		require.Equal(t, domain.TaskPriorityHigh, taskRepo.saved.Priority())
	})

	t.Run("with scheduled at", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		bus := memory.New()
		handler := NewCreateTaskHandler(taskRepo, bus)

		scheduledTime := time.Now().Add(1 * time.Hour)
		taskID, err := handler.Handle(context.Background(), CreateTaskCommand{
			TaskType:    domain.TaskTypeGenerateReport,
			Payload:     domain.TaskPayload{"report": "daily"},
			Priority:    domain.TaskPriorityNormal,
			ScheduledAt: &scheduledTime,
		})

		require.NoError(t, err)
		require.NotEmpty(t, taskID)
		require.Equal(t, scheduledTime.Unix(), taskRepo.saved.ScheduledAt().Unix())
	})

	t.Run("with max attempts", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		bus := memory.New()
		handler := NewCreateTaskHandler(taskRepo, bus)

		maxAttempts := 10
		taskID, err := handler.Handle(context.Background(), CreateTaskCommand{
			TaskType:    domain.TaskTypeSendPush,
			Payload:     domain.TaskPayload{"token": "abc123"},
			Priority:    domain.TaskPriorityNormal,
			MaxAttempts: &maxAttempts,
		})

		require.NoError(t, err)
		require.NotEmpty(t, taskID)
		require.Equal(t, 10, taskRepo.saved.MaxAttempts())
	})
}

func TestCancelTaskHandler_Handle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		bus := memory.New()

		// Create task first
		createHandler := NewCreateTaskHandler(taskRepo, bus)
		taskID, _ := createHandler.Handle(context.Background(), CreateTaskCommand{
			TaskType: domain.TaskTypeSendEmail,
			Payload:  domain.TaskPayload{"test": "data"},
			Priority: domain.TaskPriorityNormal,
		})

		// Cancel task
		cancelHandler := NewCancelTaskHandler(taskRepo)
		err := cancelHandler.Handle(context.Background(), CancelTaskCommand{
			TaskID: taskID,
			Reason: "User cancelled",
		})

		require.NoError(t, err)

		task, _ := taskRepo.GetByID(context.Background(), taskID)
		require.Equal(t, domain.TaskStatusCancelled, task.Status())
	})

	t.Run("task not found", func(t *testing.T) {
		taskRepo := newMockTaskRepo(t)
		handler := NewCancelTaskHandler(taskRepo)

		err := handler.Handle(context.Background(), CancelTaskCommand{
			TaskID: domain.NewTaskID(),
			Reason: "User cancelled",
		})

		require.Error(t, err)
	})
}

func TestCreateScheduleHandler_Handle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		scheduleRepo := newMockScheduleRepo()
		handler := NewCreateScheduleHandler(scheduleRepo)

		scheduleID, err := handler.Handle(context.Background(), CreateScheduleCommand{
			Name:     "daily-cleanup",
			TaskType: domain.TaskTypeCleanupOldData,
			Cron:     "0 0 * * *",
			Payload:  domain.TaskPayload{"older_than": "30d"},
		})

		require.NoError(t, err)
		require.NotEmpty(t, scheduleID)
	})

	t.Run("with timezone", func(t *testing.T) {
		scheduleRepo := newMockScheduleRepo()
		handler := NewCreateScheduleHandler(scheduleRepo)

		scheduleID, err := handler.Handle(context.Background(), CreateScheduleCommand{
			Name:     "daily-cleanup",
			TaskType: domain.TaskTypeCleanupOldData,
			Cron:     "0 0 * * *",
			Payload:  domain.TaskPayload{"older_than": "30d"},
			Timezone: "UTC",
		})

		require.NoError(t, err)
		require.NotEmpty(t, scheduleID)
	})
}

func TestDeleteScheduleHandler_Handle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		scheduleRepo := newMockScheduleRepo()
		createHandler := NewCreateScheduleHandler(scheduleRepo)

		// Create schedule first
		scheduleID, _ := createHandler.Handle(context.Background(), CreateScheduleCommand{
			Name:     "daily-cleanup",
			TaskType: domain.TaskTypeCleanupOldData,
			Cron:     "0 0 * * *",
			Payload:  domain.TaskPayload{"older_than": "30d"},
		})

		// Delete schedule
		deleteHandler := NewDeleteScheduleHandler(scheduleRepo)
		err := deleteHandler.Handle(context.Background(), DeleteScheduleCommand{ScheduleID: scheduleID})

		require.NoError(t, err)

		// Verify it's deleted
		_, err = scheduleRepo.GetByID(context.Background(), scheduleID)
		require.Error(t, err)
	})
}

func TestActivateScheduleHandler_Handle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		scheduleRepo := newMockScheduleRepo()
		createHandler := NewCreateScheduleHandler(scheduleRepo)

		// Create schedule first
		scheduleID, _ := createHandler.Handle(context.Background(), CreateScheduleCommand{
			Name:     "daily-cleanup",
			TaskType: domain.TaskTypeCleanupOldData,
			Cron:     "0 0 * * *",
			Payload:  domain.TaskPayload{"older_than": "30d"},
		})

		// Deactivate
		deactivateHandler := NewDeactivateScheduleHandler(scheduleRepo)
		err := deactivateHandler.Handle(context.Background(), DeactivateScheduleCommand{ScheduleID: scheduleID})
		require.NoError(t, err)

		schedule, _ := scheduleRepo.GetByID(context.Background(), scheduleID)
		require.False(t, schedule.IsActive())

		// Activate
		activateHandler := NewActivateScheduleHandler(scheduleRepo)
		err = activateHandler.Handle(context.Background(), ActivateScheduleCommand{ScheduleID: scheduleID})
		require.NoError(t, err)

		schedule, _ = scheduleRepo.GetByID(context.Background(), scheduleID)
		require.True(t, schedule.IsActive())
	})
}
