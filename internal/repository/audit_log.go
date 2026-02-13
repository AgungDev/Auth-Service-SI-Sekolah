package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"encoding/json"
)

// AuditLogRepository defines audit log repository interface
type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, log *entity.AuditLog) error
	GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error)
}

// AuditLogRepositoryImpl implements AuditLogRepository
type AuditLogRepositoryImpl struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sql.DB) AuditLogRepository {
	return &AuditLogRepositoryImpl{db: db}
}

// CreateAuditLog creates an audit log
func (r *AuditLogRepositoryImpl) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor_id, tenant_id, action, target, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var metadata interface{}
	if log.Metadata != nil {
		metadata = log.Metadata
	} else {
		// Use empty JSON object for null metadata
		metadata = json.RawMessage(`{}`)
	}

	_, err := r.db.ExecContext(ctx, query,
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

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
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
