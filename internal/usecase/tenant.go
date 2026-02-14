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
	SuspendTenant(ctx context.Context, id string, actorID string) error
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
		Address:   req.Address,
		Status:    func() string { if req.Status != "" { return req.Status }; return "ACTIVE" }(),
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

// SuspendTenant sets the tenant status to SUSPENDED and records an audit log
func (u *tenantUseCase) SuspendTenant(ctx context.Context, id string, actorID string) error {
	// check tenant exists
	t, err := u.tenantRepo.GetTenantByID(ctx, id)
	if err != nil {
		return err
	}

	if t.Status == "SUSPENDED" {
		return nil
	}

	if err := u.tenantRepo.UpdateTenantStatus(ctx, id, "SUSPENDED"); err != nil {
		return err
	}

	// create audit log
	u.auditLogRepo.CreateAuditLog(ctx, &entity.AuditLog{
		ID:        uuid.New().String(),
		ActorID:   actorID,
		TenantID:  id,
		Action:    "TENANT_SUSPENDED",
		Target:    id,
		Metadata:  map[string]interface{}{"previous_status": t.Status},
		CreatedAt: time.Now(),
	})

	return nil
}