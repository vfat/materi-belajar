---
topik: Golang + PostgreSQL Setup & JSONB
urutan: 7 dari 20
posisi: lanjutan
sebelumnya: MySQL Migrations & Environment Sync
prerequisites:
  - MySQL Migrations & Environment Sync (06-mysql-migration)
  - Golang + MySQL Connection & Config (04-mysql-setup)
level: menengah
---

> 🔗 **Lanjutan dari:** MySQL Migrations & Environment Sync
> ← Kembali ke: `06-mysql-migration.md`

# Golang + PostgreSQL: Setup & JSONB

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Menjalankan PostgreSQL via Docker Compose
- Menghubungkan Go ke PostgreSQL menggunakan driver `pgx` (dan kompatibilitas `database/sql`)
- Memahami perbedaan DSN/URL PostgreSQL vs MySQL
- Mengonfigurasi connection pool, timeout, dan SSL mode
- Melakukan CRUD standar dengan PostgreSQL
- Menyimpan dan mengquery data dengan tipe **JSONB**
- Menangani tipe khas PostgreSQL: `UUID`, `ARRAY`, `timestamptz`

---

## 1. Mengapa PostgreSQL?

PostgreSQL adalah RDBMS open-source yang kaya fitur — sering disebut *"the most advanced open source database"*.

| Fitur | MySQL | PostgreSQL |
|---|---|---|
| JSONB (binary JSON + index) | JSON terbatas | ✅ Native, bisa diindex |
| Array type native | ❌ | ✅ `INTEGER[]`, `TEXT[]` |
| UUID native | ❌ | ✅ `gen_random_uuid()` |
| Full Text Search | Terbatas | ✅ `tsvector`, `tsquery` |
| Window Functions | Sebagian | ✅ Lengkap |
| CTE (Common Table Expressions) | Sebagian | ✅ Lengkap + recursive |
| Transactional DDL | ❌ | ✅ (`ALTER TABLE` bisa di-rollback) |
| Partial Index | ❌ | ✅ |
| Listen/Notify | ❌ | ✅ Event push ke client |

> 💡 **Pilih PostgreSQL** untuk aplikasi yang butuh query kompleks, tipe data kaya, atau fleksibilitas schema (JSONB).

---

## 2. Setup Docker Compose

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    container_name: golang_postgres_demo
    restart: unless-stopped
    environment:
      POSTGRES_USER: demo
      POSTGRES_PASSWORD: demo
      POSTGRES_DB: golang_demo
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U demo -d golang_demo"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
    driver: local
```

```bash
# Jalankan
docker compose up -d

# Cek status (tunggu healthy)
docker compose ps

# Masuk ke psql (opsional, untuk verifikasi)
docker exec -it golang_postgres_demo psql -U demo -d golang_demo
```

---

## 3. Driver: pgx vs lib/pq

Ada dua driver utama untuk Go + PostgreSQL:

| Driver | Import | Kelebihan | Digunakan |
|---|---|---|---|
| `pgx/v5` | `github.com/jackc/pgx/v5` | Native, performa tinggi, semua fitur PG | Direkomendasikan |
| `lib/pq` | `github.com/lib/pq` | Lama, `database/sql` compat | Legacy, kurang aktif |

Kita pakai **`pgx/v5`** — bisa digunakan langsung (`pgx.Connect`) atau via adapter `database/sql` (`pgxpool`, `pgx/stdlib`).

```bash
go get github.com/jackc/pgx/v5
```

---

## 4. Koneksi & DSN

### 4.1 Format DSN / Connection URL

```
postgres://user:password@host:port/dbname?sslmode=disable
```

| Parameter | Contoh | Keterangan |
|---|---|---|
| `sslmode=disable` | lokal/docker | Matikan TLS (hanya untuk dev) |
| `sslmode=require` | production | Wajib TLS |
| `connect_timeout=5` | | Timeout koneksi (detik) |
| `pool_max_conns=10` | pgxpool | Max koneksi pool |

### 4.2 Koneksi dengan `pgxpool` (Direkomendasikan)

```go
package db

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// Config menyimpan konfigurasi koneksi PostgreSQL
type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
    MaxConns int32
    MinConns int32
    MaxLife  time.Duration
}

func DefaultConfig() Config {
    return Config{
        Host:     getEnv("PG_HOST", "localhost"),
        Port:     getEnvInt("PG_PORT", 5432),
        User:     getEnv("PG_USER", "demo"),
        Password: getEnv("PG_PASSWORD", "demo"),
        DBName:   getEnv("PG_DB", "golang_demo"),
        SSLMode:  getEnv("PG_SSLMODE", "disable"),
        MaxConns: 10,
        MinConns: 2,
        MaxLife:  30 * time.Minute,
    }
}

