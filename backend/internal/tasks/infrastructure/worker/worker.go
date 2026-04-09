// Package worker provides background worker implementations for the tasks domain.
// This package contains the task worker that processes pending and scheduled tasks.
package worker

import (
	"context"
	"sync"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// Worker processes pending tasks concurrently.
// It runs as a background worker, polling for tasks to execute.
type Worker struct {
	taskRepo        domain.TaskRepository
	deadLetterRepo  domain.DeadLetterRepository
	handlerRegistry domain.TaskHandlerRegistry
	eventBus        eventbus.Bus
	pollInterval    time.Duration
	maxConcurrent   int
	taskTimeout     time.Duration

	stopChan  chan struct{}
	semaphore chan struct{}
	wg        sync.WaitGroup
}

// NewWorker creates a new task worker with the provided dependencies and configuration.
func NewWorker(
	taskRepo domain.TaskRepository,
	deadLetterRepo domain.DeadLetterRepository,
	handlerRegistry domain.TaskHandlerRegistry,
	eventBus eventbus.Bus,
	pollInterval time.Duration,
	maxConcurrent int,
	taskTimeout time.Duration,
) *Worker {
	return &Worker{
		taskRepo:        taskRepo,
		deadLetterRepo:  deadLetterRepo,
		handlerRegistry: handlerRegistry,
		eventBus:        eventBus,
		pollInterval:    pollInterval,
		maxConcurrent:   maxConcurrent,
		taskTimeout:     taskTimeout,
		stopChan:        make(chan struct{}),
		semaphore:       make(chan struct{}, maxConcurrent),
	}
}

// Start begins the worker's main processing loop.
// It continuously polls for tasks to execute until the context is cancelled.
func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.stopChan:
			return nil
		case <-ticker.C:
			w.processPendingTasks(ctx)
		}
	}
}

// Stop gracefully stops the worker and waits for all running tasks to complete.
func (w *Worker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

// processPendingTasks retrieves and executes pending tasks concurrently.
func (w *Worker) processPendingTasks(ctx context.Context) {
	tasks, err := w.taskRepo.GetPendingTasks(ctx, w.maxConcurrent)
	if err != nil {
		return
	}

	for _, task := range tasks {
		w.semaphore <- struct{}{}
		w.wg.Add(1)

		go func(t *domain.Task) {
			defer func() {
				<-w.semaphore
				w.wg.Done()
			}()

			w.executeTask(ctx, t)
		}(task)
	}
}

// executeTask executes a single task with timeout handling.
func (w *Worker) executeTask(ctx context.Context, task *domain.Task) {
	if err := task.Start(); err != nil {
		return
	}

	if err := w.taskRepo.Update(ctx, task); err != nil {
		return
	}

	w.eventBus.Publish(ctx, domain.NewTaskStarted(task.ID()))

	handler, err := w.handlerRegistry.Get(task.Type())
	if err != nil {
		w.handleTaskFailure(ctx, task, err)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, w.taskTimeout)
	defer cancel()

	startTime := time.Now()
	result, err := handler.Execute(timeoutCtx, task.Payload())
	durationMs := time.Since(startTime).Milliseconds()

	if err != nil {
		w.handleTaskFailure(ctx, task, err)
		return
	}

	task.Complete(result)

	if err := w.taskRepo.Update(ctx, task); err != nil {
		return
	}

	w.eventBus.Publish(ctx, domain.NewTaskCompleted(
		task.ID(),
		task.Type(),
		result,
		durationMs,
	))
}

// handleTaskFailure handles task execution failures, either scheduling retry or moving to dead letter.
func (w *Worker) handleTaskFailure(ctx context.Context, task *domain.Task, taskErr error) {
	task.IncrementAttempts()

	if task.CanRetry() {
		delay := task.NextRetryDelay()
		task.Retry(delay)

		if err := w.taskRepo.Update(ctx, task); err != nil {
			return
		}

		w.eventBus.Publish(ctx, domain.NewTaskRetrying(
			task.ID(),
			task.Attempts(),
			task.ScheduledAt(),
		))
	} else {
		task.Fail(taskErr.Error())

		if err := w.taskRepo.Update(ctx, task); err != nil {
			return
		}

		deadLetter := domain.NewDeadLetterTask(task, taskErr.Error())
		w.deadLetterRepo.Create(ctx, deadLetter)

		w.eventBus.Publish(ctx, domain.NewTaskFailed(
			task.ID(),
			task.Type(),
			&domain.TaskError{
				Message: taskErr.Error(),
			},
			task.Attempts(),
			false,
			nil,
		))
	}
}
