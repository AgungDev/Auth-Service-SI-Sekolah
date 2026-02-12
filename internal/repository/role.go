package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"errors"
)

// RoleRepository defines role repository interface
type RoleRepository interface {
	CreateRole(ctx context.Context, role *entity.Role) (*entity.Role, error)
	GetRoleByID(ctx context.Context, id string) (*entity.Role, error)
	GetAllRoles(ctx context.Context) ([]*entity.Role, error)
	GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error)
}

// RoleRepositoryImpl implements RoleRepository
type RoleRepositoryImpl struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &RoleRepositoryImpl{db: db}
}

// CreateRole creates a new role
func (r *RoleRepositoryImpl) CreateRole(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	query := `
		INSERT INTO roles (id, name, scope, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, scope, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query,
		role.ID, role.Name, role.Scope, role.CreatedAt, role.UpdatedAt)

	var ro entity.Role
	err := row.Scan(&ro.ID, &ro.Name, &ro.Scope, &ro.CreatedAt, &ro.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &ro, nil
}

// GetRoleByID gets a role by ID
func (r *RoleRepositoryImpl) GetRoleByID(ctx context.Context, id string) (*entity.Role, error) {
	query := `
		SELECT id, name, scope, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var role entity.Role
	err := row.Scan(&role.ID, &role.Name, &role.Scope, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// GetAllRoles gets all roles
func (r *RoleRepositoryImpl) GetAllRoles(ctx context.Context) ([]*entity.Role, error) {
	query := `
		SELECT id, name, scope, created_at, updated_at
		FROM roles
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		var role entity.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Scope, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// GetRolesByUserID gets roles for a user
func (r *RoleRepositoryImpl) GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.scope, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		var role entity.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Scope, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}
