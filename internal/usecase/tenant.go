package usecase

import (
	"context"
	"time"

	"auth-service/internal/entity"
	"auth-service/internal/entity/dto"
	"auth-service/internal/repository"

	"github.com/google/uuid"
)

type TenantUseCaseInterface interface {
	CreateTenant(ctx context.Context, req dto.CreateTenantRequestBody) (*entity.Tenant, error)
}

// TenantUseCase handles tenant-related logic
type tenantUseCase struct {
	tenantRepo repository.TenantRepository
	auditLogRepo repository.AuditLogRepository
}

// NewTenantUseCase creates a new tenant usecase
func NewTenantUseCase(
	tenantRepo repository.TenantRepository,
	auditLogRepo repository.AuditLogRepository,
) TenantUseCaseInterface {
	return &tenantUseCase{
		tenantRepo: tenantRepo,
		auditLogRepo: auditLogRepo,
	}
}	

// CreateTenant creates a new tenant
func (u *tenantUseCase) CreateTenant(ctx context.Context, req dto.CreateTenantRequestBody) (*entity.Tenant, error) {
	tenant := &entity.Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdTenant, err := u.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}

	// Audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:       uuid.New().String(),
		ActorID:  "system",
		TenantID: createdTenant.ID,
		Action:   "TENANT_CREATED",
		Target:   createdTenant.ID,
		Metadata: map[string]interface{}{
			"tenant_name": createdTenant.Name,
		},
		CreatedAt: time.Now(),
	})

	return createdTenant, nil
}