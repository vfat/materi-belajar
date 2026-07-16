---
topik: PostgreSQL Migrations & Backup
urutan: 10 dari 20
posisi: lanjutan
sebelumnya: PostgreSQL Concurrency & Locking
prerequisites:
  - PostgreSQL Concurrency & Locking (09-postgres-concurrency)
  - MySQL Migrations & Environment Sync (06-mysql-migration)
level: menengah
---

> 🔗 **Lanjutan dari:** PostgreSQL Concurrency & Locking
> ← Kembali ke: `09-postgres-concurrency.md`

# PostgreSQL: Migrations & Backup

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami workflow **schema versioning** untuk PostgreSQL
- Membuat **migration files** (up/down) dengan `golang-migrate`
- Mengintegrasikan migrasi ke aplikasi Go + PostgreSQL
- Melakukan **backup & restore** dengan `pg_dump` / `pg_restore`
- Mengelola **environment** (dev/staging/prod) dengan aman
- Menangani **dirty state**, rollback, dan recovery migration
- Membangun **auto-migrate** pada startup aplikasi
- Menggunakan **seed data** untuk development

---

## 1. Recap: Kenapa Migrations?

Migration adalah **version control untuk schema database**. Setiap perubahan struktur disimpan sebagai file terpisah, bisa diaplikasikan (`up`) atau di-rollback (`down`) secara terprogram.

| Tanpa Migrations | Dengan Migrations |
|---|---|
| Manual SQL via CLI/UI | File versioned + automated |
| Gampang lupa langkah | Idempotent & repeatable |
| Dev vs prod tidak sinkron | Sinkron antar environment |
| Rollback manual & riskan | Rollback via `down` file |
| Tidak ada audit trail | Riwayat lengkap perubahan |

> 💡 **Catatan:** Pendekatan ini mirip dengan materi #06 (MySQL), tapi PostgreSQL punya kelebihan: DDL fully transactional, `IF NOT EXISTS` untuk semua DDL, dan `pg_dump` untuk backup.

---

## 2. Tooling: golang-migrate

### 2.1 Instalasi

```bash
# Install CLI (dengan PostgreSQL driver)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Cek instalasi
migrate -version

# Tambahkan dependency ke project Go
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

### 2.2 Struktur Project

```
project/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_products_table.up.sql
│   ├── 000002_add_products_table.down.sql
│   ├── 000003_seed_initial_data.up.sql
│   ├── 000003_seed_initial_data.down.sql
│   └── ...
├── internal/
│   └── db/
│       ├── config.go
│       ├── postgres.go
│       └── migrate.go
├── cmd/
│   ├── migrate/
│   │   └── main.go
│   └── server/
│       └── main.go
├── scripts/
│   ├── backup.sh
│   └── restore.sh
├── docker-compose.yml
└── .env
```

### 2.3 Naming Convention

Format: `{version}_{description}.{direction}.sql`

| Bagian | Contoh | Keterangan |
|--------|--------|------------|
| version | `000001` | 6 digit, urut naik |
| description | `create_users_table` | Snake-case, deskriptif |
| direction | `up` / `down` | Apply / rollback |

```bash
# Buat migration menggunakan CLI
migrate create -ext sql -dir migrations -seq create_users_table
# Output:
#   migrations/000001_create_users_table.up.sql
#   migrations/000001_create_users_table.down.sql
```

---

## 3. Migration Files untuk PostgreSQL

### 3.1 CREATE TABLE (DDL)

**`migrations/000001_create_users_table.up.sql`**

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
```

**`migrations/000001_create_users_table.down.sql`**

```sql
-- +migrate Down
DROP TABLE IF EXISTS users;
```

### 3.2 Menambahkan Kolom (ALTER TABLE)

**`migrations/000002_add_phone_to_users.up.sql`**

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone VARCHAR(20) NULL,
    ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500) NULL;

CREATE INDEX IF NOT EXISTS idx_users_phone ON users (phone);
```

**`migrations/000002_add_phone_to_users.down.sql`**

```sql
-- +migrate Down
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS phone;
```

### 3.3 Membuat & Relasi Tabel

**`migrations/000003_create_products.up.sql`**

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL UNIQUE,
    description TEXT NULL,
    price DECIMAL(12,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_active ON products (is_active);
CREATE INDEX IF NOT EXISTS idx_products_metadata ON products USING GIN (metadata);
```

