package catalog

import (
	"errors"
	"fmt"
	"time"
)

type VariantID string

func NewVariantID() VariantID {
	return VariantID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id VariantID) String() string {
	return string(id)
}

type VariantStatus string

const (
	VariantStatusActive       VariantStatus = "active"
	VariantStatusInactive     VariantStatus = "inactive"
	VariantStatusDiscontinued VariantStatus = "discontinued"
)

func (s VariantStatus) String() string {
	return string(s)
}

type ProductVariant struct {
	id           VariantID
	itemID       ItemID
	sku          string
	attributes   Attributes
	priceAdjust  float64
	stock        int
	stockTracked bool
	status       VariantStatus
	createdAt    time.Time
	updatedAt    time.Time
	events       []DomainEvent
}

func NewProductVariant(itemID ItemID, sku string, attributes Attributes, priceAdjust float64) (*ProductVariant, error) {
	if itemID.IsZero() {
		return nil, errors.New("item ID is required")
	}
	if sku == "" {
		return nil, errors.New("SKU is required")
	}
	if len(attributes) == 0 {
		return nil, errors.New("at least one attribute is required")
	}

	now := time.Now().UTC()
	variant := &ProductVariant{
		id:           NewVariantID(),
		itemID:       itemID,
		sku:          sku,
		attributes:   attributes,
		priceAdjust:  priceAdjust,
		stock:        0,
		stockTracked: false,
		status:       VariantStatusActive,
		createdAt:    now,
		updatedAt:    now,
		events:       make([]DomainEvent, 0),
	}

	variant.events = append(variant.events, VariantCreated{
		VariantID:   variant.id,
		ItemID:      variant.itemID,
		SKU:         variant.sku,
		Attributes:  variant.attributes,
		PriceAdjust: variant.priceAdjust,
		occurredAt:  now,
	})

	return variant, nil
}

func (v *ProductVariant) GetID() VariantID          { return v.id }
func (v *ProductVariant) GetItemID() ItemID         { return v.itemID }
func (v *ProductVariant) GetSKU() string            { return v.sku }
func (v *ProductVariant) GetAttributes() Attributes { return v.attributes }
func (v *ProductVariant) GetPriceAdjust() float64   { return v.priceAdjust }
func (v *ProductVariant) GetStock() int             { return v.stock }
func (v *ProductVariant) IsStockTracked() bool      { return v.stockTracked }
func (v *ProductVariant) GetStatus() VariantStatus  { return v.status }
func (v *ProductVariant) GetCreatedAt() time.Time   { return v.createdAt }
func (v *ProductVariant) GetUpdatedAt() time.Time   { return v.updatedAt }

func (v *ProductVariant) UpdatePriceAdjust(priceAdjust float64) error {
	if v.status == VariantStatusDiscontinued {
		return errors.New("cannot update discontinued variant")
	}

	oldAdjust := v.priceAdjust
	v.priceAdjust = priceAdjust
	v.updatedAt = time.Now().UTC()

	if oldAdjust != priceAdjust {
		v.events = append(v.events, VariantPriceAdjustChanged{
			VariantID:  v.id,
			OldAdjust:  oldAdjust,
			NewAdjust:  priceAdjust,
			occurredAt: v.updatedAt,
		})
	}

	return nil
}

func (v *ProductVariant) UpdateStock(quantity int) error {
	if !v.stockTracked {
		return errors.New("stock tracking is not enabled for this variant")
	}
	if quantity < 0 {
		return errors.New("stock cannot be negative")
	}

	oldStock := v.stock
	v.stock = quantity
	v.updatedAt = time.Now().UTC()

	v.events = append(v.events, VariantStockUpdated{
		VariantID:  v.id,
		OldStock:   oldStock,
		NewStock:   quantity,
		occurredAt: v.updatedAt,
	})

	return nil
}

func (v *ProductVariant) AdjustStock(delta int) error {
	if !v.stockTracked {
		return errors.New("stock tracking is not enabled for this variant")
	}

	newStock := v.stock + delta
	if newStock < 0 {
		return fmt.Errorf("insufficient stock: current %d, adjust by %d", v.stock, delta)
	}

	return v.UpdateStock(newStock)
}

func (v *ProductVariant) EnableStockTracking() {
	if v.stockTracked {
		return
	}
	v.stockTracked = true
	v.updatedAt = time.Now().UTC()
}

func (v *ProductVariant) DisableStockTracking() {
	if !v.stockTracked {
		return
	}
	v.stockTracked = false
	v.stock = 0
	v.updatedAt = time.Now().UTC()
}

func (v *ProductVariant) UpdateAttribute(key string, value interface{}) {
	v.attributes[key] = value
	v.updatedAt = time.Now().UTC()
}

func (v *ProductVariant) Activate() {
	if v.status == VariantStatusActive {
		return
	}

	oldStatus := v.status
	v.status = VariantStatusActive
	v.updatedAt = time.Now().UTC()

	v.events = append(v.events, VariantStatusChanged{
		VariantID:  v.id,
		OldStatus:  oldStatus,
		NewStatus:  VariantStatusActive,
		occurredAt: v.updatedAt,
	})
}

func (v *ProductVariant) Deactivate() {
	if v.status == VariantStatusInactive {
		return
	}

	oldStatus := v.status
	v.status = VariantStatusInactive
	v.updatedAt = time.Now().UTC()

	v.events = append(v.events, VariantStatusChanged{
		VariantID:  v.id,
		OldStatus:  oldStatus,
		NewStatus:  VariantStatusInactive,
		occurredAt: v.updatedAt,
	})
}

func (v *ProductVariant) Discontinue() {
	if v.status == VariantStatusDiscontinued {
		return
	}

	oldStatus := v.status
	v.status = VariantStatusDiscontinued
	v.updatedAt = time.Now().UTC()

	v.events = append(v.events, VariantStatusChanged{
		VariantID:  v.id,
		OldStatus:  oldStatus,
		NewStatus:  VariantStatusDiscontinued,
		occurredAt: v.updatedAt,
	})
}

func (v *ProductVariant) PullEvents() []DomainEvent {
	events := v.events
	v.events = make([]DomainEvent, 0)
	return events
}

func (v *ProductVariant) String() string {
	return fmt.Sprintf("ProductVariant{id=%s, sku=%s, attributes=%v, status=%s}",
		v.id, v.sku, v.attributes, v.status)
}

func ReconstituteProductVariant(
	id VariantID,
	itemID ItemID,
	sku string,
	attributes Attributes,
	priceAdjust float64,
	stock int,
	stockTracked bool,
	status VariantStatus,
	createdAt, updatedAt time.Time,
) *ProductVariant {
	return &ProductVariant{
		id:           id,
		itemID:       itemID,
		sku:          sku,
		attributes:   attributes,
		priceAdjust:  priceAdjust,
		stock:        stock,
		stockTracked: stockTracked,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		events:       make([]DomainEvent, 0),
	}
}
