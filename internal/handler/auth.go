package handler

import (
	"net/http"

	"auth-service/internal/entity"
	"auth-service/internal/entity/dto"
	"auth-service/internal/middleware"
	"auth-service/internal/usecase"
	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase usecase.AuthUseCaseInterface
	rg          *gin.RouterGroup
	jwtService  service.JwtServiceImpl
	mid middleware.AuthMiddleware
}

func (h *AuthHandler) Routes() {
    // Grup publik tanpa middleware
    h.rg.POST("/login", h.Login)
    h.rg.POST("/refresh", h.RefreshToken)
	// Logout requires a valid access token (any authenticated user)
	h.rg.POST("/logout", h.mid.RequiredToken(), h.Logout)
    h.rg.GET("/health", h.Health)

    // Grup untuk route yang membutuhkan token TENANT_ADMIN atau SUPER_ADMIN
    tenantAdminGroup := h.rg.Group("/")
    tenantAdminGroup.Use(h.mid.RequiredToken("TENANT_ADMIN", "SUPER_ADMIN"))
    {
        tenantAdminGroup.POST("/users", h.CreateUser)
    }
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase usecase.AuthUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		jwtService:  jwtService,
		rg:          rg,
		mid:         am,
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

// Note: Middleware already validates TENANT_ADMIN or SUPER_ADMIN role
func (h *AuthHandler) CreateUser(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get user info from context (injected by middleware)
	userClaims, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not found in context"})
		return
	}

	claims := userClaims.(*entity.AccessTokenClaims)

	var req dto.CreateUserRequestBody
	err := ctx.ShouldBind(&req)
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

// Logout handles POST /logout
func (h *AuthHandler) Logout(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, dto.ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get user info from context (injected by middleware)
	userClaims, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "User not found in context"})
		return
	}

	claims := userClaims.(*entity.AccessTokenClaims)

	var req dto.RefreshTokenRequestBody
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "refresh_token is required"})
		return
	}

	if err := h.authUseCase.Logout(ctx, claims.Sub, req.RefreshToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
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
