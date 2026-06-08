# Manual Test: Fondasi Database di Go

Folder ini berisi script Go untuk menguji semua konsep dari materi **01-db-fundamentals.md**.

## Struktur File

```
01-db-fundamentals/
├── main.go          ← Script testing
├── README.md        ← File ini
├── go.mod           ← Module definition + dependency
└── 01-test          ← Hasil compile (setelah build, opsional)
```

## Cara Menjalankan

### Prasyarat

Pastikan kamu punya compiler C (gcc/clang) karena driver SQLite (`go-sqlite3`) menggunakan CGO.

### 1. Setup & Jalankan Langsung

```bash
cd 01-db-fundamentals

# Install dependency (jika belum)
go mod tidy

# Jalankan
go run main.go
```

### 2. Build & Jalankan

```bash
cd 01-db-fundamentals

go mod tidy

# Compile
go build -o 01-test main.go

# Jalankan
./01-test
```

### 3. Cleanup (opsional)

Script akan otomatis menghapus file database sementara (`/tmp/minilab01-*.db`) setelah selesai.

## Apa yang Ditest?

File `main.go` berisi test cases untuk mendemonstrasikan konsep kunci:

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Koneksi & Config | Pattern `New(cfg)` dengan pool config + PingContext |
| 2 | DSN | Contoh DSN untuk SQLite, Postgres, MySQL |
| 3 | Context & Timeout | QueryRowContext, context.WithTimeout, cancellation |
| 4 | Connection Pooling | SetMaxOpenConns, SetMaxIdleConns, db.Stats() |
| 5 | Error Handling | sql.ErrNoRows → sentinel error (ErrUserNotFound), errors.Is |
| 6 | Transaksi | BeginTx + defer Rollback + Commit pattern |
| 7 | CRUD Lengkap | Create, GetByID, List, Update, Delete dengan context |
| 8 | Resource Cleanup | defer db.Close(), graceful pattern |

## Expected Output

Ketika dijalankan, program akan menampilkan:

```
=== MANUAL TEST MATERI 01: FONDASI DATABASE DI GO ===

--- Test 1: Koneksi Database dengan Pattern Benar ---
✅ Berhasil konek ke SQLite
✅ PingContext sukses
Pool config: MaxOpen=25, MaxIdle=5, MaxLifetime=5m0s

--- Test 2: DSN Examples ---
SQLite DSN: file:app.db?_foreign_keys=1
Postgres DSN: postgres://user:pass@localhost:5432/mydb?sslmode=disable
MySQL DSN: user:pass@tcp(localhost:3306)/mydb?parseTime=true

--- Test 3: Context dalam Database Operations ---
✅ Query dengan context sukses
✅ Timeout context bekerja (simulasi)

... dan seterusnya ...

=== SEMUA TEST SELESAI ===
✅ Database ditutup dengan bersih
```

## Catatan Penting

- **SELALU** gunakan `Context` pada `QueryContext`, `ExecContext`, `QueryRowContext`, `BeginTx` dll.
- Jangan hardcode DSN di production — gunakan env vars.
- Konfigurasi pool **sebelum** digunakan untuk query.
- Gunakan `errors.Is(err, sql.ErrNoRows)` atau custom sentinel error.
- Pattern transaksi: `BeginTx` → `defer Rollback` → `tx = nil` setelah Commit sukses.
- Untuk graceful shutdown di server: `srv.Shutdown(ctx)` lalu `db.Close()`.

## Referensi

- 📖 Materi: `bahan/materi-belajar/teknologi/pemrograman/golang/02-golang-plus-database/01-db-fundamentals.md`
- 🌐 database/sql: https://pkg.go.dev/database/sql
- 🐘 Driver SQLite: https://github.com/mattn/go-sqlite3
