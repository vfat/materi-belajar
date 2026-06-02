# Manual Test: Package & Module

Folder ini berisi script Go untuk menguji konsep pada materi **13-package-module.md**.

## Struktur File

```
13-package-module/
├── go.mod           ← module file
├── main.go          ← Script testing (import subpackage)
├── util/
│   └── util.go      ← contoh subpackage
└── 13-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/13-package-module
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/13-package-module

# Compile
go build -o 13-test main.go

# Jalankan
./13-test
```

## Apa yang Ditest?

File `main.go` dan subpackage `util` berisi test untuk:

- Struktur package lokal (`util`)
- Penggunaan subpackage dari `main`
- Contoh `go.mod` pada root module

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 13: PACKAGE & MODULE ===

--- Test 1: Local subpackage usage ---
Add 2 + 3 = 5
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/13-package-module.md`
- 🌐 Dokumentasi: https://pkg.go.dev/
