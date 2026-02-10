# Troubleshooting Guide

## Common Issues and Solutions

### 1. Database Connection Issues

#### Error: "Failed to connect to database"

**Causes:**
- PostgreSQL service not running
- Wrong host/port/credentials
- Database doesn't exist

**Solutions:**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Verify credentials in .env
cat .env

# Test connection
psql -h localhost -p 5433 -U postgres -d auth_db

# Create database if not exists
createdb -h localhost -p 5433 -U postgres auth_db
```

#### Error: "pq: password authentication failed"

**Cause:** Wrong database password

**Solution:**
```bash
# Update .env with correct password
DB_PASS=your_actual_password

# Test connection
psql -h localhost -p 5433 -U postgres -W -d auth_db
```

### 2. JWT Token Issues

#### Error: "Invalid token" when using Access Token

**Causes:**
- Token has expired
- JWT_SECRET doesn't match between services
- Token format incorrect

**Solutions:**
```bash
# Verify token format (should have 3 parts separated by dots)
echo $TOKEN | tr '.' '\n'

# Decode and check expiry
# Visit https://jwt.io and paste your token

# Check JWT_SECRET matches
echo $JWT_SECRET
```

#### Error: "Token expired"

**Cause:** Access token lifetime exceeded (default 30 minutes)

**Solution:**
```bash
# Use refresh token to get new access token
curl -X POST http://localhost:8001/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "your_refresh_token"}'
```

### 3. Authentication Issues

#### Error: "Invalid credentials" on login

**Causes:**
- Wrong email/password
- User doesn't exist in tenant
- User account inactive

**Solutions:**
```bash
# Verify user exists
psql -h localhost -d auth_db -c "SELECT * FROM users WHERE email = 'test@example.com';"

# Check user status
psql -h localhost -d auth_db -c "SELECT email, status FROM users;"

# Verify tenant ID is correct
psql -h localhost -d auth_db -c "SELECT * FROM tenants;"
```

#### Error: "Unauthorized" - Missing Authorization header

**Cause:** Request missing Bearer token

**Solution:**
```bash
# Correct format
curl -X POST http://localhost:8001/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Wrong format - WILL FAIL
curl -X POST http://localhost:8001/users \
  -H "Authorization: YOUR_ACCESS_TOKEN"
```

### 4. Permission Issues

#### Error: "Insufficient permissions" / "Forbidden"

**Causes:**
- User doesn't have required role
- Wrong tenant ID
- Role not assigned to user

**Solutions:**
```bash
# Check user roles
psql -h localhost -d auth_db -c "
SELECT u.email, r.name 
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'user@example.com';"

# Assign role to user
psql -h localhost -d auth_db -c "
INSERT INTO user_roles (user_id, role_id) 
VALUES ('user-id', 'admin-sekolah');"

# Verify token contains correct roles
# Decode JWT at https://jwt.io and check 'roles' claim
```

### 5. API Request Issues

#### Error: "Invalid request body"

**Cause:** Malformed JSON or missing required fields

**Solution:**
```bash
# Validate JSON syntax
echo '{"email": "test@example.com"}' | jq .

# Check required fields for each endpoint
# See API.md for complete documentation

# Example with all required fields for login
curl -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "tenant_id": "tenant-1"
  }'
```

#### Error: "Method not allowed"

**Cause:** Wrong HTTP method

**Solution:**
```bash
# Check documentation for correct method
# POST /login (not GET)
# POST /refresh (not GET)
# GET /health (not POST)

# Verify method
curl -X POST http://localhost:8001/health  # WRONG
curl -X GET http://localhost:8001/health   # CORRECT
```

### 6. Database Schema Issues

#### Error: "relation \"users\" does not exist"

**Cause:** Migrations not run

**Solution:**
```bash
# Run migrations
psql -h localhost -U postgres -d auth_db < database/migrations/001_init.sql

# Verify tables created
psql -h localhost -d auth_db -c "\dt"

