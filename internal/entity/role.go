package entity

import (
	"time"
)

type Role struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Scope     string    `db:"scope"` // SYSTEM, TENANT
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
