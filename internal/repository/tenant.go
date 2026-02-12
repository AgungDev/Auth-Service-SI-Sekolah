package repository

import (
	"auth-service/internal/entity"
	"context"
	"database/sql"
	"errors"
	"time"
)

type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error)
	GetTenantByID(ctx context.Context, id string) (*entity.Tenant, error)
	GetAllTenants(ctx context.Context) ([]*entity.Tenant, error)
	UpdateTenantStatus(ctx context.Context, id, status string) error
}

type TenantRepositoryImpl struct {
	db *sql.DB
}

func NewTenantRepository(db *sql.DB) TenantRepository {
	return &TenantRepositoryImpl{db: db}
}

func (r *TenantRepositoryImpl) CreateTenant(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	query := `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, status, created_at, updated_at
	`

	row := r.db.QueryRowContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Status, tenant.CreatedAt, tenant.UpdatedAt)

	var t entity.Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// GetTenantByID gets a tenant by ID
func (r *TenantRepositoryImpl) GetTenantByID(ctx context.Context, id string) (*entity.Tenant, error) {
	query := `
		SELECT id, name, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var tenant entity.Tenant
	err := row.Scan(&tenant.ID, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	return &tenant, nil
}

// GetAllTenants gets all tenants
func (r *TenantRepositoryImpl) GetAllTenants(ctx context.Context) ([]*entity.Tenant, error) {
	query := `
		SELECT id, name, status, created_at, updated_at
		FROM tenants
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		var tenant entity.Tenant
		err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// UpdateTenantStatus updates tenant status
func (r *TenantRepositoryImpl) UpdateTenantStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE tenants
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("tenant not found")
	}

	return nil
}
