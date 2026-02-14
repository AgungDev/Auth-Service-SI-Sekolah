# 📘 Auth Service – API Contract

## Multi-Tenant School System (ARKAS Scenario)

Auth Service adalah Identity & Access Management (IAM) untuk sistem multi-tenant sekolah.

Service ini bertanggung jawab atas:

- Tenant (Sekolah)
- User
- Role
- Permission
- JWT Access & Refresh Token
- Audit Log aktivitas sensitif

Auth Service **tidak mengandung logic ARKAS**.

---

# 🧱 Architectural Role

Auth berfungsi sebagai:

- Identity Provider (IdP)
- Token Issuer
- Role & Permission Authority
- Security Boundary Enforcement

Service lain (contoh: ARKAS) hanya:

- Memverifikasi JWT
- Membaca claims
- Mengecek permission

---

# 🔐 Authentication Strategy

- JWT (RS256 recommended)
- Access Token: short-lived (15 menit)
- Refresh Token: long-lived (7 hari)
- Stateless access validation
- Refresh token disimpan di DB (rotating recommended)

---

# 📦 Base URL

```
/api/v1
```

---

# 🏫 Multi-Tenant Model

Semua user terikat ke:

```
tenant_id
```

SUPER_ADMIN tidak terikat tenant.

---

# 📄 JWT Claims Structure

Access Token payload:

```json
{
  "sub": "user_uuid",
  "tenant_id": "tenant_uuid",
  "role": "BENDAHARA",
  "permissions": ["budget:create", "realization:create", "report:view"],
  "is_super_admin": false,
  "exp": 1730000000,
  "iat": 1729990000,
  "iss": "auth-service"
}
```

ARKAS membaca:

- tenant_id
- permissions
- role
- is_super_admin

---

# 🟢 Public Endpoints

## 1️⃣ Login

### POST `/auth/login`

Request:

```json
{
  "email": "bendahara@sekolah.id",
  "password": "password123"
}
```

Response 200:

```json
{
  "access_token": "jwt_token",
  "refresh_token": "refresh_token",
  "expires_in": 900
}
```

Error 401:

```json
{
  "error": "invalid_credentials"
}
```

---

## 2️⃣ Refresh Token

### POST `/auth/refresh`

Request:

```json
{
  "refresh_token": "refresh_token"
}
```

Response:

```json
{
  "access_token": "new_jwt_token",
  "refresh_token": "new_refresh_token",
  "expires_in": 900
}
```

Jika refresh token invalid → 401.

---

## 3️⃣ Logout

### POST `/auth/logout`

Header:

```
Authorization: Bearer <access_token>
```

Request:

```json
{
  "refresh_token": "refresh_token"
}
```

Action:

- Revoke refresh token
- Insert audit log

Response:

```
204 No Content
```

---

# 🔐 Protected Endpoints (Require Access Token)

Semua endpoint berikut memerlukan:

```
Authorization: Bearer <access_token>
```

---

# 🏫 Tenant Management

## Create Tenant (SUPER_ADMIN only)

### POST `/tenants`

```json
{
  "name": "SMA Negeri 1",
  "address": "Jakarta",
  "status": "active"
}
```

Response:

```json
{
  "id": "tenant_uuid",
  "name": "SMA Negeri 1",
  "status": "active"
}
```

---

## Suspend Tenant (SUPER_ADMIN)

### PATCH `/tenants/{id}/suspend`

---

# 👤 User Management

## Create User

### POST `/users`

```json
{
  "email": "bendahara@sekolah.id",
  "password": "password123",
  "role_id": "uuid",
  "tenant_id": "uuid"
}
```

Rules:

- SUPER_ADMIN bisa buat user lintas tenant
- ADMIN_SEKOLAH hanya bisa buat user dalam tenant sendiri

---

## Update User

### PATCH `/users/{id}`

---

## Disable User

### PATCH `/users/{id}/disable`

Jika user disabled:

- Refresh token invalidated
- Access token tetap valid sampai expired

---

# 🏷 Role Management

## Create Role

### POST `/roles`

```json
{
  "name": "BENDAHARA",
  "tenant_id": "uuid"
}
```

Role bisa:

- Global (tenant_id null, only SUPER_ADMIN)
- Tenant-specific

---

## Assign Permission to Role

### POST `/roles/{id}/permissions`

```json
{
  "permissions": ["realization:create", "report:view"]
}
```

---

# 🔑 Permission Management

## Create Permission (SUPER_ADMIN)

### POST `/permissions`

```json
{
  "name": "realization:create",
  "description": "Create realization transaction"
}
```

Permissions global, tidak tenant-specific.

---

# 📜 Audit Log

## Get Audit Logs

### GET `/audit-logs`

Filter:

```
?tenant_id=
?user_id=
?from=
?to=
```

Audit Log Fields:

- id
- tenant_id
- user_id
- action
- resource
- ip_address
- created_at

---

# 🧠 Authorization Rules (Auth Service)

| Endpoint                 | Required Role               |
| ------------------------ | --------------------------- |
| Create Tenant            | SUPER_ADMIN                 |
| Suspend Tenant           | SUPER_ADMIN                 |
| Create Global Permission | SUPER_ADMIN                 |
| Create Role (Tenant)     | ADMIN_SEKOLAH               |
| Create User              | ADMIN_SEKOLAH               |
| View Audit Log           | SUPER_ADMIN / ADMIN_SEKOLAH |

---

# 🔄 Interaction with ARKAS

## Flow:

1. Client login ke Auth
2. Auth issue JWT
3. Client kirim JWT ke Gateway
4. Gateway forward ke ARKAS
5. ARKAS verify JWT signature
6. ARKAS cek permission

Auth tidak dipanggil saat request ARKAS berlangsung.

---

# 🛡 Security Design

- RS256 key pair
- Public key exposed at:

```
GET /.well-known/jwks.json
```

- Refresh token disimpan hashed di database
- Token rotation recommended
- Access token short-lived

---

# 🚫 Anti-Pattern

- Shared database antar service
- Call Auth setiap request ARKAS
- Simpan permission di ARKAS
- Hardcode role di service lain

---

# 📦 Error Format Standard

Semua error:

```json
{
  "error": "error_code",
  "message": "human readable message"
}
```

---

# 📈 Future Expansion

- SSO support
- OAuth2 provider
- Service-to-service token
- Key rotation
- Multi-region tenant

---

# 🎯 Boundary Summary

Auth bertanggung jawab atas:

- Identity
- Role
- Permission
- Token lifecycle

ARKAS bertanggung jawab atas:

- Business authorization
- Financial rule enforcement
- Tenant data isolation