**`migrations/000003_create_products.down.sql`**

```sql
-- +migrate Down
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
```

### 3.4 Data Seed

**`migrations/000004_seed_development_data.up.sql`**

```sql
-- +migrate Up
-- Hanya jalankan di environment development
INSERT INTO categories (name, slug, description) VALUES
    ('Elektronik', 'elektronik', 'Produk elektronik dan gadget'),
    ('Pakaian', 'pakaian', 'Pakaian pria dan wanita'),
    ('Makanan', 'makanan', 'Makanan dan minuman ringan'),
    ('Buku', 'buku', 'Buku dan media pembelajaran')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (name, email, password_hash, role) VALUES
    ('Admin', 'admin@example.com', '$2a$10$dummy_hash_for_development', 'admin'),
    ('User Demo', 'demo@example.com', '$2a$10$dummy_hash_for_development', 'user')
ON CONFLICT (email) DO NOTHING;

INSERT INTO products (category_id, name, slug, price, stock)
SELECT c.id, p.name, p.slug, p.price, p.stock
FROM (VALUES
    ('Elektronik', 'Smartphone X', 'smartphone-x', 5000000, 50),
    ('Elektronik', 'Laptop Pro', 'laptop-pro', 15000000, 20),
    ('Pakaian', 'Kaos Polos', 'kaos-polos', 75000, 200),
    ('Makanan', 'Kopi Arabika', 'kopi-arabika', 45000, 500)
) AS p(cat_name, name, slug, price, stock)
JOIN categories c ON c.name = p.cat_name
ON CONFLICT (slug) DO NOTHING;
```

**`migrations/000004_seed_development_data.down.sql`**

```sql
-- +migrate Down
DELETE FROM products WHERE slug IN ('smartphone-x', 'laptop-pro', 'kaos-polos', 'kopi-arabika');
DELETE FROM users WHERE email IN ('admin@example.com', 'demo@example.com');
DELETE FROM categories WHERE slug IN ('elektronik', 'pakaian', 'makanan', 'buku');
```

### 3.5 Perubahan Kompleks dengan Transaction

**`migrations/000005_split_user_name.up.sql`**

```sql
-- +migrate Up
-- Pisahkan kolom name menjadi first_name dan last_name
BEGIN;

-- Tambah kolom baru
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS first_name VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS last_name VARCHAR(100) NULL;

-- Copy data: ambil kata pertama sebagai first_name, sisanya last_name
UPDATE users
SET first_name = SPLIT_PART(name, ' ', 1),
    last_name = TRIM(SUBSTRING(name FROM LENGTH(SPLIT_PART(name, ' ', 1)) + 1));

-- Hapus kolom lama
ALTER TABLE users DROP COLUMN IF EXISTS name;

COMMIT;
```

**`migrations/000005_split_user_name.down.sql`**

```sql
-- +migrate Down
BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(200) NULL;

UPDATE users
SET name = TRIM(COALESCE(first_name, '') || ' ' || COALESCE(last_name, ''));

ALTER TABLE users
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name;

COMMIT;
```

> 💡 **PostgreSQL DDL fully transactional** — semua `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE` bisa di-rollback. Ini lebih aman dari MySQL untuk migration kompleks.

---

## 4. Integrasi ke Aplikasi Go

### 4.1 Config & Environment

**`internal/db/config.go`**

```go
package db

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxOpen  int
	MaxIdle  int
}

// ConfigFromEnv membaca konfigurasi dari environment variables.
// Mendukung prefix untuk membedakan environment (dev/staging/prod).
func ConfigFromEnv(prefix string) Config {
	cfg := Config{
		Host:    getEnv(prefix+"DB_HOST", "localhost"),
		Port:    getEnvInt(prefix+"DB_PORT", 5432),
		User:    getEnv(prefix+"DB_USER", "demo"),
		DBName:  getEnv(prefix+"DB_NAME", "golang_demo"),
		SSLMode: getEnv(prefix+"DB_SSLMODE", "disable"),
		MaxOpen: getEnvInt(prefix+"DB_MAX_OPEN", 25),
		MaxIdle: getEnvInt(prefix+"DB_MAX_IDLE", 5),
	}

	// Password dibaca dari file atau env (lebih aman)
	if pwdFile := os.Getenv(prefix + "DB_PASSWORD_FILE"); pwdFile != "" {
		if data, err := os.ReadFile(pwdFile); err == nil {
			cfg.Password = string(data)
		}
	} else {
		cfg.Password = getEnv(prefix+"DB_PASSWORD", "demo")
	}

	return cfg
}

// DSN menghasilkan Data Source Name untuk koneksi PostgreSQL.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// MigrationDSN menghasilkan DSN tanpa database (untuk create database).
func (c Config) MigrationDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.SSLMode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
```

