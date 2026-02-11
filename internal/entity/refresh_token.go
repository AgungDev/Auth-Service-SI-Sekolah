package entity

import (
	"time"
)

type RefreshToken struct {
	Token     string     `db:"token"`
	UserID    string     `db:"user_id"`
	TenantID  string     `db:"tenant_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}
