# Manual Test: Pointer

Folder ini berisi script Go untuk menguji semua konsep dari materi **08-pointer.md**.

## Struktur File

```
08-pointer/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 08-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/08-pointer
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/08-pointer

# Compile
go build -o 08-test main.go

# Jalankan
./08-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Pointer dasar: `&` dan `*`
- Nil pointer (`nil`) dan cara pengecekan
- Membuat pointer dengan `new()`
- Passing pointer ke fungsi untuk modifikasi
- Pointer ke elemen array
- Pointer ke struct dan pengubahan field
- Perbandingan pointer

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 08: POINTER ===

--- Test 1: Pointer basics ---
a=10, p=0xc000... , *p=10
after *p=20 -> a=20
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/08-pointer.md`
- 🌐 Dokumentasi: https://go.dev/doc/
