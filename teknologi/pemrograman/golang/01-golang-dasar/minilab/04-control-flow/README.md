# Manual Test: Control Flow (If/Else, Switch, For)

Folder ini berisi script Go untuk menguji semua konsep dari materi **04-control-flow.md**.

## Struktur File

```
04-control-flow/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 04-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/04-control-flow
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/04-control-flow

# Compile
go build -o 04-test main.go

# Jalankan
./04-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Percabangan `if`, `else if`, `else`
- `if` dengan inisialisasi dan pengecekan error
- `switch` (termasuk fallthrough)
- Berbagai bentuk `for` (standar, while-like)
- `break` dan `continue`
- Nested loop
- Label untuk break/continue bertingkat
- `for range` pada slice, map, dan string

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 04: CONTROL FLOW ===

--- Test 1: If/Else ---
Grade: B

--- Test 2: If with init ---
Nama pendek: Budi
Parsed: 123 (tipe: int)

... (dan seterusnya)

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/04-control-flow.md`
- 🌐 Dokumentasi: https://go.dev/doc/
