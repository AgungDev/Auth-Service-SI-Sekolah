# Auth Service – Multi-Tenant School System

Auth Service adalah **microservice autentikasi & otorisasi**  
untuk Sistem Informasi Sekolah berbasis **multi-tenant (sekolah)**.

Service ini bertanggung jawab penuh atas:

- Tenant (sekolah)
- User
- Role & Permission
- Token (Access & Refresh)
- Audit log aktivitas sensitif
- Boundary otoritas sistem

⚠️ Repository ini **HANYA** berisi auth-service.  
Tidak ada logic ARKAS atau service lain di sini.

---

## 🎯 Tujuan Service

- Menjadi **single source of truth** untuk identitas user
- Menyediakan **kontrak JWT ringan** untuk service lain
- Mendukung arsitektur microservice secara aman
- Siap dipakai sebagai fondasi produk SaaS

---

## 🧠 Konsep Inti

### Tenant

- 1 tenant = 1 sekolah
- Semua user **SELALU** terikat ke satu tenant
- Tidak ada data sharing antar tenant

### Token Strategy

Auth Service menggunakan **2 jenis token**:

- **Access Token** → pendek, dipakai antar service
- **Refresh Token** → panjang, hanya ke auth-service

Service lain **TIDAK PERNAH** menerima refresh token.

---

## 🔐 Token Design

### Access Token (JWT – Ringan)

Digunakan oleh service lain (ARKAS, dll).

```json
{
  "sub": "user-id",
  "tenant_id": "school-id",
  "roles": ["BENDAHARA"],
  "tenant_status": "ACTIVE",
  "exp": 1710000000
}
```

Prinsip:

- **TIDAK menyimpan permission detail**
- Role dipakai sebagai identifier
- Permission dimapping di masing-masing service

Tujuan:

- JWT kecil
- Performa stabil
- Mudah di-cache

---

### Refresh Token

- Disimpan di database auth-service
- Terkait dengan user & tenant
- Bisa direvoke kapan saja

Contoh atribut:

- token
- user_id
- tenant_id
- expires_at
- revoked_at

---

## 🏗️ Arsitektur

Auth Service menggunakan **Clean Architecture**.

```
cmd/
  api/
internal/
  entity/
  usecase/
  repository/
  handler/
  middleware/
pkg/
  config/
  logger/
```

### Aturan Boundary (WAJIB)

- `entity` → tidak tahu HTTP / DB
- `usecase` → tidak tahu framework
- `handler` → hanya adapter
- Logic bisnis **tidak boleh** di handler

---

## 👤 Role & Boundary Otoritas

### System Level (NON-SEKOLAH)

#### SYSTEM_OWNER

- Akses internal
- Tidak login via API publik
- Mengelola sistem SaaS

#### SUPER_ADMIN

- Membuat tenant
- Suspend / activate tenant
- Reset admin sekolah

⚠️ SUPER_ADMIN:

- Tidak digunakan operasional harian
- Tidak mengelola data sekolah

---

### Tenant Level (Sekolah)

| Role           | Deskripsi           |
| -------------- | ------------------- |
| ADMIN_SEKOLAH  | Admin utama sekolah |
| BENDAHARA      | Pengelola keuangan  |
| KEPALA_SEKOLAH | Approver & pengawas |
| OPERATOR       | Input data          |

---

## 🔑 Permission Strategy

Auth Service:

- Menyimpan **role & permission mapping**
- Menghasilkan token berbasis **role**

Service lain:

- Memetakan `role → permission` secara lokal
- Tidak memanggil auth-service

Contoh permission format:

```
{resource}.{action}
```

Contoh:

- transaction.create
- report.export

---

## 📡 API Endpoint (Minimal)

### POST /login

- Validasi user
- Menghasilkan access token & refresh token

Response:

```json
{
  "access_token": "jwt",
  "refresh_token": "opaque-token",
  "expires_in": 1800
}
```

---

### POST /refresh

- Mengganti access token
- Refresh token **WAJIB valid & belum direvoke**

---

### POST /tenants

- Buat tenant (SUPER_ADMIN only)

---

### POST /users

- Buat user dalam tenant

---

## 🗄️ Database Design

### auth_db Tables

#### tenants

- id
- name
- status (ACTIVE | SUSPENDED | ARCHIVED)

---

#### users

- id
- tenant_id
- email
- password_hash
- status

---

#### roles

- id
- name
- scope (SYSTEM | TENANT)

---

#### permissions

- id
- code
- description

---

#### role_permissions

- role_id
- permission_id

---

#### user_roles

- user_id
- role_id

---

#### refresh_tokens

- token
- user_id
- tenant_id
- expires_at
- revoked_at

---

#### audit_logs

Digunakan untuk **traceability & compliance**.

Field:

- id
- actor_id
- tenant_id
- action
- target
- metadata (JSON)
- created_at

Contoh action:

- USER_CREATED
- ROLE_ASSIGNED
- TENANT_SUSPENDED

---

## 🐳 Container

### Docker Compose (Ready to Run)

```yaml
version: "3.9"

services:
  auth-service:
    build: .
    ports:
      - "8001:8000"
    environment:
      - DB_HOST=auth-db
      - DB_NAME=auth_db
      - DB_USER=postgres
      - DB_PASS=postgres
    depends_on:
      - auth-db

  auth-db:
    image: postgres:15
    environment:
      POSTGRES_DB: auth_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5433:5432"
```

---

## ▶️ Cara Menjalankan

```bash
docker compose up --build
```

Auth Service tersedia di:

```
http://localhost:8001
```

---

## 🚫 Anti-Pattern (DILARANG)

- Menaruh permission detail di JWT
- Share database dengan service lain
- SUPER_ADMIN dipakai user sekolah
- Tidak mencatat audit log
- Refresh token dikirim ke service lain

---

## 🎯 Target Pembelajaran

Dengan auth-service ini:

- Boundary microservice jelas
- Token lifecycle dipahami
- Role ≠ Permission
- Sistem siap dikembangkan ke ARKAS & modul lain

---

## 📝 Catatan Akhir

Auth Service ini **dibuat untuk tumbuh**,
bukan untuk cepat-cepat jadi demo.

Kalau auth rapi,
service lain boleh ribet — sistem tetap hidup.
