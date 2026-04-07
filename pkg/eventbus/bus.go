// Package eventbus defines the event bus abstraction for publish-subscribe
// messaging patterns. It provides interfaces for event publishing and
// subscription, enabling decoupled communication between application components.
package eventbus

import (
	"context"
	"time"
)

// Event represents a domain event that can be published to the event bus.
// Implementations should provide a unique event name and timestamp.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Handler is a function type that processes events from the event bus.
// It receives the event context and the event to process, returning an error if processing fails.
type Handler func(ctx context.Context, event Event) error

// Bus defines the interface for an event bus implementation.
// Implementations must support publishing events and subscribing handlers to event types.
type Bus interface {
	// Publish sends an event to all subscribers of that event type.
	Publish(ctx context.Context, event Event) error
	// Subscribe registers a handler to be called when events of the given name are published.
	Subscribe(eventName string, handler Handler)
}
