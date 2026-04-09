package catalog

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type ItemFilter struct {
	CategoryID *CategoryID
	Status     *ItemStatus
	Search     string
	SKU        string
	MinPrice   *float64
	MaxPrice   *float64
	Cursor     string
	Limit      int
}

type CategoryFilter struct {
	ParentID *CategoryID
	IsActive *bool
	Search   string
	Cursor   string
	Limit    int
}

type ItemRepository interface {
	Save(ctx context.Context, item *Item) error
	FindByID(ctx context.Context, id ItemID) (*Item, error)
	FindBySKU(ctx context.Context, sku string) (*Item, error)
	FindByCategory(ctx context.Context, categoryID CategoryID, filter ItemFilter) (pagination.PageResult[*Item], error)
	FindAll(ctx context.Context, filter ItemFilter) (pagination.PageResult[*Item], error)
	Delete(ctx context.Context, id ItemID) error
}

type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	FindByID(ctx context.Context, id CategoryID) (*Category, error)
	FindByPath(ctx context.Context, path string) (*Category, error)
	FindAll(ctx context.Context, filter CategoryFilter) (pagination.PageResult[*Category], error)
	FindChildren(ctx context.Context, parentID CategoryID) ([]*Category, error)
	Delete(ctx context.Context, id CategoryID) error
}
