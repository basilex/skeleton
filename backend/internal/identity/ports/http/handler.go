package http

import (
	"errors"
	"net/http"

	"github.com/basilex/skeleton/internal/identity/application/command"
	"github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	registerUser *command.RegisterUserHandler
	loginUser    *command.LoginUserHandler
	logoutUser   *command.LogoutUserHandler
	assignRole   *command.AssignRoleHandler
	revokeRole   *command.RevokeRoleHandler
	getUser      *query.GetUserHandler
	listUsers    *query.ListUsersHandler
	sessionStore session.Store
	tokenService domain.TokenService
	roleRepo     domain.RoleRepository
}

func NewHandler(
	registerUser *command.RegisterUserHandler,
	loginUser *command.LoginUserHandler,
	logoutUser *command.LogoutUserHandler,
	assignRole *command.AssignRoleHandler,
	revokeRole *command.RevokeRoleHandler,
	getUser *query.GetUserHandler,
	listUsers *query.ListUsersHandler,
	sessionStore session.Store,
	tokenService domain.TokenService,
	roleRepo domain.RoleRepository,
) *Handler {
	return &Handler{
		registerUser: registerUser,
		loginUser:    loginUser,
		logoutUser:   logoutUser,
		assignRole:   assignRole,
		revokeRole:   revokeRole,
		getUser:      getUser,
		listUsers:    listUsers,
		sessionStore: sessionStore,
		tokenService: tokenService,
		roleRepo:     roleRepo,
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

	userDTO, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: result.UserID})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		UserID:       result.UserID,
		Email:        userDTO.Email,
		Roles:        userDTO.Roles,
		IsActive:     userDTO.IsActive,
		AccessToken:  "",
		RefreshToken: "",
	})
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

	userID, _ := domain.ParseUserID(result.UserID)
	sess, err := h.sessionStore.Create(
		c.Request.Context(),
		userID,
		result.Roles,
		result.Permissions,
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.SetCookie(h.sessionCookieName(), sess.ID, h.sessionTTL(), "/", "", false, true)

	c.JSON(http.StatusOK, AuthResponse{
		UserID:       result.UserID,
		Email:        result.Email,
		Roles:        result.Roles,
		IsActive:     result.IsActive,
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

	userID, err := h.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		apierror.RespondError(c, apierror.NewUnauthorized("invalid or expired refresh token", c.Request.URL.Path, getRequestID(c)))
		return
	}

	userDTO, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: userID.String()})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	if !userDTO.IsActive {
		apierror.RespondError(c, apierror.NewUnauthorized("user is inactive", c.Request.URL.Path, getRequestID(c)))
		return
	}

	userIDTyped, _ := domain.ParseUserID(userDTO.ID)
	roleIDs := make([]domain.RoleID, 0, len(userDTO.Roles))

	roles, err := h.loadRoleEntities(c, roleIDs)
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	accessToken, err := h.tokenService.GenerateAccessToken(userIDTyped, roles)
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(userIDTyped)
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	sess, err := h.sessionStore.Create(
		c.Request.Context(),
		userIDTyped,
		userDTO.Roles,
		aggregatePermissions(roles),
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	c.SetCookie(h.sessionCookieName(), sess.ID, h.sessionTTL(), "/", "", false, true)

	c.JSON(http.StatusOK, AuthResponse{
		UserID:       userDTO.ID,
		Email:        userDTO.Email,
		Roles:        userDTO.Roles,
		IsActive:     userDTO.IsActive,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if sid, ok := c.Get("session_id"); ok {
		_ = h.sessionStore.Delete(c.Request.Context(), sid.(string))
		c.SetCookie(h.sessionCookieName(), "", -1, "/", "", false, true)
	}

	_ = h.logoutUser.Handle(c.Request.Context(), command.LogoutUserCommand{
		UserID: userID.(string),
	})

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

func (h *Handler) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondError(c, apierror.NewUnauthorized("user not found in context", c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: userID.(string)})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondError(c, apierror.NewUnauthorized("user not authenticated", c.Request.URL.Path, getRequestID(c)))
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
	userID := c.Param("id")

	parsedID, err := domain.ParseUserID(userID)
	if err != nil {
		apierror.RespondError(c, apierror.NewValidation("invalid user id", c.Request.URL.Path, getRequestID(c)))
		return
	}

	user, err := h.getUser.Handle(c.Request.Context(), query.GetUserQuery{UserID: parsedID.String()})
	if err != nil {
		handleIdentityError(c, err)
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusOK, gin.H{"message": "user already inactive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deactivated", "user_id": userID})
}

func handleIdentityError(c *gin.Context, err error) {
	requestID := getRequestID(c)
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		apierror.RespondError(c, apierror.NewNotFound(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrUserAlreadyExists):
		apierror.RespondError(c, apierror.NewConflict(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrInvalidPassword):
		apierror.RespondError(c, apierror.NewUnauthorized(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrUserInactive):
		apierror.RespondError(c, apierror.NewUnauthorized(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrRoleNotFound):
		apierror.RespondError(c, apierror.NewNotFound(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrRoleAlreadyAssigned):
		apierror.RespondError(c, apierror.NewConflict(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrRoleNotAssigned):
		apierror.RespondError(c, apierror.NewNotFound(err.Error(), c.Request.URL.Path, requestID))
	case errors.Is(err, domain.ErrSessionNotFound), errors.Is(err, domain.ErrSessionExpired), errors.Is(err, domain.ErrSessionRevoked):
		apierror.RespondError(c, apierror.NewUnauthorized(err.Error(), c.Request.URL.Path, requestID))
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

func (h *Handler) sessionCookieName() string {
	return "session"
}

func (h *Handler) sessionTTL() int {
	return 86400
}

func (h *Handler) loadRoleEntities(c *gin.Context, roleIDs []domain.RoleID) ([]domain.Role, error) {
	if len(roleIDs) == 0 {
		allRoles, err := h.roleRepo.FindAll(c.Request.Context())
		if err != nil {
			return nil, err
		}
		result := make([]domain.Role, len(allRoles))
		for i, r := range allRoles {
			result[i] = *r
		}
		return result, nil
	}
	rolePtrs, err := h.roleRepo.FindByIDs(c.Request.Context(), roleIDs)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Role, len(rolePtrs))
	for i, r := range rolePtrs {
		result[i] = *r
	}
	return result, nil
}

func aggregatePermissions(roles []domain.Role) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, role := range roles {
		for _, p := range role.Permissions() {
			ps := p.String()
			if _, ok := seen[ps]; !ok {
				seen[ps] = struct{}{}
				result = append(result, ps)
			}
		}
	}
	return result
}
