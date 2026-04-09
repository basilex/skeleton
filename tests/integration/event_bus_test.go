package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
	membus "github.com/basilex/skeleton/pkg/eventbus/memory"
)

type TestEvent struct {
	ID   string
	Name string
}

func (e TestEvent) EventName() string {
	return "test.event"
}

func (e TestEvent) OccurredAt() time.Time {
	return time.Now()
}

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := membus.New()

	var received TestEvent
	var mu sync.Mutex

	bus.Subscribe("test.event", func(ctx context.Context, event eventbus.Event) error {
		mu.Lock()
		defer mu.Unlock()
		received = event.(TestEvent)
		return nil
	})

	event := TestEvent{
		ID:   "123",
		Name: "Test Event",
	}

	err := bus.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("failed to publish event: %s", err)
	}

	mu.Lock()
	if received.ID != event.ID {
		t.Errorf("expected ID %s, got %s", event.ID, received.ID)
	}
	if received.Name != event.Name {
		t.Errorf("expected Name %s, got %s", event.Name, received.Name)
	}
	mu.Unlock()
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := membus.New()

	var counter int
	var mu sync.Mutex

	bus.Subscribe("test.event", func(ctx context.Context, event eventbus.Event) error {
		mu.Lock()
		defer mu.Unlock()
		counter++
		return nil
	})

	bus.Subscribe("test.event", func(ctx context.Context, event eventbus.Event) error {
		mu.Lock()
		defer mu.Unlock()
		counter++
		return nil
	})

	event := TestEvent{ID: "123"}
	err := bus.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("failed to publish event: %s", err)
	}

	mu.Lock()
	if counter != 2 {
		t.Errorf("expected 2 handlers to be called, got %d", counter)
	}
	mu.Unlock()
}

func TestEventBus_HandlerError(t *testing.T) {
	bus := membus.New()

	bus.Subscribe("test.event", func(ctx context.Context, event eventbus.Event) error {
		return fmt.Errorf("handler error")
	})

	event := TestEvent{ID: "123"}
	err := bus.Publish(context.Background(), event)
	if err == nil {
		t.Error("expected error from handler, got nil")
	}
}
