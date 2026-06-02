# Manual Test: String & Formatting

Folder ini berisi script Go untuk menguji konsep pada materi **11-string-formatting.md**.

## Struktur File

```
11-string-formatting/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 11-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/11-string-formatting
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/11-string-formatting

# Compile
go build -o 11-test main.go

# Jalankan
./11-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Basic `fmt` formatting (`%s`, `%d`, `%.2f`)
- `fmt.Sprintf` dan berbagai verb
- Konversi dengan `strconv`
- Formatting waktu (`time.Format`)
- Width, padding dan alignment
- Type/value printing dengan `%#v` dan `%T`

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 11: STRING & FORMATTING ===

--- Test 1: Basic formatting ---
Hello, Gopher. Age: 5
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/11-string-formatting.md`
- 🌐 Dokumentasi: https://pkg.go.dev/fmt, https://pkg.go.dev/strconv
