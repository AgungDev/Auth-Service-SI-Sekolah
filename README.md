# Auth Service – Sistem Informasi Sekolah

Auth Service adalah **microservice autentikasi & otorisasi**  
untuk Sistem Informasi Sekolah berbasis **multi-tenant (sekolah)**.

Service ini bertanggung jawab penuh atas:

- User
- Tenant (sekolah)
- Role
- Permission
- JWT sebagai kontrak antar service

⚠️ Repo ini **HANYA** auth-service.  
Service lain (ARKAS, Akademik, dll) **tidak ada di repo ini**.

---

## 🎯 Tujuan Dibuat

- Menjadi **single source of truth** identitas user
- Menghasilkan JWT berisi konteks tenant & permission
- Menjadi fondasi untuk microservice lain di masa depan
- Studi kasus pembelajaran microservice nyata (bukan dummy)

---

## 🧠 Konsep Inti

### Tenant

- 1 tenant = 1 sekolah
- Semua user **SELALU** terikat ke satu tenant
- Tenant **tidak boleh** sharing data

### JWT sebagai Kontrak

- Auth Service mengeluarkan JWT
- Service lain **percaya** JWT
- Auth Service **TIDAK DIPANGGIL** di setiap request service lain

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
- `handler` → boleh framework
- Tidak ada logic bisnis di handler

---

## 👤 Role (Global)

| Role           | Scope  | Deskripsi          |
| -------------- | ------ | ------------------ |
| SUPER_ADMIN    | System | Owner / SaaS admin |
| ADMIN_SEKOLAH  | Tenant | Admin sekolah      |
| BENDAHARA      | Tenant | Pengelola keuangan |
| KEPALA_SEKOLAH | Tenant | Approver           |
| OPERATOR       | Tenant | Input data         |

---

## 🔑 Permission (Auth Service)

Format permission:

```

{resource}.{action}

```

### Daftar Permission

| Permission      | Deskripsi           |
| --------------- | ------------------- |
| tenant.create   | Buat sekolah        |
| tenant.update   | Update sekolah      |
| tenant.suspend  | Suspend sekolah     |
| user.create     | Tambah user         |
| user.update     | Edit user           |
| user.delete     | Hapus user          |
| role.assign     | Assign role ke user |
| permission.read | Lihat permission    |

> Permission service lain (ARKAS, dll) **TIDAK DIDEFINISIKAN DI SINI**  
> tapi **DIPRODUKSI** oleh Auth Service.

---

## 🔐 JWT Payload (Kontrak Resmi)

```json
{
  "sub": "user-id",
  "tenant_id": "school-id",
  "roles": ["bendahara"],
  "permissions": ["transaction.create", "transaction.update", "report.view"],
  "exp": 1710000000
}
```

Field wajib:

- `sub`
- `tenant_id`
- `permissions`

Auth Service:

- Menghasilkan token
- Menentukan isi token
  Service lain:
- Hanya membaca & memverifikasi

---

## 📡 API Endpoint (Minimal)

### POST /login

Login user dan generate JWT.

Request:

```json
{
  "email": "user@school.sch.id",
  "password": "secret"
}
```

Response:

```json
{
  "access_token": "jwt-token",
  "expires_in": 3600
}
```

---

### POST /tenants

Buat tenant (sekolah).

> Biasanya hanya SUPER_ADMIN.

---

### POST /users

Buat user dalam tenant.

---

## 🗄️ Database

Auth Service memiliki **database sendiri**.

### Tabel

- tenants
- users
- roles
- permissions
- role_permissions
- user_roles

⚠️ Tidak ada service lain yang boleh mengakses database ini.

---

## 🐳 Container

### Dockerfile (Golang)

Auth Service dijalankan sebagai container mandiri.

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

Auth Service akan berjalan di:

```
http://localhost:8001
```

---

## 🚫 Anti-Pattern (DILARANG)

- Share database dengan service lain
- Hardcode role di handler
- Call auth-service di setiap request service lain
- Mengabaikan tenant_id

---

## 🎯 Target Pembelajaran

Setelah auth-service ini selesai:

- Siap dipakai oleh service lain (ARKAS, dll)
- Boundary antar service jelas
- JWT dipahami sebagai kontrak, bukan sekadar token
- Siap naik ke microservice level berikutnya

---

## 📝 Catatan Jujur

Auth Service ini **tidak dibuat untuk pamer arsitektur**,
tapi untuk **bertahan saat sistem tumbuh**.

Kalau auth berantakan, semua service ikut tumbang.