### 4.2 Koneksi PostgreSQL

**`internal/db/postgres.go`**

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxOpen)
	poolCfg.MinConns = int32(cfg.MaxIdle)
	poolCfg.MaxConnLifetime = 5 * time.Minute
	poolCfg.MaxConnIdleTime = 3 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// EnsureDatabase membuat database jika belum ada.
func EnsureDatabase(ctx context.Context, cfg Config) error {
	// Connect ke postgres database (default)
	poolCfg, err := pgxpool.ParseConfig(cfg.MigrationDSN())
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	// Cek apakah database sudah ada
	var exists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.DBName,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database: %w", err)
	}

	if exists {
		return nil
	}

	// Buat database (tidak bisa dalam transaksi)
	_, err = pool.Exec(ctx, fmt.Sprintf(
		"CREATE DATABASE %s WITH ENCODING 'UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8'",
		cfg.DBName,
	))
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	return nil
}
```

### 4.3 Migration Engine

**`internal/db/migrate.go`**

```go
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, dbName string) (*Migrator, error) {
	// Source dari embedded filesystem
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	// Database driver untuk PostgreSQL
	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName:    dbName,
		MultiStatement:  true, // izinkan multiple statements per file
		SchemaName:      "public",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down: %w", err)
	}
	return nil
}

func (m *Migrator) Step(n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration step %d: %w", n, err)
	}
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}

func (m *Migrator) Drop() error {
	return m.migrate.Drop()
}

// RunMigrations adalah one-liner untuk integrasi startup.
func RunMigrations(db *sql.DB, dbName string) error {
	migrator, err := NewMigrator(db, dbName)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := migrator.Up(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err == nil {
		if dirty {
			log.Printf("⚠️  Migration version %d is DIRTY\n", version)
		} else {
			log.Printf("✅ Migration at version %d\n", version)
		}
	}

	return nil
}
```

### 4.4 CLI untuk Migration

**`cmd/migrate/main.go`**

```go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"your-project/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	var (
		up      = flag.Bool("up", false, "Run all pending migrations")
		down    = flag.Bool("down", false, "Rollback one migration")
		step    = flag.Int("step", 0, "Step N migrations (+ up / - down)")
		version = flag.Bool("version", false, "Show current migration version")
		force   = flag.Int("force", 0, "Force set version (dirty recovery)")
		dbName  = flag.String("db", "", "Database name (overrides env)")
		env     = flag.String("env", "development", "Environment prefix")
	)
	flag.Parse()

	// Pilih prefix environment
	var prefix string
	switch *env {
	case "production", "staging":
		prefix = "PROD_"
	case "test":
		prefix = "TEST_"
	default:
		prefix = "DEV_" // development
	}

	cfg := db.ConfigFromEnv(prefix)
	if *dbName != "" {
		cfg.DBName = *dbName
	}

	ctx := context.Background()

	// Pastikan database ada
	if err := db.EnsureDatabase(ctx, cfg); err != nil {
		log.Fatalf("ensure database: %v", err)
	}

	// Connect menggunakan pgx/stdlib untuk database/sql compatibility
	database, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database, cfg.DBName)
	if err != nil {
		log.Fatalf("init migrator: %v", err)
	}

	switch {
	case *up:
		if err := migrator.Up(); err != nil {
			log.Fatalf("up: %v", err)
		}
		fmt.Println("✅ Migrations applied (up)")

	case *down:
		if err := migrator.Down(); err != nil {
			log.Fatalf("down: %v", err)
		}
		fmt.Println("✅ Last migration rolled back (down)")

	case *step != 0:
		if err := migrator.Step(*step); err != nil {
			log.Fatalf("step: %v", err)
		}
		fmt.Printf("✅ Stepped %d migrations\n", *step)

	case *force != 0:
		if err := migrator.Force(*force); err != nil {
			log.Fatalf("force: %v", err)
		}
		fmt.Printf("✅ Version forced to %d\n", *force)

	case *version:
		v, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", v, dirty)

	default:
		flag.Usage()
		os.Exit(1)
	}
}
```

### 4.5 Auto-Migrate pada Startup Server

**`cmd/server/main.go`**

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"your-project/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Tentukan environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Load config sesuai environment
	prefix := ""
	switch env {
	case "production":
		prefix = "PROD_"
	case "staging":
		prefix = "STAGING_"
	default:
		prefix = "DEV_"
	}

	cfg := db.ConfigFromEnv(prefix)
	ctx := context.Background()

	// Pastikan database ada
	if err := db.EnsureDatabase(ctx, cfg); err != nil {
		log.Fatalf("[%s] ensure db: %v", env, err)
	}

	// Connect
	database, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		log.Fatalf("[%s] connect: %v", env, err)
	}
	defer database.Close()

	// Auto-migrate
	if err := db.RunMigrations(database, cfg.DBName); err != nil {
		log.Fatalf("[%s] migrate: %v", env, err)
	}

	log.Printf("[%s] Server started on port %s", env, os.Getenv("PORT"))
	// ... start server ...
}
```

