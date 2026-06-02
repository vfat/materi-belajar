# Manual Test: Interface

Folder ini berisi script Go untuk menguji semua konsep dari materi **09-interface.md**.

## Struktur File

```
09-interface/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 09-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/09-interface
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/09-interface

# Compile
go build -o 09-test main.go

# Jalankan
./09-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Interface dasar dan implicit implementation
- Empty interface (`interface{}` / `any`)
- Type assertion dan type switch
- Perilaku nil pada interface
- Implementasi `fmt.Stringer`
- Slice of interfaces dan iterasi

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 09: INTERFACE ===

--- Test 1: Basic interface assignment ---
Woof, I'm Rex
say: Woof, I'm Buddy
say: Meow, I'm Kitty
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/09-interface.md`
- 🌐 Dokumentasi: https://go.dev/doc/
