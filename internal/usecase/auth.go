package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"auth-service/internal/entity"
	"auth-service/internal/entity/dto"
	"auth-service/internal/repository"
	"auth-service/pkg/service"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCaseInterface interface {
	Login(ctx context.Context, req dto.LoginRequestBody) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequestBody) (*dto.LoginResponse, error)
	CreateUser(ctx context.Context, req dto.CreateUserRequestBody) (*entity.User, error)
	Logout(ctx context.Context, actorID, refreshToken string) error
}

// AuthUseCase handles authentication logic
type authUseCase struct {
	userRepo           repository.UserRepository
	tenantRepo         repository.TenantRepository
	roleRepo           repository.RoleRepository
	permissionRepo     repository.PermissionRepository
	userRoleRepo       repository.UserRoleRepository
	refreshTokenRepo   repository.RefreshTokenRepository
	auditLogRepo       repository.AuditLogRepository
	jwtService         service.JwtServiceImpl
	refreshTokenExpiry int
}

// NewAuthUseCase creates a new auth usecase
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	userRoleRepo repository.UserRoleRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	auditLogRepo repository.AuditLogRepository,
	jwtService service.JwtServiceImpl,
	refreshTokenExpiry int,
) AuthUseCaseInterface {
	return &authUseCase{
		userRepo:           userRepo,
		tenantRepo:         tenantRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
		userRoleRepo:       userRoleRepo,
		refreshTokenRepo:   refreshTokenRepo,
		auditLogRepo:       auditLogRepo,
		jwtService:         jwtService,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Login handles user login
func (u *authUseCase) Login(ctx context.Context, req dto.LoginRequestBody) (*dto.LoginResponse, error) {
	// Defensive checks to avoid nil-interface panics when running in different environments
	if u == nil {
		return nil, errors.New("internal error")
	}
	// Report which dependency is missing to aid debugging
	if u.userRepo == nil {
		return nil, errors.New("missing userRepo")
	}
	if u.tenantRepo == nil {
		return nil, errors.New("missing tenantRepo")
	}
	if u.roleRepo == nil {
		return nil, errors.New("missing roleRepo")
	}
	if u.refreshTokenRepo == nil {
		return nil, errors.New("missing refreshTokenRepo")
	}
	if u.auditLogRepo == nil {
		return nil, errors.New("missing auditLogRepo")
	}
	if u.jwtService == nil {
		return nil, errors.New("missing jwtService")
	}
	// Get user by email and tenant
	user, err := u.userRepo.GetUserByEmail(ctx, req.Email, req.TenantID)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if user.PasswordHash == "" {
		return nil, errors.New("invalid credentials")
	}
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

	// Aggregate permissions for all roles (permissionRepo may be nil in some setups)
	permissions := []string{}
	if u.permissionRepo != nil {
		for _, r := range roles {
			perms, err := u.permissionRepo.GetPermissionsByRoleID(ctx, r.ID)
			if err != nil {
				return nil, err
			}
			for _, p := range perms {
				permissions = append(permissions, p.Code)
			}
		}
	}

	// Generate access token
	accessToken, err := u.jwtService.GenerateAccessToken(user, roles, tenant, permissions)
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

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    u.jwtService.AccessTokenLifeTime(),
	}, nil
}

// RefreshToken refreshes access token
func (u *authUseCase) RefreshToken(ctx context.Context, req dto.RefreshTokenRequestBody) (*dto.LoginResponse, error) {
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

	// Aggregate permissions for roles (permissionRepo may be nil)
	permissions := []string{}
	if u.permissionRepo != nil {
		for _, r := range roles {
			perms, err := u.permissionRepo.GetPermissionsByRoleID(ctx, r.ID)
			if err != nil {
				return nil, err
			}
			for _, p := range perms {
				permissions = append(permissions, p.Code)
			}
		}
	}

	// Generate new access token
	accessToken, err := u.jwtService.GenerateAccessToken(user, roles, tenant, permissions)
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

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    u.jwtService.AccessTokenLifeTime(),
	}, nil
}

// CreateUser creates a new user
func (u *authUseCase) CreateUser(ctx context.Context, req dto.CreateUserRequestBody) (*entity.User, error) {
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
			"email":    createdUser.Email,
			"role_ids": req.RoleIDs,
		},
		CreatedAt: time.Now(),
	})

	return createdUser, nil
}

// generateRefreshToken generates a random refresh token
func (u *authUseCase) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Logout revokes the provided refresh token and records an audit log
func (u *authUseCase) Logout(ctx context.Context, actorID, refreshToken string) error {
	// Revoke the refresh token
	if err := u.refreshTokenRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return err
	}

	// Record audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:       uuid.New().String(),
		ActorID:  actorID,
		TenantID: "",
		Action:   "USER_LOGOUT",
		Target:   refreshToken,
		CreatedAt: time.Now(),
	})

	return nil
}
