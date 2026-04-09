package http

type CreateOrderRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
	CustomerID  string `json:"customer_id" binding:"required"`
	SupplierID  string `json:"supplier_id" binding:"required"`
	ContractID  string `json:"contract_id"`
	Currency    string `json:"currency" binding:"required"`
}

type AddOrderLineRequest struct {
	ItemID    string  `json:"item_id" binding:"required"`
	ItemName  string  `json:"item_name" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unit_price" binding:"required,gt=0"`
	Discount  float64 `json:"discount"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=confirmed completed cancelled"`
	Reason string `json:"reason"`
}

type ListOrdersRequest struct {
	CustomerID *string `form:"customer_id"`
	SupplierID *string `form:"supplier_id"`
	Status     *string `form:"status" binding:"omitempty,oneof=draft pending confirmed processing completed cancelled refunded"`
	StartDate  *string `form:"start_date"`
	EndDate    *string `form:"end_date"`
	Cursor     string  `form:"cursor"`
	Limit      int     `form:"limit"`
}
