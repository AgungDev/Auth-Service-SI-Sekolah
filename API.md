# Auth Service API Documentation

## Endpoints

### Health Check
- **GET** `/health`
  - Check if the service is running
  - No authentication required

### Authentication

#### Login
- **POST** `/login`
  - Login user and get tokens
  - **Request Body:**
    ```json
    {
      "email": "user@example.com",
      "password": "password123",
      "tenant_id": "school-id"
    }
    ```
  - **Response:**
    ```json
    {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "opaque-token",
      "expires_in": 1800
    }
    ```

#### Refresh Token
- **POST** `/refresh`
  - Get a new access token using refresh token
  - **Request Body:**
    ```json
    {
      "refresh_token": "opaque-token"
    }
    ```
  - **Response:** Same as login

### Tenant Management (SUPER_ADMIN only)

#### Create Tenant
- **POST** `/tenants`
  - Create a new school/tenant
  - **Authorization:** Bearer token (SUPER_ADMIN role required)
  - **Request Body:**
    ```json
    {
      "name": "SMA Negeri 1"
    }
    ```
  - **Response:**
    ```json
    {
      "message": "Tenant created successfully",
      "data": {
        "id": "tenant-id",
        "name": "SMA Negeri 1",
        "status": "ACTIVE",
        "created_at": "2024-01-10T10:00:00Z",
        "updated_at": "2024-01-10T10:00:00Z"
      }
    }
    ```

### User Management

#### Create User
- **POST** `/users`
  - Create a new user in a tenant
  - **Authorization:** Bearer token (must be ADMIN_SEKOLAH in the same tenant)
  - **Request Body:**
    ```json
    {
      "email": "user@example.com",
      "password": "password123",
      "tenant_id": "school-id",
      "role_ids": ["admin-sekolah"]
    }
    ```
  - **Response:**
    ```json
    {
      "message": "User created successfully",
      "data": {
        "id": "user-id",
        "tenant_id": "school-id",
        "email": "user@example.com",
        "status": "ACTIVE",
        "created_at": "2024-01-10T10:00:00Z",
        "updated_at": "2024-01-10T10:00:00Z"
      }
    }
    ```

## Access Token Claims

The JWT access token contains the following claims:

```json
{
  "sub": "user-id",
  "tenant_id": "school-id",
  "email": "user@example.com",
  "roles": ["BENDAHARA"],
  "tenant_status": "ACTIVE",
  "exp": 1710000000,
  "iat": 1710000000
}
```

## Roles

### System Level
- **SYSTEM_OWNER**: Internal access, manages the entire system
- **SUPER_ADMIN**: Creates and suspends tenants, resets admin accounts

### Tenant Level
- **ADMIN_SEKOLAH**: Can manage users, roles, and school data
- **BENDAHARA**: Financial management
- **KEPALA_SEKOLAH**: Approver and supervisor
- **OPERATOR**: Data entry

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message"
}
```

Common HTTP status codes:
- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Lacks required permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Testing with cURL

### Login
```bash
curl -X POST http://localhost:8001/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123",
    "tenant_id": "school-1"
  }'
```

### Refresh Token
```bash
curl -X POST http://localhost:8001/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "token-from-login"
  }'
```

### Create Tenant (as SUPER_ADMIN)
```bash
curl -X POST http://localhost:8001/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer access-token" \
  -d '{
    "name": "SMA Negeri 2"
  }'
```

### Create User
```bash
curl -X POST http://localhost:8001/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer access-token" \
  -d '{
    "email": "operator@example.com",
    "password": "password123",
    "tenant_id": "school-id",
    "role_ids": ["operator"]
  }'
```

### Health Check
```bash
curl http://localhost:8001/health
```