func (c Config) DSN() string {
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s",
        c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
    )
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
    poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
    if err != nil {
        return nil, fmt.Errorf("parse dsn: %w", err)
    }

    poolCfg.MaxConns = cfg.MaxConns
    poolCfg.MinConns = cfg.MinConns
    poolCfg.MaxConnLifetime = cfg.MaxLife

    pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    // Ping untuk verifikasi
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("ping postgres: %w", err)
    }

    return pool, nil
}
```

### 4.3 Retry Helper

```go
func WaitForPostgres(ctx context.Context, cfg Config, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
    var pool *pgxpool.Pool
    var lastErr error

    for i := 0; i < attempts; i++ {
        pool, lastErr = NewPool(ctx, cfg)
        if lastErr == nil {
            return pool, nil
        }
        log.Printf("⚠️  Attempt %d/%d: %v (retry in %v)", i+1, attempts, lastErr, delay)
        time.Sleep(delay)
    }
    return nil, fmt.Errorf("postgres tidak tersedia setelah %d percobaan: %w", attempts, lastErr)
}
```

---

## 5. CRUD Dasar

### 5.1 Create Table

```go
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(200) UNIQUE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

func CreateTable(ctx context.Context, pool *pgxpool.Pool) error {
    _, err := pool.Exec(ctx, createUsersTable)
    return err
}
```

> 💡 **Perhatikan perbedaan dari MySQL:**
> - `UUID` sebagai primary key (bukan `INT AUTO_INCREMENT`)
> - `gen_random_uuid()` fungsi bawaan PostgreSQL 13+
> - `TIMESTAMPTZ` — timestamp dengan timezone (rekomendasi vs `TIMESTAMP`)

### 5.2 Insert

```go
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

func InsertUser(ctx context.Context, pool *pgxpool.Pool, name, email string) (*User, error) {
    u := &User{}
    err := pool.QueryRow(ctx,
        `INSERT INTO users (name, email) VALUES ($1, $2)
         RETURNING id, name, email, created_at`,
        name, email,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("insert user: %w", err)
    }
    return u, nil
}
```

> 💡 **`RETURNING`** — PostgreSQL mendukung `RETURNING` clause untuk langsung mengambil data yang baru di-insert/update. Sangat berguna vs MySQL yang harus `LastInsertId()` lalu query lagi.

### 5.3 Query

```go
var ErrNotFound = errors.New("not found")

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id string) (*User, error) {
    u := &User{}
    err := pool.QueryRow(ctx,
        `SELECT id, name, email, created_at FROM users WHERE id = $1`, id,
    ).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return u, nil
}

func ListUsers(ctx context.Context, pool *pgxpool.Pool) ([]*User, error) {
    rows, err := pool.Query(ctx,
        `SELECT id, name, email, created_at FROM users ORDER BY created_at`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        u := &User{}
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}
```

> ⚠️ **Placeholder PostgreSQL:** Gunakan `$1, $2, ...` (bukan `?` seperti MySQL).

### 5.4 Update & Delete

```go
func UpdateUser(ctx context.Context, pool *pgxpool.Pool, id, name, email string) error {
    res, err := pool.Exec(ctx,
        `UPDATE users SET name = $1, email = $2 WHERE id = $3`,
        name, email, id,
    )
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }
    if res.RowsAffected() == 0 {
        return ErrNotFound
    }
    return nil
}

func DeleteUser(ctx context.Context, pool *pgxpool.Pool, id string) error {
    res, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }
    if res.RowsAffected() == 0 {
        return ErrNotFound
    }
    return nil
}
```

---

## 6. JSONB: Menyimpan & Query Data Fleksibel

**JSONB** adalah tipe data PostgreSQL yang menyimpan JSON dalam format binary, mendukung indexing, dan operator query khusus.

### 6.1 Kapan Pakai JSONB?

| Kasus | Pakai JSONB? |
|---|---|
| Atribut produk yang berbeda-beda per kategori | ✅ Ya |
| Data konfigurasi/settings user | ✅ Ya |
| Event log dengan schema fleksibel | ✅ Ya |
| Data relasional terstruktur | ❌ Pakai kolom biasa |
| Data yang sering di-JOIN | ❌ Pakai kolom biasa |

### 6.2 Tabel dengan Kolom JSONB

```sql
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    price       NUMERIC(12, 2) NOT NULL,
    metadata    JSONB,          -- data fleksibel: tags, attributes, dll
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index GIN untuk JSONB (mempercepat query @>, ?, ?|, ?&)
CREATE INDEX IF NOT EXISTS idx_products_metadata ON products USING GIN (metadata);
```

### 6.3 Struct Go untuk JSONB

```go
import "encoding/json"

