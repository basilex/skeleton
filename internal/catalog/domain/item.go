package catalog

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Attributes map[string]interface{}

func (a Attributes) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("marshal attributes: %w", err)
	}
	return json.RawMessage(data), nil
}

func AttributesFromJSON(data json.RawMessage) (Attributes, error) {
	var a Attributes
	if len(data) == 0 {
		return make(Attributes), nil
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("unmarshal attributes: %w", err)
	}
	return a, nil
}

type Item struct {
	id          ItemID
	categoryID  *CategoryID
	sku         string
	name        string
	description string
	basePrice   float64
	currency    string
	status      ItemStatus
	attributes  Attributes
	metadata    map[string]interface{}
	createdAt   time.Time
	updatedAt   time.Time
	events      []eventbus.Event
}

func NewItem(
	categoryID *CategoryID,
	sku, name, description string,
	basePrice float64,
	currency string,
) (*Item, error) {
	if name == "" {
		return nil, fmt.Errorf("item name is required")
	}
	if basePrice < 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now().UTC()
	item := &Item{
		id:          NewItemID(),
		categoryID:  categoryID,
		sku:         sku,
		name:        name,
		description: description,
		basePrice:   basePrice,
		currency:    currency,
		status:      ItemStatusActive,
		attributes:  make(Attributes),
		metadata:    make(map[string]interface{}),
		createdAt:   now,
		updatedAt:   now,
		events:      make([]eventbus.Event, 0),
	}

	item.events = append(item.events, ItemCreated{
		ItemID:      item.id,
		CategoryID:  item.categoryID,
		SKU:         item.sku,
		Name:        item.name,
		Description: item.description,
		BasePrice:   item.basePrice,
		Currency:    item.currency,
		Status:      item.status,
		occurredAt:  now,
	})

	return item, nil
}

func (i *Item) GetID() ItemID              { return i.id }
func (i *Item) GetCategoryID() *CategoryID { return i.categoryID }
func (i *Item) GetSKU() string             { return i.sku }
func (i *Item) GetName() string            { return i.name }
func (i *Item) GetDescription() string     { return i.description }
func (i *Item) GetBasePrice() float64      { return i.basePrice }
func (i *Item) GetCurrency() string        { return i.currency }
func (i *Item) GetStatus() ItemStatus      { return i.status }
func (i *Item) GetAttributes() Attributes  { return i.attributes }
func (i *Item) GetCreatedAt() time.Time    { return i.createdAt }
func (i *Item) GetUpdatedAt() time.Time    { return i.updatedAt }

func (i *Item) UpdatePrice(price float64) error {
	if price < 0 {
		return ErrInvalidPrice
	}

	oldPrice := i.basePrice
	i.basePrice = price
	i.updatedAt = time.Now().UTC()

	if oldPrice != price {
		i.events = append(i.events, ItemPriceChanged{
			ItemID:     i.id,
			OldPrice:   oldPrice,
			NewPrice:   price,
			Currency:   i.currency,
			occurredAt: i.updatedAt,
		})
	}

	return nil
}

func (i *Item) UpdateCategory(categoryID *CategoryID) {
	oldCategoryID := i.categoryID
	i.categoryID = categoryID
	i.updatedAt = time.Now().UTC()

	// Publish event only if category actually changed
	if (oldCategoryID != nil || categoryID != nil) &&
		(oldCategoryID == nil || categoryID == nil || *oldCategoryID != *categoryID) {
		i.events = append(i.events, ItemCategoryChanged{
			ItemID:        i.id,
			OldCategoryID: oldCategoryID,
			NewCategoryID: categoryID,
			occurredAt:    i.updatedAt,
		})
	}
}

func (i *Item) SetAttribute(key string, value interface{}) {
	i.attributes[key] = value
	i.updatedAt = time.Now().UTC()
}

func (i *Item) RemoveAttribute(key string) {
	delete(i.attributes, key)
	i.updatedAt = time.Now().UTC()
}

func (i *Item) Activate() {
	if i.status == ItemStatusActive {
		return
	}

	i.status = ItemStatusActive
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, ItemActivated{
		ItemID:     i.id,
		occurredAt: i.updatedAt,
	})
}

func (i *Item) Deactivate() {
	if i.status == ItemStatusInactive {
		return
	}

	i.status = ItemStatusInactive
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, ItemDeactivated{
		ItemID:     i.id,
		Reason:     "Item deactivated",
		occurredAt: i.updatedAt,
	})
}

func (i *Item) Discontinue() {
	if i.status == ItemStatusDiscontinued {
		return
	}

	oldStatus := i.status
	i.status = ItemStatusDiscontinued
	i.updatedAt = time.Now().UTC()

	i.events = append(i.events, ItemStatusChanged{
		ItemID:     i.id,
		OldStatus:  oldStatus,
		NewStatus:  ItemStatusDiscontinued,
		Reason:     "Item discontinued",
		occurredAt: i.updatedAt,
	})
}

func ReconstituteItem(
	id ItemID,
	categoryID *CategoryID,
	sku, name, description string,
	basePrice float64,
	currency string,
	status ItemStatus,
	attributes Attributes,
	metadata map[string]interface{},
	createdAt, updatedAt time.Time,
) (*Item, error) {
	return &Item{
		id:          id,
		categoryID:  categoryID,
		sku:         sku,
		name:        name,
		description: description,
		basePrice:   basePrice,
		currency:    currency,
		status:      status,
		attributes:  attributes,
		metadata:    metadata,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]eventbus.Event, 0),
	}, nil
}

// PullEvents returns all pending domain events and clears the event list
func (i *Item) PullEvents() []eventbus.Event {
	events := i.events
	i.events = make([]eventbus.Event, 0)
	return events
}

// Additional methods for application layer

func (i *Item) UpdateName(name string) {
	if name != "" {
		i.name = name
		i.updatedAt = time.Now().UTC()
	}
}

func (i *Item) UpdateDescription(description string) {
	i.description = description
	i.updatedAt = time.Now().UTC()
}

type Category struct {
	id          CategoryID
	name        string
	description string
	parentID    *CategoryID
	path        string
	isActive    bool
	metadata    map[string]interface{}
	createdAt   time.Time
	updatedAt   time.Time
}

func NewCategory(name, description string, parentID *CategoryID) (*Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	now := time.Now().UTC()
	return &Category{
		id:          NewCategoryID(),
		name:        name,
		description: description,
		parentID:    parentID,
		path:        "", // Will be set by repository
		isActive:    true,
		metadata:    make(map[string]interface{}),
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func (c *Category) GetID() CategoryID        { return c.id }
func (c *Category) GetName() string          { return c.name }
func (c *Category) GetDescription() string   { return c.description }
func (c *Category) GetParentID() *CategoryID { return c.parentID }
func (c *Category) GetPath() string          { return c.path }
func (c *Category) IsActive() bool           { return c.isActive }
func (c *Category) GetCreatedAt() time.Time  { return c.createdAt }
func (c *Category) GetUpdatedAt() time.Time  { return c.updatedAt }

func (c *Category) Deactivate() {
	c.isActive = false
	c.updatedAt = time.Now().UTC()
}

func (c *Category) Activate() {
	c.isActive = true
	c.updatedAt = time.Now().UTC()
}

func ReconstituteCategory(
	id CategoryID,
	name, description string,
	parentID *CategoryID,
	path string,
	isActive bool,
	metadata map[string]interface{},
	createdAt, updatedAt time.Time,
) (*Category, error) {
	return &Category{
		id:          id,
		name:        name,
		description: description,
		parentID:    parentID,
		path:        path,
		isActive:    isActive,
		metadata:    metadata,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}
