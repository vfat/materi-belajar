---
topik: SQLite Migrations & Schema
urutan: 3 dari 20
posisi: setelah SQLite setup
tag:
  - sqlite
  - migration
  - golang-migrate
  - schema
  - version-control
prerequisites:
  - Golang + SQLite Setup & CRUD (02-sqlite-setup)
level: menengah
---

> 🚀 **Materi #03** — Manajemen schema versioning dengan migrations di SQLite.

# SQLite: Migrations & Schema

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami pentingnya schema versioning di aplikasi production
- Menggunakan `golang-migrate` untuk manajemen migrations
- Membuat migration files yang idempotent dan aman
- Mengimplementasikan rollback strategies
- Mengintegrasikan migrations ke aplikasi Go startup

---

## 1. Kenapa Perlu Migrations?

### 1.1 Masalah Tanpa Migrations

Bayangkan kamu punya aplikasi yang sudah di-deploy. Beberapa minggu kemudian, kamu perlu menambahkan kolom baru atau tabel baru. Tanpa sistem migrations:

```
❌ Tanpa migrations:
├── Manual SQL di server (gampang lupa)
├── Schema tidak sinkron antara dev/staging/prod
├── Rollback rumit kalau ada error
├── Tim lain tidak bisa reproduce schema
└── Deploy berantakan, data hilang
```

### 1.2 Benefits Migrations

```
✅ Dengan migrations:
├── Version-controlled schema changes
├── Automated deploy process
├── Rollback otomatis kalau gagal
├── Audit trail perubahan schema
├── Tim bisa reproduce environment
└── Zero-downtime deploy (untuk DB yang mendukung)
```

---

## 2. golang-migrate: Tool Migrations

### 2.1 Instalasi

```bash
# Install golang-migrate CLI
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Atau via package manager
# macOS
brew install migrate

# Linux (binary)
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xz
sudo mv migrate.linux-amd64 /usr/local/bin/migrate
```

### 2.2 Struktur Project Migrations

```
project/
├── migrations/
│   ├── 000001_create_categories_table.up.sql
│   ├── 000001_create_categories_table.down.sql
│   ├── 000002_create_tasks_table.up.sql
│   ├── 000002_create_tasks_table.down.sql
│   └── ...
├── cmd/
│   └── migrate/
│       └── main.go
└── internal/
    └── db/
        └── migrate.go
```

---

## 3. Membuat Migration Files

### 3.1 Migration File Naming

Format: `{version}_{description}.{direction}.sql`

| Komponen | Contoh | Keterangan |
|----------|--------|------------|
| version | `000001` | Urutan migrations (6 digit) |
| description | `create_categories_table` | Deskripsi singkat |
| direction | `up` atau `down` | `up` untuk apply, `down` untuk rollback |

### 3.2 Migration Up (Apply)

Buat file `migrations/000001_create_categories_table.up.sql`:

```sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT DEFAULT '#3B82F6',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
```

Buat file `migrations/000001_create_categories_table.down.sql`:

```sql
-- +migrate Down
DROP INDEX IF EXISTS idx_categories_name;
DROP TABLE IF EXISTS categories;
```

### 3.3 Migration dengan Data Seed

Buat file `migrations/000002_seed_categories.up.sql`:

```sql
-- +migrate Up
INSERT OR IGNORE INTO categories (name, description, color) VALUES
    ('Work', 'Work-related tasks', '#EF4444'),
    ('Personal', 'Personal tasks', '#10B981'),
    ('Learning', 'Learning and education', '#3B82F6');
```

Buat file `migrations/000002_seed_categories.down.sql`:

```sql
-- +migrate Down
DELETE FROM categories WHERE name IN ('Work', 'Personal', 'Learning');
```

### 3.4 Migration Alter Table

Buat file `migrations/000003_add_task_priority.up.sql`:

```sql
-- +migrate Up
-- SQLite tidak mendukung ALTER COLUMN, jadi perlu recreate table
BEGIN TRANSACTION;

-- Backup data lama
CREATE TEMPORARY TABLE tasks_backup AS SELECT * FROM tasks;

-- Drop table lama
DROP TABLE tasks;

-- Create table baru dengan kolom priority
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 2,
    status TEXT DEFAULT 'pending',
    category_id INTEGER,
    due_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE SET NULL
);

-- Restore data
INSERT INTO tasks (id, title, description, status, category_id, due_date, created_at, updated_at)
SELECT id, title, description, status, category_id, due_date, created_at, updated_at
FROM tasks_backup;

DROP TABLE tasks_backup;
COMMIT;
```

Buat file `migrations/000003_add_task_priority.down.sql`:

```sql
-- +migrate Down
BEGIN TRANSACTION;

CREATE TEMPORARY TABLE tasks_backup AS SELECT * FROM tasks;
DROP TABLE tasks;

CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'pending',
    category_id INTEGER,
    due_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE SET NULL
);

INSERT INTO tasks (id, title, description, status, category_id, due_date, created_at, updated_at)
SELECT id, title, description, status, category_id, due_date, created_at, updated_at
FROM tasks_backup;

DROP TABLE tasks_backup;
COMMIT;
```

---

## 4. CLI Commands

### 4.1 Basic Commands

```bash
# Buat file migration baru
migrate create -ext sql -dir migrations create_tasks_table

# Apply semua migrations
migrate -path migrations -database "sqlite3://app.db" up

# Rollback 1 migration
migrate -path migrations -database "sqlite3://app.db" down

# Rollback semua migrations
migrate -path migrations -database "sqlite3://app.db" down -all

# Cek status migrations
migrate -path migrations -database "sqlite3://app.db" version

# Force set version (untuk recovery)
migrate -path migrations -database "sqlite3://app.db" force 20240101000000
```

