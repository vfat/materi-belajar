# Manual Test: Error Handling

Folder ini berisi script Go untuk menguji semua konsep dari materi **10-error-handling.md**.

## Struktur File

```
10-error-handling/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 10-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/10-error-handling
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/10-error-handling

# Compile
go build -o 10-test main.go

# Jalankan
./10-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Pembuatan dan wrapping error (fmt.Errorf)
- Multiple-return functions mengembalikan error
- Sentinel error dan `errors.Is`
- Custom error type dan `errors.As`
- Panic & recover pattern
- Error dari operasi OS (os.Open)
- Error wrapping/unwrapping

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 10: ERROR HANDLING ===

--- Test 1: Basic error & wrapping ---
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/10-error-handling.md`
- 🌐 Dokumentasi: https://go.dev/doc/
