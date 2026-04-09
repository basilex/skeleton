package domain

import (
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type OrderCreated struct {
	OrderID    OrderID
	CustomerID string
	SupplierID string
	Total      money.Money
	Currency   string
	occurredAt time.Time
}

func (e OrderCreated) EventName() string     { return "ordering.order_created" }
func (e OrderCreated) OccurredAt() time.Time { return e.occurredAt }

type OrderStatusChanged struct {
	OrderID    OrderID
	OldStatus  OrderStatus
	NewStatus  OrderStatus
	occurredAt time.Time
}

func (e OrderStatusChanged) EventName() string     { return "ordering.order_status_changed" }
func (e OrderStatusChanged) OccurredAt() time.Time { return e.occurredAt }

type OrderConfirmed struct {
	OrderID     OrderID
	CustomerID  string
	SupplierID  string
	WarehouseID string
	Lines       []OrderConfirmedLine
	Total       money.Money
	Currency    string
	occurredAt  time.Time
}

type OrderConfirmedLine struct {
	ItemID    string
	ItemName  string
	Quantity  float64
	Unit      string
	UnitPrice money.Money
	Discount  money.Money
	Total     money.Money
}

func (e OrderConfirmed) EventName() string     { return "ordering.order_confirmed" }
func (e OrderConfirmed) OccurredAt() time.Time { return e.occurredAt }

type OrderCancelled struct {
	OrderID    OrderID
	CustomerID string
	Reason     string
	occurredAt time.Time
}

func (e OrderCancelled) EventName() string     { return "ordering.order_cancelled" }
func (e OrderCancelled) OccurredAt() time.Time { return e.occurredAt }

type OrderCompleted struct {
	OrderID    OrderID
	CustomerID string
	Total      money.Money
	occurredAt time.Time
}

func (e OrderCompleted) EventName() string     { return "ordering.order_completed" }
func (e OrderCompleted) OccurredAt() time.Time { return e.occurredAt }

type QuoteCreated struct {
	QuoteID    string
	CustomerID string
	SupplierID string
	Total      money.Money
	Currency   string
	occurredAt time.Time
}

func (e QuoteCreated) EventName() string     { return "ordering.quote_created" }
func (e QuoteCreated) OccurredAt() time.Time { return e.occurredAt }
