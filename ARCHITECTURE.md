# Auth Service Architecture

## System Overview

```
┌─────────────────┐
│  API Gateway    │
│  or Service     │
└────────┬────────┘
         │
         │ HTTP/REST
         ▼
┌─────────────────────────────────────┐
│   Auth Service (This Service)       │
├─────────────────────────────────────┤
│  Handlers (HTTP)                    │
│  ├─ POST /login                     │
│  ├─ POST /refresh                   │
│  ├─ POST /tenants                   │
│  └─ POST /users                     │
├─────────────────────────────────────┤
│  Middleware                         │
│  ├─ CORS                            │
│  ├─ Logging                         │
│  └─ Content-Type Validation         │
├─────────────────────────────────────┤
│  Usecases (Business Logic)          │
│  ├─ AuthUseCase (Login, Refresh)    │
│  └─ RoleUseCase (Manage Roles)      │
├─────────────────────────────────────┤
│  Repositories (Data Access)         │
│  ├─ UserRepository                  │
│  ├─ TenantRepository                │
│  ├─ RoleRepository                  │
│  ├─ PermissionRepository            │
│  ├─ RefreshTokenRepository          │
│  └─ AuditLogRepository              │
├─────────────────────────────────────┤
│  Database (PostgreSQL)              │
│  ├─ tenants                         │
│  ├─ users                           │
│  ├─ roles                           │
│  ├─ permissions                     │
│  ├─ user_roles                      │
│  ├─ role_permissions                │
│  ├─ refresh_tokens                  │
│  └─ audit_logs                      │
└─────────────────────────────────────┘
```

## Clean Architecture Layers

### 1. Entity Layer
**Location:** `internal/entity/`

**Responsibility:**
- Define domain models
- Contain business rules
- No dependencies on frameworks or databases

**Key Entities:**
- `User` - Represents a user in the system
- `Tenant` - Represents a school
- `Role` - Represents user roles
- `Permission` - Represents permissions
- `RefreshToken` - Token management
- `AuditLog` - Activity tracking

### 2. Repository Layer
**Location:** `internal/repository/`

**Responsibility:**
- Define interfaces for data access
- Implement database operations
- Abstract database details from business logic

**Interfaces:**
- `UserRepository` - User data access
- `TenantRepository` - Tenant data access
- `RoleRepository` - Role data access
- `PermissionRepository` - Permission data access
- `RefreshTokenRepository` - Token management
- `AuditLogRepository` - Audit log management

**Implementations:**
- `PostgresDB` - PostgreSQL wrapper
- `*RepositoryImpl` - Concrete implementations

### 3. Usecase Layer
**Location:** `internal/usecase/`

**Responsibility:**
- Implement application business logic
- Orchestrate repositories
- Independent of frameworks

**Usecases:**
- `AuthUseCase` - Login, refresh token, user/tenant creation
- `JWTService` - JWT token generation and validation
- `RoleUseCase` - Role management

### 4. Handler Layer
**Location:** `internal/handler/`

**Responsibility:**
- Handle HTTP requests/responses
- Parse input, validate data
- Map domain errors to HTTP status codes

**Handlers:**
- `AuthHandler` - Auth endpoints
- Request/Response DTOs

### 5. Middleware Layer
**Location:** `internal/middleware/`

**Responsibility:**
- Cross-cutting concerns
- Authentication/Authorization helpers
- Request/Response processing

**Middleware:**
- `CORS` - CORS header handling
- `Logger` - Request logging
- `ContentTypeJSON` - JSON content type

## Dependency Injection Flow

```
main.go
├─ Create database connection
├─ Instantiate repositories
│  ├─ Create PostgresDB instance
│  └─ Create repository implementations
├─ Instantiate JWT service
├─ Instantiate usecases
│  ├─ Create AuthUseCase with repositories
│  └─ Create RoleUseCase with repositories
├─ Instantiate handlers
│  └─ Create AuthHandler with usecases
└─ Register routes and middleware
```

## Data Flow Examples

