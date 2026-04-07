package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/stretchr/testify/require"
)

type testEvent struct {
	name string
	time time.Time
}

func (e testEvent) EventName() string     { return e.name }
func (e testEvent) OccurredAt() time.Time { return e.time }

func TestBusPublishSubscribe(t *testing.T) {
	bus := New()
	ctx := context.Background()

	var received eventbus.Event
	bus.Subscribe("test.event", func(ctx context.Context, e eventbus.Event) error {
		received = e
		return nil
	})

	evt := testEvent{name: "test.event", time: time.Now()}
	err := bus.Publish(ctx, evt)
	require.NoError(t, err)
	require.NotNil(t, received)
	require.Equal(t, "test.event", received.EventName())
}

func TestBusMultipleHandlers(t *testing.T) {
	bus := New()
	ctx := context.Background()

	var results []string
	var mu sync.Mutex

	bus.Subscribe("test.event", func(ctx context.Context, e eventbus.Event) error {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, "handler1")
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, e eventbus.Event) error {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, "handler2")
		return nil
	})

	evt := testEvent{name: "test.event", time: time.Now()}
	err := bus.Publish(ctx, evt)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestBusPanicRecovery(t *testing.T) {
	bus := New()
	ctx := context.Background()

	bus.Subscribe("test.event", func(ctx context.Context, e eventbus.Event) error {
		panic("test panic")
	})

	evt := testEvent{name: "test.event", time: time.Now()}
	err := bus.Publish(ctx, evt)
	require.NoError(t, err)
}

func TestBusHandlerError(t *testing.T) {
	bus := New()
	ctx := context.Background()

	bus.Subscribe("test.event", func(ctx context.Context, e eventbus.Event) error {
		return errors.New("handler error")
	})

	evt := testEvent{name: "test.event", time: time.Now()}
	err := bus.Publish(ctx, evt)
	require.NoError(t, err)
}

func TestBusNoHandlers(t *testing.T) {
	bus := New()
	ctx := context.Background()

	evt := testEvent{name: "unknown.event", time: time.Now()}
	err := bus.Publish(ctx, evt)
	require.NoError(t, err)
}
