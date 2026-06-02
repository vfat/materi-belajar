# Manual Test: File I/O

Folder ini berisi script Go untuk menguji konsep pada materi **14-file-io.md**.

## Struktur File

```
14-file-io/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 14-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/14-file-io
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/14-file-io

# Compile
go build -o 14-test main.go

# Jalankan
./14-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Menulis file (`os.WriteFile`)
- Membaca file (`os.ReadFile`)
- Append file (`os.OpenFile` + `O_APPEND`)
- Mendapatkan info file (`os.Stat`)
- Menggunakan `os.CreateTemp`
- Menangani error ketika membuka file yang tidak ada
- Membersihkan file setelah test

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 14: FILE I/O ===

--- Test 1: Write file ---
Wrote to example.txt
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/14-file-io.md`
- 🌐 Dokumentasi: https://pkg.go.dev/os
