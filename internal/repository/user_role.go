package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"time"
)

// UserRoleRepository defines user-role repository interface
type UserRoleRepository interface {
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error)
}

// UserRoleRepositoryImpl implements UserRoleRepository
type UserRoleRepositoryImpl struct {
	db *sql.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return &UserRoleRepositoryImpl{db: db}
}

// AssignRoleToUser assigns a role to a user
func (r *UserRoleRepositoryImpl) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, userID, roleID, time.Now())
	return err
}

// RemoveRoleFromUser removes a role from a user
func (r *UserRoleRepositoryImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// GetRolesByUserID gets roles for a user
func (r *UserRoleRepositoryImpl) GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error) {
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