type ProductMeta struct {
    Tags       []string          `json:"tags,omitempty"`
    Attributes map[string]string `json:"attributes,omitempty"`
    Stock      int               `json:"stock,omitempty"`
}

type Product struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    Price     float64     `json:"price"`
    Metadata  ProductMeta `json:"metadata"`
    CreatedAt time.Time   `json:"created_at"`
}
```

### 6.4 Insert dengan JSONB

```go
func InsertProduct(ctx context.Context, pool *pgxpool.Pool, name string, price float64, meta ProductMeta) (*Product, error) {
    metaJSON, err := json.Marshal(meta)
    if err != nil {
        return nil, fmt.Errorf("marshal metadata: %w", err)
    }

    p := &Product{}
    var metaRaw []byte
    err = pool.QueryRow(ctx,
        `INSERT INTO products (name, price, metadata)
         VALUES ($1, $2, $3)
         RETURNING id, name, price, metadata, created_at`,
        name, price, metaJSON,
    ).Scan(&p.ID, &p.Name, &p.Price, &metaRaw, &p.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("insert product: %w", err)
    }

    if err := json.Unmarshal(metaRaw, &p.Metadata); err != nil {
        return nil, fmt.Errorf("unmarshal metadata: %w", err)
    }
    return p, nil
}
```

### 6.5 Query dengan Operator JSONB

PostgreSQL punya operator khusus untuk query JSONB:

| Operator | Arti | Contoh |
|---|---|---|
| `->` | Ambil nilai JSON (sebagai JSON) | `metadata->'tags'` |
| `->>` | Ambil nilai JSON (sebagai text) | `metadata->>'stock'` |
| `@>` | Contains (JSONB mengandung...) | `metadata @> '{"tags":["sale"]}'` |
| `?` | Key exists? | `metadata ? 'tags'` |
| `?|` | Salah satu key exists? | `metadata ?| array['a','b']` |
| `#>>` | Ambil nilai nested (sebagai text) | `metadata #>> '{attributes,color}'` |

```go
// Cari produk yang punya tag tertentu (pakai @>)
func FindByTag(ctx context.Context, pool *pgxpool.Pool, tag string) ([]*Product, error) {
    query := `
        SELECT id, name, price, metadata, created_at
        FROM products
        WHERE metadata @> $1::jsonb
        ORDER BY name`

    // Build filter JSON: {"tags": ["tag-yang-dicari"]}
    filter, _ := json.Marshal(map[string][]string{"tags": {tag}})

    rows, err := pool.Query(ctx, query, filter)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []*Product
    for rows.Next() {
        p := &Product{}
        var metaRaw []byte
        if err := rows.Scan(&p.ID, &p.Name, &p.Price, &metaRaw, &p.CreatedAt); err != nil {
            return nil, err
        }
        json.Unmarshal(metaRaw, &p.Metadata)
        products = append(products, p)
    }
    return products, rows.Err()
}

// Query nilai dari field JSONB sebagai text
func GetStockByID(ctx context.Context, pool *pgxpool.Pool, id string) (int, error) {
    var stock int
    err := pool.QueryRow(ctx,
        `SELECT (metadata->>'stock')::int FROM products WHERE id = $1`, id,
    ).Scan(&stock)
    return stock, err
}

// Update nilai dalam JSONB (jsonb_set)
func UpdateStock(ctx context.Context, pool *pgxpool.Pool, id string, newStock int) error {
    _, err := pool.Exec(ctx,
        `UPDATE products
         SET metadata = jsonb_set(metadata, '{stock}', $1::text::jsonb)
         WHERE id = $2`,
        newStock, id,
    )
    return err
}
```

---

## 7. Tipe Data Khas PostgreSQL

### 7.1 UUID

```go
// PostgreSQL: gen_random_uuid() (PG 13+)
// Atau: uuid_generate_v4() jika pakai extension uuid-ossp

// Di Go, scan UUID sebagai string (pgx handle otomatis)
var id string
pool.QueryRow(ctx, `SELECT gen_random_uuid()`).Scan(&id)
```

### 7.2 ARRAY

```go
// Kolom: tags TEXT[]
// Insert array:
pool.Exec(ctx,
    `INSERT INTO items (tags) VALUES ($1)`,
    []string{"golang", "postgresql", "backend"},
)

// Scan array:
var tags []string
pool.QueryRow(ctx, `SELECT tags FROM items WHERE id = $1`, id).Scan(&tags)
```