---

## 5. Backup & Restore

### 5.1 pg_dump — Backup Database

```bash
# Backup seluruh database (plain SQL)
pg_dump -h localhost -U demo -d golang_demo -F p -f backup.sql

# Backup dalam format custom (compressed, bisa selective restore)
pg_dump -h localhost -U demo -d golang_demo -F c -f backup.dump

# Backup hanya schema (tanpa data)
pg_dump -h localhost -U demo -d golang_demo -s -f schema.sql

# Backup hanya data (tanpa schema)
pg_dump -h localhost -U demo -d golang_demo -a -f data.sql

# Backup tabel tertentu
pg_dump -h localhost -U demo -d golang_demo -t users -t products -f tables.sql

# Backup dengan options
pg_dump -h localhost -U demo -d golang_demo \
    --no-owner \
    --no-privileges \
    --clean \
    --if-exists \
    -F c -f backup.dump
```

### 5.2 pg_restore — Restore Database

```bash
# Buat database kosong terlebih dahulu
createdb -h localhost -U demo golang_demo_restored

# Restore dari format custom
pg_restore -h localhost -U demo -d golang_demo_restored backup.dump

# Restore hanya schema
pg_restore -h localhost -U demo -d golang_demo_restored -s backup.dump

# Restore hanya data
pg_restore -h localhost -U demo -d golang_demo_restored -a backup.dump

# Restore dengan options
pg_restore -h localhost -U demo -d golang_demo_restored \
    --clean \
    --if-exists \
    --no-owner \
    backup.dump
```

### 5.3 Backup Script untuk Go

**`scripts/backup.sh`**

```bash
#!/bin/bash
# scripts/backup.sh — Backup PostgreSQL database

set -euo pipefail

# Config
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-demo}"
DB_NAME="${DB_NAME:-golang_demo}"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.dump"

# Buat direktori backup
mkdir -p "$BACKUP_DIR"

echo "📦 Backing up $DB_HOST:$DB_PORT/$DB_NAME ..."

# Backup dengan pg_dump (format custom, compressed)
PGPASSWORD="${DB_PASSWORD:-demo}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -F c \
    -Z 9 \
    --no-owner \
    --no-privileges \
    -f "$BACKUP_FILE"

# Verifikasi
if [ -f "$BACKUP_FILE" ]; then
    SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✅ Backup saved: $BACKUP_FILE ($SIZE)"
else
    echo "❌ Backup failed!"
    exit 1
fi

# Cleanup backup lama (simpan 7 hari terakhir)
find "$BACKUP_DIR" -name "*.dump" -mtime +7 -delete
echo "🧹 Old backups cleaned"
```

### 5.4 Restore Script

**`scripts/restore.sh`**

