package repository

import (
	"context"
	"database/sql"
	"time"
)

// RolePermissionRepository defines role-permission repository interface
type RolePermissionRepository interface {
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
}

// RolePermissionRepositoryImpl implements RolePermissionRepository
type RolePermissionRepositoryImpl struct {
	db *sql.DB
}

// NewRolePermissionRepository creates a new role permission repository
func NewRolePermissionRepository(db *sql.DB) RolePermissionRepository {
	return &RolePermissionRepositoryImpl{db: db}
}

// AssignPermissionToRole assigns a permission to a role
func (r *RolePermissionRepositoryImpl) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	return err
}

// RemovePermissionFromRole removes a permission from a role
func (r *RolePermissionRepositoryImpl) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}
