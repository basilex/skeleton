package http

type CreateAccountRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	AccountType string  `json:"account_type" binding:"required,oneof=asset liability equity revenue expense"`
	Currency    string  `json:"currency" binding:"required,oneof=UAH USD EUR GBP"`
	ParentID    *string `json:"parent_id"`
}

type RecordTransactionRequest struct {
	FromAccountID string  `json:"from_account_id" binding:"required"`
	ToAccountID   string  `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,oneof=UAH USD EUR GBP"`
	Reference     string  `json:"reference"`
	Description   string  `json:"description"`
}

type ListAccountsRequest struct {
	AccountType *string `form:"account_type" binding:"omitempty,oneof=asset liability equity revenue expense"`
	IsActive    *bool   `form:"is_active"`
	Search      string  `form:"search"`
	Cursor      string  `form:"cursor"`
	Limit       int     `form:"limit"`
}
