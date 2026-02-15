package handler

import (
	"net/http"
	"strconv"

	"auth-service/internal/entity"
	"auth-service/internal/entity/dto"
	"auth-service/internal/middleware"
	"auth-service/internal/usecase"
	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	auditLogUseCase usecase.AuditLogUseCaseInterface
	rg              *gin.RouterGroup
	jwtService      service.JwtServiceImpl
	mid             middleware.AuthMiddleware
}

func NewAuditLogHandler(a usecase.AuditLogUseCaseInterface, jwtService service.JwtServiceImpl, rg *gin.RouterGroup, am middleware.AuthMiddleware) *AuditLogHandler {
	return &AuditLogHandler{auditLogUseCase: a, rg: rg, jwtService: jwtService, mid: am}
}

func (h *AuditLogHandler) Routes() {
	group := h.rg.Group("/")
	group.Use(h.mid.RequiredToken("TENANT_ADMIN", "SUPER_ADMIN"))
	{
		group.GET("/audit-logs", h.ListAuditLogs)
	}
}

// ListAuditLogs handles GET /audit-logs with pagination
// Query params: limit (default 20, max 100), offset (default 0)
// For SUPER_ADMIN, can optionally query tenant_id; for TENANT_ADMIN restricted to their tenant
func (h *AuditLogHandler) ListAuditLogs(ctx *gin.Context) {
	userClaims, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "user not found in context"})
		return
	}
	claims := userClaims.(*entity.AccessTokenClaims)

	// Get pagination params
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Determine tenant ID based on role
	tenantID := claims.TenantID
	if claims.IsSuperAdmin {
		// SUPER_ADMIN can optionally query a specific tenant; default to all
		if queryTenant := ctx.Query("tenant_id"); queryTenant != "" {
			tenantID = queryTenant
		}
	}

	if tenantID == "" {
		// If no tenant context, cannot proceed (SUPER_ADMIN must specify or have context)
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_request", Message: "tenant_id is required"})
		return
	}

	logs, err := h.auditLogUseCase.GetAuditLogs(ctx, tenantID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Return with pagination metadata
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Audit logs fetched",
		"data":    logs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}
