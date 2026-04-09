package http

type CreateItemRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	SKU          string                 `json:"sku" binding:"required"`
	BasePrice    float64                 `json:"base_price" binding:"required,gt=0"`
	Currency     string                 `json:"currency" binding:"required,oneof=UAH USD EUR GBP"`
	CategoryID   *string                `json:"category_id"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type UpdateItemRequest struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	BasePrice    *float64 `json:"base_price" binding:"omitempty,gt=0"`
}

type ListItemsRequest struct {
	CategoryID *string `form:"category_id"`
	Status     *string `form:"status" binding:"omitempty,oneof=active inactive discontinued"`
	Search     string  `form:"search"`
	Cursor     string  `form:"cursor"`
	Limit      int     `form:"limit"`
}
