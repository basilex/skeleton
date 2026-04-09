package http

type CreateContractRequest struct {
	ContractType  string  `json:"contract_type" binding:"required"`
	PartyID       string  `json:"party_id" binding:"required"`
	PaymentType   string  `json:"payment_type" binding:"required"`
	CreditDays    int     `json:"credit_days"`
	Currency      string  `json:"currency" binding:"required"`
	DeliveryType  string  `json:"delivery_type" binding:"required"`
	EstimatedDays int     `json:"estimated_days"`
	StartDate     string  `json:"start_date" binding:"required"`
	EndDate       string  `json:"end_date" binding:"required"`
	CreditLimit   float64 `json:"credit_limit"`
}

type ListContractsRequest struct {
	PartyID      *string `form:"party_id"`
	ContractType *string `form:"contract_type"`
	Status       *string `form:"status"`
	ActiveOnly   bool    `form:"active_only"`
	Cursor       string  `form:"cursor"`
	Limit        int     `form:"limit"`
}

type ActivateContractRequest struct {
	SignedAt string `json:"signed_at" binding:"required"`
}

type TerminateContractRequest struct {
	Reason string `json:"reason" binding:"required"`
}