### 7.3 TIMESTAMPTZ vs TIMESTAMP

```go
// Selalu gunakan TIMESTAMPTZ untuk menyimpan waktu dengan timezone
// pgx otomatis scan ke time.Time dengan timezone

var createdAt time.Time
pool.QueryRow(ctx, `SELECT created_at FROM users WHERE id = $1`, id).Scan(&createdAt)

// Untuk insert: kirim time.Time, pgx akan otomatis convert
pool.Exec(ctx, `INSERT INTO events (happened_at) VALUES ($1)`, time.Now())
```

### 7.4 NUMERIC / Decimal

```go
// Untuk money/harga, hindari FLOAT — pakai NUMERIC(12,2)
// pgx scan ke float64 atau string (untuk presisi penuh)
var price float64
pool.QueryRow(ctx, `SELECT price FROM products WHERE id = $1`, id).Scan(&price)
```

---

## 8. Error Handling di pgx

```go
import (
    "errors"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

func handlePgError(err error) {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            fmt.Println("Duplicate key:", pgErr.Detail)
        case "23503": // foreign_key_violation
            fmt.Println("Foreign key violation:", pgErr.Detail)
        case "23502": // not_null_violation
            fmt.Println("Not null violation:", pgErr.ColumnName)
        default:
            fmt.Printf("PG Error %s: %s\n", pgErr.Code, pgErr.Message)
        }
    }
    if errors.Is(err, pgx.ErrNoRows) {
        fmt.Println("Record not found")
    }
}
```

> 💡 **Kode error PostgreSQL** (SQLSTATE) lengkap di: https://www.postgresql.org/docs/current/errcodes-appendix.html

---

## 9. Connection Pool Stats

```go
// pgxpool.Pool menyediakan stats koneksi
stats := pool.Stat()
fmt.Printf("Total: %d, InUse: %d, Idle: %d\n",
    stats.TotalConns(), stats.AcquiredConns(), stats.IdleConns())
```

---

## 10. Perbandingan pgx vs MySQL Driver

| Aspek | pgx/v5 | go-sql-driver/mysql |
|---|---|---|
| `database/sql` compat | Opsional (via stdlib adapter) | Ya (default) |
| Native types | UUID, JSONB, ARRAY | Terbatas |
| Placeholder | `$1, $2, ...` | `?` |
| RETURNING | ✅ | ❌ |
| Batch operations | ✅ `pgxpool.SendBatch` | ❌ |
| Copy protocol | ✅ `pgx.CopyFrom` (bulk insert) | ❌ |
| `ErrNoRows` | `pgx.ErrNoRows` | `sql.ErrNoRows` |

---

## 11. Checklist

- [x] Docker Compose + PostgreSQL 16 berjalan.
- [x] Koneksi via `pgxpool` dengan retry.
- [x] CRUD dengan UUID primary key dan `RETURNING`.
- [x] CRUD dengan kolom JSONB + operator `@>`, `->>`.
- [x] Insert/update nilai dalam JSONB (`jsonb_set`).
- [x] Cari produk berdasarkan konten JSONB.
- [x] Error handling PostgreSQL (SQLSTATE codes).
- [x] Connection pool stats.

---

## 12. Catatan & Gotchas

- **Placeholder `$1` bukan `?`** — kesalahan paling umum saat migrasi dari MySQL.
- **Selalu pakai `TIMESTAMPTZ`** bukan `TIMESTAMP` untuk menghindari masalah timezone.
- **`rows.Err()`** — selalu cek setelah loop `rows.Next()`.
- **JSONB vs JSON** — selalu pakai `JSONB` (bisa diindex), bukan `JSON` (disimpan as-is text).
- **UUID** — lebih baik dari auto-increment integer untuk distributed system (tidak bisa ditebak, aman untuk expose ke client).
- **`pgxpool`** bersifat *goroutine-safe* — aman digunakan dari multiple goroutine.

---

## 13. Referensi

- https://github.com/jackc/pgx — dokumentasi pgx v5
- https://www.postgresql.org/docs/current/datatype-json.html — JSONB di PostgreSQL
- https://www.postgresql.org/docs/current/functions-json.html — fungsi & operator JSON
- https://hub.docker.com/_/postgres — Docker image PostgreSQL
- https://www.postgresql.org/docs/current/errcodes-appendix.html — SQLSTATE error codes

---

> ⏭️ **Selanjutnya:** `08-postgres-advanced.md` — Advanced Queries: CTE, Window Functions, Full Text Search
