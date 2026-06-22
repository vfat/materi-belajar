---
topik: MySQL Migrations & Environment Sync
urutan: 6 dari 20
posisi: lanjutan
sebelumnya: MySQL Transactions & Query Patterns
prerequisites:
  - MySQL Transactions & Query Patterns (05-mysql-queries)
  - SQLite Migrations & Schema (03-sqlite-migration)
level: menengah
---

> 🔗 **Lanjutan dari:** MySQL Transactions & Query Patterns
> ← Kembali ke: `05-mysql-queries.md`

# MySQL: Migrations & Environment Sync

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami konsep schema versioning dan migration workflow untuk MySQL
- Membuat migration files (up/down) untuk berbagai skenario perubahan schema
- Mengintegrasikan `golang-migrate` ke aplikasi Go + MySQL
- Mengelola konfigurasi environment (dev/staging/prod) dengan aman
- Mengimplementasikan auto-migrate pada startup aplikasi
- Melakukan seeding data untuk development environment
- Menangani dirty state, rollback, dan recovery migration

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

> 💡 **Catatan:** Pendekatan ini mirip dengan materi #03 (SQLite), tapi MySQL punya kelebihan: DDL transactional (InnoDB), ALTER TABLE native, dan user management.

---

## 2. Tooling: golang-migrate

### 2.1 Instalasi

```bash
# Install CLI (dengan MySQL driver)
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Cek instalasi
migrate -version

# Tambahkan dependency ke project Go
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/mysql
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
│       ├── mysql.go
│       └── migrate.go
├── cmd/
│   ├── migrate/
│   │   └── main.go
│   └── server/
│       └── main.go
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

## 3. Migration Files untuk MySQL

### 3.1 CREATE TABLE (DDL)

**`migrations/000001_create_users_table.up.sql`**

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('admin', 'user') NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_users_email (email),
    INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
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
    ADD COLUMN phone VARCHAR(20) NULL AFTER email,
    ADD COLUMN avatar_url VARCHAR(500) NULL AFTER phone,
    ADD INDEX idx_users_phone (phone);
```

**`migrations/000002_add_phone_to_users.down.sql`**

```sql
-- +migrate Down
ALTER TABLE users
    DROP INDEX idx_users_phone,
    DROP COLUMN avatar_url,
    DROP COLUMN phone;
```

### 3.3 Membuat & Relasi Tabel

