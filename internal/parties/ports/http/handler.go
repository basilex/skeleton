package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/parties/application/command"
	"github.com/basilex/skeleton/internal/parties/application/query"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createCustomer *command.CreateCustomerHandler
	updateCustomer *command.UpdateCustomerHandler
	getCustomer    *query.GetCustomerHandler
	listCustomers  *query.ListCustomersHandler
	createSupplier *command.CreateSupplierHandler
	getSupplier    *query.GetSupplierHandler
	listSuppliers  *query.ListSuppliersHandler
}

func NewHandler(
	createCustomer *command.CreateCustomerHandler,
	updateCustomer *command.UpdateCustomerHandler,
	getCustomer *query.GetCustomerHandler,
	listCustomers *query.ListCustomersHandler,
	createSupplier *command.CreateSupplierHandler,
	getSupplier *query.GetSupplierHandler,
	listSuppliers *query.ListSuppliersHandler,
) *Handler {
	return &Handler{
		createCustomer: createCustomer,
		updateCustomer: updateCustomer,
		getCustomer:    getCustomer,
		listCustomers:  listCustomers,
		createSupplier: createSupplier,
		getSupplier:    getSupplier,
		listSuppliers:  listSuppliers,
	}
}

// CreateCustomer godoc
// @Summary Create a new customer
// @Description Creates a new customer party with contact information
// @Tags customers
// @Accept json
// @Produce json
// @Param request body CreateCustomerRequest true "Customer data"
// @Success 201 {object} map[string]string "Customer created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/customers [post]
func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createCustomer.Handle(c.Request.Context(), command.CreateCustomerCommand{
		Name:        req.Name,
		TaxID:       req.TaxID,
		Email:       req.Email,
		Phone:       req.Phone,
		Street:      req.Address.Street,
		City:        req.Address.City,
		Region:      req.Address.Region,
		PostalCode:  req.Address.PostalCode,
		Country:     req.Address.Country,
		Website:     req.Website,
		SocialMedia: req.SocialMedia,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.CustomerID,
	})
}

// UpdateCustomer godoc
// @Summary Update customer information
// @Description Updates customer contact information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body UpdateCustomerRequest true "Customer data"
// @Success 200 {object} map[string]string "Customer updated"
// @Failure 404 {object} apierror.APIError "Customer not found"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/customers/{id} [put]
func (h *Handler) UpdateCustomer(c *gin.Context) {
	customerID := c.Param("id")

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.updateCustomer.Handle(c.Request.Context(), command.UpdateCustomerCommand{
		CustomerID:  customerID,
		Name:        req.Name,
		TaxID:       req.TaxID,
		Email:       req.Email,
		Phone:       req.Phone,
		Street:      req.Address.Street,
		City:        req.Address.City,
		Region:      req.Address.Region,
		PostalCode:  req.Address.PostalCode,
		Country:     req.Address.Country,
		Website:     req.Website,
		SocialMedia: req.SocialMedia,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer updated"})
}

// GetCustomer godoc
// @Summary Get customer by ID
// @Description Retrieves detailed information about a customer
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} query.CustomerDTO "Customer details"
// @Failure 404 {object} apierror.APIError "Customer not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/customers/{id} [get]
func (h *Handler) GetCustomer(c *gin.Context) {
	customerID := c.Param("id")

	result, err := h.getCustomer.Handle(c.Request.Context(), query.GetCustomerQuery{
		CustomerID: customerID,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListCustomers godoc
// @Summary List customers with filtering
// @Description Retrieves paginated list of customers with optional filtering
// @Tags customers
// @Produce json
// @Param status query string false "Filter by status (active, inactive, blacklisted)"
// @Param search query string false "Search by name or email"
// @Param tax_id query string false "Filter by tax ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20)"
// @Success 200 {object} pagination.PageResult[query.CustomerDTO] "Customer list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/customers [get]
func (h *Handler) ListCustomers(c *gin.Context) {
	var req ListCustomersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listCustomers.Handle(c.Request.Context(), query.ListCustomersQuery{
		Status: req.Status,
		Search: req.Search,
		TaxID:  req.TaxID,
		Cursor: req.Cursor,
		Limit:  req.Limit,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateSupplier godoc
// @Summary Create a new supplier
// @Description Creates a new supplier party with contact information
// @Tags suppliers
// @Accept json
// @Produce json
// @Param request body CreateSupplierRequest true "Supplier data"
// @Success 201 {object} map[string]string "Supplier created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/suppliers [post]
func (h *Handler) CreateSupplier(c *gin.Context) {
	var req CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createSupplier.Handle(c.Request.Context(), command.CreateSupplierCommand{
		Name:        req.Name,
		TaxID:       req.TaxID,
		Email:       req.Email,
		Phone:       req.Phone,
		Street:      req.Address.Street,
		City:        req.Address.City,
		Region:      req.Address.Region,
		PostalCode:  req.Address.PostalCode,
		Country:     req.Address.Country,
		Website:     req.Website,
		SocialMedia: req.SocialMedia,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.SupplierID,
	})
}

// GetSupplier godoc
// @Summary Get supplier by ID
// @Description Retrieves detailed information about a supplier including rating and contracts
// @Tags suppliers
// @Produce json
// @Param id path string true "Supplier ID"
// @Success 200 {object} query.SupplierDTO "Supplier details"
// @Failure 404 {object} apierror.APIError "Supplier not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/suppliers/{id} [get]
func (h *Handler) GetSupplier(c *gin.Context) {
	supplierID := c.Param("id")

	result, err := h.getSupplier.Handle(c.Request.Context(), query.GetSupplierQuery{
		SupplierID: supplierID,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListSuppliers godoc
// @Summary List suppliers with filtering
// @Description Retrieves paginated list of suppliers with optional filtering
// @Tags suppliers
// @Produce json
// @Param status query string false "Filter by status (active, inactive, blacklisted)"
// @Param search query string false "Search by name or email"
// @Param tax_id query string false "Filter by tax ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20)"
// @Success 200 {object} pagination.PageResult[query.SupplierDTO] "Supplier list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/suppliers [get]
func (h *Handler) ListSuppliers(c *gin.Context) {
	var req ListSuppliersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listSuppliers.Handle(c.Request.Context(), query.ListSuppliersQuery{
		Status: req.Status,
		Search: req.Search,
		TaxID:  req.TaxID,
		Cursor: req.Cursor,
		Limit:  req.Limit,
	})
	if err != nil {
		handlePartiesError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func handlePartiesError(c *gin.Context, err error) {
	requestID := getRequestID(c)
	switch err {
	default:
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, requestID))
	}
}

func getRequestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
