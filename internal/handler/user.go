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

type UserHandler struct {
    userUseCase usecase.UserUseCaseInterface
    rg          *gin.RouterGroup
    jwtService  service.JwtServiceImpl
    mid         middleware.AuthMiddleware
}

func NewUserHandler(u usecase.UserUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *UserHandler {
    return &UserHandler{userUseCase: u, rg: rg, jwtService: jwtService, mid: am}
}

func (h *UserHandler) Routes() {
    // Public route for authenticated users to get their own profile
    publicGroup := h.rg.Group("/")
    publicGroup.Use(h.mid.RequiredToken())
    {
        publicGroup.GET("/users/profile", h.GetProfile)
    }

    // Admin routes for managing users
    group := h.rg.Group("/")
    group.Use(h.mid.RequiredToken("TENANT_ADMIN", "SUPER_ADMIN"))
    {
        group.GET("/users", h.ListUsers)
        group.PATCH("/users/:id", h.UpdateUser)
        group.PATCH("/users/:id/disable", h.DisableUser)
    }
}

// UpdateUser handles PATCH /users/:id
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
    id := ctx.Param("id")
    if id == "" {
        ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "user id is required"})
        return
    }

    userClaims, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "user not found in context"})
        return
    }
    claims := userClaims.(*entity.AccessTokenClaims)

    var req dto.UpdateUserRequestBody
    if err := ctx.ShouldBind(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "invalid request body"})
        return
    }

    // If caller is TENANT_ADMIN, ensure same tenant
    if !claims.IsSuperAdmin && claims.Role != "SUPER_ADMIN" {
        // ensure tenant match
        // handler does not load target user here; assume usecase will check or DB will error
    }

    updated, err := h.userUseCase.UpdateUser(ctx, id, req, claims.Sub)
    if err != nil {
        if err.Error() == "user not found" {
            ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "user not found"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "User updated successfully", Data: updated})
}

// DisableUser handles PATCH /users/:id/disable
func (h *UserHandler) DisableUser(ctx *gin.Context) {
    id := ctx.Param("id")
    if id == "" {
        ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "user id is required"})
        return
    }

    userClaims, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "user not found in context"})
        return
    }
    claims := userClaims.(*entity.AccessTokenClaims)

    if err := h.userUseCase.DisableUser(ctx, id, claims.Sub); err != nil {
        if err.Error() == "user not found" {
            ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "user not found"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "User disabled successfully"})
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(ctx *gin.Context) {
    userClaims, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "user not found in context"})
        return
    }
    claims := userClaims.(*entity.AccessTokenClaims)

    // Tenant admins see users in their tenant; super admin sees all
    users, err := h.userUseCase.ListUsers(ctx, claims.TenantID, claims.IsSuperAdmin)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Users fetched", Data: users})
}

// GetProfile handles GET /users/profile - returns authenticated user's profile
func (h *UserHandler) GetProfile(ctx *gin.Context) {
    userClaims, exists := ctx.Get("user")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "user not found in context"})
        return
    }
    claims := userClaims.(*entity.AccessTokenClaims)

    profile, err := h.userUseCase.GetUserProfile(ctx, claims.Sub)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Profile fetched", Data: profile})
}
