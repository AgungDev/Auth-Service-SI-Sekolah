# Fast Start Guide

## ⚡ 30-Minute Quick Start

### Prerequisites
- Docker & Docker Compose installed
- OR PostgreSQL + Go 1.21

### Option 1: With Docker (Recommended)

```bash
# 1. Navigate to project
cd Auth-Service-SI-Sekolah

# 2. Start services
docker compose up --build

# 3. Wait for "database migrations complete" message

# 4. Test health endpoint
curl http://localhost:8001/health

# Done! API is ready
```

### Option 2: Local Development

```bash
# 1. Setup environment
cp .env.example .env

# 2. Start PostgreSQL
docker run --name auth-db -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 -d postgres:15

# 3. Run migrations
sleep 5
psql -h localhost -U postgres -d auth_db < database/migrations/001_init.sql

# 4. Install dependencies & run
go mod download
go run ./cmd/api

# Service runs on http://localhost:8000
```

## 📝 Basic API Usage

### 1. Check Service Health
```bash
curl http://localhost:8001/health
```

**Response:**
```json
{
  "message": "Auth Service is running",
  "data": null
}
```

### 2. Create a Tenant (School)
First, you need a user with SUPER_ADMIN role. For testing, insert directly:

```bash
psql -h localhost -d auth_db << 'EOF'
-- Create tenant
INSERT INTO tenants (id, name, status) VALUES 
('school-1', 'SMA Negeri 1', 'ACTIVE');

-- Create user with SUPER_ADMIN role
INSERT INTO users (id, tenant_id, email, password_hash, status) VALUES 
('admin-1', 'school-1', 'admin@test.com', 
'$2a$10$N9qo8uLOickgx2zE2jE8PuS7Y8LvxNz6nEPqC6rK7ZC0.ZD9Rqnwe', 'ACTIVE');

INSERT INTO user_roles (user_id, role_id) VALUES 
('admin-1', 'super-admin');
EOF
```

### 3. Login
```bash
curl -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@test.com",
    "password": "password",
    "tenant_id": "school-1"
  }'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "opaque-token",
  "expires_in": 1800
}
```

### 4. Create Another School
```bash
TOKEN="your-access-token-from-login"

curl -X POST http://localhost:8001/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "SMA Negeri 2"
  }'
```

### 5. Add User to School
```bash
TOKEN="your-access-token"
SCHOOL_ID="from-previous-response"

curl -X POST http://localhost:8001/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "email": "operator@school.com",
    "password": "secure123",
    "tenant_id": "'$SCHOOL_ID'",
    "role_ids": ["operator"]
  }'
```

## 📚 Useful Commands

```bash
# View logs
docker compose logs -f auth-service

# Stop services
docker compose down

# Clean everything
docker compose down -v

# Connect to database
docker compose exec auth-db psql -U postgres -d auth_db

# View all users
docker compose exec auth-db psql -U postgres -d auth_db \
  -c "SELECT id, email, status FROM users;"

# View audit logs  
docker compose exec auth-db psql -U postgres -d auth_db \
  -c "SELECT * FROM audit_logs ORDER BY created_at DESC;"
```

## 🔐 Default Test Credentials

After running migrations with seed data:
- **Email:** admin@test.com
- **Password:** password (bcrypt hashed)
- **Tenant:** Check `tenants` table

## 📖 Full Documentation

- **[API.md](API.md)** - Complete API reference
- **[SETUP.md](SETUP.md)** - Detailed setup instructions
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues

## 🚀 Next Steps

1. Read [API.md](API.md) for all endpoints
2. Create test users with different roles
3. Test permission-based access
4. Integrate with other services
5. Deploy to production

## ⚠️ Important

- Change `JWT_SECRET` in production
- Don't commit `.env` to version control
- Use HTTPS in production
- Implement proper logging/monitoring
- Rotate refresh tokens regularly

## 💡 Tips

- Use `jq` to format JSON: `curl ... | jq .`
- Store token in variable: `TOKEN=$(curl ... | jq -r '.access_token')`
- Test endpoints with provided scripts: `./scripts/test-api.sh`

## 🆘 Issues?

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common problems and solutions.
