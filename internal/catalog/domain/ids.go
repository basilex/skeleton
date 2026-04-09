package catalog

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type ItemID uuid.UUID

func NewItemID() ItemID {
	return ItemID(uuid.NewV7())
}

func ParseItemID(s string) (ItemID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ItemID{}, fmt.Errorf("invalid item id: %w", err)
	}
	return ItemID(u), nil
}

func (id ItemID) String() string {
	return uuid.UUID(id).String()
}

type CategoryID uuid.UUID

func NewCategoryID() CategoryID {
	return CategoryID(uuid.NewV7())
}

func ParseCategoryID(s string) (CategoryID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return CategoryID{}, fmt.Errorf("invalid category id: %w", err)
	}
	return CategoryID(u), nil
}

func (id CategoryID) String() string {
	return uuid.UUID(id).String()
}

type ItemStatus string

const (
	ItemStatusActive       ItemStatus = "active"
	ItemStatusInactive     ItemStatus = "inactive"
	ItemStatusDiscontinued ItemStatus = "discontinued"
)

func (is ItemStatus) String() string {
	return string(is)
}

func ParseItemStatus(s string) (ItemStatus, error) {
	switch ItemStatus(s) {
	case ItemStatusActive, ItemStatusInactive, ItemStatusDiscontinued:
		return ItemStatus(s), nil
	default:
		return "", fmt.Errorf("invalid item status: %s", s)
	}
}

type PriceType string

const (
	PriceTypeBase      PriceType = "base"
	PriceTypeSale      PriceType = "sale"
	PriceTypeWholesale PriceType = "wholesale"
	PriceTypePartner   PriceType = "partner"
)

func (pt PriceType) String() string {
	return string(pt)
}
