package handler

import (
	"net/http"

	"auth-service/internal/entity/dto"
	"auth-service/internal/middleware"
	"auth-service/internal/usecase"
	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleUseCase usecase.RoleUseCaseInterface
	rg          *gin.RouterGroup
	jwtService  service.JwtServiceImpl
	mid         middleware.AuthMiddleware
}

func NewRoleHandler(r usecase.RoleUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *RoleHandler {
	return &RoleHandler{roleUseCase: r, rg: rg, jwtService: jwtService, mid: am}
}

func (h *RoleHandler) Routes() {
	group := h.rg.Group("/")
	group.Use(h.mid.RequiredToken("SUPER_ADMIN"))
	{
		group.POST("/roles", h.CreateRole)
		group.GET("/roles", h.ListRoles)
		group.GET("/roles/:id", h.GetRole)
		group.PATCH("/roles/:id", h.UpdateRole)
		group.DELETE("/roles/:id", h.DeleteRole)
		group.POST("/roles/:id/permissions/:permissionId", h.AssignPermission)
		group.DELETE("/roles/:id/permissions/:permissionId", h.RemovePermission)
		group.GET("/roles/:id/permissions", h.GetRolePermissions)
	}
}

// CreateRole handles POST /roles
func (h *RoleHandler) CreateRole(ctx *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "invalid request body"})
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "name is required"})
		return
	}

	role, err := h.roleUseCase.CreateRole(ctx, req.Name, req.Scope)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.SuccessResponse{Message: "Role created successfully", Data: role})
}

// GetRole handles GET /roles/:id
func (h *RoleHandler) GetRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id is required"})
		return
	}

	role, err := h.roleUseCase.GetRoleByID(ctx, id)
	if err != nil {
		if err.Error() == "role not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "role not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Role fetched", Data: role})
}

// ListRoles handles GET /roles
func (h *RoleHandler) ListRoles(ctx *gin.Context) {
	roles, err := h.roleUseCase.GetAllRoles(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Roles fetched", Data: roles})
}

// UpdateRole handles PATCH /roles/:id
func (h *RoleHandler) UpdateRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id is required"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	}

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "invalid request body"})
		return
	}

	role, err := h.roleUseCase.UpdateRole(ctx, id, req.Name, req.Scope)
	if err != nil {
		if err.Error() == "role not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "role not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Role updated successfully", Data: role})
}

// DeleteRole handles DELETE /roles/:id
func (h *RoleHandler) DeleteRole(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id is required"})
		return
	}

	err := h.roleUseCase.DeleteRole(ctx, id)
	if err != nil {
		if err.Error() == "role not found" {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "role not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Role deleted successfully"})
}

// AssignPermission handles POST /roles/:id/permissions/:permissionId
func (h *RoleHandler) AssignPermission(ctx *gin.Context) {
	roleID := ctx.Param("id")
	permissionID := ctx.Param("permissionId")

	if roleID == "" || permissionID == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id and permission id are required"})
		return
	}

	err := h.roleUseCase.AssignPermissionToRole(ctx, roleID, permissionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permission assigned successfully"})
}

// RemovePermission handles DELETE /roles/:id/permissions/:permissionId
func (h *RoleHandler) RemovePermission(ctx *gin.Context) {
	roleID := ctx.Param("id")
	permissionID := ctx.Param("permissionId")

	if roleID == "" || permissionID == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id and permission id are required"})
		return
	}

	err := h.roleUseCase.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permission removed successfully"})
}

// GetRolePermissions handles GET /roles/:id/permissions
func (h *RoleHandler) GetRolePermissions(ctx *gin.Context) {
	roleID := ctx.Param("id")
	if roleID == "" {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "role id is required"})
		return
	}

	permissions, err := h.roleUseCase.GetRolePermissions(ctx, roleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Message: "Permissions fetched", Data: permissions})
}
