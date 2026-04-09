package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/accounting/application/command"
	"github.com/basilex/skeleton/internal/accounting/application/query"
	"github.com/basilex/skeleton/internal/accounting/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createAccount     *command.CreateAccountHandler
	recordTransaction *command.RecordTransactionHandler
	getAccount        *query.GetAccountHandler
	listAccounts      *query.ListAccountsHandler
}

func NewHandler(
	createAccount *command.CreateAccountHandler,
	recordTransaction *command.RecordTransactionHandler,
	getAccount *query.GetAccountHandler,
	listAccounts *query.ListAccountsHandler,
) *Handler {
	return &Handler{
		createAccount:     createAccount,
		recordTransaction: recordTransaction,
		getAccount:        getAccount,
		listAccounts:      listAccounts,
	}
}

// CreateAccount godoc
// @Summary Create a new account
// @Description Creates a new account in the chart of accounts
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body CreateAccountRequest true "Account data"
// @Success 201 {object} map[string]string "Account created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/accounts [post]
func (h *Handler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createAccount.Handle(c.Request.Context(), command.CreateAccountCommand{
		Code:        req.Code,
		Name:        req.Name,
		AccountType: req.AccountType,
		Currency:    req.Currency,
		ParentID:    req.ParentID,
	})
	if err != nil {
		handleAccountingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.AccountID,
	})
}

// GetAccount godoc
// @Summary Get account by ID
// @Description Retrieves an account by its ID
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} query.AccountDTO "Account details"
// @Failure 404 {object} apierror.APIError "Account not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/accounts/{id} [get]
func (h *Handler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")

	result, err := h.getAccount.Handle(c.Request.Context(), query.GetAccountQuery{
		AccountID: accountID,
	})
	if err != nil {
		handleAccountingError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListAccounts godoc
// @Summary List accounts
// @Description Lists accounts with optional filtering
// @Tags accounts
// @Produce json
// @Param account_type query string false "Filter by account type (asset, liability, equity, revenue, expense)"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search by account name"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} pagination.PageResult[query.AccountDTO] "List of accounts"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/accounts [get]
func (h *Handler) ListAccounts(c *gin.Context) {
	var req ListAccountsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listAccounts.Handle(c.Request.Context(), query.ListAccountsQuery{
		AccountType: req.AccountType,
		IsActive:    req.IsActive,
		Search:      req.Search,
		Cursor:      req.Cursor,
		Limit:       req.Limit,
	})
	if err != nil {
		handleAccountingError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// RecordTransaction godoc
// @Summary Record a transaction
// @Description Records a double-entry accounting transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body RecordTransactionRequest true "Transaction data"
// @Success 201 {object} map[string]string "Transaction recorded"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/transactions [post]
func (h *Handler) RecordTransaction(c *gin.Context) {
	var req RecordTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.recordTransaction.Handle(c.Request.Context(), command.RecordTransactionCommand{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Reference:     req.Reference,
		Description:   req.Description,
		PostedBy:      getUserID(c),
	})
	if err != nil {
		handleAccountingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.TransactionID,
	})
}

func handleAccountingError(c *gin.Context, err error) {
	switch err {
	case domain.ErrAccountNotFound:
		apierror.RespondError(c, apierror.NewNotFound("account not found", c.Request.URL.Path, getRequestID(c)))
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
