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


type TenantHandler struct {
	tenantUseCase usecase.TenantUseCaseInterface
	rg            *gin.RouterGroup
	jwtService    service.JwtServiceImpl
	mid middleware.AuthMiddleware
}

func (h *TenantHandler) Routes() {
    // Grup untuk route yang membutuhkan middleware token SUPER_ADMIN
    superAdminGroup := h.rg.Group("/")
    superAdminGroup.Use(h.mid.RequiredToken("SUPER_ADMIN"))
    {
        superAdminGroup.POST("/tenants", h.CreateTenant)
    }
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantUseCase usecase.TenantUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *TenantHandler {
	return &TenantHandler{
		tenantUseCase: tenantUseCase,
		jwtService:    jwtService,
		rg:            rg,
		mid:           am,
	}
}


// CreateTenant handles POST /tenants
// Note: Middleware already validates SUPER_ADMIN role
func (h *TenantHandler) CreateTenant(ctx *gin.Context) {
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

	_ = userClaims.(*entity.AccessTokenClaims)

	var req dto.CreateTenantRequestBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	tenant, err := h.tenantUseCase.CreateTenant(ctx, req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}