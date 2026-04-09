package eventhandler

import (
	"context"
	"testing"

	"github.com/basilex/skeleton/pkg/eventbus"
)

func TestIdentityEventHandler_Subscriptions(t *testing.T) {
	bus := &mockBus{}
	handler := NewIdentityEventHandler(nil)

	handler.Register(bus)

	expectedEvents := []string{
		"identity.user_registered",
		"identity.role_assigned",
		"identity.role_revoked",
		"identity.login",
		"identity.logout",
	}

	if len(bus.subscriptions) != len(expectedEvents) {
		t.Errorf("expected %d subscriptions, got %d", len(expectedEvents), len(bus.subscriptions))
	}

	for _, eventName := range expectedEvents {
		if _, exists := bus.subscriptions[eventName]; !exists {
			t.Errorf("expected subscription for event %s", eventName)
		}
	}
}

type mockBus struct {
	subscriptions map[string]eventbus.Handler
}

func (m *mockBus) Publish(ctx context.Context, event eventbus.Event) error {
	return nil
}

func (m *mockBus) Subscribe(eventName string, handler eventbus.Handler) {
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]eventbus.Handler)
	}
	m.subscriptions[eventName] = handler
}
