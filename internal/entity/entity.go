package entity

import (
	"time"
)

// Tenant represents a school in the system
type Tenant struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Status    string    `db:"status"` // ACTIVE, SUSPENDED, ARCHIVED
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID           string    `db:"id"`
	TenantID     string    `db:"tenant_id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Status       string    `db:"status"` // ACTIVE, INACTIVE, SUSPENDED
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// Role represents a role in the system
type Role struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Scope     string    `db:"scope"` // SYSTEM, TENANT
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          string    `db:"id"`
	Code        string    `db:"code"` // e.g., transaction.create
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	RoleID       string    `db:"role_id"`
	PermissionID string    `db:"permission_id"`
	CreatedAt    time.Time `db:"created_at"`
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	UserID    string    `db:"user_id"`
	RoleID    string    `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token     string     `db:"token"`
	UserID    string     `db:"user_id"`
	TenantID  string     `db:"tenant_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

// AuditLog represents audit logging
type AuditLog struct {
	ID        string                 `db:"id"`
	ActorID   string                 `db:"actor_id"`
	TenantID  string                 `db:"tenant_id"`
	Action    string                 `db:"action"` // USER_CREATED, ROLE_ASSIGNED, etc
	Target    string                 `db:"target"`
	Metadata  map[string]interface{} `db:"metadata"`
	CreatedAt time.Time              `db:"created_at"`
}

// AccessTokenClaims represents JWT claims for access token
type AccessTokenClaims struct {
	Sub          string   `json:"sub"`           // user-id
	TenantID     string   `json:"tenant_id"`     // school-id
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	TenantStatus string   `json:"tenant_status"`
	ExpiresAt    int64    `json:"exp"`
}
