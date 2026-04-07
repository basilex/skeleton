package eventbus

import (
	"context"
	"time"
)

type Event interface {
	EventName() string
	OccurredAt() time.Time
}

type Handler func(ctx context.Context, event Event) error

type Bus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventName string, handler Handler)
}
