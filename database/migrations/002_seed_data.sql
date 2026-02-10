-- Sample data for testing
-- This file contains INSERT statements for testing the API

-- Create a test tenant
INSERT INTO tenants (id, name, status) 
VALUES ('system-tenant', 'System', 'ACTIVE'), ('tenant-test-1', 'SMA Negeri 1 Testing', 'ACTIVE');

-- Create test users
-- Password: password (hashed with bcrypt)
INSERT INTO users (id, tenant_id, email, password_hash, status)
VALUES 
('user-admin-1', 'tenant-test-1', 'admin@test.com', '$2a$12$FDSYPHx7vVauhqtdsXHDQuSd95L2yBvgEfTdWkBKeeD0DOFgbC/kS', 'ACTIVE'),
('user-bendahara-1', 'tenant-test-1', 'bendahara@test.com', '$2a$12$FDSYPHx7vVauhqtdsXHDQuSd95L2yBvgEfTdWkBKeeD0DOFgbC/kS', 'ACTIVE'),
('user-operator-1', 'tenant-test-1', 'operator@test.com', '$2a$12$FDSYPHx7vVauhqtdsXHDQuSd95L2yBvgEfTdWkBKeeD0DOFgbC/kS', 'ACTIVE'),
('user-super-admin', 'system-tenant', 'superadmin@system.com', '$2a$12$FDSYPHx7vVauhqtdsXHDQuSd95L2yBvgEfTdWkBKeeD0DOFgbC/kS', 'ACTIVE');

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id)
VALUES 
('user-admin-1', 'admin-sekolah'),
('user-bendahara-1', 'bendahara'),
('user-operator-1', 'operator'),
('user-super-admin', 'super-admin');

-- Assign permissions to roles
INSERT INTO role_permissions (role_id, permission_id)
VALUES 
('admin-sekolah', 'perm-1'),
('admin-sekolah', 'perm-2'),
('admin-sekolah', 'perm-3'),
('admin-sekolah', 'perm-4'),
('bendahara', 'perm-1'),
('bendahara', 'perm-2'),
('bendahara', 'perm-5'),
('bendahara', 'perm-7'),
('operator', 'perm-2'),
('operator', 'perm-9');

-- Note: To create actual test users, you need to:
-- 1. Hash the password using bcrypt
-- 2. Use the /users endpoint or insert directly with hashed password

-- Example of creating a user with the API:
-- curl -X POST http://localhost:8001/users \
--   -H "Content-Type: application/json" \
--   -H "Authorization: Bearer YOUR_TOKEN" \
--   -d '{
--     "email": "test@example.com",
--     "password": "password123",
--     "tenant_id": "tenant-test-1",
--     "role_ids": ["admin-sekolah"]
--   }'
