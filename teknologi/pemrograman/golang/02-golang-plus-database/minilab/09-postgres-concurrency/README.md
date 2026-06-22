# Manual Test: PostgreSQL Concurrency & Locking

Folder ini berisi script Go untuk menguji semua konsep dari materi **09-postgres-concurrency.md**.

## Struktur File

```
09-postgres-concurrency/
├── docker-compose.yml   ← PostgreSQL container
├── main.go              ← Script testing (8 test scenario)
├── README.md            ← File ini
├── go.mod               ← Module + pgx/v5
└── go.sum               ← Checksum
```

## Prasyarat

- Docker & Docker Compose
- Go 1.22+
- Port 5432 tidak dipakai

## Cara Menjalankan

```bash
cd 09-postgres-concurrency
docker compose up -d

go mod tidy
go run main.go

# Atau build:
go build -o 09-test main.go && ./09-test

# Cleanup:
docker compose down -v
```

## Apa yang Ditest?

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Transaksi Dasar | Transfer saldo Alice→Bob dengan `withTx`, rollback otomatis saat saldo tidak cukup |
| 2 | Concurrent Transfers | 10 goroutine transfer bersamaan, `withRetryTx` untuk deadlock retry, verifikasi total saldo konsisten |
| 3 | Optimistic Locking | Kolom `version`: simulasi conflict (2 goroutine baca versi sama), retry otomatis dengan `updateStockWithRetry` |
| 4 | Pessimistic Locking | `SELECT FOR UPDATE`: 5 goroutine concurrent update stok, tidak ada race condition |
| 5 | Job Queue SKIP LOCKED | 10 jobs dienqueue, 3 worker concurrent claim+process via `FOR UPDATE SKIP LOCKED` |
| 6 | Advisory Locks | `pg_try_advisory_lock`: hanya 1 koneksi bisa dapat lock, koneksi lain ditolak |
| 7 | Savepoint | Batch insert dengan duplikat: yang duplikat di-rollback ke savepoint, sisanya tetap masuk |
| 8 | Isolation Levels | Demo perbedaan `READ COMMITTED` (non-repeatable read) vs `REPEATABLE READ` (snapshot) |

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PG_HOST` | `localhost` | Host PostgreSQL |
| `PG_PORT` | `5432` | Port PostgreSQL |
| `PG_USER` | `demo` | User |
| `PG_PASSWORD` | `demo` | Password |
| `PG_DB` | `golang_demo` | Database |
| `PG_SSLMODE` | `disable` | SSL mode |

## Catatan Penting

- **`defer tx.Rollback(ctx)`** aman dipanggil setelah `Commit()` — pgx mengabaikannya (no-op).
- **Consistent lock order** (`ORDER BY id FOR UPDATE`) mencegah deadlock di concurrent transfer.
- **Optimistic vs Pessimistic**: optimistic cocok untuk read-heavy (jarang konflik), pessimistic cocok untuk write-heavy (sering konflik).
- **`SKIP LOCKED`** memungkinkan multiple worker ambil job tanpa saling blokir.
- **Advisory lock** adalah lock level aplikasi — tidak ada hubungan dengan baris/tabel, berguna untuk koordinasi proses (cron, scheduler).
- Test 8 mendemonstrasikan non-repeatable read di `READ COMMITTED` dan snapshot isolation di `REPEATABLE READ`.
