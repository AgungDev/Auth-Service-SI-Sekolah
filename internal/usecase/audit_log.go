package usecase

import (
	"context"

	"auth-service/internal/entity"
	"auth-service/internal/repository"
)

type AuditLogUseCaseInterface interface {
	GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error)
}

type auditLogUseCase struct {
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogUseCase(
	auditLogRepo repository.AuditLogRepository,
) AuditLogUseCaseInterface {
	return &auditLogUseCase{
		auditLogRepo: auditLogRepo,
	}
}

// GetAuditLogs retrieves audit logs for a tenant with pagination
func (u *auditLogUseCase) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error) {
	// Default pagination
	if limit == 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return u.auditLogRepo.GetAuditLogs(ctx, tenantID, limit, offset)
}
