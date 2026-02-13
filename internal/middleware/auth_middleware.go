package middleware

import (
	"auth-service/pkg/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware interface {
	RequiredToken(roles ...string) gin.HandlerFunc
}

type authMiddleware struct {
	jwtService service.JwtServiceImpl
}

type authHeader struct {
	AuthorizationHeader string `header:"Authorization" binding:"required"`
}

// Middleware for validating access token and role restrictions
func (a *authMiddleware) RequiredToken(roles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var aH authHeader
		err := ctx.ShouldBindHeader(&aH)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			ctx.Abort()
			return
		}

		token := strings.Replace(aH.AuthorizationHeader, "Bearer ", "", 1)
		tokenClaim, err := a.jwtService.VerifyAccessToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			ctx.Abort()
			return
		}

		ctx.Set("user", tokenClaim)

		// Check if the user has the required role
		validRole := false
		for _, role := range roles {
			if contains(tokenClaim.Roles, role) {
				validRole = true
				break
			}
		}

		if !validRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this resource"})
			ctx.Abort()
			return
		}

		// Inject tenant_id into context for tenant isolation
		ctx.Set("tenant_id", tokenClaim.TenantID)
		ctx.Next()
	}
}

// Helper function to check if a role exists in the roles slice
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func NewAuthMiddleware(jwtService service.JwtServiceImpl) AuthMiddleware {
	return &authMiddleware{jwtService: jwtService}
}