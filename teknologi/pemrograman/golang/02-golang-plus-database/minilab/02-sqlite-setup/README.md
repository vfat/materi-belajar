# Manual Test: Golang + SQLite: Setup & CRUD

Folder ini berisi script Go untuk menguji semua konsep dari materi **02-sqlite-setup.md**.

## Struktur File

```
02-sqlite-setup/
├── main.go          ← Script testing
├── README.md        ← File ini
├── go.mod           ← Module definition + dependency
└── 02-test          ← Hasil compile (setelah build, opsional)
```

## Cara Menjalankan

### Prasyarat

Pastikan kamu punya compiler C (gcc/clang) karena driver SQLite (`go-sqlite3`) menggunakan CGO.

### 1. Setup & Jalankan Langsung

```bash
cd 02-sqlite-setup

# Install dependency (jika belum)
go mod tidy

# Jalankan
go run main.go
```

### 2. Build & Jalankan

```bash
cd 02-sqlite-setup

go mod tidy

# Compile
go build -o 02-test main.go

# Jalankan
./02-test
```

### 3. Cleanup

Script akan otomatis menghapus file database sementara (`/tmp/minilab02-*.db`) setelah selesai.

## Apa yang Ditest?

File `main.go` berisi test cases untuk mendemonstrasikan:

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Koneksi SQLite | DSN dengan opsi (`_foreign_keys=1`, `_busy_timeout`, dll), pool config, PingContext |
| 2 | Init Schema | Membuat tabel categories + tasks + index |
| 3 | Category CRUD | Create, GetByID, GetByName, List, Update, Delete |
| 4 | Task CRUD | Create, GetByID, GetByIDWithCategory (join), List, ListByStatus, Update, UpdateStatus, Delete |
| 5 | Query Lanjutan | Count, CountByStatus, ListByCategory |
| 6 | Transaksi | BulkCreateTasks dan TransferTask menggunakan BeginTx + Commit/Rollback |
| 7 | Context & Error | Penggunaan context.WithTimeout, sentinel errors (ErrCategoryNotFound, ErrTaskNotFound), sql.ErrNoRows |
| 8 | Resource Cleanup | defer db.Close() + hapus file temp |

## Expected Output

Ketika dijalankan, program akan menampilkan:

```
=== MANUAL TEST MATERI 02: GOLANG + SQLITE: SETUP & CRUD ===

Menggunakan temporary DB: /tmp/minilab02-....db

--- Test 1: Koneksi SQLite ---
✅ Berhasil konek ke SQLite (file + options)
✅ PingContext sukses
DSN: file:/tmp/...db?_foreign_keys=1&_busy_timeout=5000&_journal_mode=WAL

--- Test 2: Init Schema ---
✅ Schema categories + tasks + indexes berhasil dibuat

--- Test 3: Category CRUD ---
✅ Create category: id=1 name=Work
✅ GetByID: Work
✅ GetByName: Work
✅ List: 3 categories
✅ Update category sukses
✅ Delete category sukses

--- Test 4: Task CRUD ---
✅ Create task: id=1 "Finish report"
✅ GetByIDWithCategory (join): task + category name
✅ List tasks: 5 tasks
✅ ListByStatus (pending): 3 tasks
✅ UpdateStatus: done
✅ Delete task sukses

... dan seterusnya ...

=== SEMUA TEST SELESAI ===
✅ Database ditutup dengan bersih
🎉 SQLite CRUD + repository pattern + transaction sudah dipahami!
```

## Catatan Penting

- Selalu pakai `?_foreign_keys=1` di DSN untuk enforce foreign key.
- Gunakan `context.Context` di semua operasi DB.
- Untuk testing: pakai temp file atau `:memory:` (tapi `:memory:` butuh perhatian untuk shared connection).
- Repository pattern memisahkan logic akses data dari business logic.
- Transaction penting untuk operasi yang melibatkan multiple statements (atomicity).
- `sql.NullString`, `sql.NullInt64`, `sql.NullTime` untuk handle kolom yang boleh NULL.

## Referensi

- 📖 Materi: `bahan/materi-belajar/teknologi/pemrograman/golang/02-golang-plus-database/02-sqlite-setup.md`
- 🌐 SQLite docs: https://www.sqlite.org/docs.html
- 🐘 Driver: https://github.com/mattn/go-sqlite3
