package command

import (
	"context"
	"fmt"

	catalog "github.com/basilex/skeleton/internal/catalog/domain"
)

type CreateItemHandler struct {
	items catalog.ItemRepository
}

func NewCreateItemHandler(items catalog.ItemRepository) *CreateItemHandler {
	return &CreateItemHandler{items: items}
}

type CreateItemCommand struct {
	CategoryID  *string
	SKU         string
	Name        string
	Description string
	BasePrice   float64
	Currency    string
	Attributes  map[string]interface{}
}

type CreateItemResult struct {
	ItemID string
}

func (h *CreateItemHandler) Handle(ctx context.Context, cmd CreateItemCommand) (CreateItemResult, error) {
	var categoryID *catalog.CategoryID
	if cmd.CategoryID != nil {
		cid, err := catalog.ParseCategoryID(*cmd.CategoryID)
		if err != nil {
			return CreateItemResult{}, fmt.Errorf("parse category id: %w", err)
		}
		categoryID = &cid
	}

	item, err := catalog.NewItem(categoryID, cmd.SKU, cmd.Name, cmd.Description, cmd.BasePrice, cmd.Currency)
	if err != nil {
		return CreateItemResult{}, fmt.Errorf("create item: %w", err)
	}

	if cmd.Attributes != nil {
		for key, value := range cmd.Attributes {
			item.SetAttribute(key, value)
		}
	}

	if err := h.items.Save(ctx, item); err != nil {
		return CreateItemResult{}, fmt.Errorf("save item: %w", err)
	}

	return CreateItemResult{
		ItemID: item.GetID().String(),
	}, nil
}
