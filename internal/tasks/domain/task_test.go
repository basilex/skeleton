package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	payload := TaskPayload{
		"email":   "test@example.com",
		"subject": "Test",
	}

	task, err := NewTask(TaskTypeSendEmail, payload)
	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotEmpty(t, task.ID())
	require.Equal(t, TaskTypeSendEmail, task.Type())
	require.Equal(t, TaskStatusPending, task.Status())
	require.Equal(t, TaskPriorityNormal, task.Priority())
	require.Equal(t, 0, task.Attempts())
	require.Equal(t, 5, task.MaxAttempts())
	require.False(t, task.CreatedAt().IsZero())
}

func TestNewTaskWithPriority(t *testing.T) {
	payload := TaskPayload{"key": "value"}

	task, err := NewTask(TaskTypeSendSMS, payload, WithTaskPriority(TaskPriorityHigh))
	require.NoError(t, err)
	require.Equal(t, TaskPriorityHigh, task.Priority())
}

func TestNewTaskWithMaxAttempts(t *testing.T) {
	payload := TaskPayload{"key": "value"}

	task, err := NewTask(TaskTypeSendPush, payload, WithTaskMaxAttempts(10))
	require.NoError(t, err)
	require.Equal(t, 10, task.MaxAttempts())
}

func TestNewTaskWithScheduledAt(t *testing.T) {
	payload := TaskPayload{"key": "value"}
	scheduledTime := time.Now().Add(1 * time.Hour)

	task, err := NewTask(TaskTypeGenerateReport, payload, WithTaskScheduledAt(scheduledTime))
	require.NoError(t, err)
	require.Equal(t, scheduledTime.Unix(), task.ScheduledAt().Unix())
}

func TestNewTaskEmptyType(t *testing.T) {
	payload := TaskPayload{"key": "value"}

	_, err := NewTask("", payload)
	require.Error(t, err)
	require.Contains(t, err.Error(), "task type cannot be empty")
}

func TestTaskStart(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	err := task.Start()
	require.NoError(t, err)
	require.Equal(t, TaskStatusRunning, task.Status())
	require.NotNil(t, task.StartedAt())
}

func TestTaskStartFromWrongStatus(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	_ = task.Start()

	err := task.Start()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot start task")
}

func TestTaskComplete(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	_ = task.Start()

	result := &TaskResult{
		Data:       map[string]interface{}{"sent": true},
		DurationMs: 150,
	}

	err := task.Complete(result)
	require.NoError(t, err)
	require.Equal(t, TaskStatusCompleted, task.Status())
	require.NotNil(t, task.CompletedAt())
	require.Equal(t, result, task.Result())
}

func TestTaskCompleteFromWrongStatus(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	err := task.Complete(&TaskResult{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot complete task")
}

func TestTaskFail(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	_ = task.Start()

	err := task.Fail("Connection refused")
	require.NoError(t, err)
	require.Equal(t, TaskStatusFailed, task.Status())
	require.NotNil(t, task.CompletedAt())
	require.NotNil(t, task.Error())
	require.Equal(t, "Connection refused", task.Error().Message)
}

func TestTaskRetry(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	_ = task.Start()
	_ = task.Fail("Connection error")

	delay := 5 * time.Minute
	err := task.Retry(delay)
	require.NoError(t, err)
	require.Equal(t, TaskStatusPending, task.Status())
	require.Equal(t, 1, task.Attempts())
	require.Nil(t, task.Error())
}

func TestTaskRetryFromWrongStatus(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	err := task.Retry(5 * time.Minute)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can only retry failed tasks")
}

func TestTaskCancel(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	err := task.Cancel("user request")
	require.NoError(t, err)
	require.Equal(t, TaskStatusCancelled, task.Status())
	require.NotNil(t, task.Error())
	require.Equal(t, "user request", task.Error().Message)
}

func TestTaskCancelFromWrongStatus(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	_ = task.Start()
	_ = task.Complete(&TaskResult{})

	err := task.Cancel("user request")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot cancel task")
}

func TestTaskCanRetry(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{}, WithTaskMaxAttempts(3))

	// Pending tasks cannot retry
	require.False(t, task.CanRetry())

	_ = task.Start()
	// Running tasks cannot retry
	require.False(t, task.CanRetry())

	_ = task.Fail("error")
	// Failed tasks can retry if attempts < maxAttempts
	require.True(t, task.CanRetry())

	// After Retry, attempts increased to 1
	_ = task.Retry(1 * time.Second)
	require.Equal(t, 1, task.Attempts())
	// Retry sets status to Pending, so cannot retry yet
	require.False(t, task.CanRetry())

	_ = task.Start()
	_ = task.Fail("error")
	// Failed again, can retry (attempts=1 < 3)
	require.True(t, task.CanRetry())

	_ = task.Retry(1 * time.Second)
	require.Equal(t, 2, task.Attempts())

	_ = task.Start()
	_ = task.Fail("error")
	// attempts=2 < 3, can still retry
	require.True(t, task.CanRetry())

	_ = task.Retry(1 * time.Second)
	require.Equal(t, 3, task.Attempts())

	_ = task.Start()
	_ = task.Fail("final error")
	// attempts=3, maxAttempts=3, cannot retry
	require.False(t, task.CanRetry())
}

func TestTaskNextRetryDelay(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	// First attempt (attempts=0)
	require.Equal(t, 1*time.Second, task.NextRetryDelay())

	// Second attempt (attempts=1)
	task.IncrementAttempts()
	require.Equal(t, 5*time.Second, task.NextRetryDelay())

	// Third attempt (attempts=2)
	task.IncrementAttempts()
	require.Equal(t, 15*time.Second, task.NextRetryDelay())

	// Fourth attempt (attempts=3)
	task.IncrementAttempts()
	require.Equal(t, 1*time.Minute, task.NextRetryDelay())

	// Fifth attempt (attempts=4)
	task.IncrementAttempts()
	require.Equal(t, 5*time.Minute, task.NextRetryDelay())

	// Sixth attempt (attempts=5)
	task.IncrementAttempts()
	require.Equal(t, 15*time.Minute, task.NextRetryDelay())

	// Seventh attempt (attempts=6)
	task.IncrementAttempts()
	require.Equal(t, 1*time.Hour, task.NextRetryDelay())

	// Eighth attempt (attempts=7)
	task.IncrementAttempts()
	require.Equal(t, 6*time.Hour, task.NextRetryDelay())

	// Beyond limit (attempts=8+) returns last delay
	task.IncrementAttempts()
	require.Equal(t, 6*time.Hour, task.NextRetryDelay())
}

func TestTaskIncrementAttempts(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})

	require.Equal(t, 0, task.Attempts())

	task.IncrementAttempts()
	require.Equal(t, 1, task.Attempts())

	task.IncrementAttempts()
	require.Equal(t, 2, task.Attempts())
}

