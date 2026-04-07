// Package memory provides an in-memory implementation of the event bus interface.
// It is suitable for single-process applications and testing scenarios
// where distributed messaging is not required.
package memory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/basilex/skeleton/pkg/eventbus"
)

// Bus implements the eventbus.Bus interface using in-memory storage.
// It is thread-safe and suitable for single-process applications.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]eventbus.Handler
}

// New creates a new in-memory event bus.
func New() *Bus {
	return &Bus{
		handlers: make(map[string][]eventbus.Handler),
	}
}

// Publish synchronously calls all registered handlers for the event's name.
// Errors from handlers are logged but do not prevent other handlers from being called.
func (b *Bus) Publish(ctx context.Context, event eventbus.Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "eventbus: handler panic recovered",
						"event", event.EventName(),
						"recover", r,
					)
				}
			}()
			if err := h(ctx, event); err != nil {
				slog.ErrorContext(ctx, "eventbus: handler error",
					"event", event.EventName(),
					"error", err,
				)
			}
		}()
	}
	return nil
}

// Subscribe registers a handler for the specified event name.
// Multiple handlers can be registered for the same event.
func (b *Bus) Subscribe(eventName string, handler eventbus.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}
