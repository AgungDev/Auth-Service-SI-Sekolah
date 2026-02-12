package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"errors"
	"time"
)

// UserRepository defines user repository interface
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email, tenantID string) (*entity.User, error)
	GetAllUsers(ctx context.Context, tenantID string) ([]*entity.User, error)
	UpdateUserStatus(ctx context.Context, id, status string) error
}

// UserRepositoryImpl implements UserRepository
type UserRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, email, password_hash, status, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Status, user.CreatedAt, user.UpdatedAt)

	var u entity.User
	err := row.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUserByID gets a user by ID
func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var user entity.User
	err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail gets a user by email and tenant ID
func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email, tenantID string) (*entity.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND tenant_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, email, tenantID)
	var user entity.User
	err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetAllUsers gets all users for a tenant
func (r *UserRepositoryImpl) GetAllUsers(ctx context.Context, tenantID string) ([]*entity.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// UpdateUserStatus updates user status
func (r *UserRepositoryImpl) UpdateUserStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
