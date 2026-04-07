package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	registerUser *command.RegisterUserHandler
	loginUser    *command.LoginUserHandler
	assignRole   *command.AssignRoleHandler
	revokeRole   *command.RevokeRoleHandler
	getUser      *query.GetUserHandler
	listUsers    *query.ListUsersHandler
}

func NewHandler(
	registerUser *command.RegisterUserHandler,
	loginUser *command.LoginUserHandler,
	assignRole *command.AssignRoleHandler,
	revokeRole *command.RevokeRoleHandler,
	getUser *query.GetUserHandler,
	listUsers *query.ListUsersHandler,
) *Handler {
	return &Handler{
		registerUser: registerUser,
		loginUser:    loginUser,
		assignRole:   assignRole,
		revokeRole:   revokeRole,
		getUser:      getUser,
		listUsers:    listUsers,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.registerUser.Handle(c.Request.Context(), command.RegisterUserCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": result.UserID})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.loginUser.Handle(c.Request.Context(), command.LoginUserCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refresh not yet implemented"})
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	result, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: userID})
	if err != nil {
		handleIdentityError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) ListUsers(c *gin.Context) {
	var req UserFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listUsers.Handle(c.Request.Context(), query.ListUsersQuery{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Search,
		IsActive: req.IsActive,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetMyProfile(c *gin.Context) {
	userID, ok := c.Get(string(ContextKeyUserID))
	if !ok {
		apierror.RespondError(c, apierror.NewUnauthorized("missing user context", c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: userID.(string)})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) AssignRole(c *gin.Context) {
	userID := c.Param("id")
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.assignRole.Handle(c.Request.Context(), command.AssignRoleCommand{
		UserID: userID,
		RoleID: req.RoleID,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
}

func (h *Handler) RevokeRole(c *gin.Context) {
	userID := c.Param("id")
	roleID := c.Param("rid")

	err := h.revokeRole.Handle(c.Request.Context(), command.RevokeRoleCommand{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role revoked"})
}

func (h *Handler) DeactivateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "deactivate not yet implemented"})
}

func handleIdentityError(c *gin.Context, err error) {
	requestID := getRequestID(c)
	switch err {
	case nil:
		return
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
