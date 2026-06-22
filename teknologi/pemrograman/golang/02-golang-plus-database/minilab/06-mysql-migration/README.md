# Manual Test: MySQL Migrations & Environment Sync

Folder ini berisi script Go untuk menguji semua konsep dari materi **06-mysql-migration.md**.

## Materi yang Dicakup

- **Migration Framework** — Menggunakan `golang-migrate` untuk migrasi database
- **Embedded SQL** — File `.sql` diembed ke binary dengan `//go:embed`
- **Auto-Migrate** — Migrasi otomatis saat aplikasi startup (`main.go`)
- **Environment Sync** — Konfigurasi terpisah per environment (`DEV_`, `STAGING_`, `PROD_`)
- **Database Creation** — Buat database otomatis jika belum ada (`EnsureDatabase`)
- **CLI Migration Tool** — Command-line untuk manage migrasi (`cmd/migrate`)
- **Migration Operations** — `up`, `down`, `step`, `force`, `version`
- **Dirty Recovery** — `force` untuk recovery dari dirty state

## Struktur File

```
06-mysql-migration/
├── docker-compose.yml              ← MySQL container
├── main.go                         ← Auto-migrate + verifikasi data
├── cmd/
│   └── migrate/
│       └── main.go                 ← CLI tool untuk manage migrasi
├── internal/
│   └── db/
│       ├── config.go               ← Config from env (DEV_/STAGING_/PROD_)
│       ├── mysql.go                ← Koneksi + ensure database
│       ├── migrate.go              ← Wrapper golang-migrate + embed
│       └── migrations/             ← File SQL migrasi
│           ├── 000001_create_users_table.up.sql
│           ├── 000001_create_users_table.down.sql
│           ├── 000002_create_products.up.sql
│           ├── 000002_create_products.down.sql
│           ├── 000003_seed_data.up.sql
│           └── 000003_seed_data.down.sql
├── README.md
├── go.mod
└── 06-test                         ← Hasil compile (setelah build, opsional)
```

## Prasyarat

- Docker & Docker Compose
- Go 1.22+
- Port 3306 tidak dipakai oleh service lain

## Cara Menjalankan

### 1. Start MySQL Container

```bash
cd 06-mysql-migration
docker compose up -d

# Tunggu sampai healthcheck lulus
docker compose ps
# Status harus "healthy"
```

### 2. Jalankan Auto-Migrate

```bash
cd 06-mysql-migration

# Install dependency
go mod tidy

# Jalankan (development environment default)
go run main.go

# Atau dengan environment spesifik
APP_ENV=development go run main.go
APP_ENV=staging     go run main.go
APP_ENV=production  go run main.go
```

### 3. Gunakan CLI Migration Tool

```bash
cd 06-mysql-migration

# Melihat semua migration yang pending
go run ./cmd/migrate -up

# Rollback 1 migration terakhir
go run ./cmd/migrate -down

# Step 2 migrations ke atas
go run ./cmd/migrate -step 2

# Step 1 migration ke bawah (rollback)
go run ./cmd/migrate -step -1

# Cek version saat ini
go run ./cmd/migrate -version

# Force version (recovery dirty state)
go run ./cmd/migrate -force 1

# Dengan environment berbeda
go run ./cmd/migrate -env staging -up
go run ./cmd/migrate -env production -version

# Pakai database berbeda
go run ./cmd/migrate -db nama_database_lain -up
```

### 4. Custom Environment Variables

| Variable              | Default       | Keterangan                           |
|-----------------------|---------------|--------------------------------------|
| `DEV_DB_HOST`         | `localhost`   | Host MySQL development               |
| `DEV_DB_PORT`         | `3306`        | Port MySQL development               |
| `DEV_DB_USER`         | `demo`        | User MySQL development               |
| `DEV_DB_PASSWORD`     | `demo`        | Password development                 |
| `DEV_DB_NAME`         | `golang_demo` | Nama database development            |
| `DEV_DB_MAX_OPEN`     | `25`          | Max open connections                 |
| `DEV_DB_MAX_IDLE`     | `5`           | Max idle connections                 |
| `DEV_DB_PASSWORD_FILE`| —             | Alternatif: baca password dari file  |
| `APP_ENV`             | `development` | Pilih prefix: development/staging/production |

Ganti prefix `DEV_` dengan `STAGING_` atau `PROD_` untuk environment lain.

### 5. Cleanup

```bash
cd 06-mysql-migration
docker compose down -v
```

## Output yang Diharapkan (Auto-Migrate)

```
🔧 Environment: development
📦 Database: demo@localhost:3306/golang_demo
✅ Database ensured
✅ Connected to MySQL
📋 Running migrations...
✅ Migrations complete

📊 Verifying seed data...
  Tables:
    - schema_migrations
    - users
    - categories
    - products
  Categories:
    1. Elektronik (elektronik)
    2. Pakaian (pakaian)
    3. Makanan (makanan)
  Products:
    1. Smartphone X | Rp5000000 | stock: 50 | Elektronik
    2. Laptop Pro | Rp15000000 | stock: 20 | Elektronik
    3. Kaos Polos | Rp75000 | stock: 200 | Pakaian
    4. Kopi Arabika | Rp45000 | stock: 500 | Makanan

🔄 Demo: menambah produk baru...
  ✅ Produk 'Mouse Gaming Pro' berhasil ditambahkan!

✅ All checks passed!
```

## Catatan

- Migrasi menggunakan `embed.FS` — semua file SQL ter-compile ke dalam binary, tidak perlu deploy folder migrations terpisah.
- Prefix environment (`DEV_` / `STAGING_` / `PROD_`) memudahkan manajemen konfigurasi terpusat.
- Password bisa dibaca dari file (docker secret) via `*_DB_PASSWORD_FILE`.
- Tabel `schema_migrations` dibuat otomatis oleh `golang-migrate` untuk tracking version.
