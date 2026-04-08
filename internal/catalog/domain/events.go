package catalog

import "time"

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// ItemCreated is published when a new catalog item is created
type ItemCreated struct {
	ItemID      ItemID
	CategoryID  *CategoryID
	SKU         string
	Name        string
	Description string
	BasePrice   float64
	Currency    string
	Status      ItemStatus
	occurredAt  time.Time
}

func (e ItemCreated) EventName() string     { return "catalog.item_created" }
func (e ItemCreated) OccurredAt() time.Time { return e.occurredAt }

// ItemPriceChanged is published when item base price is updated
type ItemPriceChanged struct {
	ItemID     ItemID
	OldPrice   float64
	NewPrice   float64
	Currency   string
	occurredAt time.Time
}

func (e ItemPriceChanged) EventName() string     { return "catalog.item_price_changed" }
func (e ItemPriceChanged) OccurredAt() time.Time { return e.occurredAt }

// ItemStatusChanged is published when item status changes (active/inactive/discontinued)
type ItemStatusChanged struct {
	ItemID     ItemID
	OldStatus  ItemStatus
	NewStatus  ItemStatus
	Reason     string
	occurredAt time.Time
}

func (e ItemStatusChanged) EventName() string     { return "catalog.item_status_changed" }
func (e ItemStatusChanged) OccurredAt() time.Time { return e.occurredAt }

// ItemDeactivated is published when item is deactivated (discontinued)
type ItemDeactivated struct {
	ItemID     ItemID
	Reason     string
	occurredAt time.Time
}

func (e ItemDeactivated) EventName() string     { return "catalog.item_deactivated" }
func (e ItemDeactivated) OccurredAt() time.Time { return e.occurredAt }

// ItemActivated is published when item is reactivated
type ItemActivated struct {
	ItemID     ItemID
	occurredAt time.Time
}

func (e ItemActivated) EventName() string     { return "catalog.item_activated" }
func (e ItemActivated) OccurredAt() time.Time { return e.occurredAt }

// ItemCategoryChanged is published when item is moved to different category
type ItemCategoryChanged struct {
	ItemID        ItemID
	OldCategoryID *CategoryID
	NewCategoryID *CategoryID
	occurredAt    time.Time
}

func (e ItemCategoryChanged) EventName() string     { return "catalog.item_category_changed" }
func (e ItemCategoryChanged) OccurredAt() time.Time { return e.occurredAt }
