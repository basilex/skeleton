package memory

import (
	"context"
	"log/slog"
	"sync"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]eventbus.Handler
}

func New() *Bus {
	return &Bus{
		handlers: make(map[string][]eventbus.Handler),
	}
}

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

func (b *Bus) Subscribe(eventName string, handler eventbus.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}
