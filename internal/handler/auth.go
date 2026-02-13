package handler

import (
	"net/http"
	"strings"

	"auth-service/internal/entity/dto"
	"auth-service/internal/usecase"
	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase usecase.AuthUseCaseInterface
	rg          *gin.RouterGroup
	jwtService  service.JwtServiceImpl
}

func (h *AuthHandler) Routes() {
	h.rg.POST("/login", h.Login)
	h.rg.POST("/refresh", h.RefreshToken)
	h.rg.POST("/tenants", h.CreateTenant)
	h.rg.POST("/users", h.CreateUser)
	h.rg.GET("/health", h.Health)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase usecase.AuthUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		jwtService:  jwtService,
		rg:          rg,
	}
}

// Login handles POST /login
func (h *AuthHandler) Login(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req dto.LoginRequestBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	resp, err := h.authUseCase.Login(ctx, req)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// RefreshToken handles POST /refresh
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req dto.RefreshTokenRequestBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	resp, err := h.authUseCase.RefreshToken(ctx, req)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateTenant handles POST /tenants
func (h *AuthHandler) CreateTenant(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization - SUPER_ADMIN only
	token := strings.TrimPrefix(ctx.Request.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
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
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Insufficient permissions"})
		return
	}

	var req dto.CreateTenantRequestBody
	err = ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	tenant, err := h.authUseCase.CreateTenant(ctx, req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}

// CreateUser handles POST /users
func (h *AuthHandler) CreateUser(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Check authorization
	token := strings.TrimPrefix(ctx.Request.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Missing authorization token"})
		return
	}

	claims, err := h.jwtService.VerifyAccessToken(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return
	}

	var req dto.CreateUserRequestBody
	err = ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if user has permission to create users in this tenant
	if claims.TenantID != req.TenantID {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Cannot create users in other tenants"})
		return
	}

	user, err := h.authUseCase.CreateUser(ctx, req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "User created successfully",
		Data:    user,
	})
}

// Health handles GET /health
func (h *AuthHandler) Health(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Auth Service is running",
		Data:    nil,
	})
}
