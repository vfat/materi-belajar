# Manual Test: Function

Folder ini berisi script Go untuk menguji semua konsep dari materi **06-function.md**.

## Struktur File

```
06-function/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 06-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/06-function
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/06-function

# Compile
go build -o 06-test main.go

# Jalankan
./06-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Deklarasi & pemanggilan fungsi
- Fungsi dengan return value
- Multiple return values
- Named return values
- Variadic functions
- `defer` dan urutannya
- Anonymous functions dan closures
- Rekursi (factorial)
- Fungsi sebagai nilai / parameter

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 06: FUNCTION ===

--- Test 1: Function Declaration ---
Halo, Budi

... (dan seterusnya)

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/06-function.md`
- 🌐 Dokumentasi: https://go.dev/doc/
