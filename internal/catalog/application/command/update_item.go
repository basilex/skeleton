package command

import (
	"context"
	"fmt"

	catalog "github.com/basilex/skeleton/internal/catalog/domain"
)

type UpdateItemHandler struct {
	items catalog.ItemRepository
}

func NewUpdateItemHandler(items catalog.ItemRepository) *UpdateItemHandler {
	return &UpdateItemHandler{items: items}
}

type UpdateItemCommand struct {
	ItemID      string
	Name        *string
	Description *string
	BasePrice   *float64
}

func (h *UpdateItemHandler) Handle(ctx context.Context, cmd UpdateItemCommand) error {
	itemID, err := catalog.ParseItemID(cmd.ItemID)
	if err != nil {
		return fmt.Errorf("parse item id: %w", err)
	}

	item, err := h.items.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("find item: %w", err)
	}

	if cmd.Name != nil {
		item.UpdateName(*cmd.Name)
	}
	if cmd.Description != nil {
		item.UpdateDescription(*cmd.Description)
	}
	if cmd.BasePrice != nil {
		if err := item.UpdatePrice(*cmd.BasePrice); err != nil {
			return fmt.Errorf("update price: %w", err)
		}
	}

	if err := h.items.Save(ctx, item); err != nil {
		return fmt.Errorf("save item: %w", err)
	}

	return nil
}
