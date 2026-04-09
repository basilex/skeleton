package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/contracts/application/command"
	"github.com/basilex/skeleton/internal/contracts/application/query"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createContract    *command.CreateContractHandler
	activateContract  *command.ActivateContractHandler
	terminateContract *command.TerminateContractHandler
	getContract       *query.GetContractHandler
	listContracts     *query.ListContractsHandler
}

func NewHandler(
	createContract *command.CreateContractHandler,
	activateContract *command.ActivateContractHandler,
	terminateContract *command.TerminateContractHandler,
	getContract *query.GetContractHandler,
	listContracts *query.ListContractsHandler,
) *Handler {
	return &Handler{
		createContract:    createContract,
		activateContract:  activateContract,
		terminateContract: terminateContract,
		getContract:       getContract,
		listContracts:     listContracts,
	}
}

// CreateContract godoc
// @Summary Create a new contract
// @Description Creates a new contract between parties
// @Tags contracts
// @Accept json
// @Produce json
// @Param request body CreateContractRequest true "Contract data"
// @Success 201 {object} map[string]string "Contract created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/contracts [post]
func (h *Handler) CreateContract(c *gin.Context) {
	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createContract.Handle(c.Request.Context(), command.CreateContractCommand{
		ContractType:  req.ContractType,
		PartyID:       req.PartyID,
		PaymentType:   req.PaymentType,
		CreditDays:    req.CreditDays,
		Currency:      req.Currency,
		DeliveryType:  req.DeliveryType,
		EstimatedDays: req.EstimatedDays,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		CreditLimit:   req.CreditLimit,
		CreatedBy:     c.GetString("user_id"),
	})
	if err != nil {
		handleContractsError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.ContractID,
	})
}

// GetContract godoc
// @Summary Get contract by ID
// @Description Retrieves detailed information about a contract
// @Tags contracts
// @Produce json
// @Param id path string true "Contract ID"
// @Success 200 {object} query.ContractDTO "Contract details"
// @Failure 404 {object} apierror.APIError "Contract not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/contracts/{id} [get]
func (h *Handler) GetContract(c *gin.Context) {
	contractID := c.Param("id")

	result, err := h.getContract.Handle(c.Request.Context(), query.GetContractQuery{
		ContractID: contractID,
	})
	if err != nil {
		handleContractsError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListContracts godoc
// @Summary List contracts with filtering
// @Description Retrieves paginated list of contracts with optional filtering
// @Tags contracts
// @Produce json
// @Param party_id query string false "Filter by party ID"
// @Param contract_type query string false "Filter by type (supply, service, employment, partnership, lease, license)"
// @Param status query string false "Filter by status (draft, pending_approval, active, expired, terminated)"
// @Param active_only query bool false "Show only active contracts"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size (default 20)"
// @Success 200 {object} pagination.PageResult[query.ContractDTO] "Contract list"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/contracts [get]
func (h *Handler) ListContracts(c *gin.Context) {
	var req ListContractsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listContracts.Handle(c.Request.Context(), query.ListContractsQuery{
		PartyID:      req.PartyID,
		ContractType: req.ContractType,
		Status:       req.Status,
		ActiveOnly:   req.ActiveOnly,
		Cursor:       req.Cursor,
		Limit:        req.Limit,
	})
	if err != nil {
		handleContractsError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ActivateContract godoc
// @Summary Activate a contract
// @Description Changes contract status from draft/pending_approval to active
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID"
// @Param request body ActivateContractRequest true "Activation data"
// @Success 200 {object} map[string]string "Contract activated"
// @Failure 404 {object} apierror.APIError "Contract not found"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/contracts/{id}/activate [post]
func (h *Handler) ActivateContract(c *gin.Context) {
	contractID := c.Param("id")

	var req ActivateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.activateContract.Handle(c.Request.Context(), command.ActivateContractCommand{
		ContractID: contractID,
		SignedAt:   req.SignedAt,
	})
	if err != nil {
		handleContractsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contract activated"})
}

// TerminateContract godoc
// @Summary Terminate a contract
// @Description Changes contract status to terminated with a reason
// @Tags contracts
// @Accept json
// @Produce json
// @Param id path string true "Contract ID"
// @Param request body TerminateContractRequest true "Termination data"
// @Success 200 {object} map[string]string "Contract terminated"
// @Failure 404 {object} apierror.APIError "Contract not found"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/contracts/{id}/terminate [post]
func (h *Handler) TerminateContract(c *gin.Context) {
	contractID := c.Param("id")

	var req TerminateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.terminateContract.Handle(c.Request.Context(), command.TerminateContractCommand{
		ContractID: contractID,
		Reason:     req.Reason,
	})
	if err != nil {
		handleContractsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contract terminated"})
}

func handleContractsError(c *gin.Context, err error) {
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
