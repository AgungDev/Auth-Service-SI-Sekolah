package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"auth-service/internal/entity"
	"auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication logic
type AuthUseCase struct {
	userRepo              repository.UserRepository
	tenantRepo            repository.TenantRepository
	roleRepo              repository.RoleRepository
	userRoleRepo          repository.UserRoleRepository
	refreshTokenRepo      repository.RefreshTokenRepository
	auditLogRepo          repository.AuditLogRepository
	jwtService            JWTService
	refreshTokenExpiry    int
}

// JWTService interface for JWT operations
type JWTService interface {
	GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant) (string, error)
	VerifyAccessToken(token string) (*entity.AccessTokenClaims, error)
}

// NewAuthUseCase creates a new auth usecase
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	roleRepo repository.RoleRepository,
	userRoleRepo repository.UserRoleRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	auditLogRepo repository.AuditLogRepository,
	jwtService JWTService,
	refreshTokenExpiry int,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:           userRepo,
		tenantRepo:         tenantRepo,
		roleRepo:           roleRepo,
		userRoleRepo:       userRoleRepo,
		refreshTokenRepo:   refreshTokenRepo,
		auditLogRepo:       auditLogRepo,
		jwtService:         jwtService,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// LoginRequest represents login input
type LoginRequest struct {
	Email    string
	Password string
	TenantID string
}

// LoginResponse represents login output
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Login handles user login
func (u *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by email and tenant
	user, err := u.userRepo.GetUserByEmail(ctx, req.Email, req.TenantID)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check user status
	if user.Status != "ACTIVE" {
		return nil, errors.New("user is not active")
	}

	// Get tenant
	tenant, err := u.tenantRepo.GetTenantByID(ctx, req.TenantID)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	// Check tenant status
	if tenant.Status != "ACTIVE" {
		return nil, errors.New("tenant is not active")
	}

	// Get user roles
	roles, err := u.roleRepo.GetRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, err := u.jwtService.GenerateAccessToken(user, roles, tenant)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := u.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	rt := &entity.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		TenantID:  req.TenantID,
		ExpiresAt: time.Now().Add(time.Duration(u.refreshTokenExpiry) * time.Second),
		CreatedAt: time.Now(),
	}
	if err := u.refreshTokenRepo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, err
	}

	// Create audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:        uuid.New().String(),
		ActorID:   user.ID,
		TenantID:  req.TenantID,
		Action:    "USER_LOGIN",
		Target:    user.ID,
		CreatedAt: time.Now(),
	})

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    1800, // 30 minutes
	}, nil
}

// RefreshTokenRequest represents refresh token input
type RefreshTokenRequest struct {
	RefreshToken string
}

// RefreshToken refreshes access token
func (u *AuthUseCase) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	// Get refresh token
	rt, err := u.refreshTokenRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	// Check if token is revoked
	if rt.RevokedAt != nil {
		return nil, errors.New("refresh token is revoked")
	}

	// Check if token is expired
	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token is expired")
	}

	// Get user
	user, err := u.userRepo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Get tenant
	tenant, err := u.tenantRepo.GetTenantByID(ctx, rt.TenantID)
	if err != nil {
		return nil, errors.New("tenant not found")
	}

	// Get user roles
	roles, err := u.roleRepo.GetRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := u.jwtService.GenerateAccessToken(user, roles, tenant)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := u.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	if err := u.refreshTokenRepo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

	// Save new refresh token
	newRT := &entity.RefreshToken{
		Token:     newRefreshToken,
		UserID:    rt.UserID,
		TenantID:  rt.TenantID,
		ExpiresAt: time.Now().Add(time.Duration(u.refreshTokenExpiry) * time.Second),
		CreatedAt: time.Now(),
	}
	if err := u.refreshTokenRepo.SaveRefreshToken(ctx, newRT); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    1800,
	}, nil
}

// CreateTenantRequest represents create tenant input
type CreateTenantRequest struct {
	Name string
}

// CreateTenant creates a new tenant
func (u *AuthUseCase) CreateTenant(ctx context.Context, req CreateTenantRequest) (*entity.Tenant, error) {
	tenant := &entity.Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdTenant, err := u.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	// Audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:       uuid.New().String(),
		ActorID:  "system",
		TenantID: createdTenant.ID,
		Action:   "TENANT_CREATED",
		Target:   createdTenant.ID,
		Metadata: map[string]interface{}{
			"tenant_name": createdTenant.Name,
		},
		CreatedAt: time.Now(),
	})

	return createdTenant, nil
}

// CreateUserRequest represents create user input
type CreateUserRequest struct {
	Email    string
	Password string
	TenantID string
	RoleIDs  []string
}

// CreateUser creates a new user
func (u *AuthUseCase) CreateUser(ctx context.Context, req CreateUserRequest) (*entity.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New().String(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUser, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Assign roles
	for _, roleID := range req.RoleIDs {
		if err := u.userRoleRepo.AssignRoleToUser(ctx, createdUser.ID, roleID); err != nil {
			return nil, err
		}
	}

	// Audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:       uuid.New().String(),
		ActorID:  "system",
		TenantID: req.TenantID,
		Action:   "USER_CREATED",
		Target:   createdUser.ID,
		Metadata: map[string]interface{}{
			"email":   createdUser.Email,
			"role_ids": req.RoleIDs,
		},
		CreatedAt: time.Now(),
	})

	return createdUser, nil
}

// generateRefreshToken generates a random refresh token
func (u *AuthUseCase) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