```bash
#!/bin/bash
# scripts/restore.sh — Restore PostgreSQL database

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup_file.dump>"
    echo ""
    echo "Available backups:"
    ls -lh ./backups/*.dump 2>/dev/null || echo "  No backups found"
    exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ File not found: $BACKUP_FILE"
    exit 1
fi

# Config
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-demo}"
DB_NAME="${DB_NAME:-golang_demo}"

echo "⚠️  This will OVERWRITE database $DB_NAME!"
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

echo "🔄 Dropping existing database..."
PGPASSWORD="${DB_PASSWORD:-demo}" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" --if-exists "$DB_NAME"

echo "🆕 Creating new database..."
PGPASSWORD="${DB_PASSWORD:-demo}" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"

echo "📥 Restoring from $BACKUP_FILE ..."
PGPASSWORD="${DB_PASSWORD:-demo}" pg_restore \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --clean \
    --if-exists \
    --no-owner \
    "$BACKUP_FILE"

echo "✅ Restore complete!"
```

### 5.5 Backup dari Go Application

```go
package db

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type BackupConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	FilePath string
}

func BackupDatabase(ctx context.Context, cfg BackupConfig) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", cfg.Host,
		"-p", fmt.Sprintf("%d", cfg.Port),
		"-U", cfg.User,
		"-d", cfg.DBName,
		"-F", "c", // custom format
		"-Z", "9", // max compression
		"--no-owner",
		"--no-privileges",
		"-f", cfg.FilePath,
	)

	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func RestoreDatabase(ctx context.Context, cfg BackupConfig) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pg_restore",
		"-h", cfg.Host,
		"-p", fmt.Sprintf("%d", cfg.Port),
		"-U", cfg.User,
		"-d", cfg.DBName,
		"--clean",
		"--if-exists",
		"--no-owner",
		cfg.FilePath,
	)

	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
```

---

## 6. Environment Management

### 6.1 Strategy Multi-Environment

```
┌─────────────────────────────────────────────────────┐
│                   Application                        │
├─────────────────────────────────────────────────────┤
│   APP_ENV = development / staging / production      │
├──────────┬──────────┬───────────────────────────────┤
│   DEV_   │  STAGING_│          PROD_                │
│  DB_HOST │ DB_HOST  │  DB_HOST                      │
│  DB_USER │ DB_USER  │  DB_USER                      │
│  DB_PASS │ DB_PASS  │  DB_PASSWORD_FILE (lebih aman)│
│  DB_NAME │ DB_NAME  │  DB_NAME                      │
└──────────┴──────────┴───────────────────────────────┘
```

**`.env` (development)**

```bash
# Development — nyalakan fitur debug
APP_ENV=development
DEV_DB_HOST=localhost
DEV_DB_PORT=5432
DEV_DB_USER=demo
DEV_DB_PASSWORD=demo
DEV_DB_NAME=golang_demo
DEV_DB_SSLMODE=disable
DEV_DB_MAX_OPEN=10
DEV_DB_MAX_IDLE=5

# Debug & seeding
DB_DEBUG=true
DB_SEED=true
```

**`.env.staging`**

```bash
# Staging — mirror production, tapi credential tetap aman
APP_ENV=staging
STAGING_DB_HOST=staging-db.internal
STAGING_DB_PORT=5432
STAGING_DB_USER=app_service
STAGING_DB_PASSWORD=s3cret-staging
STAGING_DB_NAME=golang_app
STAGING_DB_SSLMODE=require
STAGING_DB_MAX_OPEN=20
STAGING_DB_MAX_IDLE=5

# Seed hanya data dummy
DB_SEED=true
```

**`.env.production`**

```bash
# Production — pakai password via file (docker secret)
APP_ENV=production
PROD_DB_HOST=prod-db.cluster.internal
PROD_DB_PORT=5432
PROD_DB_USER=app_service
PROD_DB_PASSWORD_FILE=/run/secrets/db_password
PROD_DB_NAME=golang_app
PROD_DB_SSLMODE=require
PROD_DB_MAX_OPEN=50
PROD_DB_MAX_IDLE=10
```

### 6.2 Docker Compose dengan Multi-Environment

**`docker-compose.yml`**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: golang_postgres_demo
    restart: unless-stopped
    environment:
      POSTGRES_DB: golang_demo
      POSTGRES_USER: demo
      POSTGRES_PASSWORD: demo
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U demo -d golang_demo"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    build: .
    container_name: golang_app
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8080:8080"
    env_file:
      - .env
    environment:
      # Bisa dioverride per service
      DEV_DB_HOST: postgres
    volumes:
      - ./migrations:/app/migrations