### Login Flow
```
1. Client sends POST /login
   ├─ Email, Password, TenantID
   
2. AuthHandler.Login()
   ├─ Parse JSON request
   ├─ Validate input
   └─ Call AuthUseCase.Login()
   
3. AuthUseCase.Login()
   ├─ Query user via UserRepository
   ├─ Compare password hashes
   ├─ Get tenant status via TenantRepository
   ├─ Get user roles via RoleRepository
   ├─ Generate JWT via JWTService
   ├─ Generate refresh token
   ├─ Save refresh token via RefreshTokenRepository
   ├─ Create audit log via AuditLogRepository
   └─ Return access & refresh tokens
   
4. AuthHandler formats response
   └─ Return JSON with tokens
```

### Token Refresh Flow
```
1. Client sends POST /refresh
   ├─ RefreshToken
   
2. AuthHandler.RefreshToken()
   ├─ Parse JSON request
   └─ Call AuthUseCase.RefreshToken()
   
3. AuthUseCase.RefreshToken()
   ├─ Get refresh token via RefreshTokenRepository
   ├─ Validate token (check expiry, revocation)
   ├─ Get user via UserRepository
   ├─ Get tenant via TenantRepository
   ├─ Get user roles via RoleRepository
   ├─ Generate new access token
   ├─ Generate new refresh token
   ├─ Revoke old token via RefreshTokenRepository
   ├─ Save new token via RefreshTokenRepository
   └─ Return new tokens
   
4. AuthHandler formats response
   └─ Return JSON with new tokens
```

## Database Schema

### Users & Tenants
```
tenants
├─ id (PK)
├─ name
├─ status
└─ timestamps

users
├─ id (PK)
├─ tenant_id (FK)
├─ email
├─ password_hash
├─ status
└─ timestamps
```

### Roles & Permissions
```
roles
├─ id (PK)
├─ name
├─ scope (SYSTEM/TENANT)
└─ timestamps

permissions
├─ id (PK)
├─ code (e.g., "transaction.create")
├─ description
└─ timestamps

role_permissions (Junction)
├─ role_id (FK)
├─ permission_id (FK)
└─ timestamp

user_roles (Junction)
├─ user_id (FK)
├─ role_id (FK)
└─ timestamp
```

### Token Management
```
refresh_tokens
├─ token (PK)
├─ user_id (FK)
├─ tenant_id (FK)
├─ expires_at
├─ revoked_at
└─ created_at
```

### Audit
```
audit_logs
├─ id (PK)
├─ actor_id
├─ tenant_id (FK)
├─ action
├─ target
├─ metadata (JSONB)
└─ created_at
```

## Security Considerations

### Authentication
- JWT for stateless authentication between services
- Refresh tokens stored in database (can be revoked)
- Bcrypt for password hashing

### Authorization
- Role-based access control (RBAC)
- Tenant isolation (users can only access their tenant)
- System-level roles for platform administration

### Audit Trail
- All sensitive operations logged
- Metadata stored as JSON for flexibility
- Immutable audit log

## Performance Optimizations

### Database Indexes
- `users.tenant_id` - Filter by tenant
- `users.email` - User lookup
- `user_roles.user_id` - User role queries
- `refresh_tokens.user_id` - Token lookups
- `refresh_tokens.expires_at` - Expired token cleanup
- `audit_logs.tenant_id` - Audit queries

### Connection Pooling
- Max 25 open connections
- 5 idle connections
- 5-minute connection lifetime

### Caching Opportunities
- Roles (relatively static)
- Permissions (relatively static)
- May implement Redis for future scaling

## Testing Strategy

### Unit Tests
- Test business logic in usecases
- Mock repositories
- Test edge cases

### Integration Tests
- Test full flow with real database
- Use test database container
- Verify data persistence

### API Tests
- Test HTTP handlers
- Verify request/response formats
- Check error handling

## Scalability Considerations

### Horizontal Scaling
- Service is stateless (except refresh tokens)
- Load balance across multiple instances
- Shared PostgreSQL database

### Vertical Scaling
- Increase database connection pool
- Optimize queries
- Add caching layer

### Future Enhancements
- Redis for token caching/blacklisting
- Message queue for async audit logging
- Service discovery
- Distributed tracing
- Rate limiting
