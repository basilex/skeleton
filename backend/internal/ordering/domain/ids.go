package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type OrderID uuid.UUID

func NewOrderID() OrderID {
	return OrderID(uuid.NewV7())
}

func ParseOrderID(s string) (OrderID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return OrderID{}, fmt.Errorf("invalid order id: %w", err)
	}
	return OrderID(u), nil
}

func MustParseOrderID(s string) OrderID {
	id, err := ParseOrderID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (id OrderID) String() string {
	return uuid.UUID(id).String()
}

type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

func (os OrderStatus) String() string {
	return string(os)
}

func ParseOrderStatus(s string) (OrderStatus, error) {
	switch OrderStatus(s) {
	case OrderStatusDraft, OrderStatusPending, OrderStatusConfirmed,
		OrderStatusProcessing, OrderStatusCompleted, OrderStatusCancelled, OrderStatusRefunded:
		return OrderStatus(s), nil
	default:
		return "", fmt.Errorf("invalid order status: %s", s)
	}
}

type QuoteStatus string

const (
	QuoteStatusDraft    QuoteStatus = "draft"
	QuoteStatusSent     QuoteStatus = "sent"
	QuoteStatusAccepted QuoteStatus = "accepted"
	QuoteStatusRejected QuoteStatus = "rejected"
	QuoteStatusExpired  QuoteStatus = "expired"
)

func (qs QuoteStatus) String() string {
	return string(qs)
}

func ParseQuoteStatus(s string) (QuoteStatus, error) {
	switch QuoteStatus(s) {
	case QuoteStatusDraft, QuoteStatusSent, QuoteStatusAccepted,
		QuoteStatusRejected, QuoteStatusExpired:
		return QuoteStatus(s), nil
	default:
		return "", fmt.Errorf("invalid quote status: %s", s)
	}
}

type OrderLineID uuid.UUID

func NewOrderLineID() OrderLineID {
	return OrderLineID(uuid.NewV7())
}

func ParseOrderLineID(s string) (OrderLineID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return OrderLineID{}, fmt.Errorf("invalid order line id: %w", err)
	}
	return OrderLineID(u), nil
}

func MustParseOrderLineID(s string) OrderLineID {
	id, err := ParseOrderLineID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func (id OrderLineID) String() string {
	return uuid.UUID(id).String()
}