volumes:
  postgres_data:
```

---

## 7. Advanced Migration Patterns

### 7.1 Conditional Migration (hanya untuk environment tertentu)

```go
// internal/db/migrate.go
func RunMigrations(db *sql.DB, dbName, env string) error {
	migrator, err := NewMigrator(db, dbName)
	if err != nil {
		return err
	}

	// Jalankan migrasi struktural
	if err := migrator.Step(3); err != nil { // 3 migration DDL
		return err
	}

	// Seed data hanya untuk development/staging
	if env == "development" || env == "staging" {
		if err := migrator.Step(1); err != nil { // seed migration
			return err
		}
	}

	return nil
}
```

### 7.2 Backup Before Migration

```bash
#!/bin/bash
# scripts/migrate-with-backup.sh

DB_NAME="${DB_NAME:-golang_demo}"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.dump"

mkdir -p "$BACKUP_DIR"

echo "📦 Backing up $DB_NAME to $BACKUP_FILE ..."
docker compose exec -T postgres pg_dump \
    -U demo \
    -d "$DB_NAME" \
    -F c \
    -Z 9 \
    --no-owner \
    -f "/tmp/backup.dump"

docker compose cp postgres:/tmp/backup.dump "$BACKUP_FILE"

echo "✅ Backup saved: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

echo "🚀 Running migrations ..."
go run ./cmd/migrate -up -env development

echo "✅ Done!"
```

### 7.3 Version Checking & Guard

```go
// internal/db/guard.go
package db

import (
	"database/sql"
	"fmt"
)

// RequireVersion memastikan database berada di version tertentu.
func RequireVersion(db *sql.DB, dbName string, minVersion uint) error {
	migrator, err := NewMigrator(db, dbName)
	if err != nil {
		return err
	}

	v, _, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	if v < minVersion {
		return fmt.Errorf(
			"database version %d is below minimum %d. Run migration first",
			v, minVersion,
		)
	}

	return nil
}
```

### 7.4 Multi-Statement Migration

Untuk migration kompleks yang punya banyak statement, set `MultiStatement: true` di config driver:

```sql
-- +migrate Up
-- Migration ini akan dijalankan dalam satu transaksi

ALTER TABLE orders ADD COLUMN discount DECIMAL(12,2) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN coupon_code VARCHAR(50) NULL;
CREATE INDEX IF NOT EXISTS idx_orders_coupon ON orders (coupon_code);

-- Update data existing
UPDATE orders SET discount = 0 WHERE discount IS NULL;
```

---

## 8. Troubleshooting

### 8.1 Dirty State Recovery

Migration **dirty** berarti gagal di tengah jalan dan perlu diperbaiki manual.

```bash
# Cek dirty state
go run ./cmd/migrate -version

# Output: Version: 3, Dirty: true

# Investigasi: cek apa yang gagal
# 1. Lihat log error
# 2. Cek apakah perubahan sudah terapply sebagian

# Solusi 1: Force ke version sebelumnya (jika perubahan belum terapply penuh)
go run ./cmd/migrate -force=2

# Solusi 2: Fix manual, lalu force ke version selanjutnya
psql -h localhost -U demo -d golang_demo < migrations/000003_xxx.up.sql
go run ./cmd/migrate -force=3
```

### 8.2 Common Errors

| Error | Penyebab | Solusi |
|-------|----------|--------|
| `dirty database` | Migration gagal di tengah | Force version, fix manual |
| `no change` | Semua migration sudah terapply | Aman, tidak perlu action |
| `relation already exists` | Table sudah ada | Cek `IF NOT EXISTS`, atau force version |
| `foreign key constraint` | Data reference tidak ada | Urutkan migration dengan benar |
| `lock timeout` | Query lain sedang hold lock | Perpanjang timeout, atau cek `pg_locks` |
| `permission denied` | User tidak punya hak | Grant privileges: `GRANT ALL ON DATABASE` |

### 8.3 Migration Dry Run

```go
func DryRun(db *sql.DB, dbName string) error {
	migrator, err := NewMigrator(db, dbName)
	if err != nil {
		return err
	}

	v, dirty, err := migrator.Version()
	if err != nil {
		return err
	}

	if dirty {
		fmt.Printf("⚠️  Current version %d is DIRTY\n", v)
	} else {
		fmt.Printf("✅ Current version: %d (clean)\n", v)
	}

	// Cek migration files yang belum dijalankan
	nextVersion := v + 1
	fmt.Printf("📋 Pending migrations starting from: %d\n", nextVersion)

	return nil
}
```

---

## 9. Best Practices

### 9.1 Panduan Penulisan Migration

```sql
-- ✅ SELALU gunakan IF NOT EXISTS / IF EXISTS
CREATE TABLE IF NOT EXISTS users (...);
DROP TABLE IF EXISTS users;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);

