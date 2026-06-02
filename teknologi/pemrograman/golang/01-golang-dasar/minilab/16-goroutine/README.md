# Manual Test: Goroutine & Channel

Folder ini berisi script Go untuk menguji konsep pada materi **16-goroutine.md**.

## Struktur File

```
16-goroutine/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 16-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/16-goroutine
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/16-goroutine

# Compile
go build -o 16-test main.go

# Jalankan
./16-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Menjalankan goroutine sederhana
- Unbuffered vs buffered channels
- Worker pool dengan `sync.WaitGroup`
- `select` dengan timeout
- Race condition demo dan penggunaan `sync.Mutex`

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 16: GOROUTINE & CHANNEL ===

--- Test 1: Simple goroutine and done channel ---
hello from goroutine
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/16-goroutine.md`
- 🌐 Dokumentasi: https://pkg.go.dev/runtime, https://pkg.go.dev/sync
