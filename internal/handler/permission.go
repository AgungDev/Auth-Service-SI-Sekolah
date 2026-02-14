package handler

import (
	"net/http"

	"auth-service/internal/entity/dto"
	"auth-service/internal/middleware"
	"auth-service/internal/usecase"
	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionUseCase usecase.PermissionUseCaseInterface
	rg                *gin.RouterGroup
	jwtService        service.JwtServiceImpl
	mid               middleware.AuthMiddleware
}

func NewPermissionHandler(p usecase.PermissionUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *PermissionHandler {
	return &PermissionHandler{permissionUseCase: p, rg: rg, jwtService: jwtService, mid: am}
}

func (h *PermissionHandler) Routes() {
	group := h.rg.Group("/")
	group.Use(h.mid.RequiredToken("SUPER_ADMIN"))
	{
		group.POST("/permissions", h.CreatePermission)
		group.GET("/permissions", h.ListPermissions)
		group.GET("/permissions/:id", h.GetPermission)
		group.PATCH("/permissions/:id", h.UpdatePermission)
		group.DELETE("/permissions/:id", h.DeletePermission)
	}
}

// CreatePermission handles POST /permissions
func (h *PermissionHandler) CreatePermission(ctx *gin.Context) {
	var req struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "invalid request body"})
		return
	}

	if req.Code == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "code is required"})
		return
	}

	permission, err := h.permissionUseCase.CreatePermission(ctx, req.Code, req.Description)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{Message: "Permission created successfully", Data: permission})
}

// GetPermission handles GET /permissions/:id
func (h *PermissionHandler) GetPermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "permission id is required"})
		return
	}

	permission, err := h.permissionUseCase.GetPermissionByID(ctx, id)
	if err != nil {
		if err.Error() == "permission not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "permission not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permission fetched", Data: permission})
}

// ListPermissions handles GET /permissions
func (h *PermissionHandler) ListPermissions(ctx *gin.Context) {
	permissions, err := h.permissionUseCase.GetAllPermissions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permissions fetched", Data: permissions})
}

// UpdatePermission handles PATCH /permissions/:id
func (h *PermissionHandler) UpdatePermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "permission id is required"})
		return
	}

	var req struct {
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "invalid request body"})
		return
	}

	permission, err := h.permissionUseCase.UpdatePermission(ctx, id, req.Code, req.Description)
	if err != nil {
		if err.Error() == "permission not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "permission not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permission updated successfully", Data: permission})
}

// DeletePermission handles DELETE /permissions/:id
func (h *PermissionHandler) DeletePermission(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "permission id is required"})
		return
	}

	err := h.permissionUseCase.DeletePermission(ctx, id)
	if err != nil {
		if err.Error() == "permission not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "permission not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permission deleted successfully"})
}
