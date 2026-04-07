// Package handler provides task handler registry implementations.
// This package contains an in-memory registry for mapping task types to their handlers.
package handler

import (
	"fmt"
	"sync"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

// InMemoryHandlerRegistry provides a thread-safe in-memory registry for task handlers.
// It maps task types to their corresponding handlers.
type InMemoryHandlerRegistry struct {
	handlers map[domain.TaskType]domain.TaskHandler
	mu       sync.RWMutex
}

// NewInMemoryHandlerRegistry creates a new empty handler registry.
func NewInMemoryHandlerRegistry() *InMemoryHandlerRegistry {
	return &InMemoryHandlerRegistry{
		handlers: make(map[domain.TaskType]domain.TaskHandler),
	}
}

// Register associates a task type with its handler.
// Returns an error if a handler is already registered for the task type.
func (r *InMemoryHandlerRegistry) Register(taskType domain.TaskType, handler domain.TaskHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("handler already registered for task type: %s", taskType)
	}

	r.handlers[taskType] = handler
	return nil
}

// Get retrieves the handler for a task type.
// Returns domain.ErrHandlerNotRegistered if no handler is registered for the task type.
func (r *InMemoryHandlerRegistry) Get(taskType domain.TaskType) (domain.TaskHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[taskType]
	if !exists {
		return nil, domain.NewErrHandlerNotRegistered(taskType)
	}

	return handler, nil
}

// Exists checks if a handler is registered for the given task type.
func (r *InMemoryHandlerRegistry) Exists(taskType domain.TaskType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.handlers[taskType]
	return exists
}
