package entity

import (
	"time"
)

type UserRole struct {
	UserID    string    `db:"user_id"`
	RoleID    string    `db:"role_id"`
	CreatedAt time.Time `db:"created_at"`
}