func TestTaskIsScheduled(t *testing.T) {
	task, _ := NewTask(TaskTypeSendEmail, TaskPayload{})
	require.False(t, task.IsScheduled())

	scheduledTime := time.Now().Add(1 * time.Hour)
	task2, _ := NewTask(TaskTypeSendEmail, TaskPayload{}, WithTaskScheduledAt(scheduledTime))
	require.True(t, task2.IsScheduled())
}

func TestTaskPayload(t *testing.T) {
	payload := TaskPayload{
		"email":   "test@example.com",
		"subject": "Test Subject",
		"data": map[string]interface{}{
			"key": "value",
		},
	}

	task, err := NewTask(TaskTypeSendEmail, payload)
	require.NoError(t, err)
	require.Equal(t, payload, task.Payload())
}

func TestParseTaskID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ID",
			input:   "019d65d6-de90-7200-b1cf-4f8745597e0a",
			wantErr: false,
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseTaskID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.input, id.String())
			}
		})
	}
}

func TestParseTaskType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TaskType
		wantErr bool
	}{
		{name: "send_email", input: "send_email", want: TaskTypeSendEmail, wantErr: false},
		{name: "send_sms", input: "send_sms", want: TaskTypeSendSMS, wantErr: false},
		{name: "send_push", input: "send_push", want: TaskTypeSendPush, wantErr: false},
		{name: "send_in_app", input: "send_in_app", want: TaskTypeSendInApp, wantErr: false},
		{name: "process_file", input: "process_file", want: TaskTypeProcessFile, wantErr: false},
		{name: "generate_thumbnail", input: "generate_thumbnail", want: TaskTypeGenerateThumbnail, wantErr: false},
		{name: "cleanup_old_data", input: "cleanup_old_data", want: TaskTypeCleanupOldData, wantErr: false},
		{name: "generate_report", input: "generate_report", want: TaskTypeGenerateReport, wantErr: false},
		{name: "sync_external_api", input: "sync_external_api", want: TaskTypeSyncExternalAPI, wantErr: false},
		{name: "batch_operation", input: "batch_operation", want: TaskTypeBatchOperation, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskType, err := ParseTaskType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, taskType)
			}
		})
	}
}

func TestParseTaskStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TaskStatus
		wantErr bool
	}{
		{name: "pending", input: "pending", want: TaskStatusPending, wantErr: false},
		{name: "queued", input: "queued", want: TaskStatusQueued, wantErr: false},
		{name: "running", input: "running", want: TaskStatusRunning, wantErr: false},
		{name: "completed", input: "completed", want: TaskStatusCompleted, wantErr: false},
		{name: "failed", input: "failed", want: TaskStatusFailed, wantErr: false},
		{name: "cancelled", input: "cancelled", want: TaskStatusCancelled, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParseTaskStatus(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, status)
			}
		})
	}
}

func TestParseTaskPriority(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TaskPriority
		wantErr bool
	}{
		{name: "low", input: "low", want: TaskPriorityLow, wantErr: false},
		{name: "normal", input: "normal", want: TaskPriorityNormal, wantErr: false},
		{name: "high", input: "high", want: TaskPriorityHigh, wantErr: false},
		{name: "critical", input: "critical", want: TaskPriorityCritical, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority, err := ParseTaskPriority(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, priority)
			}
		})
	}
}

func TestTaskTypeString(t *testing.T) {
	require.Equal(t, "send_email", TaskTypeSendEmail.String())
	require.Equal(t, "send_sms", TaskTypeSendSMS.String())
}

func TestTaskStatusString(t *testing.T) {
	require.Equal(t, "pending", TaskStatusPending.String())
	require.Equal(t, "running", TaskStatusRunning.String())
}

func TestTaskPriorityString(t *testing.T) {
	require.Equal(t, "low", TaskPriorityLow.String())
	require.Equal(t, "critical", TaskPriorityCritical.String())
}
