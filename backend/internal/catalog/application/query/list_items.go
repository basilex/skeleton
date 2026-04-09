package query

import (
	"context"

	catalog "github.com/basilex/skeleton/internal/catalog/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

type ListItemsHandler struct {
	items catalog.ItemRepository
}

func NewListItemsHandler(items catalog.ItemRepository) *ListItemsHandler {
	return &ListItemsHandler{items: items}
}

type ListItemsQuery struct {
	CategoryID *string
	Status     *string
	Search     string
	Cursor     string
	Limit      int
}

func (h *ListItemsHandler) Handle(ctx context.Context, q ListItemsQuery) (pagination.PageResult[ItemDTO], error) {
	filter := catalog.ItemFilter{
		Search: q.Search,
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}

	if q.CategoryID != nil {
		categoryID, err := catalog.ParseCategoryID(*q.CategoryID)
		if err == nil {
			filter.CategoryID = &categoryID
		}
	}
	if q.Status != nil {
		status, err := catalog.ParseItemStatus(*q.Status)
		if err == nil {
			filter.Status = &status
		}
	}

	result, err := h.items.FindAll(ctx, filter)
	if err != nil {
		return pagination.PageResult[ItemDTO]{}, err
	}

	dtos := make([]ItemDTO, 0, len(result.Items))
	for _, item := range result.Items {
		var categoryID *string
		if item.GetCategoryID() != nil {
			cid := item.GetCategoryID().String()
			categoryID = &cid
		}

		dtos = append(dtos, ItemDTO{
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
		})
	}

	return pagination.PageResult[ItemDTO]{
		Items:      dtos,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		Limit:      result.Limit,
	}, nil
}
