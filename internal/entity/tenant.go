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
