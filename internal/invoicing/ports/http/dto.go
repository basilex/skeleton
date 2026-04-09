package http

type CreateInvoiceRequest struct {
	InvoiceNumber string  `json:"invoice_number" binding:"required"`
	OrderID       *string `json:"order_id"`
	ContractID    *string `json:"contract_id"`
	CustomerID    string  `json:"customer_id" binding:"required"`
	SupplierID    *string `json:"supplier_id"`
	Currency      string  `json:"currency" binding:"required"`
	DueDate       string  `json:"due_date" binding:"required"`
	Notes         *string `json:"notes"`
}

type AddInvoiceLineRequest struct {
	Description string  `json:"description" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" binding:"required,gte=0"`
	Unit        string  `json:"unit" binding:"required"`
	Discount    float64 `json:"discount" binding:"gte=0"`
}

type SendInvoiceRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"`
}

type RecordPaymentRequest struct {
	InvoiceID string  `json:"invoice_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Method    string  `json:"method" binding:"required"`
	Reference string  `json:"reference"`
	Notes     string  `json:"notes"`
}

type CancelInvoiceRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

type InvoiceResponse struct {
	ID            string                `json:"id"`
	InvoiceNumber string                `json:"invoice_number"`
	OrderID       *string               `json:"order_id"`
	ContractID    *string               `json:"contract_id"`
	CustomerID    string                `json:"customer_id"`
	SupplierID    *string               `json:"supplier_id"`
	IssueDate     string                `json:"issue_date"`
	DueDate       string                `json:"due_date"`
	Status        string                `json:"status"`
	Lines         []InvoiceLineResponse `json:"lines"`
	Subtotal      float64               `json:"subtotal"`
	TaxAmount     float64               `json:"tax_amount"`
	Discount      float64               `json:"discount"`
	Total         float64               `json:"total"`
	Currency      string                `json:"currency"`
	Notes         *string               `json:"notes"`
	PaidAmount    float64               `json:"paid_amount"`
	Payments      []PaymentResponse     `json:"payments"`
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
}

type InvoiceLineResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Unit        string  `json:"unit"`
	Discount    float64 `json:"discount"`
	Total       float64 `json:"total"`
}

type PaymentResponse struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method"`
	Reference string  `json:"reference"`
	PaidAt    string  `json:"paid_at"`
	Notes     string  `json:"notes"`
}
