# Contributing to Auth Service

## Development Guidelines

### Code Style
- Follow Go conventions and idioms
- Use `gofmt` to format code
- Keep functions small and focused
- Write clear variable and function names

### Testing
- Write tests for business logic (usecase layer)
- Test edge cases and error scenarios
- Use table-driven tests for multiple conditions
- Aim for >80% code coverage

### Commit Messages
Use clear, descriptive commit messages:
```
feat: Add user login functionality
fix: Resolve JWT token validation issue
docs: Update API documentation
refactor: Simplify password hashing
test: Add tests for RefreshToken usecase
```

## Architecture Principles

### Clean Architecture Layers

1. **Entity Layer**
   - Contains domain models
   - No external dependencies
   - Pure business rules

2. **Repository Layer**
   - Data access abstraction
   - Interfaces define contracts
   - Multiple implementations possible

3. **Usecase Layer**
   - Application business logic
   - Orchestrates repositories
   - Framework independent

4. **Handler Layer**
   - HTTP request/response handling
   - Input validation
   - Maps domain errors to HTTP status codes

### Dependency Direction
```
Handler → Usecase → Repository → [Database]
         ↓
       Entity
```

## Adding New Features

### Example: Add Password Reset Feature

1. **Add to Entity** (`internal/entity/entity.go`)
   ```go
   type PasswordReset struct {
       ID        string
       UserID    string
       Token     string
       ExpiresAt time.Time
       UsedAt    *time.Time
   }
   ```

2. **Add Repository Interface** (`internal/repository/repository.go`)
   ```go
   type PasswordResetRepository interface {
       CreateResetToken(ctx context.Context, reset *entity.PasswordReset) error
       GetResetToken(ctx context.Context, token string) (*entity.PasswordReset, error)
       MarkAsUsed(ctx context.Context, token string) error
   }
   ```

3. **Implement Repository** (`internal/repository/postgres.go`)
   ```go
   type PasswordResetRepositoryImpl struct {
       db *PostgresDB
   }
   
   func (r *PasswordResetRepositoryImpl) CreateResetToken(...) error {
       // Implementation
   }
   ```

4. **Add Usecase** (`internal/usecase/password_reset.go`)
   ```go
   type PasswordResetUseCase struct {
       userRepo repository.UserRepository
       resetRepo repository.PasswordResetRepository
   }
   
   func (u *PasswordResetUseCase) RequestReset(...) error {
       // Implementation
   }
   ```

5. **Add Handler** (`internal/handler/password_reset.go`)
   ```go
   func (h *PasswordHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

6. **Add Routes** (`cmd/api/main.go`)
   ```go
   mux.HandleFunc("/password-reset/request", passwordHandler.ForgotPassword)
   mux.HandleFunc("/password-reset/confirm", passwordHandler.ResetPassword)
   ```

## Database Migrations

When adding new tables or columns:

1. Create new migration file: `database/migrations/003_add_feature.sql`
2. Use `IF NOT EXISTS` for idempotence
3. Include rollback comments
4. Document the changes

Example:
```sql
-- Migration: Add password_reset table
-- Run: psql -U postgres -d auth_db < database/migrations/003_add_password_reset.sql

CREATE TABLE IF NOT EXISTS password_resets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rollback:
-- DROP TABLE IF EXISTS password_resets;
```

## Testing

### Unit Tests
```go
func TestCreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    uc := usecase.NewUserUseCase(mockRepo)
    
    // Act
    user, err := uc.CreateUser(context.Background(), req)
    
    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if user.Email != req.Email {
        t.Errorf("expected email %s, got %s", req.Email, user.Email)
    }
}
```

### Integration Tests
```go
func TestLoginFlow(t *testing.T) {
    // Setup database
    db := setupTestDB(t)
    defer db.Close()
    
    // Run test
    // ...
}
```

## Performance Considerations

1. **Indexes**: Add database indexes for frequently queried columns
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Caching**: Consider caching roles and permissions
4. **Pagination**: Implement pagination for list endpoints
5. **Query Optimization**: Use efficient SQL queries

## Security Best Practices

1. **Password Hashing**: Always use bcrypt with salt
2. **JWT Secret**: Use strong, randomly generated secrets
3. **HTTPS**: Enforce HTTPS in production
4. **Rate Limiting**: Implement rate limiting for auth endpoints
5. **Audit Logging**: Log all sensitive operations
6. **SQL Injection**: Use parameterized queries (always done with database/sql)
7. **CORS**: Carefully configure CORS headers

## Deployment Checklist

- [ ] Change JWT_SECRET to strong random value
- [ ] Set ENVIRONMENT=production
- [ ] Use managed PostgreSQL service
- [ ] Enable HTTPS
- [ ] Setup logging and monitoring
- [ ] Configure backup strategy
- [ ] Setup CI/CD pipeline
- [ ] Configure secrets management
- [ ] Test disaster recovery
- [ ] Setup alerts for errors

## Common Tasks

### Adding a New Endpoint

1. Add handler method in `internal/handler/`
2. Add usecase logic in `internal/usecase/`
3. Add repository methods if needed
4. Register route in `cmd/api/main.go`
5. Add tests
6. Document in `API.md`

### Modifying Database Schema

1. Create new migration file
2. Test migration locally
3. Update entity definitions
4. Update repository queries
5. Add new tests

### Debugging

Enable debug logging:
```go
log.Debug("Processing login for user: %s", email)
```

Check JWT claims:
```bash
curl -s http://localhost:8001/login ... | jq '.access_token' | jwt decode -
```

## Need Help?

- Check existing code for patterns
- Review tests for examples
- Read API.md for endpoint details
- Check SETUP.md for environment setup
