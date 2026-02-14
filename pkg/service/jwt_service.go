package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/entity"
	"auth-service/internal/repository"
	"auth-service/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewJwtService(myConfig config.TokenConfig) JwtServiceImpl {
	return &jWTServiceImpl{cfg: myConfig}
}

type JwtServiceImpl interface {
	GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant, permissions []string) (string, error)
	VerifyAccessToken(token string) (*entity.AccessTokenClaims, error)
	AccessTokenLifeTime() int
}

// JWTServiceImpl handles JWT operations
type jWTServiceImpl struct {
	cfg config.TokenConfig
}

// GenerateAccessToken generates a JWT access token
func (s *jWTServiceImpl) GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant, permissions []string) (string, error) {
	// determine primary role (use first role if present)
	primaryRole := ""
	isSuper := false
	if len(roles) > 0 {
		primaryRole = roles[0].Name
		for _, r := range roles {
			if r.Name == "SUPER_ADMIN" || r.Name == "super-admin" {
				isSuper = true
				break
			}
		}
	}

	claims := jwt.MapClaims{
		"sub":            user.ID,
		"tenant_id":      tenant.ID,
		"role":           primaryRole,
		"permissions":    permissions,
		"is_super_admin": isSuper,
		"exp":            time.Now().Add(time.Duration(s.cfg.AccessTokenLifeTime) * time.Second).Unix(),
		"iat":            time.Now().Unix(),
		"iss":            s.cfg.AplicationName,
	}

	token := jwt.NewWithClaims(s.cfg.JwtSigningMethod, claims)
	return token.SignedString(s.cfg.JwtSignatureKey)
}

// VerifyAccessToken verifies and parses a JWT access token
func (s *jWTServiceImpl) VerifyAccessToken(tokenString string) (*entity.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.cfg.JwtSignatureKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract role
	roleStr, _ := claims["role"].(string)

	// Extract permissions
	perms := []string{}
	if pi, ok := claims["permissions"].([]interface{}); ok {
		for _, v := range pi {
			if ps, ok := v.(string); ok {
				perms = append(perms, ps)
			}
		}
	}

	// is_super_admin
	isSuper := false
	if v, ok := claims["is_super_admin"].(bool); ok {
		isSuper = v
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid sub claim")
	}

	tenantID, ok := claims["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tenant_id claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid exp claim")
	}

	iat, _ := claims["iat"].(float64)

	issuer, _ := claims["iss"].(string)

	return &entity.AccessTokenClaims{
		Sub:          sub,
		TenantID:     tenantID,
		Role:         roleStr,
		Permissions:  perms,
		IsSuperAdmin: isSuper,
		ExpiresAt:    int64(exp),
		IssuedAt:     int64(iat),
		Issuer:       issuer,
	}, nil
}

func (s *jWTServiceImpl) AccessTokenLifeTime() int {
	return s.cfg.AccessTokenLifeTime
}

// RoleUseCase handles role operations
type RoleUseCase struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	rolepermRepo   repository.RolePermissionRepository
	auditLogRepo   repository.AuditLogRepository
}

// NewRoleUseCase creates a new role usecase
func NewRoleUseCase(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	rolepermRepo repository.RolePermissionRepository,
	auditLogRepo repository.AuditLogRepository,
) *RoleUseCase {
	return &RoleUseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		rolepermRepo:   rolepermRepo,
		auditLogRepo:   auditLogRepo,
	}
}

// CreateRoleRequest represents create role input
type CreateRoleRequest struct {
	Name  string
	Scope string // SYSTEM or TENANT
}

// CreateRole creates a new role
func (u *RoleUseCase) CreateRole(ctx context.Context, req CreateRoleRequest) (*entity.Role, error) {
	role := &entity.Role{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Scope:     req.Scope,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return u.roleRepo.CreateRole(ctx, role)
}

// Error definitions
var (
	ErrInvalidToken = errors.New("invalid token")
)
