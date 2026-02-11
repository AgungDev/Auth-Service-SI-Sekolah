package entity

import (
	"time"
)

type AuditLog struct {
	ID        string                 `db:"id"`
	ActorID   string                 `db:"actor_id"`
	TenantID  string                 `db:"tenant_id"`
	Action    string                 `db:"action"` // USER_CREATED, ROLE_ASSIGNED, etc
	Target    string                 `db:"target"`
	Metadata  map[string]interface{} `db:"metadata"`
	CreatedAt time.Time              `db:"created_at"`
}
