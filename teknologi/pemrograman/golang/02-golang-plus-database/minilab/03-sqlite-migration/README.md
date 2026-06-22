# Manual Test: SQLite Migrations & Schema

Folder ini berisi script Go untuk menguji semua konsep dari materi **03-sqlite-migration.md**.

## Struktur File

```
03-sqlite-migration/
├── main.go          ← Script testing
├── README.md        ← File ini
├── go.mod           ← Module definition + dependency
└── 03-test          ← Hasil compile (setelah build, opsional)
```

## Cara Menjalankan

### Prasyarat

Pastikan kamu punya compiler C (gcc/clang) karena driver SQLite (`go-sqlite3`) menggunakan CGO.

### 1. Setup & Jalankan Langsung

```bash
cd 03-sqlite-migration

# Install dependency (jika belum)
go mod tidy

# Jalankan
go run main.go
```

### 2. Build & Jalankan

```bash
cd 03-sqlite-migration

go mod tidy

# Compile
go build -o 03-test main.go

# Jalankan
./03-test
```

### 3. Cleanup

Script akan otomatis menghapus file database sementara (`/tmp/minilab03-*.db`) setelah selesai.

## Apa yang Ditest?

File `main.go` berisi test cases untuk mendemonstrasikan:

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Koneksi Database | SQLite connection dengan DSN + pool config |
| 2 | Migration 000001 | Membuat tabel `categories` + index (`CREATE TABLE IF NOT EXISTS`) |
| 3 | Migration 000002 | Membuat tabel `tasks` + 3 indexes |
| 4 | Migration 000003 | Seed data kategori (Work, Personal, Learning) via `INSERT OR IGNORE` |
| 5 | Insert Data Demo | Menambahkan 3 tasks untuk persiapan alter table |
| 6 | Migration 000004 (Alter) | Menambah kolom `priority` via recreate table + data migration |
| 7 | Rollback 000004 | Menghapus kolom `priority` via recreate table |
| 8 | Re-apply 000004 | Menambah kembali kolom `priority` |
| 9 | Rollback Seed | Menghapus data seed kategori |
| 10 | Rollback All | Drop semua tabel + indexes (clean state) |
| 11 | Idempotent | Membuktikan `IF NOT EXISTS` dan `OR IGNORE` aman dijalankan 2x |
| 12 | Version Tracking | Simulasi tabel `schema_migrations`, dirty state, force recovery |
| 13 | PRAGMA Checks | Cek `journal_mode` (WAL) dan `foreign_keys` |
| 14 | Resource Cleanup | `defer db.Close()` + hapus file temp |

## Expected Output

Ketika dijalankan, program akan menampilkan:

```
=== MANUAL TEST MATERI 03: SQLITE MIGRATIONS & SCHEMA ===

Menggunakan temporary DB: /tmp/minilab03-....db

--- Test 1: Koneksi Database ---
✅ Berhasil konek ke SQLite

--- Test 2: Apply Migration 000001 (Create Categories Table) ---
✅ Migration 000001 up sukses (tabel categories + index dibuat)
✅ Verifikasi: tabel categories terdaftar di sqlite_master

--- Test 3: Apply Migration 000002 (Create Tasks Table) ---
✅ Migration 000002 up sukses (tabel tasks + indexes dibuat)
✅ Tabel di database: [categories tasks]

--- Test 4: Seed Data via Migration 000003 ---
✅ Migration 000003 seed sukses (3 kategori ditambahkan)
✅ Kategori di database: 3
✅ Work category ID: 1, Personal category ID: 2

--- Test 5: Insert Data Demo ---
✅ 3 tasks berhasil diinsert
📋 Tasks Sebelum migration 000004: ...

--- Test 6: Migration 000004 - Add Priority Column ---
✅ Migration 000004 up sukses (kolom priority ditambahkan via recreate table)
✅ Verifikasi: kolom priority ada di tabel tasks

--- Test 7: Rollback Migration 000004 ---
✅ Rollback 000004 sukses (kolom priority dihapus)
✅ Verifikasi: kolom priority sudah tidak ada

... dan seterusnya ...

--- Test 12: Migration Version Tracking ---
✅ Current migration version: 4
✅ Applied versions: [1 2 3 4]
✅ Dirty state detected for version 5: true
✅ Force recovery: version 5 dirty = false (recovered)

--- Test 13: PRAGMA Checks ---
✅ Journal mode: wal
✅ Foreign keys: 1

=== SEMUA TEST SELESAI ===
🎉 SQLite migrations, schema versioning, alter table via recreate,
   rollback, idempotent migrations, dirty state recovery sudah dipahami!
```

## Catatan Penting

- SQLite **tidak mendukung** `ALTER COLUMN`. Untuk alter schema, gunakan recreate table + data migration.
- Selalu gunakan `CREATE TABLE IF NOT EXISTS` dan `INSERT OR IGNORE` untuk idempotent migrations.
- Di production, gunakan tool seperti `golang-migrate` untuk version control.
- Backup database sebelum menjalankan migration di production.
- `PRAGMA journal_mode = WAL` meningkatkan concurrency untuk read/write.
- `PRAGMA foreign_keys = ON` harus diaktifkan setiap koneksi (via DSN atau pragma).
