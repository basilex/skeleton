package persistence

import (
	"encoding/json"
	"time"

	catalog "github.com/basilex/skeleton/internal/catalog/domain"
)

type itemDTO struct {
	ID          string          `db:"id"`
	CategoryID   *string         `db:"category_id"`
	Name         string          `db:"name"`
	Description  string          `db:"description"`
	SKU          string          `db:"sku"`
	BasePrice    float64         `db:"base_price"`
	Currency     string          `db:"currency"`
	Status       string          `db:"status"`
	Attributes   json.RawMessage `db:"attributes"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

type categoryDTO struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Path        string    `db:"path"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (dto *itemDTO) toDomain() (*catalog.Item, error) {
	itemID, err := catalog.ParseItemID(dto.ID)
	if err != nil {
		return nil, err
	}

	status, err := catalog.ParseItemStatus(dto.Status)
	if err != nil {
		return nil, err
	}

	var categoryID *catalog.CategoryID
	if dto.CategoryID != nil {
		cid, err := catalog.ParseCategoryID(*dto.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = &cid
	}

	var attributes catalog.Attributes
	if len(dto.Attributes) > 0 {
		attributes, err = catalog.AttributesFromJSON(dto.Attributes)
		if err != nil {
			return nil, err
		}
	} else {
		attributes = make(catalog.Attributes)
	}

	return catalog.ReconstituteItem(
		itemID,
		categoryID,
		dto.SKU,
		dto.Name,
		dto.Description,
		dto.BasePrice,
		dto.Currency,
		status,
		attributes,
		make(map[string]interface{}),
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func (dto *categoryDTO) toDomain() (*catalog.Category, error) {
	categoryID, err := catalog.ParseCategoryID(dto.ID)
	if err != nil {
		return nil, err
	}

	return catalog.ReconstituteCategory(
		categoryID,
		dto.Name,
		dto.Description,
		nil, // parentID - not stored in this DTO
		dto.Path,
		dto.IsActive,
		make(map[string]interface{}),
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

func toItemDTO(item *catalog.Item) *itemDTO {
	var categoryID *string
	if item.GetCategoryID() != nil {
		cid := item.GetCategoryID().String()
		categoryID = &cid
	}

	attrs, _ := item.GetAttributes().ToJSON()

	return &itemDTO{
		ID:          item.GetID().String(),
		CategoryID:  categoryID,
		Name:        item.GetName(),
		Description: item.GetDescription(),
		SKU:         item.GetSKU(),
		BasePrice:   item.GetBasePrice(),
		Currency:    item.GetCurrency(),
		Status:      item.GetStatus().String(),
		Attributes:  attrs,
		CreatedAt:   item.GetCreatedAt(),
		UpdatedAt:   item.GetUpdatedAt(),
	}
}

func toCategoryDTO(category *catalog.Category) *categoryDTO {
	return &categoryDTO{
		ID:          category.GetID().String(),
		Name:        category.GetName(),
		Description: category.GetDescription(),
		Path:        category.GetPath(),
		IsActive:    category.IsActive(),
		CreatedAt:   category.GetCreatedAt(),
		UpdatedAt:   category.GetUpdatedAt(),
	}
}
