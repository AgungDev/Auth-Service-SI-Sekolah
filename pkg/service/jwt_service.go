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
	GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant) (string, error)
	VerifyAccessToken(token string) (*entity.AccessTokenClaims, error)
}

// JWTServiceImpl handles JWT operations
type jWTServiceImpl struct {
	cfg config.TokenConfig
}

// type JWTServiceImpl struct {
// 	secretKey string
// 	expiry    int
// }

// NewJWTService creates a new JWT service
// func NewJWTService(secretKey string, expiry int) *JWTServiceImpl {
// 	return &JWTServiceImpl{
// 		secretKey: secretKey,
// 		expiry:    expiry,
// 	}
// }

// GenerateAccessToken generates a JWT access token
func (s *jWTServiceImpl) GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant) (string, error) {
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	claims := jwt.MapClaims{
		"sub":           user.ID,
		"tenant_id":     tenant.ID,
		"email":         user.Email,
		"roles":         roleNames,
		"tenant_status": tenant.Status,
		"exp":           time.Now().Add(time.Duration(s.cfg.AccessTokenLifeTime) * time.Second).Unix(),
		"iat":           time.Now().Unix(),
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
		return []byte(s.cfg.JwtSignatureKey), nil
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

	// Extract roles as interface slice then convert to string slice
	rolesInterface, ok := claims["roles"].([]interface{})
	roles := make([]string, len(rolesInterface))
	if ok {
		for i, v := range rolesInterface {
			if roleStr, ok := v.(string); ok {
				roles[i] = roleStr
			}
		}
	}

	// Safely extract string claims with type checking
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid sub claim")
	}

	tenantID, ok := claims["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tenant_id claim")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email claim")
	}

	tenantStatus, ok := claims["tenant_status"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tenant_status claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid exp claim")
	}

	return &entity.AccessTokenClaims{
		Sub:          sub,
		TenantID:     tenantID,
		Email:        email,
		Roles:        roles,
		TenantStatus: tenantStatus,
		ExpiresAt:    int64(exp),
	}, nil
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
