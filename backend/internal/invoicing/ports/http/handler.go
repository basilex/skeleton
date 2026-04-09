package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	invoicingCommand "github.com/basilex/skeleton/internal/invoicing/application/command"
	invoicingQuery "github.com/basilex/skeleton/internal/invoicing/application/query"
	invoicingDomain "github.com/basilex/skeleton/internal/invoicing/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createInvoice  *invoicingCommand.CreateInvoiceHandler
	addInvoiceLine *invoicingCommand.AddInvoiceLineHandler
	sendInvoice    *invoicingCommand.SendInvoiceHandler
	recordPayment  *invoicingCommand.RecordPaymentHandler
	cancelInvoice  *invoicingCommand.CancelInvoiceHandler
	getInvoice     *invoicingQuery.GetInvoiceHandler
	listInvoices   *invoicingQuery.ListInvoicesHandler
}

func NewHandler(
	createInvoice *invoicingCommand.CreateInvoiceHandler,
	addInvoiceLine *invoicingCommand.AddInvoiceLineHandler,
	sendInvoice *invoicingCommand.SendInvoiceHandler,
	recordPayment *invoicingCommand.RecordPaymentHandler,
	cancelInvoice *invoicingCommand.CancelInvoiceHandler,
	getInvoice *invoicingQuery.GetInvoiceHandler,
	listInvoices *invoicingQuery.ListInvoicesHandler,
) *Handler {
	return &Handler{
		createInvoice:  createInvoice,
		addInvoiceLine: addInvoiceLine,
		sendInvoice:    sendInvoice,
		recordPayment:  recordPayment,
		cancelInvoice:  cancelInvoice,
		getInvoice:     getInvoice,
		listInvoices:   listInvoices,
	}
}

// CreateInvoice godoc
// @Summary Create a new invoice
// @Description Creates a new invoice in draft status
// @Tags invoices
// @Accept json
// @Produce json
// @Param request body CreateInvoiceRequest true "Invoice data"
// @Success 201 {object} map[string]string "Invoice created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices [post]
func (h *Handler) CreateInvoice(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid due_date format", c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createInvoice.Handle(c.Request.Context(), invoicingCommand.CreateInvoiceCommand{
		InvoiceNumber: req.InvoiceNumber,
		OrderID:       req.OrderID,
		ContractID:    req.ContractID,
		CustomerID:    req.CustomerID,
		SupplierID:    req.SupplierID,
		Currency:      req.Currency,
		DueDate:       dueDate,
		Notes:         req.Notes,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InvoiceID})
}

// AddInvoiceLine godoc
// @Summary Add line to invoice
// @Description Adds a new line item to an existing invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body AddInvoiceLineRequest true "Line data"
// @Success 200 {object} map[string]string "Line added"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Invoice not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices/{id}/lines [post]
func (h *Handler) AddInvoiceLine(c *gin.Context) {
	invoiceID := c.Param("id")

	var req AddInvoiceLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.addInvoiceLine.Handle(c.Request.Context(), invoicingCommand.AddInvoiceLineCommand{
		InvoiceID:   invoiceID,
		Description: req.Description,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
		Unit:        req.Unit,
		Discount:    req.Discount,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "line added"})
}

// SendInvoice godoc
// @Summary Send invoice
// @Description Sends an invoice (changes status from draft to sent)
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} map[string]string "Invoice sent"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Invoice not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices/{id}/send [post]
func (h *Handler) SendInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	err := h.sendInvoice.Handle(c.Request.Context(), invoicingCommand.SendInvoiceCommand{
		InvoiceID: invoiceID,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice sent"})
}

// RecordPayment godoc
// @Summary Record payment for invoice
// @Description Records a payment for an invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body RecordPaymentRequest true "Payment data"
// @Success 201 {object} map[string]string "Payment recorded"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Invoice not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices/{id}/payments [post]
func (h *Handler) RecordPayment(c *gin.Context) {
	invoiceID := c.Param("id")

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.recordPayment.Handle(c.Request.Context(), invoicingCommand.RecordPaymentCommand{
		InvoiceID: invoiceID,
		Amount:    req.Amount,
		Method:    req.Method,
		Reference: req.Reference,
		Notes:     req.Notes,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payment_id": result.PaymentID})
}

// CancelInvoice godoc
// @Summary Cancel invoice
// @Description Cancels an invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param request body CancelInvoiceRequest true "Cancellation data"
// @Success 200 {object} map[string]string "Invoice cancelled"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Invoice not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices/{id}/cancel [post]
func (h *Handler) CancelInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	var req CancelInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.cancelInvoice.Handle(c.Request.Context(), invoicingCommand.CancelInvoiceCommand{
		InvoiceID: invoiceID,
		Reason:    req.Reason,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice cancelled"})
}

// GetInvoice godoc
// @Summary Get invoice by ID
// @Description Retrieves an invoice with its lines and payments by ID
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} InvoiceResponse "Invoice details"
// @Failure 404 {object} apierror.APIError "Invoice not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices/{id} [get]
func (h *Handler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")

	result, err := h.getInvoice.Handle(c.Request.Context(), invoicingQuery.GetInvoiceQuery{
		InvoiceID: invoiceID,
	})
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListInvoices godoc
// @Summary List invoices
// @Description Lists invoices with filtering and pagination
// @Tags invoices
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param status query string false "Filter by status"
// @Param overdue query bool false "Filter overdue invoices"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{} "List of invoices"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) {
	var query invoicingQuery.ListInvoicesQuery

	if customerID := c.Query("customer_id"); customerID != "" {
		query.CustomerID = &customerID
	}
	if status := c.Query("status"); status != "" {
		query.Status = &status
	}
	if overdue := c.Query("overdue"); overdue == "true" {
		t := true
		query.Overdue = &t
	}
	if cursor := c.Query("cursor"); cursor != "" {
		query.Cursor = cursor
	}
	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := fmt.Sscanf(limit, "%d", &l); err == nil && l > 0 {
			query.Limit = l
		}
	}

	result, err := h.listInvoices.Handle(c.Request.Context(), query)
	if err != nil {
		handleInvoicingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       result.Items,
		"next_cursor": result.NextCursor,
		"has_more":    result.HasMore,
	})
}

func handleInvoicingError(c *gin.Context, err error) {
	if errors.Is(err, invoicingDomain.ErrInvoiceNotFound) {
		apierror.RespondError(c, apierror.NewNotFound("invoice not found", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, invoicingDomain.ErrInvoiceAlreadySent) {
		apierror.RespondError(c, apierror.NewConflict("invoice already sent", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, invoicingDomain.ErrInvoiceAlreadyPaid) {
		apierror.RespondError(c, apierror.NewConflict("invoice already paid", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, invoicingDomain.ErrInvoiceAlreadyCancelled) {
		apierror.RespondError(c, apierror.NewConflict("invoice already cancelled", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, invoicingDomain.ErrPaymentExceedsAmount) {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
