package entity

import (
	"time"
)

type RolePermission struct {
	RoleID       string    `db:"role_id"`
	PermissionID string    `db:"permission_id"`
	CreatedAt    time.Time `db:"created_at"`
}
