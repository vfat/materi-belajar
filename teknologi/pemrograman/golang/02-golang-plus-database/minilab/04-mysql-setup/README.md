# Manual Test: Golang + MySQL: Connection & Config

Folder ini berisi script Go untuk menguji semua konsep dari materi **04-mysql-setup.md**.

## Struktur File

```
04-mysql-setup/
├── docker-compose.yml   ← MySQL container
├── main.go              ← Script testing
├── README.md            ← File ini
├── go.mod               ← Module definition + dependency
└── 04-test              ← Hasil compile (setelah build, opsional)
```

## Prasyarat

- Docker & Docker Compose
- Go 1.21+
- Port 3306 tidak dipakai oleh service lain

## Cara Menjalankan

### 1. Start MySQL Container

```bash
cd 04-mysql-setup
docker compose up -d

# Tunggu sampai healthcheck lulus
docker compose ps
# Status harus "healthy"
```

### 2. Jalankan Test

```bash
cd 04-mysql-setup

# Install dependency
go mod tidy

# Jalankan langsung
go run main.go

# Atau build dulu
go build -o 04-test main.go && ./04-test
```

### 3. Cleanup

```bash
cd 04-mysql-setup
docker compose down -v    # -v untuk hapus volume juga
```

## Apa yang Ditest?

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | DSN & Connection Pool | Config dengan `parseTime=true`, `MaxOpen`, `MaxIdle`, `MaxLife`, verifikasi MySQL version |
| 2 | Create Table | `CREATE TABLE IF NOT EXISTS users` + verifikasi di `information_schema` |
| 3 | Insert Users | Insert 3 users (Alice, Bob, Charlie) |
| 4 | Query Users | `GetByID`, `List` dengan `QueryRowContext` / `QueryContext` |
| 5 | Update User | `UPDATE` dengan rows affected check |
| 6 | Delete User | `DELETE` dengan verifikasi not found |
| 7 | Count | `SELECT COUNT(*)` |
| 8 | Error & Retry | `withRetry` helper + `ErrUserNotFound` sentinel error |
| 9 | Pool Stats | `db.Stats()` untuk monitor koneksi |
| 10 | Resource Cleanup | `defer db.Close()` + cleanup pattern |

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `MYSQL_USER` | `demo` | User MySQL |
| `MYSQL_PASSWORD` | `demo` | Password MySQL |
| `MYSQL_HOST` | `localhost` | Host MySQL |
| `MYSQL_PORT` | `3306` | Port MySQL |
| `MYSQL_DB` | `golang_demo` | Database name |

## Expected Output

```
=== MANUAL TEST MATERI 04: GOLANG + MYSQL: CONNECTION & CONFIG ===

MySQL Config:
  Host: localhost:3306
  DB:   golang_demo
  User: demo
  Pass: ****

⏳ Menunggu MySQL siap (retry up to 10x, setiap 2 detik)...

✅ Berhasil konek ke MySQL!

--- Test 1: DSN & Connection Pool ---
✅ Connection pool: MaxOpen=25, MaxIdle=5, MaxLife=5m0s
✅ MySQL version: 8.0.36

--- Test 2: Create Table ---
✅ Tabel 'users' berhasil dibuat
✅ Verifikasi: tabel 'users' terdaftar di information_schema

... dan seterusnya ...

=== SEMUA TEST SELESAI ===
🎉 MySQL connection, config, CRUD, retry, connection pool sudah dipahami!
```

## Catatan Penting

- **WAJIB** jalankan `docker compose up -d` dahulu sebelum test.
- Gunakan environment variables untuk config di production, bukan hardcode.
- `parseTime=true` penting agar `time.Time` ter-parse otomatis dari `TIMESTAMP` / `DATETIME`.
- `withRetry` pattern sangat berguna untuk menunggu service database siap.
- Selalu gunakan `QueryRowContext` / `ExecContext` dengan context untuk timeout & cancellation.
