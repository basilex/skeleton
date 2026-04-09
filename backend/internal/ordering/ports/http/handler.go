package http

import (
	"net/http"

	orderingCommand "github.com/basilex/skeleton/internal/ordering/application/command"
	orderingQuery "github.com/basilex/skeleton/internal/ordering/application/query"
	"github.com/basilex/skeleton/internal/ordering/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createOrder       *orderingCommand.CreateOrderHandler
	addOrderLine      *orderingCommand.AddOrderLineHandler
	updateOrderStatus *orderingCommand.UpdateOrderStatusHandler
	getOrder          *orderingQuery.GetOrderHandler
	listOrders        *orderingQuery.ListOrdersHandler
}

func NewHandler(
	createOrder *orderingCommand.CreateOrderHandler,
	addOrderLine *orderingCommand.AddOrderLineHandler,
	updateOrderStatus *orderingCommand.UpdateOrderStatusHandler,
	getOrder *orderingQuery.GetOrderHandler,
	listOrders *orderingQuery.ListOrdersHandler,
) *Handler {
	return &Handler{
		createOrder:       createOrder,
		addOrderLine:      addOrderLine,
		updateOrderStatus: updateOrderStatus,
		getOrder:          getOrder,
		listOrders:        listOrders,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Creates a new order in draft status
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order data"
// @Success 201 {object} map[string]string "Order created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createOrder.Handle(c.Request.Context(), orderingCommand.CreateOrderCommand{
		OrderNumber: req.OrderNumber,
		CustomerID:  req.CustomerID,
		SupplierID:  req.SupplierID,
		ContractID:  req.ContractID,
		Currency:    req.Currency,
		CreatedBy:   getUserID(c),
	})
	if err != nil {
		handleOrderingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.OrderID,
	})
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieves an order with its lines by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} orderingQuery.OrderDTO "Order details"
// @Failure 404 {object} apierror.APIError "Order not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	result, err := h.getOrder.Handle(c.Request.Context(), orderingQuery.GetOrderQuery{
		OrderID: orderID,
	})
	if err != nil {
		handleOrderingError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListOrders godoc
// @Summary List orders
// @Description Lists orders with filtering and pagination
// @Tags orders
// @Produce json
// @Param customer_id query string false "Filter by customer ID"
// @Param supplier_id query string false "Filter by supplier ID"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Filter by start date"
// @Param end_date query string false "Filter by end date"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} pagination.PageResult[orderingQuery.OrderDTO] "List of orders"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	var req ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listOrders.Handle(c.Request.Context(), orderingQuery.ListOrdersQuery{
		CustomerID: req.CustomerID,
		SupplierID: req.SupplierID,
		Status:     req.Status,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Cursor:     req.Cursor,
		Limit:      req.Limit,
	})
	if err != nil {
		handleOrderingError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// AddOrderLine godoc
// @Summary Add line to order
// @Description Adds a line item to an existing draft order
// @Tags orders
// @Accept json
// @Param id path string true "Order ID"
// @Param request body AddOrderLineRequest true "Line data"
// @Success 200 {object} map[string]string "Line added"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Order not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/orders/{id}/lines [post]
func (h *Handler) AddOrderLine(c *gin.Context) {
	orderID := c.Param("id")

	var req AddOrderLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.addOrderLine.Handle(c.Request.Context(), orderingCommand.AddOrderLineCommand{
		OrderID:   orderID,
		ItemID:    req.ItemID,
		ItemName:  req.ItemName,
		Quantity:  req.Quantity,
		Unit:      req.Unit,
		UnitPrice: req.UnitPrice,
		Discount:  req.Discount,
	})
	if err != nil {
		handleOrderingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "line added"})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Updates order status (confirm, complete, cancel)
// @Tags orders
// @Accept json
// @Param id path string true "Order ID"
// @Param request body UpdateOrderStatusRequest true "Status data"
// @Success 200 {object} map[string]string "Status updated"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Order not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/orders/{id}/status [put]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.updateOrderStatus.Handle(c.Request.Context(), orderingCommand.UpdateOrderStatusCommand{
		OrderID: orderID,
		Status:  req.Status,
		Reason:  req.Reason,
	})
	if err != nil {
		handleOrderingError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func handleOrderingError(c *gin.Context, err error) {
	switch err {
	case domain.ErrOrderNotFound:
		apierror.RespondError(c, apierror.NewNotFound("order not found", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrOrderLineNotFound:
		apierror.RespondError(c, apierror.NewNotFound("order line not found", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrOrderCannotComplete:
		apierror.RespondError(c, apierror.NewValidation("order cannot be completed", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrOrderCannotCancel:
		apierror.RespondError(c, apierror.NewValidation("order cannot be cancelled", c.Request.URL.Path, getRequestID(c)))
	case domain.ErrInvalidQuantity:
		apierror.RespondError(c, apierror.NewValidation("invalid quantity", c.Request.URL.Path, getRequestID(c)))
	default:
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
	}
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

func getUserID(c *gin.Context) string {
	if id, exists := c.Get("user_id"); exists {
		if userID, ok := id.(string); ok {
			return userID
		}
	}
	return ""
}
