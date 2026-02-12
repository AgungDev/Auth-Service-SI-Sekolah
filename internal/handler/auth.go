package handler

import (
	"net/http"
	"strings"

	"auth-service/internal/entity/dto"
	"auth-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase usecase.AuthUseCaseInterface
	rg          *gin.RouterGroup
	jwtService  usecase.JWTService
}

func (h *AuthHandler) Routes() {
	h.rg.POST("/login", h.Login)
	h.rg.POST("/refresh", h.RefreshToken)
	h.rg.POST("/tenants", h.CreateTenant)
	h.rg.POST("/users", h.CreateUser)
	h.rg.GET("/health", h.Health)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase usecase.AuthUseCaseInterface, jwtService usecase.JWTService, rg *gin.RouterGroup) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		jwtService:  jwtService,
		rg:          rg,
	}
}

// func NewAuthHandler(authUseCase *usecase.AuthUseCase, jwtService usecase.JWTService) *AuthHandler {
// 	return &AuthHandler{
// 		authUseCase: authUseCase,
// 		jwtService:  jwtService,
// 	}
// }

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// LoginRequest represents the login request body
type LoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
}

// Login handles POST /login
func (h *AuthHandler) Login(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req LoginRequestBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	resp, err := h.authUseCase.Login(ctx, dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		TenantID: req.TenantID,
	})

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RefreshTokenRequestBody represents the refresh token request body
type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken handles POST /refresh
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req RefreshTokenRequestBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	resp, err := h.authUseCase.RefreshToken(ctx, dto.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateTenantRequestBody represents the create tenant request body
type CreateTenantRequestBody struct {
	Name string `json:"name"`
}

// CreateTenant handles POST /tenants
func (h *AuthHandler) CreateTenant(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization - SUPER_ADMIN only
	token := strings.TrimPrefix(ctx.Request.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	// Check if user has SUPER_ADMIN role
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "SUPER_ADMIN" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "Insufficient permissions"})
		return
	}

	var req CreateTenantRequestBody
	err = ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	tenant, err := h.authUseCase.CreateTenant(ctx, dto.CreateTenantRequest{
		Name: req.Name,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}

// CreateUserRequestBody represents the create user request body
type CreateUserRequestBody struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	TenantID string   `json:"tenant_id"`
	RoleIDs  []string `json:"role_ids"`
}

// CreateUser handles POST /users
func (h *AuthHandler) CreateUser(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization
	token := strings.TrimPrefix(ctx.Request.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
		return
	}

	var req CreateUserRequestBody
	err = ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if user has permission to create users in this tenant
	if claims.TenantID != req.TenantID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "Cannot create users in other tenants"})
		return
	}

	user, err := h.authUseCase.CreateUser(ctx, dto.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		TenantID: req.TenantID,
		RoleIDs:  req.RoleIDs,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, SuccessResponse{
		Message: "User created successfully",
		Data:    user,
	})
}

// Health handles GET /health
func (h *AuthHandler) Health(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Auth Service is running",
		Data:    nil,
	})
}
