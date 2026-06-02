# Manual Test: Time & Date

Folder ini berisi script Go untuk menguji konsep pada materi **12-time-date.md**.

## Struktur File

```
12-time-date/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 12-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/12-time-date
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/12-time-date

# Compile
go build -o 12-test main.go

# Jalankan
./12-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- `time.Now()` dan format RFC3339
- Parsing waktu dengan layout custom
- `time.ParseDuration` dan operasi add/sub
- Timezone / `time.LoadLocation`
- Unix timestamps (`Unix`, `UnixNano`)
- Common layouts (`ANSIC`, `RFC1123Z`)

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 12: TIME & DATE ===

--- Test 1: Now & RFC3339 ---
Now (RFC3339): 2026-05-30T14:00:00Z
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/12-time-date.md`
- 🌐 Dokumentasi: https://pkg.go.dev/time
