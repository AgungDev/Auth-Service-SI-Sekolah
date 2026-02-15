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
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetAllUsers(ctx context.Context, tenantID string) ([]*entity.User, error)
	GetAllUsersAll(ctx context.Context) ([]*entity.User, error)
	UpdateUserStatus(ctx context.Context, id, status string) error
	UpdateUser(ctx context.Context, user *entity.User) error
	// UpdateUserWithRoles updates user fields and role assignments atomically
	UpdateUserWithRoles(ctx context.Context, user *entity.User, roleIDs []string) error
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
func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)
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

// GetAllUsersAll returns users across all tenants (for SUPER_ADMIN)
func (r *UserRepositoryImpl) GetAllUsersAll(ctx context.Context) ([]*entity.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, status, created_at, updated_at
		FROM users
	`

	rows, err := r.db.QueryContext(ctx, query)
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

// UpdateUser updates user fields such as email and password_hash
func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.PasswordHash, time.Now(), user.ID)
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

// UpdateUserWithRoles updates user fields and replaces role assignments inside a transaction.
// If roleIDs is nil, role assignments are left unchanged. If roleIDs is empty slice, roles are cleared.
func (r *UserRepositoryImpl) UpdateUserWithRoles(ctx context.Context, user *entity.User, roleIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// update user including tenant_id
	query := `
		UPDATE users
		SET tenant_id = $1, email = $2, password_hash = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = tx.ExecContext(ctx, query, user.TenantID, user.Email, user.PasswordHash, time.Now(), user.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if roleIDs != nil {
		// remove all existing roles
		if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, user.ID); err != nil {
			tx.Rollback()
			return err
		}

		// validate role ids exist to avoid FK errors
		for _, rid := range roleIDs {
			var one int
			err := tx.QueryRowContext(ctx, `SELECT 1 FROM roles WHERE id = $1`, rid).Scan(&one)
			if err != nil {
				tx.Rollback()
				if err == sql.ErrNoRows {
					return errors.New("role not found: " + rid)
				}
				return err
			}
		}

		// insert new roles
		for _, rid := range roleIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1,$2,$3)`, user.ID, rid, time.Now()); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
