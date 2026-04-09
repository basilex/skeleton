package query

import (
	"context"
	"fmt"

	catalog "github.com/basilex/skeleton/internal/catalog/domain"
)

type GetItemHandler struct {
	items catalog.ItemRepository
}

func NewGetItemHandler(items catalog.ItemRepository) *GetItemHandler {
	return &GetItemHandler{items: items}
}

type GetItemQuery struct {
	ItemID string
}

type ItemDTO struct {
	ID           string                 `json:"id"`
	CategoryID   *string                `json:"category_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	SKU          string                 `json:"sku"`
	BasePrice    float64                `json:"base_price"`
	Currency     string                 `json:"currency"`
	Status       string                 `json:"status"`
	Attributes   map[string]interface{} `json:"attributes"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

func (h *GetItemHandler) Handle(ctx context.Context, q GetItemQuery) (ItemDTO, error) {
	itemID, err := catalog.ParseItemID(q.ItemID)
	if err != nil {
		return ItemDTO{}, fmt.Errorf("parse item id: %w", err)
	}

	item, err := h.items.FindByID(ctx, itemID)
	if err != nil {
		return ItemDTO{}, fmt.Errorf("find item: %w", err)
	}

	var categoryID *string
	if item.GetCategoryID() != nil {
		cid := item.GetCategoryID().String()
		categoryID = &cid
	}

	return ItemDTO{
		ID:          item.GetID().String(),
		CategoryID:  categoryID,
		Name:        item.GetName(),
		Description: item.GetDescription(),
		SKU:         item.GetSKU(),
		BasePrice:   item.GetBasePrice(),
		Currency:    item.GetCurrency(),
		Status:      item.GetStatus().String(),
		Attributes:  item.GetAttributes(),
		CreatedAt:   item.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   item.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
