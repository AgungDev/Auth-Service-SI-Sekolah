package repository

import (
	"context"
	"auth-service/internal/entity"
)

// TenantRepository defines tenant repository interface
type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error)
	GetTenantByID(ctx context.Context, id string) (*entity.Tenant, error)
	GetAllTenants(ctx context.Context) ([]*entity.Tenant, error)
	UpdateTenantStatus(ctx context.Context, id, status string) error
}

// UserRepository defines user repository interface
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email, tenantID string) (*entity.User, error)
	GetAllUsers(ctx context.Context, tenantID string) ([]*entity.User, error)
	UpdateUserStatus(ctx context.Context, id, status string) error
}

// RoleRepository defines role repository interface
type RoleRepository interface {
	CreateRole(ctx context.Context, role *entity.Role) (*entity.Role, error)
	GetRoleByID(ctx context.Context, id string) (*entity.Role, error)
	GetAllRoles(ctx context.Context) ([]*entity.Role, error)
	GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error)
}

// PermissionRepository defines permission repository interface
type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *entity.Permission) (*entity.Permission, error)
	GetPermissionByID(ctx context.Context, id string) (*entity.Permission, error)
	GetAllPermissions(ctx context.Context) ([]*entity.Permission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
}

// RolePermissionRepository defines role-permission repository interface
type RolePermissionRepository interface {
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
}

// UserRoleRepository defines user-role repository interface
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error)
}

// RefreshTokenRepository defines refresh token repository interface
type RefreshTokenRepository interface {
	SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshTokensByUserID(ctx context.Context, userID string) error
}

// AuditLogRepository defines audit log repository interface
type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, log *entity.AuditLog) error
	GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error)
}
