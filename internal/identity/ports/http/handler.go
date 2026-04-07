package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
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
	sessionStore session.Store
}

func NewHandler(
	registerUser *command.RegisterUserHandler,
	loginUser *command.LoginUserHandler,
	assignRole *command.AssignRoleHandler,
	revokeRole *command.RevokeRoleHandler,
	getUser *query.GetUserHandler,
	listUsers *query.ListUsersHandler,
	sessionStore session.Store,
) *Handler {
	return &Handler{
		registerUser: registerUser,
		loginUser:    loginUser,
		assignRole:   assignRole,
		revokeRole:   revokeRole,
		getUser:      getUser,
		listUsers:    listUsers,
		sessionStore: sessionStore,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} map[string]string "User created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/auth/register [post]
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

// Login godoc
// @Summary Login user
// @Description Authenticates user and returns access + refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} TokenResponse "Tokens"
// @Failure 400 {object} apierror.APIError "Bad request"
// @Failure 401 {object} apierror.APIError "Invalid credentials"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/auth/login [post]
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

// Refresh godoc
// @Summary Refresh access token
// @Description Issues new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh request"
// @Success 200 {object} TokenResponse "New tokens"
// @Failure 400 {object} apierror.APIError "Bad request"
// @Failure 401 {object} apierror.APIError "Invalid refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refresh not yet implemented"})
}

// Logout godoc
// @Summary Logout user
// @Description Destroys current user session
// @Tags auth
// @Produce json
// @Security SessionAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	sid, ok := c.Get("session_id")
	if ok {
		_ = h.sessionStore.Delete(c.Request.Context(), sid.(string))
	}
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// GetUser godoc
// @Summary Get user by ID
// @Description Returns user details by ID
// @Tags users
// @Produce json
// @Security SessionAuth
// @Param id path string true "User ID"
// @Success 200 {object} query.UserDTO "User details"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 404 {object} apierror.APIError "User not found"
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	result, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: userID})
	if err != nil {
		handleIdentityError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListUsers godoc
// @Summary List users
// @Description Returns paginated list of users
// @Tags users
// @Produce json
// @Security SessionAuth
// @Param cursor query string false "Pagination cursor (UUID v7)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Param search query string false "Search by email"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} pagination.PageResult "Paginated users"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Router /api/v1/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	var req UserFilterRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.listUsers.Handle(c.Request.Context(), query.ListUsersQuery{
		Cursor:   req.Cursor,
		Limit:    req.Limit,
		Search:   req.Search,
		IsActive: req.IsActive,
	})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyProfile godoc
// @Summary Get current user profile
// @Description Returns authenticated user's profile
// @Tags users
// @Produce json
// @Security SessionAuth
// @Success 200 {object} query.UserDTO "User profile"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Router /api/v1/users/me [get]
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

// AssignRole godoc
// @Summary Assign role to user
// @Description Assigns a role to the specified user
// @Tags roles
// @Accept json
// @Produce json
// @Security SessionAuth
// @Param id path string true "User ID"
// @Param request body AssignRoleRequest true "Role assignment"
// @Success 200 {object} map[string]string "Role assigned"
// @Failure 400 {object} apierror.APIError "Bad request"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 404 {object} apierror.APIError "User or role not found"
// @Router /api/v1/users/{id}/roles [post]
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

// RevokeRole godoc
// @Summary Revoke role from user
// @Description Removes a role from the specified user
// @Tags roles
// @Produce json
// @Security SessionAuth
// @Param id path string true "User ID"
// @Param rid path string true "Role ID"
// @Success 200 {object} map[string]string "Role revoked"
// @Failure 401 {object} apierror.APIError "Unauthorized"
// @Failure 403 {object} apierror.APIError "Forbidden"
// @Failure 404 {object} apierror.APIError "User not found"
// @Router /api/v1/users/{id}/roles/{rid} [delete]
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
