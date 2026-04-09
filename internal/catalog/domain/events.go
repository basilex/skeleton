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

// VariantCreated is published when a new product variant is created
type VariantCreated struct {
	VariantID   VariantID
	ItemID      ItemID
	SKU         string
	Attributes  Attributes
	PriceAdjust float64
	occurredAt  time.Time
}

func (e VariantCreated) EventName() string     { return "catalog.variant_created" }
func (e VariantCreated) OccurredAt() time.Time { return e.occurredAt }

// VariantPriceAdjustChanged is published when variant price adjustment changes
type VariantPriceAdjustChanged struct {
	VariantID  VariantID
	OldAdjust  float64
	NewAdjust  float64
	occurredAt time.Time
}

func (e VariantPriceAdjustChanged) EventName() string     { return "catalog.variant_price_adjust_changed" }
func (e VariantPriceAdjustChanged) OccurredAt() time.Time { return e.occurredAt }

// VariantStockUpdated is published when variant stock is updated
type VariantStockUpdated struct {
	VariantID  VariantID
	OldStock   int
	NewStock   int
	occurredAt time.Time
}

func (e VariantStockUpdated) EventName() string     { return "catalog.variant_stock_updated" }
func (e VariantStockUpdated) OccurredAt() time.Time { return e.occurredAt }

// VariantStatusChanged is published when variant status changes
type VariantStatusChanged struct {
	VariantID  VariantID
	OldStatus  VariantStatus
	NewStatus  VariantStatus
	occurredAt time.Time
}

func (e VariantStatusChanged) EventName() string     { return "catalog.variant_status_changed" }
func (e VariantStatusChanged) OccurredAt() time.Time { return e.occurredAt }

// PricingRuleCreated is published when a new pricing rule is created
type PricingRuleCreated struct {
	RuleID      PricingRuleID
	Name        string
	RuleType    PricingRuleType
	MinQuantity int
	occurredAt  time.Time
}

func (e PricingRuleCreated) EventName() string     { return "catalog.pricing_rule_created" }
func (e PricingRuleCreated) OccurredAt() time.Time { return e.occurredAt }

// ItemVariantAdded is published when a variant is added to an item
type ItemVariantAdded struct {
	ItemID     ItemID
	VariantID  VariantID
	occurredAt time.Time
}

func (e ItemVariantAdded) EventName() string     { return "catalog.item_variant_added" }
func (e ItemVariantAdded) OccurredAt() time.Time { return e.occurredAt }

// ItemVariantRemoved is published when a variant is removed from an item
type ItemVariantRemoved struct {
	ItemID     ItemID
	VariantID  VariantID
	occurredAt time.Time
}

func (e ItemVariantRemoved) EventName() string     { return "catalog.item_variant_removed" }
func (e ItemVariantRemoved) OccurredAt() time.Time { return e.occurredAt }
