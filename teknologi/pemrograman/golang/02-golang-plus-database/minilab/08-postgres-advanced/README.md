# Manual Test: PostgreSQL Advanced Queries

Folder ini berisi script Go untuk menguji semua konsep dari materi **08-postgres-advanced.md**.

## Struktur File

```
08-postgres-advanced/
├── docker-compose.yml   ← PostgreSQL container
├── main.go              ← Script testing (schema setup + 11 test)
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
cd 08-postgres-advanced
docker compose up -d

# Tunggu sampai healthy
docker compose ps
```

### 2. Jalankan Test

```bash
cd 08-postgres-advanced

# Install dependency
go mod tidy

# Jalankan
go run main.go

# Atau build dulu
go build -o 08-test main.go && ./08-test
```

### 3. Cleanup

```bash
cd 08-postgres-advanced
docker compose down -v
```

## Apa yang Ditest?

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | CTE Dasar | `WITH sales_summary AS (...), ranked AS (...)` — top sellers dengan multi-step CTE |
| 2 | Recursive CTE | `WITH RECURSIVE category_tree` — traversal hierarki kategori 3 level |
| 3 | ROW_NUMBER / RANK / DENSE_RANK | Window function ranking produk per kategori |
| 4 | LAG & LEAD | Price history dengan perubahan harga antar baris |
| 5 | Running Total | `SUM OVER (ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)` |
| 6 | NTILE | Bagi gaji karyawan ke 4 kuartil |
| 7 | Full Text Search | `tsvector`, `websearch_to_tsquery`, `ts_rank`, GIN index |
| 8 | FTS Highlight | `ts_headline` — snippet dengan kata kunci di-highlight |
| 9 | Bulk Insert CopyFrom | Insert 100 produk sekaligus via PostgreSQL COPY protocol |
| 10 | Batch Insert SendBatch | `pgx.Batch` + `pool.SendBatch` — 3 insert dalam 1 round-trip |
| 11 | EXPLAIN ANALYZE | Baca query plan dari Go, cek apakah index terpakai |

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PG_HOST` | `localhost` | Host PostgreSQL |
| `PG_PORT` | `5432` | Port PostgreSQL |
| `PG_USER` | `demo` | User PostgreSQL |
| `PG_PASSWORD` | `demo` | Password PostgreSQL |
| `PG_DB` | `golang_demo` | Database name |
| `PG_SSLMODE` | `disable` | SSL mode |

## Catatan Penting

- Schema dan data di-setup otomatis saat `go run main.go` dijalankan.
- `CopyFrom` membutuhkan `pool.Acquire()` untuk mendapat koneksi dedicated.
- `SendBatch` hasil harus dikonsumsi berurutan sesuai urutan query dalam batch.
- FTS menggunakan kolom `search_vec` (GENERATED ALWAYS AS ... STORED) — di-update otomatis.
- EXPLAIN ANALYZE dengan data kecil mungkin masih pakai Seq Scan (normal untuk < ~100 baris).
- Placeholder PostgreSQL: `$1, $2, ...` — bukan `?`.
