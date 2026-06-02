# Manual Test: Struct

Folder ini berisi script Go untuk menguji semua konsep dari materi **07-struct.md**.

## Struktur File

```
07-struct/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 07-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/07-struct
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/07-struct

# Compile
go build -o 07-test main.go

# Jalankan
./07-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Deklarasi struct dasar
- Inisialisasi literal & pointer ke struct
- Anonymous struct
- Value vs Pointer receiver pada method
- Embedded struct (composition)
- JSON marshal/unmarshal sederhana
- Perbandingan struct dan zero-value

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 07: STRUCT ===

--- Test 1: Basic struct ---
 p: {Name:Budi Age:25}
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/07-struct.md`
- 🌐 Dokumentasi: https://go.dev/doc/
