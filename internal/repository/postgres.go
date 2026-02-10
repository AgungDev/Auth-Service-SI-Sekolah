package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "time"

    "auth-service/internal/entity"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	conn *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(db *sql.DB) *PostgresDB {
	return &PostgresDB{
		conn: db,
	}
}

// TenantRepositoryImpl implements TenantRepository
type TenantRepositoryImpl struct {
	db *PostgresDB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *PostgresDB) TenantRepository {
	return &TenantRepositoryImpl{db: db}
}

// CreateTenant creates a new tenant
func (r *TenantRepositoryImpl) CreateTenant(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	query := `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, status, created_at, updated_at
	`

	row := r.db.conn.QueryRowContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Status, tenant.CreatedAt, tenant.UpdatedAt)

	var t entity.Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// GetTenantByID gets a tenant by ID
func (r *TenantRepositoryImpl) GetTenantByID(ctx context.Context, id string) (*entity.Tenant, error) {
	query := `
		SELECT id, name, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	row := r.db.conn.QueryRowContext(ctx, query, id)
	var tenant entity.Tenant
	err := row.Scan(&tenant.ID, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	return &tenant, nil
}

// GetAllTenants gets all tenants
func (r *TenantRepositoryImpl) GetAllTenants(ctx context.Context) ([]*entity.Tenant, error) {
	query := `
		SELECT id, name, status, created_at, updated_at
		FROM tenants
	`

	rows, err := r.db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		var tenant entity.Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// UpdateTenantStatus updates tenant status
func (r *TenantRepositoryImpl) UpdateTenantStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE tenants
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.conn.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("tenant not found")
	}

	return nil
}

// UserRepositoryImpl implements UserRepository
type UserRepositoryImpl struct {
	db *PostgresDB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *PostgresDB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, tenant_id, email, password_hash, status, created_at, updated_at
	`

	row := r.db.conn.QueryRowContext(ctx, query,
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

	row := r.db.conn.QueryRowContext(ctx, query, id)
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

	row := r.db.conn.QueryRowContext(ctx, query, email, tenantID)
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

	rows, err := r.db.conn.QueryContext(ctx, query, tenantID)
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

	result, err := r.db.conn.ExecContext(ctx, query, status, time.Now(), id)
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

// RoleRepositoryImpl implements RoleRepository
type RoleRepositoryImpl struct {
	db *PostgresDB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *PostgresDB) RoleRepository {
	return &RoleRepositoryImpl{db: db}
}

// CreateRole creates a new role
func (r *RoleRepositoryImpl) CreateRole(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	query := `
		INSERT INTO roles (id, name, scope, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, scope, created_at, updated_at
	`

	row := r.db.conn.QueryRowContext(ctx, query,
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

	row := r.db.conn.QueryRowContext(ctx, query, id)
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

	rows, err := r.db.conn.QueryContext(ctx, query)
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

	rows, err := r.db.conn.QueryContext(ctx, query, userID)
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

// RefreshTokenRepositoryImpl implements RefreshTokenRepository
type RefreshTokenRepositoryImpl struct {
	db *PostgresDB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *PostgresDB) RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{db: db}
}

// SaveRefreshToken saves a refresh token
func (r *RefreshTokenRepositoryImpl) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, tenant_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.conn.ExecContext(ctx, query,
		token.Token, token.UserID, token.TenantID, token.ExpiresAt, token.CreatedAt)
	return err
}

// GetRefreshToken gets a refresh token
func (r *RefreshTokenRepositoryImpl) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := `
		SELECT token, user_id, tenant_id, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	row := r.db.conn.QueryRowContext(ctx, query, token)
	var rt entity.RefreshToken
	err := row.Scan(&rt.Token, &rt.UserID, &rt.TenantID, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return &rt, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *RefreshTokenRepositoryImpl) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE token = $2
	`

	result, err := r.db.conn.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

// RevokeAllRefreshTokensByUserID revokes all refresh tokens for a user
func (r *RefreshTokenRepositoryImpl) RevokeAllRefreshTokensByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL
	`

	_, err := r.db.conn.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// AuditLogRepositoryImpl implements AuditLogRepository
type AuditLogRepositoryImpl struct {
	db *PostgresDB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *PostgresDB) AuditLogRepository {
	return &AuditLogRepositoryImpl{db: db}
}

// CreateAuditLog creates an audit log
func (r *AuditLogRepositoryImpl) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor_id, tenant_id, action, target, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var metadata []byte
	if log.Metadata != nil {
		var err error
		metadata, err = json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
	}

	_, err := r.db.conn.ExecContext(ctx, query,
		log.ID, log.ActorID, log.TenantID, log.Action, log.Target, metadata, log.CreatedAt)
	return err
}

// GetAuditLogs gets audit logs
func (r *AuditLogRepositoryImpl) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, actor_id, tenant_id, action, target, metadata, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.conn.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		var log entity.AuditLog
		var metadata []byte
		err := rows.Scan(&log.ID, &log.ActorID, &log.TenantID, &log.Action, &log.Target, &metadata, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			err = json.Unmarshal(metadata, &log.Metadata)
			if err != nil {
				return nil, err
			}
		}

		logs = append(logs, &log)
	}

	return logs, nil
}

// UserRoleRepositoryImpl implements UserRoleRepository
type UserRoleRepositoryImpl struct {
	db *PostgresDB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *PostgresDB) UserRoleRepository {
	return &UserRoleRepositoryImpl{db: db}
}

// AssignRoleToUser assigns a role to a user
func (r *UserRoleRepositoryImpl) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.conn.ExecContext(ctx, query, userID, roleID, time.Now())
	return err
}

// RemoveRoleFromUser removes a role from a user
func (r *UserRoleRepositoryImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	_, err := r.db.conn.ExecContext(ctx, query, userID, roleID)
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

	rows, err := r.db.conn.QueryContext(ctx, query, userID)
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

// RolePermissionRepositoryImpl implements RolePermissionRepository
type RolePermissionRepositoryImpl struct {
	db *PostgresDB
}

// NewRolePermissionRepository creates a new role permission repository
func NewRolePermissionRepository(db *PostgresDB) RolePermissionRepository {
	return &RolePermissionRepositoryImpl{db: db}
}

// AssignPermissionToRole assigns a permission to a role
func (r *RolePermissionRepositoryImpl) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.conn.ExecContext(ctx, query, roleID, permissionID, time.Now())
	return err
}

// RemovePermissionFromRole removes a permission from a role
func (r *RolePermissionRepositoryImpl) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	_, err := r.db.conn.ExecContext(ctx, query, roleID, permissionID)
	return err
}

// PermissionRepositoryImpl implements PermissionRepository
type PermissionRepositoryImpl struct {
	db *PostgresDB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *PostgresDB) PermissionRepository {
	return &PermissionRepositoryImpl{db: db}
}

// CreatePermission creates a new permission
func (r *PermissionRepositoryImpl) CreatePermission(ctx context.Context, permission *entity.Permission) (*entity.Permission, error) {
	query := `
		INSERT INTO permissions (id, code, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, description, created_at, updated_at
	`

	row := r.db.conn.QueryRowContext(ctx, query,
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

	row := r.db.conn.QueryRowContext(ctx, query, id)
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

	rows, err := r.db.conn.QueryContext(ctx, query)
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

	rows, err := r.db.conn.QueryContext(ctx, query, roleID)
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
