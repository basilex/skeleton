package handler

import (
	"fmt"
	"sync"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type InMemoryHandlerRegistry struct {
	handlers map[domain.TaskType]domain.TaskHandler
	mu       sync.RWMutex
}

func NewInMemoryHandlerRegistry() *InMemoryHandlerRegistry {
	return &InMemoryHandlerRegistry{
		handlers: make(map[domain.TaskType]domain.TaskHandler),
	}
}

func (r *InMemoryHandlerRegistry) Register(taskType domain.TaskType, handler domain.TaskHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("handler already registered for task type: %s", taskType)
	}

	r.handlers[taskType] = handler
	return nil
}

func (r *InMemoryHandlerRegistry) Get(taskType domain.TaskType) (domain.TaskHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[taskType]
	if !exists {
		return nil, domain.NewErrHandlerNotRegistered(taskType)
	}

	return handler, nil
}

func (r *InMemoryHandlerRegistry) Exists(taskType domain.TaskType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.handlers[taskType]
	return exists
}
