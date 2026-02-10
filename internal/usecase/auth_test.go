package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth-service/internal/entity"
)

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*entity.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	if _, exists := m.users[user.ID]; exists {
		return nil, errors.New("user already exists")
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email, tenantID string) (*entity.User, error) {
	for _, user := range m.users {
		if user.Email == email && user.TenantID == tenantID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, tenantID string) ([]*entity.User, error) {
	var users []*entity.User
	for _, user := range m.users {
		if user.TenantID == tenantID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *MockUserRepository) UpdateUserStatus(ctx context.Context, id, status string) error {
	user, exists := m.users[id]
	if !exists {
		return errors.New("user not found")
	}
	user.Status = status
	return nil
}

// MockTenantRepository for testing
type MockTenantRepository struct {
	tenants map[string]*entity.Tenant
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{
		tenants: make(map[string]*entity.Tenant),
	}
}

func (m *MockTenantRepository) CreateTenant(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error) {
	if _, exists := m.tenants[tenant.ID]; exists {
		return nil, errors.New("tenant already exists")
	}
	m.tenants[tenant.ID] = tenant
	return tenant, nil
}

func (m *MockTenantRepository) GetTenantByID(ctx context.Context, id string) (*entity.Tenant, error) {
	tenant, exists := m.tenants[id]
	if !exists {
		return nil, errors.New("tenant not found")
	}
	return tenant, nil
}

func (m *MockTenantRepository) GetAllTenants(ctx context.Context) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	for _, tenant := range m.tenants {
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

func (m *MockTenantRepository) UpdateTenantStatus(ctx context.Context, id, status string) error {
	tenant, exists := m.tenants[id]
	if !exists {
		return errors.New("tenant not found")
	}
	tenant.Status = status
	return nil
}

// Additional mock repositories would go here...

// MockJWTService for testing
type MockJWTService struct {
	secret string
}

func NewMockJWTService(secret string) *MockJWTService {
	return &MockJWTService{secret: secret}
}

func (m *MockJWTService) GenerateAccessToken(user *entity.User, roles []*entity.Role, tenant *entity.Tenant) (string, error) {
	return "mock-access-token", nil
}

func (m *MockJWTService) VerifyAccessToken(token string) (*entity.AccessTokenClaims, error) {
	if token == "invalid" {
		return nil, errors.New("invalid token")
	}
	return &entity.AccessTokenClaims{
		Sub:          "user-1",
		TenantID:     "tenant-1",
		Roles:        []string{"ADMIN_SEKOLAH"},
		TenantStatus: "ACTIVE",
		ExpiresAt:    time.Now().Add(30 * time.Minute).Unix(),
	}, nil
}

// Test: CreateTenant
func TestCreateTenant(t *testing.T) {
	// Arrange
	tenantRepo := NewMockTenantRepository()
	userRepo := NewMockUserRepository()
	roleRepo := &MockRoleRepository{}
	userRoleRepo := &MockUserRoleRepository{}
	refreshTokenRepo := &MockRefreshTokenRepository{}
	auditLogRepo := &MockAuditLogRepository{}
	jwtService := NewMockJWTService("secret")

	authUseCase := NewAuthUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		refreshTokenRepo,
		auditLogRepo,
		jwtService,
		604800,
	)

	ctx := context.Background()

	// Act
	tenant, err := authUseCase.CreateTenant(ctx, CreateTenantRequest{
		Name: "SMA Negeri 1",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant == nil {
		t.Fatal("expected tenant, got nil")
	}
	if tenant.Name != "SMA Negeri 1" {
		t.Errorf("expected name 'SMA Negeri 1', got '%s'", tenant.Name)
	}
	if tenant.Status != "ACTIVE" {
		t.Errorf("expected status 'ACTIVE', got '%s'", tenant.Status)
	}
}

// Test: CreateUser
func TestCreateUser(t *testing.T) {
	// Arrange
	tenantID := "tenant-1"
	tenantRepo := NewMockTenantRepository()
	userRepo := NewMockUserRepository()
	roleRepo := &MockRoleRepository{}
	userRoleRepo := &MockUserRoleRepository{}
	refreshTokenRepo := &MockRefreshTokenRepository{}
	auditLogRepo := &MockAuditLogRepository{}
	jwtService := NewMockJWTService("secret")

	// Add test tenant
	tenantRepo.CreateTenant(context.Background(), &entity.Tenant{
		ID:        tenantID,
		Name:      "Test School",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	authUseCase := NewAuthUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		refreshTokenRepo,
		auditLogRepo,
		jwtService,
		604800,
	)

	ctx := context.Background()

	// Act
	user, err := authUseCase.CreateUser(ctx, CreateUserRequest{
		Email:    "user@example.com",
		Password: "password123",
		TenantID: tenantID,
		RoleIDs:  []string{"admin-sekolah"},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", user.Email)
	}
	if user.Status != "ACTIVE" {
		t.Errorf("expected status 'ACTIVE', got '%s'", user.Status)
	}
}

// Placeholder mock repositories (would be fully implemented in real code)
type MockRoleRepository struct{}

func (m *MockRoleRepository) CreateRole(ctx context.Context, role *entity.Role) (*entity.Role, error) {
	return role, nil
}

func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id string) (*entity.Role, error) {
	return &entity.Role{ID: id, Name: "ADMIN"}, nil
}

func (m *MockRoleRepository) GetAllRoles(ctx context.Context) ([]*entity.Role, error) {
	return []*entity.Role{}, nil
}

func (m *MockRoleRepository) GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error) {
	return []*entity.Role{&entity.Role{ID: "role-1", Name: "ADMIN"}}, nil
}

type MockUserRoleRepository struct{}

func (m *MockUserRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *MockUserRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *MockUserRoleRepository) GetRolesByUserID(ctx context.Context, userID string) ([]*entity.Role, error) {
	return []*entity.Role{}, nil
}

type MockRefreshTokenRepository struct{}

func (m *MockRefreshTokenRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	return nil
}

func (m *MockRefreshTokenRepository) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	return &entity.RefreshToken{
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

func (m *MockRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return nil
}

func (m *MockRefreshTokenRepository) RevokeAllRefreshTokensByUserID(ctx context.Context, userID string) error {
	return nil
}

type MockAuditLogRepository struct{}

func (m *MockAuditLogRepository) CreateAuditLog(ctx context.Context, log *entity.AuditLog) error {
	return nil
}

func (m *MockAuditLogRepository) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*entity.AuditLog, error) {
	return []*entity.AuditLog{}, nil
}
