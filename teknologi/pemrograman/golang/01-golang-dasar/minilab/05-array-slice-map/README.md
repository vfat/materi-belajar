# Manual Test: Array, Slice & Map

Folder ini berisi script Go untuk menguji semua konsep dari materi **05-array-slice-map.md**.

## Struktur File

```
05-array-slice-map/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 05-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/05-array-slice-map
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/05-array-slice-map

# Compile
go build -o 05-test main.go

# Jalankan
./05-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Array (akses, len)
- Slice (make, append, len/cap, sub-slice, remove, insert, copy)
- Map (create, add, get, delete, iterate, sorted iteration)
- Mencari nilai max/min dari slice

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 05: ARRAY, SLICE & MAP ===

--- Test 1: Array ---
arr: [10 20 30 40 50]
arr[0]: 10 arr[4]: 50
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/05-array-slice-map.md`
- 🌐 Dokumentasi: https://go.dev/doc/
