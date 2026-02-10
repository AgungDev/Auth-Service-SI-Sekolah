# Getting Started

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL 15 (if running locally without Docker)

## Quick Start with Docker Compose

1. **Clone/Setup the project**
   ```bash
   cd Auth-Service-SI-Sekolah
   ```

2. **Build and run with Docker Compose**
   ```bash
   docker compose up --build
   ```

   The service will be available at: `http://localhost:8001`

3. **Check health**
   ```bash
   curl http://localhost:8001/health
   ```

## Local Development (without Docker)

### 1. Setup Environment Variables
```bash
cp .env.example .env
```

Edit `.env` with your local PostgreSQL credentials:
```env
DB_HOST=localhost
DB_PORT=5433
DB_NAME=auth_db
DB_USER=postgres
DB_PASS=postgres
PORT=8000
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRY=1800
REFRESH_TOKEN_EXPIRY=604800
ENVIRONMENT=development
```

### 2. Setup PostgreSQL Database

Start a PostgreSQL instance:
```bash
docker run --name auth-db \
  -e POSTGRES_DB=auth_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:15
```

### 3. Run Database Migrations
```bash
psql -h localhost -U postgres -d auth_db < database/migrations/001_init.sql
```

### 4. Install Dependencies
```bash
go mod download
```

### 5. Run the Application
```bash
go run ./cmd/api
```

The service will start on `http://localhost:8000`

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
├── internal/
│   ├── entity/                      # Domain entities
│   │   └── entity.go
│   ├── handler/                     # HTTP handlers (controllers)
│   │   └── auth.go
│   ├── middleware/                  # HTTP middleware
│   │   └── middleware.go
│   ├── repository/                  # Data access layer
│   │   ├── repository.go            # Interfaces
│   │   └── postgres.go              # PostgreSQL implementation
│   └── usecase/                     # Business logic layer
│       ├── auth.go
│       └── jwt.go
├── pkg/
│   ├── config/                      # Configuration management
│   │   └── config.go
│   └── logger/                      # Logging utility
│       └── logger.go
├── database/
│   └── migrations/
│       └── 001_init.sql             # Database schema
├── .env                             # Environment variables (local)
├── .env.example                     # Environment variables template
├── docker-compose.yml               # Docker Compose configuration
├── Dockerfile                       # Docker image definition
├── go.mod                           # Go module definition
├── API.md                           # API documentation
├── SETUP.md                         # Setup instructions
└── README.md                        # Project overview
```

## Clean Architecture Layers

### 1. **Entity Layer** (`internal/entity/`)
- Pure domain objects
- No dependencies on other layers
- Contains business rules

### 2. **Repository Layer** (`internal/repository/`)
- Interfaces that define data access contracts
- Implementations for PostgreSQL
- Independent of business logic

### 3. **Usecase Layer** (`internal/usecase/`)
- Application business logic
- Orchestrates repositories
- Independent of frameworks

### 4. **Handler Layer** (`internal/handler/`)
- HTTP request/response handling
- Input validation
- Error mapping

### 5. **Middleware Layer** (`internal/middleware/`)
- Cross-cutting concerns (CORS, logging, etc.)
- Authentication/Authorization helpers

## Testing the API

### 1. Login
```bash
curl -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123",
    "tenant_id": "school-1"
  }'
```

### 2. Create Tenant (requires SUPER_ADMIN role)
```bash
curl -X POST http://localhost:8001/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "SMA Negeri 1"
  }'
```

### 3. Create User
```bash
curl -X POST http://localhost:8001/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "tenant_id": "tenant-id",
    "role_ids": ["admin-sekolah"]
  }'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | Database host |
| DB_PORT | 5433 | Database port |
| DB_NAME | auth_db | Database name |
| DB_USER | postgres | Database user |
| DB_PASS | postgres | Database password |
| PORT | 8000 | Server port |
| JWT_SECRET | your-secret-key | JWT signing key (CHANGE IN PRODUCTION) |
| JWT_EXPIRY | 1800 | Access token expiry in seconds (30 minutes) |
| REFRESH_TOKEN_EXPIRY | 604800 | Refresh token expiry in seconds (7 days) |
| ENVIRONMENT | development | Run mode (development/production) |

## Production Deployment

1. Generate a strong JWT_SECRET
2. Use a managed PostgreSQL service (AWS RDS, Google Cloud SQL, etc.)
3. Set ENVIRONMENT=production
4. Use HTTPS for all communications
5. Implement proper logging and monitoring
6. Set up CI/CD pipeline
7. Use secrets management (AWS Secrets Manager, HashiCorp Vault, etc.)

## Troubleshooting

### Database Connection Errors
- Verify PostgreSQL is running and accessible
- Check DB_HOST, DB_PORT, DB_USER, DB_PASS in .env
- Ensure database exists: `createdb auth_db`

### JWT Token Invalid
- Check JWT_SECRET matches between requests
- Verify token hasn't expired
- Ensure Authorization header format: `Bearer <token>`

### Permission Denied Errors
- Verify user has required role in the tenant
- Check token contains the correct tenant_id
- Ensure user status is ACTIVE

## Support & Documentation

- See [API.md](API.md) for detailed API documentation
- See [README.md](README.md) for project overview
