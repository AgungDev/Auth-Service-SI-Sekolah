package usecase

import (
	"context"
	"time"

	"auth-service/internal/entity"
	"auth-service/internal/entity/dto"
	"auth-service/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCaseInterface interface {
    UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequestBody, actorID string) (*entity.User, error)
    DisableUser(ctx context.Context, id string, actorID string) error
    ListUsers(ctx context.Context, tenantID string, isSuperAdmin bool) ([]*entity.User, error)
    GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error)
}

type userUseCase struct {
    userRepo         repository.UserRepository
    userRoleRepo     repository.UserRoleRepository
    refreshTokenRepo repository.RefreshTokenRepository
    auditLogRepo     repository.AuditLogRepository
}

func NewUserUseCase(
    userRepo repository.UserRepository,
    userRoleRepo repository.UserRoleRepository,
    refreshTokenRepo repository.RefreshTokenRepository,
    auditLogRepo repository.AuditLogRepository,
) UserUseCaseInterface {
    return &userUseCase{
        userRepo:         userRepo,
        userRoleRepo:     userRoleRepo,
        refreshTokenRepo: refreshTokenRepo,
        auditLogRepo:     auditLogRepo,
    }
}

// ListUsers returns users for a tenant or all users if isSuperAdmin is true
func (u *userUseCase) ListUsers(ctx context.Context, tenantID string, isSuperAdmin bool) ([]*entity.User, error) {
    if isSuperAdmin {
        return u.userRepo.GetAllUsersAll(ctx)
    }
    return u.userRepo.GetAllUsers(ctx, tenantID)
}

// GetUserProfile returns full profile of authenticated user with roles
func (u *userUseCase) GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error) {
    user, err := u.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    roles, err := u.userRoleRepo.GetRolesByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    roleNames := []string{}
    for _, r := range roles {
        roleNames = append(roleNames, r.Name)
    }

    return &dto.UserProfileResponse{
        ID:        user.ID,
        TenantID:  user.TenantID,
        Email:     user.Email,
        Status:    user.Status,
        Roles:     roleNames,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }, nil
}


// UpdateUser updates email/password and role assignments
func (u *userUseCase) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequestBody, actorID string) (*entity.User, error) {
    user, err := u.userRepo.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Apply updates
    if req.Email != "" {
        user.Email = req.Email
    }
    if req.Password != "" {
        hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        user.PasswordHash = string(hashed)
    }
    if req.TenantID != "" {
        user.TenantID = req.TenantID
    }
    user.UpdatedAt = time.Now()

    // perform update and role changes atomically using repository transaction
    if err := u.userRepo.UpdateUserWithRoles(ctx, user, req.RoleIDs); err != nil {
        return nil, err
    }

    // Audit log
    u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
        ID:        uuid.New().String(),
        ActorID:   actorID,
        TenantID:  user.TenantID,
        Action:    "USER_UPDATED",
        Target:    user.ID,
        CreatedAt: time.Now(),
    })

    // return updated user
    updated, err := u.userRepo.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return updated, nil
}

// DisableUser disables a user and revokes refresh tokens
func (u *userUseCase) DisableUser(ctx context.Context, id string, actorID string) error {
    user, err := u.userRepo.GetUserByID(ctx, id)
    if err != nil {
        return err
    }

    // set to INACTIVE
    if err := u.userRepo.UpdateUserStatus(ctx, id, "INACTIVE"); err != nil {
        return err
    }

    // revoke refresh tokens
    if err := u.refreshTokenRepo.RevokeAllRefreshTokensByUserID(ctx, id); err != nil {
        return err
    }

    u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
        ID:        uuid.New().String(),
        ActorID:   actorID,
        TenantID:  user.TenantID,
        Action:    "USER_DISABLED",
        Target:    user.ID,
        CreatedAt: time.Now(),
    })

    return nil
}
