package http

import (
	"net/http"

	inventoryCommand "github.com/basilex/skeleton/internal/inventory/application/command"
	inventoryQuery "github.com/basilex/skeleton/internal/inventory/application/query"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createWarehouse    *inventoryCommand.CreateWarehouseHandler
	updateWarehouse    *inventoryCommand.UpdateWarehouseHandler
	getWarehouse       *inventoryQuery.GetWarehouseHandler
	listWarehouses     *inventoryQuery.ListWarehousesHandler
	createStock        *inventoryCommand.CreateStockHandler
	adjustStock        *inventoryCommand.AdjustStockHandler
	receiptStock       *inventoryCommand.ReceiptStockHandler
	issueStock         *inventoryCommand.IssueStockHandler
	transferStock      *inventoryCommand.TransferStockHandler
	reserveStock       *inventoryCommand.ReserveStockHandler
	fulfillReservation *inventoryCommand.FulfillReservationHandler
	cancelReservation  *inventoryCommand.CancelReservationHandler
	getStock           *inventoryQuery.GetStockHandler
	listStock          *inventoryQuery.ListStockHandler
	getStockMovement   *inventoryQuery.GetStockMovementHandler
	listStockMovements *inventoryQuery.ListStockMovementsHandler
	getReservation     *inventoryQuery.GetReservationHandler
	listReservations   *inventoryQuery.ListReservationsHandler
}

func NewHandler(
	createWarehouse *inventoryCommand.CreateWarehouseHandler,
	updateWarehouse *inventoryCommand.UpdateWarehouseHandler,
	getWarehouse *inventoryQuery.GetWarehouseHandler,
	listWarehouses *inventoryQuery.ListWarehousesHandler,
	createStock *inventoryCommand.CreateStockHandler,
	adjustStock *inventoryCommand.AdjustStockHandler,
	receiptStock *inventoryCommand.ReceiptStockHandler,
	issueStock *inventoryCommand.IssueStockHandler,
	transferStock *inventoryCommand.TransferStockHandler,
	reserveStock *inventoryCommand.ReserveStockHandler,
	fulfillReservation *inventoryCommand.FulfillReservationHandler,
	cancelReservation *inventoryCommand.CancelReservationHandler,
	getStock *inventoryQuery.GetStockHandler,
	listStock *inventoryQuery.ListStockHandler,
	getStockMovement *inventoryQuery.GetStockMovementHandler,
	listStockMovements *inventoryQuery.ListStockMovementsHandler,
	getReservation *inventoryQuery.GetReservationHandler,
	listReservations *inventoryQuery.ListReservationsHandler,
) *Handler {
	return &Handler{
		createWarehouse:    createWarehouse,
		updateWarehouse:    updateWarehouse,
		getWarehouse:       getWarehouse,
		listWarehouses:     listWarehouses,
		createStock:        createStock,
		adjustStock:        adjustStock,
		receiptStock:       receiptStock,
		issueStock:         issueStock,
		transferStock:      transferStock,
		reserveStock:       reserveStock,
		fulfillReservation: fulfillReservation,
		cancelReservation:  cancelReservation,
		getStock:           getStock,
		listStock:          listStock,
		getStockMovement:   getStockMovement,
		listStockMovements: listStockMovements,
		getReservation:     getReservation,
		listReservations:   listReservations,
	}
}

// CreateWarehouse godoc
// @Summary Create a new warehouse
// @Description Creates a new warehouse
// @Tags warehouses
// @Accept json
// @Produce json
// @Param request body CreateWarehouseRequest true "Warehouse data"
// @Success 201 {object} map[string]string "Warehouse created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/warehouses [post]
func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createWarehouse.Handle(c.Request.Context(), inventoryCommand.CreateWarehouseCommand{
		Name:     req.Name,
		Code:     req.Code,
		Location: req.Location,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.WarehouseID})
}