**`migrations/000003_create_orders.up.sql`**

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS categories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS products (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id BIGINT UNSIGNED NULL,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(220) NOT NULL UNIQUE,
    description TEXT NULL,
    price DECIMAL(12,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_products_category (category_id),
    INDEX idx_products_active (is_active),
    FULLTEXT INDEX idx_products_search (name, description),
    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**`migrations/000003_create_orders.down.sql`**

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
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO users (name, email, password_hash, role) VALUES
    ('Admin', 'admin@example.com', '$2a$10$dummy_hash_for_development', 'admin'),
    ('User Demo', 'demo@example.com', '$2a$10$dummy_hash_for_development', 'user')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO products (category_id, name, slug, price, stock) VALUES
    (1, 'Smartphone X', 'smartphone-x', 5000000, 50),
    (1, 'Laptop Pro', 'laptop-pro', 15000000, 20),
    (2, 'Kaos Polos', 'kaos-polos', 75000, 200),
    (3, 'Kopi Arabika', 'kopi-arabika', 45000, 500)
ON DUPLICATE KEY UPDATE name = VALUES(name);
```

**`migrations/000004_seed_development_data.down.sql`**

```sql
-- +migrate Down
DELETE FROM products WHERE slug IN ('smartphone-x', 'laptop-pro', 'kaos-polos', 'kopi-arabika');
DELETE FROM users WHERE email IN ('admin@example.com', 'demo@example.com');
DELETE FROM categories WHERE slug IN ('elektronik', 'pakaian', 'makanan', 'buku');
```

### 3.5 Perubahan Kompleks dengan Transaction

**`migrations/000005_split_product_name.up.sql`**

```sql
-- +migrate Up
-- Pisahkan kolom name menjadi first_name dan last_name
BEGIN;

-- Tambah kolom baru
ALTER TABLE users
    ADD COLUMN first_name VARCHAR(100) NULL AFTER name,
    ADD COLUMN last_name VARCHAR(100) NULL AFTER first_name;

-- Copy data: ambil kata pertama sebagai first_name, sisanya last_name
UPDATE users
SET first_name = SUBSTRING_INDEX(name, ' ', 1),
    last_name = TRIM(SUBSTRING(name, LENGTH(SUBSTRING_INDEX(name, ' ', 1)) + 1));

-- Hapus kolom lama
ALTER TABLE users DROP COLUMN name;

COMMIT;
```

**`migrations/000005_split_product_name.down.sql`**

```sql
-- +migrate Down
BEGIN;

ALTER TABLE users
    ADD COLUMN name VARCHAR(200) NULL AFTER email;

UPDATE users
SET name = CONCAT_WS(' ', first_name, last_name);

ALTER TABLE users
    DROP COLUMN last_name,
    DROP COLUMN first_name;

COMMIT;
```

> ⚠️ **Peringatan:** Di MySQL, DDL bisa transactional (InnoDB). Tapi untuk ALTER TABLE besar, lakukan di jam sepi dan backup dulu.

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
	Params   map[string]string
	MaxOpen  int
	MaxIdle  int
}

// ConfigFromEnv membaca konfigurasi dari environment variables.
// Mendukung prefix untuk membedakan environment (dev/staging/prod).
func ConfigFromEnv(prefix string) Config {
	cfg := Config{
		Host:   getEnv(prefix+"DB_HOST", "localhost"),
		Port:   getEnvInt(prefix+"DB_PORT", 3306),
		User:   getEnv(prefix+"DB_USER", "demo"),
		DBName: getEnv(prefix+"DB_NAME", "golang_demo"),
		Params: map[string]string{
			"charset":   "utf8mb4",
			"parseTime": "true",
			"loc":       "Asia/Jakarta",
		},
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

// DSN menghasilkan Data Source Name untuk koneksi MySQL.
func (c Config) DSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?",
		c.User, c.Password, c.Host, c.Port, c.DBName)
	for k, v := range c.Params {
		dsn += fmt.Sprintf("%s=%s&", k, v)
	}
	return dsn[:len(dsn)-1]
}

// MigrationDSN menghasilkan DSN tanpa database (untuk create database).
func (c Config) MigrationDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?",
		c.User, c.Password, c.Host, c.Port)
	for k, v := range c.Params {
		dsn += fmt.Sprintf("%s=%s&", k, v)
	}
	return dsn[:len(dsn)-1]
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

### 4.2 Koneksi MySQL

**`internal/db/mysql.go`**

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// EnsureDatabase membuat database jika belum ada.
func EnsureDatabase(cfg Config) error {
	db, err := sql.Open("mysql", cfg.MigrationDSN())
	if err != nil {
		return fmt.Errorf("open mysql (no db): %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
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
	"github.com/golang-migrate/migrate/v4/database/mysql"
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

	// Database driver untuk MySQL
	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName:    dbName,
		MultiStatement:  true, // izinkan multiple statements per file
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChanges {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChanges {
		return fmt.Errorf("migration down: %w", err)
	}
	return nil
}

func (m *Migrator) Step(n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChanges {
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
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"your-project/internal/db"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		up       = flag.Bool("up", false, "Run all pending migrations")
		down     = flag.Bool("down", false, "Rollback one migration")
		step     = flag.Int("step", 0, "Step N migrations (+ up / - down)")
		version  = flag.Bool("version", false, "Show current migration version")
		force    = flag.Int("force", 0, "Force set version (dirty recovery)")
		dbName   = flag.String("db", "", "Database name (overrides env)")
		env      = flag.String("env", "development", "Environment prefix")
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

	// Pastikan database ada
	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("ensure database: %v", err)
	}

	database, err := db.Connect(cfg)
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
	"log"
	"os"

	"your-project/internal/db"
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

	// Pastikan database ada
	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("[%s] ensure db: %v", env, err)
	}

	// Connect
	database, err := db.Connect(cfg)
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

## 5. Environment Management

### 5.1 Strategy Multi-Environment

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
DEV_DB_PORT=3306
DEV_DB_USER=demo
DEV_DB_PASSWORD=demo
DEV_DB_NAME=golang_demo
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
STAGING_DB_PORT=3306
STAGING_DB_USER=app_service
STAGING_DB_PASSWORD=s3cret-staging
STAGING_DB_NAME=golang_app
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
PROD_DB_PORT=3306
PROD_DB_USER=app_service
PROD_DB_PASSWORD_FILE=/run/secrets/db_password
PROD_DB_NAME=golang_app
PROD_DB_MAX_OPEN=50
PROD_DB_MAX_IDLE=10
```

### 5.2 Docker Compose dengan Multi-Environment

**`docker-compose.yml`**

```yaml
services:
  mysql:
    image: mysql:8.0
    container_name: golang_mysql_demo
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: golang_demo
      MYSQL_USER: demo
      MYSQL_PASSWORD: demo
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-proot"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    build: .
    container_name: golang_app
    depends_on:
      mysql:
        condition: service_healthy
    ports:
      - "8080:8080"
    env_file:
      - .env
    environment:
      # Bisa dioverride per service
      DEV_DB_HOST: mysql
    volumes:
      - ./migrations:/app/migrations

volumes:
  mysql_data:
```

**`Dockerfile`**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
```

---

## 6. Advanced Migration Patterns

### 6.1 Conditional Migration (hanya untuk environment tertentu)

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

### 6.2 Backup Before Migration

```bash
#!/bin/bash
# scripts/migrate-with-backup.sh

DB_NAME="${DB_NAME:-golang_demo}"
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"

mkdir -p "$BACKUP_DIR"

echo "📦 Backing up $DB_NAME to $BACKUP_FILE ..."
docker compose exec -T mysql mysqldump \
    -u root -proot \
    --single-transaction \
    --routines \
    --triggers \
    "$DB_NAME" > "$BACKUP_FILE"

echo "✅ Backup saved: $BACKUP_FILE ($(wc -c < "$BACKUP_FILE") bytes)"

echo "🚀 Running migrations ..."
go run ./cmd/migrate -up -env development

echo "✅ Done!"
```

### 6.3 Version Checking & Guard

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

### 6.4 Multi-Statement Migration

Untuk migration kompleks yang punya banyak statement, set `MultiStatement: true` di config driver:

```sql
-- +migrate Up
-- Migration ini akan dijalankan dalam satu transaksi

ALTER TABLE orders ADD COLUMN discount DECIMAL(12,2) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN coupon_code VARCHAR(50) NULL;
ALTER TABLE orders ADD INDEX idx_coupon (coupon_code);

-- Update data existing
UPDATE orders SET discount = 0 WHERE discount IS NULL;
```

---

## 7. Troubleshooting

### 7.1 Dirty State Recovery

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
mysql -h localhost -u root -proot golang_demo < migrations/000003_xxx.up.sql
go run ./cmd/migrate -force=3
```

### 7.2 Common Errors

| Error | Penyebab | Solusi |
|-------|----------|--------|
| `dirty database` | Migration gagal di tengah | Force version, fix manual |
| `no change` | Semua migration sudah terapply | Aman, tidak perlu action |
| `duplicate column` | Migration dijalankan 2x | Cek `schema_migrations` table |
| `foreign key constraint` | Data reference tidak ada | Urutkan migration dengan benar |
| `lock wait timeout` | Query lain sedang running | Perpanjang timeout, atau cek proses MySQL |

### 7.3 Migration Dry Run

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

## 8. Best Practices

### 8.1 Panduan Penulisan Migration

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

### 8.2 Git Workflow

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

### 8.3 Production Migration Checklist

- [ ] Backup database sebelum migrate
- [ ] Test migration di staging dulu
- [ ] Jalankan `down` untuk verifikasi rollback
- [ ] Monitor `schema_migrations` table
- [ ] Siapkan rollback plan
- [ ] Migrasi di jam sepi (low traffic)
- [ ] Pastikan cukup disk space untuk backup

---

## 9. Ringkasan

| Topik | Kunci Utama |
|-------|-------------|
| Migration Files | `{version}_{desc}.up.sql` + `.down.sql` |
| MySQL DDL | `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE`, `DROP IF EXISTS` |
| golang-migrate | Source: `iofs` / `file`, Driver: `mysql` |
| Multi-Environment | Prefix env: `DEV_`, `STAGING_`, `PROD_` |
| Auto-Migrate | Panggil `RunMigrations` di startup server |
| Dirty State | `Force(version)` untuk recovery |
| Seeding | Migration terpisah untuk data awal |
| Git | Migration = immutable setelah merge |

---

## 10. Latihan

1. Buat 3 migration files untuk schema `blogs` (users, posts, comments)
2. Buat CLI migration yang bisa `up`, `down`, dan `version`
3. Implementasikan `ConfigFromEnv` dengan prefix untuk development
4. Buat docker-compose untuk MySQL + auto migrate
5. Simulasikan dirty state dan recovery

```bash
# Bantuan cepat
migrate create -ext sql -dir migrations -seq create_posts_table
# Apply
migrate -path migrations -database "mysql://demo:demo@tcp(localhost:3306)/golang_demo" up
# Rollback
migrate -path migrations -database "mysql://demo:demo@tcp(localhost:3306)/golang_demo" down 1
```

---

> 🏁 **Selesai!** Selanjutnya: **[Golang + PostgreSQL: Setup & JSONB]**  
> → Lanjut ke: `07-postgres-setup.md`
