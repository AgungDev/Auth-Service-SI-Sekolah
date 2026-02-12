package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"errors"
)

// PermissionRepository defines permission repository interface
type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *entity.Permission) (*entity.Permission, error)
	GetPermissionByID(ctx context.Context, id string) (*entity.Permission, error)
	GetAllPermissions(ctx context.Context) ([]*entity.Permission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error)
}

// PermissionRepositoryImpl implements PermissionRepository
type PermissionRepositoryImpl struct {
	db *sql.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &PermissionRepositoryImpl{db: db}
}

// CreatePermission creates a new permission
func (r *PermissionRepositoryImpl) CreatePermission(ctx context.Context, permission *entity.Permission) (*entity.Permission, error) {
	query := `
		INSERT INTO permissions (id, code, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, description, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query,
		permission.ID, permission.Code, permission.Description, permission.CreatedAt, permission.UpdatedAt)

	var p entity.Permission
	err := row.Scan(&p.ID, &p.Code, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetPermissionByID gets a permission by ID
func (r *PermissionRepositoryImpl) GetPermissionByID(ctx context.Context, id string) (*entity.Permission, error) {
	query := `
		SELECT id, code, description, created_at, updated_at
		FROM permissions
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var permission entity.Permission
	err := row.Scan(&permission.ID, &permission.Code, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}

	return &permission, nil
}

// GetAllPermissions gets all permissions
func (r *PermissionRepositoryImpl) GetAllPermissions(ctx context.Context) ([]*entity.Permission, error) {
	query := `
		SELECT id, code, description, created_at, updated_at
		FROM permissions
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*entity.Permission
	for rows.Next() {
		var permission entity.Permission
		err := rows.Scan(&permission.ID, &permission.Code, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}

	return permissions, nil
}

// GetPermissionsByRoleID gets permissions for a role
func (r *PermissionRepositoryImpl) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	query := `
		SELECT p.id, p.code, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*entity.Permission
	for rows.Next() {
		var permission entity.Permission
		err := rows.Scan(&permission.ID, &permission.Code, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}

	return permissions, nil
}
