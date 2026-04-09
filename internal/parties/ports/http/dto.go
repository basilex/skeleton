package http

type AddressRequest struct {
	Street     string `json:"street" binding:"required"`
	City       string `json:"city" binding:"required"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country" binding:"required"`
}

type CreateCustomerRequest struct {
	Name        string            `json:"name" binding:"required"`
	TaxID       string            `json:"tax_id"`
	Email       string            `json:"email" binding:"required,email"`
	Phone       string            `json:"phone"`
	Address     AddressRequest    `json:"address" binding:"required"`
	Website     string            `json:"website"`
	SocialMedia map[string]string `json:"social_media"`
}

type UpdateCustomerRequest struct {
	Name        string            `json:"name"`
	TaxID       string            `json:"tax_id"`
	Email       string            `json:"email" binding:"omitempty,email"`
	Phone       string            `json:"phone"`
	Address     AddressRequest    `json:"address"`
	Website     string            `json:"website"`
	SocialMedia map[string]string `json:"social_media"`
}

type ListCustomersRequest struct {
	Status *string `form:"status"`
	Search string  `form:"search"`
	TaxID  string  `form:"tax_id"`
	Cursor string  `form:"cursor"`
	Limit  int     `form:"limit"`
}

type CreateSupplierRequest struct {
	Name        string            `json:"name" binding:"required"`
	TaxID       string            `json:"tax_id"`
	Email       string            `json:"email" binding:"required,email"`
	Phone       string            `json:"phone"`
	Address     AddressRequest    `json:"address" binding:"required"`
	Website     string            `json:"website"`
	SocialMedia map[string]string `json:"social_media"`
}

type ListSuppliersRequest struct {
	Status *string `form:"status"`
	Search string  `form:"search"`
	TaxID  string  `form:"tax_id"`
	Cursor string  `form:"cursor"`
	Limit  int     `form:"limit"`
}
