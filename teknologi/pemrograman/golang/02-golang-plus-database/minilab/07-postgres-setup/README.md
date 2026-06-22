# Manual Test: Golang + PostgreSQL: Setup & JSONB

Folder ini berisi script Go untuk menguji semua konsep dari materi **07-postgres-setup.md**.

## Struktur File

```
07-postgres-setup/
├── docker-compose.yml   ← PostgreSQL container
├── main.go              ← Script testing
├── README.md            ← File ini
├── go.mod               ← Module definition + dependency pgx/v5
└── go.sum               ← Checksum (dibuat saat go mod tidy)
```

## Prasyarat

- Docker & Docker Compose
- Go 1.22+
- Port 5432 tidak dipakai oleh service lain

## Cara Menjalankan

### 1. Start PostgreSQL Container

```bash
cd 07-postgres-setup
docker compose up -d

# Tunggu sampai healthcheck lulus
docker compose ps
# Status harus "healthy"
```

### 2. Jalankan Test

```bash
cd 07-postgres-setup

# Install dependency
go mod tidy

# Jalankan langsung
go run main.go

# Atau build dulu
go build -o 07-test main.go && ./07-test
```

### 3. Cleanup

```bash
cd 07-postgres-setup
docker compose down -v    # -v untuk hapus volume juga
```

## Apa yang Ditest?

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Koneksi & Pool | `pgxpool` config, `WaitForPostgres`, pool stats, `gen_random_uuid()` |
| 2 | Create Table users | UUID PK, TIMESTAMPTZ, `CREATE TABLE IF NOT EXISTS` |
| 3 | Insert (RETURNING) | Insert 3 users, langsung dapat UUID + created_at via `RETURNING` |
| 4 | Query | `GetByID` dengan UUID, `List` dengan `ORDER BY created_at` |
| 5 | Update & Delete | Update name/email, delete, verifikasi `ErrNotFound` |
| 6 | Create Table products | Kolom JSONB + GIN index |
| 7 | Insert JSONB | Insert produk dengan metadata `{tags, attributes, stock}` |
| 8 | FindByTag (@>) | Query JSONB dengan operator `@>` (contains) |
| 9 | GetStock (->>) | Ambil nilai field `stock` dari JSONB via `metadata->>'stock'` |
| 10 | UpdateStock | Update nilai dalam JSONB via `jsonb_set` |
| 11 | KeyExists (?) | Cek keberadaan key di JSONB via operator `?` |
| 12 | List Products | Scan JSONB ke struct `ProductMeta` |
| 13 | Error Handling | Duplicate key (23505), ErrNotFound, rows affected = 0 |
| 14 | Pool Stats | `pool.Stat()` — Total, Acquired, Idle connections |

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PG_HOST` | `localhost` | Host PostgreSQL |
| `PG_PORT` | `5432` | Port PostgreSQL |
| `PG_USER` | `demo` | User PostgreSQL |
| `PG_PASSWORD` | `demo` | Password PostgreSQL |
| `PG_DB` | `golang_demo` | Database name |
| `PG_SSLMODE` | `disable` | SSL mode (disable untuk lokal) |

## Expected Output

```
=== MANUAL TEST MATERI 07: GOLANG + POSTGRESQL: SETUP & JSONB ===

PostgreSQL Config:
  Host:   localhost:5432
  DB:     golang_demo
  User:   demo
  SSL:    disable
  DSN:    postgres://demo:****@localhost:5432/golang_demo?sslmode=disable

⏳ Menunggu PostgreSQL siap (retry up to 10x, setiap 2 detik)...
   Pastikan Docker sudah running:
   cd 07-postgres-setup && docker compose up -d

✅ Berhasil konek ke PostgreSQL!

--- Test 1: Koneksi & Connection Pool ---
✅ PostgreSQL version: PostgreSQL 16.x on x86_64-pc-linux-musl...
✅ Pool stats: Total=2, InUse=0, Idle=2
✅ gen_random_uuid(): xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

--- Test 2: Create Table (users) ---
✅ Tabel 'users' berhasil dibuat (UUID PK, TIMESTAMPTZ)
✅ Verifikasi: tabel 'users' memiliki 4 kolom

--- Test 3: Insert Users (dengan RETURNING) ---
✅ Insert Alice: id=xxxxxxxx... created_at=2026-06-19 ...
✅ Insert Bob: id=xxxxxxxx... created_at=2026-06-19 ...
✅ Insert Charlie: id=xxxxxxxx... created_at=2026-06-19 ...
  💡 RETURNING clause langsung memberikan data baru tanpa query tambahan

... dan seterusnya ...

=== SEMUA TEST SELESAI ===
🎉 PostgreSQL: koneksi, UUID, TIMESTAMPTZ, CRUD, JSONB (@>, ->>, jsonb_set) sudah dipahami!
```

## Catatan Penting

- **WAJIB** jalankan `docker compose up -d` dahulu sebelum test.
- Driver yang digunakan: `pgx/v5` (`github.com/jackc/pgx/v5`) — bukan `lib/pq`.
- Placeholder PostgreSQL menggunakan `$1, $2, ...` — **bukan** `?` seperti MySQL.
- `RETURNING` clause memungkinkan mendapatkan data baru langsung dari INSERT/UPDATE.
- `TIMESTAMPTZ` lebih disarankan daripada `TIMESTAMP` untuk menghindari masalah timezone.
- `JSONB` lebih baik dari `JSON` karena tersimpan dalam format binary dan bisa diindex.
- `pgxpool` bersifat goroutine-safe — aman dipakai dari multiple goroutine sekaligus.