# Check specific table
psql -h localhost -d auth_db -c "\d users"
```

#### Error: "duplicate key value violates unique constraint"

**Cause:** Trying to create user with duplicate email in same tenant

**Solution:**
```bash
# Check existing users in tenant
psql -h localhost -d auth_db -c "
SELECT email, tenant_id FROM users 
WHERE tenant_id = 'your-tenant-id';"

# Use different email or verify user creation logic
```

### 7. Docker Issues

#### Error: "Cannot connect to Docker daemon"

**Cause:** Docker not running

**Solution:**
```bash
# Start Docker Desktop or Docker service

# Verify Docker is running
docker ps

# Check logs
docker compose logs
```

#### Error: "Port 5433 already in use"

**Cause:** Another service using the port

**Solution:**
```bash
# Check what's using port 5433
lsof -i :5433

# Stop conflicting service or use different port
# Edit docker-compose.yml:
# ports:
#   - "5434:5432"  # Use 5434 instead

# Update .env
DB_PORT=5434
```

#### Error: Service exits immediately after starting

**Cause:** Database not ready when service starts

**Solution:**
```bash
# Check and increase wait time
# Docker Compose includes depends_on but doesn't wait for DB readiness

# Option 1: Wait a bit before restarting
docker compose down
sleep 3
docker compose up

# Option 2: Check logs
docker compose logs auth-service

# Option 3: Verify database is running
docker compose ps
```

### 8. Performance Issues

#### Service running slowly

**Causes:**
- Inefficient database queries
- Missing indexes
- Too many connections

**Solutions:**
```bash
# Check database connections
psql -h localhost -d auth_db -c "SELECT count(*) FROM pg_stat_activity;"

# Check slow queries
# Enable query logging in PostgreSQL

# Check connection pool settings in main.go
// db.SetMaxOpenConns(25)
// db.SetMaxIdleConns(5)
```

### 9. Logging and Debugging

#### Enable debug logging

Edit `cmd/api/main.go`:
```go
if cfg.Environment == "development" {
    log.Debug("Processing login for user: %s", email)
}
```

#### Check service logs

```bash
# Docker Compose logs
docker compose logs auth-service -f

# Application logs (when running locally)
go run ./cmd/api 2>&1 | grep ERROR
```

#### Database query debugging

```bash
# Enable PostgreSQL logging
# Edit postgresql.conf:
log_statement = 'all'

# Check logs
tail -f /var/log/postgresql/postgresql.log
```

### 10. Testing Tools

#### API Testing with curl

```bash
# Save access token to variable
TOKEN=$(curl -s -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"123","tenant_id":"t1"}' \
  | jq -r '.access_token')

# Use in subsequent requests
curl -X POST http://localhost:8001/users \
  -H "Authorization: Bearer $TOKEN" \
  -d '{...}'
```

#### Database testing

```bash
# Connect to database
psql -h localhost -U postgres -d auth_db

# Common queries
\dt                          # List all tables
\d users                     # Describe users table
SELECT * FROM users;        # View all users
SELECT * FROM audit_logs;   # View audit logs
```

#### JWT debugging

```bash
# Decode JWT (requires jq)
TOKEN="your-token-here"
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq .

# Verify online
# Visit https://jwt.io and paste token
```

## Getting Help

If issues persist:

1. **Check logs:** `docker compose logs`
2. **Verify environment:** `cat .env`
3. **Test connectivity:** `psql -h localhost -d auth_db`
4. **Check API docs:** Read [API.md](API.md)
5. **Review code:** Check relevant source files
6. **Consult architecture:** See [ARCHITECTURE.md](ARCHITECTURE.md)

## Health Check Endpoint

Always accessible at `GET /health`:
```bash
curl http://localhost:8001/health
```

Should return:
```json
{
  "message": "Auth Service is running",
  "data": null
}
```

If this fails, service is not running or port is wrong.
