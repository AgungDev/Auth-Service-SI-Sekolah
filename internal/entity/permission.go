package entity

import (
	"time"
)

type Permission struct {
	ID          string    `db:"id"`
	Code        string    `db:"code"` // e.g., transaction.create
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
