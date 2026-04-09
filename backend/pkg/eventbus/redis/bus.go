// Package redisbus provides a Redis-based implementation of the event bus interface.
// It uses Redis Pub/Sub for distributed event messaging across multiple service instances.
package redisbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/redis/go-redis/v9"
)

// Bus implements the eventbus.Bus interface using Redis Pub/Sub.
// It safely handles concurrent access to handlers and provides automatic
// message deserialization and dispatch.
type Bus struct {
	client   *redis.Client
	mu       sync.RWMutex
	handlers map[string][]eventbus.Handler
	pubsub   *redis.PubSub
	ctx      context.Context
	cancel   context.CancelFunc
}

// envelope wraps an event for JSON serialization over Redis.
type envelope struct {
	EventName string          `json:"event_name"`
	Occurred  string          `json:"occurred_at"`
	Payload   json.RawMessage `json:"payload"`
}

// New creates a new Redis-based event bus with the given Redis client.
// It starts a background goroutine to listen for published events.
func New(client *redis.Client) *Bus {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bus{
		client:   client,
		handlers: make(map[string][]eventbus.Handler),
		ctx:      ctx,
		cancel:   cancel,
	}
	go b.listen()
	return b
}

// Publish serializes the event and publishes it to Redis for distribution to all subscribers.
func (b *Bus) Publish(ctx context.Context, event eventbus.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	env := envelope{
		EventName: event.EventName(),
		Occurred:  event.OccurredAt().Format("2006-01-02T15:04:05Z"),
		Payload:   payload,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	return b.client.Publish(ctx, event.EventName(), data).Err()
}

// Subscribe registers a handler for the specified event name.
// Multiple handlers can be registered for the same event.
func (b *Bus) Subscribe(eventName string, handler eventbus.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// listen runs in a background goroutine to receive messages from Redis subscriptions
// and dispatch them to registered handlers.
func (b *Bus) listen() {
	b.mu.RLock()
	channels := make([]string, 0, len(b.handlers))
	for ch := range b.handlers {
		channels = append(channels, ch)
	}
	b.mu.RUnlock()

	if len(channels) == 0 {
		return
	}

	b.pubsub = b.client.Subscribe(b.ctx, channels...)
	ch := b.pubsub.Channel()

	for {
		select {
		case <-b.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var env envelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				slog.Error("redisbus: unmarshal envelope", "error", err)
				continue
			}
			b.mu.RLock()
			handlers := b.handlers[env.EventName]
			b.mu.RUnlock()

			for _, h := range handlers {
				func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("redisbus: handler panic recovered",
								"event", env.EventName,
								"recover", r,
							)
						}
					}()
					if err := h(b.ctx, &redisEvent{
						name:     env.EventName,
						occurred: env.Occurred,
						raw:      env.Payload,
					}); err != nil {
						slog.Error("redisbus: handler error",
							"event", env.EventName,
							"error", err,
						)
					}
				}()
			}
		}
	}
}

// Close stops the event bus and releases all Redis subscriptions.
// It cancels the background listener and closes the PubSub connection.
func (b *Bus) Close() error {
	b.cancel()
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	return nil
}

// redisEvent is an unmarshaled event received from Redis.
type redisEvent struct {
	name     string
	occurred string
	raw      json.RawMessage
}

// EventName returns the name of the event.
func (e *redisEvent) EventName() string {
	return e.name
}

// OccurredAt returns the timestamp when the event occurred.
func (e *redisEvent) OccurredAt() time.Time {
	t, _ := time.Parse(time.RFC3339, e.occurred)
	return t
}

// Payload returns the raw JSON payload of the event.
func (e *redisEvent) Payload() json.RawMessage {
	return e.raw
}