// UpdateWarehouse godoc
// @Summary Update warehouse
// @Description Updates warehouse properties
// @Tags warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Param request body UpdateWarehouseRequest true "Warehouse data"
// @Success 200 {object} map[string]string "Warehouse updated"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Warehouse not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/warehouses/{id} [put]
func (h *Handler) UpdateWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")
	if warehouseID == "" {
		apierror.RespondError(c, apierror.NewValidation("warehouse ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	var req UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.updateWarehouse.Handle(c.Request.Context(), inventoryCommand.UpdateWarehouseCommand{
		WarehouseID: warehouseID,
		Name:        req.Name,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Activate:    req.Activate,
		Deactivate:  req.Deactivate,
		Maintenance: req.Maintenance,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": result.WarehouseID})
}

// GetWarehouse godoc
// @Summary Get warehouse by ID
// @Description Gets warehouse details
// @Tags warehouses
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 200 {object} inventoryQuery.WarehouseDTO "Warehouse details"
// @Failure 404 {object} apierror.APIError "Warehouse not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/warehouses/{id} [get]
func (h *Handler) GetWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")
	if warehouseID == "" {
		apierror.RespondError(c, apierror.NewValidation("warehouse ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	warehouse, err := h.getWarehouse.Handle(c.Request.Context(), inventoryQuery.GetWarehouseQuery{
		WarehouseID: warehouseID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, warehouse)
}

// ListWarehouses godoc
// @Summary List warehouses
// @Description Lists all warehouses with optional filtering
// @Tags warehouses
// @Produce json
// @Param status query string false "Filter by status"
// @Param code query string false "Filter by code"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} inventoryQuery.ListWarehousesResult "Warehouses list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/warehouses [get]
func (h *Handler) ListWarehouses(c *gin.Context) {
	var query ListWarehousesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listWarehouses.Handle(c.Request.Context(), inventoryQuery.ListWarehousesQuery{
		Status: query.Status,
		Code:   query.Code,
		Cursor: query.Cursor,
		Limit:  query.Limit,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateStock godoc
// @Summary Create stock
// @Description Creates stock record for item in warehouse
// @Tags stock
// @Accept json
// @Produce json
// @Param request body CreateStockRequest true "Stock data"
// @Success 201 {object} map[string]string "Stock created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock [post]
func (h *Handler) CreateStock(c *gin.Context) {
	var req CreateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createStock.Handle(c.Request.Context(), inventoryCommand.CreateStockCommand{
		ItemID:      req.ItemID,
		WarehouseID: req.WarehouseID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.StockID})
}

// AdjustStock godoc
// @Summary Adjust stock
// @Description Adjusts stock quantity
// @Tags stock
// @Accept json
// @Produce json
// @Param id path string true "Stock ID"
// @Param request body AdjustStockRequest true "Adjustment data"
// @Success 200 {object} map[string]string "Stock adjusted"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Stock not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/{id}/adjust [post]
func (h *Handler) AdjustStock(c *gin.Context) {
	stockID := c.Param("id")
	if stockID == "" {
		apierror.RespondError(c, apierror.NewValidation("stock ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	var req AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.adjustStock.Handle(c.Request.Context(), inventoryCommand.AdjustStockCommand{
		StockID:     stockID,
		Quantity:    req.Quantity,
		Reason:      req.Reason,
		ReferenceID: req.ReferenceID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stock_id": result.StockID, "movement_id": result.MovementID})
}

// ReceiptStock godoc
// @Summary Receipt stock
// @Description Receive stock into warehouse
// @Tags stock
// @Accept json
// @Produce json
// @Param request body ReceiptStockRequest true "Receipt data"
// @Success 200 {object} map[string]string "Stock received"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/receipt [post]
func (h *Handler) ReceiptStock(c *gin.Context) {
	var req ReceiptStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.receiptStock.Handle(c.Request.Context(), inventoryCommand.ReceiptStockCommand{
		ItemID:      req.ItemID,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
		ReferenceID: req.ReferenceID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stock_id": result.StockID, "movement_id": result.MovementID})
}

// IssueStock godoc
// @Summary Issue stock
// @Description Issue stock from warehouse
// @Tags stock
// @Accept json
// @Produce json
// @Param request body IssueStockRequest true "Issue data"
// @Success 200 {object} map[string]string "Stock issued"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/issue [post]
func (h *Handler) IssueStock(c *gin.Context) {
	var req IssueStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.issueStock.Handle(c.Request.Context(), inventoryCommand.IssueStockCommand{
		ItemID:      req.ItemID,
		WarehouseID: req.WarehouseID,
		Quantity:    req.Quantity,
		OrderID:     req.OrderID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stock_id": result.StockID, "movement_id": result.MovementID})
}

// TransferStock godoc
// @Summary Transfer stock
// @Description Transfer stock between warehouses
// @Tags stock
// @Accept json
// @Produce json
// @Param request body TransferStockRequest true "Transfer data"
// @Success 200 {object} map[string]string "Stock transferred"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/transfer [post]
func (h *Handler) TransferStock(c *gin.Context) {
	var req TransferStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.transferStock.Handle(c.Request.Context(), inventoryCommand.TransferStockCommand{
		ItemID:        req.ItemID,
		FromWarehouse: req.FromWarehouse,
		ToWarehouse:   req.ToWarehouse,
		Quantity:      req.Quantity,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"movement_id": result.MovementID})
}

// ReserveStock godoc
// @Summary Reserve stock
// @Description Reserve stock for order
// @Tags stock
// @Accept json
// @Produce json
// @Param request body ReserveStockRequest true "Reservation data"
// @Success 200 {object} map[string]string "Stock reserved"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/reserve [post]
func (h *Handler) ReserveStock(c *gin.Context) {
	var req ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.reserveStock.Handle(c.Request.Context(), inventoryCommand.ReserveStockCommand{
		StockID:   req.StockID,
		OrderID:   req.OrderID,
		Quantity:  req.Quantity,
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservation_id": result.ReservationID})
}

// FulfillReservation godoc
// @Summary Fulfill reservation
// @Description Fulfill stock reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Param request body FulfillReservationRequest true "Fulfill data"
// @Success 200 {object} map[string]string "Reservation fulfilled"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Reservation not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/reservations/fulfill [post]
func (h *Handler) FulfillReservation(c *gin.Context) {
	var req FulfillReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.fulfillReservation.Handle(c.Request.Context(), inventoryCommand.FulfillReservationCommand{
		ReservationID: req.ReservationID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservation_id": result.ReservationID})
}

// CancelReservation godoc
// @Summary Cancel reservation
// @Description Cancel stock reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Param request body CancelReservationRequest true "Cancel data"
// @Success 200 {object} map[string]string "Reservation cancelled"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Reservation not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/reservations/cancel [post]
func (h *Handler) CancelReservation(c *gin.Context) {
	var req CancelReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.cancelReservation.Handle(c.Request.Context(), inventoryCommand.CancelReservationCommand{
		ReservationID: req.ReservationID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservation_id": result.ReservationID})
}

// GetStock godoc
// @Summary Get stock by ID
// @Description Gets stock details
// @Tags stock
// @Produce json
// @Param id path string true "Stock ID"
// @Success 200 {object} inventoryQuery.StockDTO "Stock details"
// @Failure 404 {object} apierror.APIError "Stock not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock/{id} [get]
func (h *Handler) GetStock(c *gin.Context) {
	stockID := c.Param("id")
	if stockID == "" {
		apierror.RespondError(c, apierror.NewValidation("stock ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	stock, err := h.getStock.Handle(c.Request.Context(), inventoryQuery.GetStockQuery{
		StockID: stockID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, stock)
}

// ListStock godoc
// @Summary List stock
// @Description Lists all stock records with optional filtering
// @Tags stock
// @Produce json
// @Param item_id query string false "Filter by item ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param available query bool false "Filter by availability"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} inventoryQuery.ListStockResult "Stock list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/stock [get]
func (h *Handler) ListStock(c *gin.Context) {
	var query ListStockQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listStock.Handle(c.Request.Context(), inventoryQuery.ListStockQuery{
		ItemID:      query.ItemID,
		WarehouseID: query.WarehouseID,
		Available:   query.Available,
		Cursor:      query.Cursor,
		Limit:       query.Limit,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStockMovement godoc
// @Summary Get stock movement by ID
// @Description Gets stock movement details
// @Tags movements
// @Produce json
// @Param id path string true "Movement ID"
// @Success 200 {object} inventoryQuery.StockMovementDTO "Movement details"
// @Failure 404 {object} apierror.APIError "Movement not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/movements/{id} [get]
func (h *Handler) GetStockMovement(c *gin.Context) {
	movementID := c.Param("id")
	if movementID == "" {
		apierror.RespondError(c, apierror.NewValidation("movement ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	movement, err := h.getStockMovement.Handle(c.Request.Context(), inventoryQuery.GetStockMovementQuery{
		MovementID: movementID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, movement)
}

// ListStockMovements godoc
// @Summary List stock movements
// @Description Lists all stock movements with optional filtering
// @Tags movements
// @Produce json
// @Param item_id query string false "Filter by item ID"
// @Param warehouse_id query string false "Filter by warehouse ID"
// @Param movement_type query string false "Filter by movement type"
// @Param reference_type query string false "Filter by reference type"
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Limit"
// @Success 200 {object} inventoryQuery.ListStockMovementsResult "Movements list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/movements [get]
func (h *Handler) ListStockMovements(c *gin.Context) {
	var query ListStockMovementsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listStockMovements.Handle(c.Request.Context(), inventoryQuery.ListStockMovementsQuery{
		ItemID:        query.ItemID,
		WarehouseID:   query.WarehouseID,
		MovementType:  query.MovementType,
		ReferenceType: query.ReferenceType,
		StartDate:     query.StartDate,
		EndDate:       query.EndDate,
		Cursor:        query.Cursor,
		Limit:         query.Limit,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetReservation godoc
// @Summary Get reservation by ID
// @Description Gets reservation details
// @Tags reservations
// @Produce json
// @Param id path string true "Reservation ID"
// @Success 200 {object} inventoryQuery.StockReservationDTO "Reservation details"
// @Failure 404 {object} apierror.APIError "Reservation not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/reservations/{id} [get]
func (h *Handler) GetReservation(c *gin.Context) {
	reservationID := c.Param("id")
	if reservationID == "" {
		apierror.RespondError(c, apierror.NewValidation("reservation ID is required", c.Request.URL.Path, getRequestID(c)))
		return
	}

	reservation, err := h.getReservation.Handle(c.Request.Context(), inventoryQuery.GetReservationQuery{
		ReservationID: reservationID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// ListReservations godoc
// @Summary List reservations
// @Description Lists all reservations for an order
// @Tags reservations
// @Produce json
// @Param order_id query string true "Order ID"
// @Success 200 {object} inventoryQuery.ListReservationsResult "Reservations list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/reservations [get]
func (h *Handler) ListReservations(c *gin.Context) {
	var query ListReservationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listReservations.Handle(c.Request.Context(), inventoryQuery.ListReservationsQuery{
		OrderID: query.OrderID,
	})
	if err != nil {
		handleInventoryError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleInventoryError(c *gin.Context, err error) {
	// Add error handling logic similar to invoicing
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