-- ✅ Gunakan transaksi eksplisit untuk perubahan multi-step
BEGIN;
ALTER TABLE ...;
UPDATE ...;
COMMIT;

-- ✅ Gunakan comment untuk menjelaskan "kenapa"
-- WHY: Memisahkan nama depan/belakang untuk fitur search
ALTER TABLE users ADD COLUMN first_name VARCHAR(100);

-- ❌ Jangan pakai nilai hardcode yang spesifik environment
-- INSERT INTO configs VALUES (1, 'api_key', 'dev-key-123'); -- TIDAK
```

### 9.2 Git Workflow

```
main
 ├── migrations/000001_create_users_table.up.sql
 ├── migrations/000001_create_users_table.down.sql
 ├── migrations/000002_add_phone_to_users.up.sql
 └── migrations/000002_add_phone_to_users.down.sql

feature/new-feature-branch
 ├── migrations/000003_add_orders_table.up.sql   ← migration baru di branch
 └── migrations/000003_add_orders_table.down.sql
```

> ⚡ **Aturan Emas:** Jangan mengubah migration yang sudah di-merge ke `main`. Buat migration baru untuk setiap perubahan.

### 9.3 Production Migration Checklist

- [ ] Backup database sebelum migrate
- [ ] Test migration di staging dulu
- [ ] Jalankan `down` untuk verifikasi rollback
- [ ] Monitor `schema_migrations` table
- [ ] Siapkan rollback plan
- [ ] Migrasi di jam sepi (low traffic)
- [ ] Pastikan cukup disk space untuk backup

---

## 10. Ringkasan

| Topik | Kunci Utama |
|-------|-------------|
| Migration Files | `{version}_{desc}.up.sql` + `.down.sql` |
| PostgreSQL DDL | `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE`, `DROP IF EXISTS` |
| golang-migrate | Source: `iofs` / `file`, Driver: `postgres` |
| Multi-Environment | Prefix env: `DEV_`, `STAGING_`, `PROD_` |
| Auto-Migrate | Panggil `RunMigrations` di startup server |
| Dirty State | `Force(version)` untuk recovery |
| Backup | `pg_dump -F c` (custom format, compressed) |
| Restore | `pg_restore --clean --if-exists` |
| Seeding | Migration terpisah untuk data awal |
| Git | Migration = immutable setelah merge |

---

## 11. Latihan

1. Buat 3 migration files untuk schema `blogs` (users, posts, comments)
2. Buat CLI migration yang bisa `up`, `down`, dan `version`
3. Implementasikan `ConfigFromEnv` dengan prefix untuk development
4. Buat docker-compose untuk PostgreSQL + auto migrate
5. Simulasikan dirty state dan recovery
6. Buat backup script yang otomatis compress dan cleanup

```bash
# Bantuan cepat
migrate create -ext sql -dir migrations -seq create_posts_table
# Apply
migrate -path migrations -database "postgres://demo:demo@localhost:5432/golang_demo?sslmode=disable" up
# Rollback
migrate -path migrations -database "postgres://demo:demo@localhost:5432/golang_demo?sslmode=disable" down 1
```

---

## 12. Referensi

- https://www.postgresql.org/docs/current/backup-dump.html — pg_dump / pg_restore
- https://www.postgresql.org/docs/current/app-createdb.html — createdb
- https://github.com/golang-migrate/migrate — golang-migrate
- https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib — pgx/stdlib (database/sql driver)
- https://www.postgresql.org/docs/current/manage-ag-templatedbs.html — Template Databases

---

> ⏭️ **Selanjutnya:** `11-mongodb-setup.md` — Golang + MongoDB: Setup & BSON
