package http

import (
	"net/http"

	catalogCommand "github.com/basilex/skeleton/internal/catalog/application/command"
	catalogQuery "github.com/basilex/skeleton/internal/catalog/application/query"
	"github.com/basilex/skeleton/internal/catalog/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createItem *catalogCommand.CreateItemHandler
	updateItem *catalogCommand.UpdateItemHandler
	getItem    *catalogQuery.GetItemHandler
	listItems  *catalogQuery.ListItemsHandler
}

func NewHandler(
	createItem *catalogCommand.CreateItemHandler,
	updateItem *catalogCommand.UpdateItemHandler,
	getItem *catalogQuery.GetItemHandler,
	listItems *catalogQuery.ListItemsHandler,
) *Handler {
	return &Handler{
		createItem: createItem,
		updateItem: updateItem,
		getItem:    getItem,
		listItems:  listItems,
	}
}

// CreateItem godoc
// @Summary Create a new catalog item
// @Description Creates a new inventory item in the catalog
// @Tags catalog
// @Accept json
// @Produce json
// @Param request body CreateItemRequest true "Item data"
// @Success 201 {object} map[string]string "Item created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/catalog/items [post]
func (h *Handler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createItem.Handle(c.Request.Context(), catalogCommand.CreateItemCommand{
		Name:         req.Name,
		Description:  req.Description,
		SKU:          req.SKU,
		BasePrice:    req.BasePrice,
		Currency:     req.Currency,
		CategoryID:   req.CategoryID,
		Attributes:   req.Attributes,
	})
	if err != nil {
		handleCatalogError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.ItemID})
}

// GetItem godoc
// @Summary Get catalog item by ID
// @Description Retrieves a catalog item by its ID
// @Tags catalog
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} catalogQuery.ItemDTO "Item details"
// @Failure 404 {object} apierror.APIError "Item not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/catalog/items/{id} [get]
func (h *Handler) GetItem(c *gin.Context) {
	itemID := c.Param("id")

	result, err := h.getItem.Handle(c.Request.Context(), catalogQuery.GetItemQuery{
		ItemID: itemID,
	})
	if err != nil {
		handleCatalogError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListItems godoc
// @Summary List catalog items
// @Description Lists catalog items with filtering and pagination
// @Tags catalog
// @Produce json
// @Param category_id query string false "Filter by category ID"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by name"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} pagination.PageResult[catalogQuery.ItemDTO] "List of items"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/catalog/items [get]
func (h *Handler) ListItems(c *gin.Context) {
	var req ListItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listItems.Handle(c.Request.Context(), catalogQuery.ListItemsQuery{
		CategoryID: req.CategoryID,
		Status:     req.Status,
		Search:     req.Search,
		Cursor:     req.Cursor,
		Limit:      req.Limit,
	})
	if err != nil {
		handleCatalogError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateItem godoc
// @Summary Update catalog item
// @Description Updates a catalog item's properties
// @Tags catalog
// @Accept json
// @Param id path string true "Item ID"
// @Param request body UpdateItemRequest true "Update data"
// @Success 200 {object} map[string]string "Item updated"
// @Failure 404 {object} apierror.APIError "Item not found"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/catalog/items/{id} [put]
func (h *Handler) UpdateItem(c *gin.Context) {
	itemID := c.Param("id")

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.updateItem.Handle(c.Request.Context(), catalogCommand.UpdateItemCommand{
		ItemID:       itemID,
		Name:         req.Name,
		Description:  req.Description,
		BasePrice:    req.BasePrice,
	})
	if err != nil {
		handleCatalogError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item updated"})
}

func handleCatalogError(c *gin.Context, err error) {
	switch err {
	case catalog.ErrItemNotFound:
		apierror.RespondError(c, apierror.NewNotFound("item not found", c.Request.URL.Path, getRequestID(c)))
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
