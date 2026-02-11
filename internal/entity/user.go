package entity

import (
	"time"
)

type User struct {
	ID           string    `db:"id"`
	TenantID     string    `db:"tenant_id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Status       string    `db:"status"` // ACTIVE, INACTIVE, SUSPENDED
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