### 4.2 Environment-based DSN

Buat file `.env`:

```env
# .env
DATABASE_URL="sqlite3://app.db"
# ATAU untuk absolute path
DATABASE_URL="sqlite3:///home/user/project/app.db"
```

Gunakan di CLI:

```bash
# Load dari .env
export $(cat .env | xargs)
migrate -path migrations -database "$DATABASE_URL" up
```

---

## 5. Integrasi ke Aplikasi Go

### 5.1 Migrate Package

Buat file `internal/db/migrate.go`:

```go
package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, dbPath string) (*Migrator, error) {
	// Create source driver dari embedded files
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create database driver
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChanges {
		return fmt.Errorf("migration up failed: %w", err)
	}
	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChanges {
		return fmt.Errorf("migration down failed: %w", err)
	}
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}
```

### 5.2 CLI Command untuk Migrate

Buat file `cmd/migrate/main.go`:

```go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"golang-sqlite/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	up := flag.Bool("up", false, "Run migrations up")
	down := flag.Bool("down", false, "Run migrations down")
	force := flag.Int("force", 0, "Force set migration version (0 = skip)")
	version := flag.Bool("version", false, "Show current migration version")
	flag.Parse()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database, dbPath)
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case *up:
		if err := migrator.Up(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✅ Migrations applied successfully")

	case *down:
		if err := migrator.Down(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("✅ Migration rolled back successfully")

	case *force != 0:
		if err := migrator.Force(*force); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("✅ Migration version forced to %d\n", *force)

	case *version:
		v, dirty, err := migrator.Version()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", v, dirty)

	default:
		flag.Usage()
	}
}
```

### 5.3 Auto Migrate pada Startup

Buat file `cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"

	"golang-sqlite/internal/db"
)

func main() {
	// Setup database
	database, err := db.NewSQLiteFromEnv()
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer database.Close()

	// Auto migrate
	migrator, err := db.NewMigrator(database, os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatal("Failed to create migrator:", err)
	}

	ctx := context.Background()
	if err := migrator.Up(); err != nil {
		log.Fatal("Failed to migrate:", err)
	}

	log.Println("✅ Database migrated successfully")

	// ... rest of application
}
```

---

## 6. Best Practices

### 6.1 Idempotent Migrations

Selalu gunakan `IF NOT EXISTS` dan `OR IGNORE`:

```sql
-- ✅ Good: Idempotent
CREATE TABLE IF NOT EXISTS users (...);
INSERT OR IGNORE INTO users (id, name) VALUES (1, 'Admin');

-- ❌ Bad: Bisa error kalau dijalankan 2x
CREATE TABLE users (...);
INSERT INTO users (id, name) VALUES (1, 'Admin');
```

### 6.2 SQLite-specific Considerations

```sql
-- SQLite tidak mendukung ALTER COLUMN
-- Solusi: Recreate table dengan data migration

-- Gunakan WAL mode untuk concurrency lebih baik
PRAGMA journal_mode = WAL;

-- Enable foreign keys
PRAGMA foreign_keys = ON;
```

### 6.3 Migration File Template

Buat script helper `scripts/create-migration.sh`:

```bash
#!/bin/bash
# scripts/create-migration.sh

if [[ -z "$1" ]]; then
    echo "Usage: ./scripts/create-migration.sh <name>"
    exit 1
fi

NAME=$1
TIMESTAMP=$(date +%Y%m%d%H%M%S)

mkdir -p migrations

migrate create -ext sql -dir migrations -seq "$NAME"

echo "✅ Created migration: $NAME"
echo "   - migrations/${TIMESTAMP}_${NAME}.up.sql"
echo "   - migrations/${TIMESTAMP}_${NAME}.down.sql"
```

---

## 7. Troubleshooting

### 7.1 Migration Dirty State

Jika migration stuck di dirty state:

```bash
# Cek version
migrate -path migrations -database "sqlite3://app.db" version

# Force ke version sebelumnya
migrate -path migrations -database "sqlite3://app.db" force <version>
```

### 7.2 SQLite Lock Issues

```sql
-- Cek locks
PRAGMA locking_mode;
PRAGMA busy_timeout;

-- Set busy timeout di connection string
-- file:app.db?_busy_timeout=5000
```

### 7.3 Rollback Gagal

SQLite memiliki keterbatasan untuk rollback. Pastikan:

1. Gunakan `BEGIN TRANSACTION` di setiap migration
2. Test rollback di development dulu
3. Backup database sebelum production migration

---

## 8. Latihan Praktik

### 8.1 Setup Project

```bash
mkdir -p ~/projects/sqlite-migrations
cd ~/projects/sqlite-migrations
go mod init sqlite-migrations
go get github.com/golang-migrate/migrate/v4
go get github.com/mattn/go-sqlite3
```

### 8.2 Buat Migration

```bash
# Buat migration untuk tabel users
migrate create -ext sql -dir migrations -seq create_users_table

# Edit file .up.sql
# Tambahkan schema users table

# Edit file .down.sql
# Tambahkan perintah drop table
```

### 8.3 Jalankan Migration

```bash
# Apply migrations
migrate -path migrations -database "sqlite3://app.db" up

# Cek hasil
sqlite3 app.db ".tables"
sqlite3 app.db "SELECT * FROM users;"
```

---

## Ringkasan

| Topik | Kunci Utama |
|-------|-------------|
| Migrations | Version-controlled schema changes |
| golang-migrate | Tool standar untuk migrations di Go |
| SQLite | Perlu recreate table untuk alter |
| Idempotent | Gunakan `IF NOT EXISTS`, `OR IGNORE` |
| Auto migrate | Jalankan di startup aplikasi |

> 📝 **Next:** Lanjut ke materi #04 untuk MySQL connection & configuration.