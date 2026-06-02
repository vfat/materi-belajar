# Manual Test: JSON Handling

Folder ini berisi script Go untuk menguji konsep pada materi **15-json-handling.md**.

## Struktur File

```
15-json-handling/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 15-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/15-json-handling
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/15-json-handling

# Compile
go build -o 15-test main.go

# Jalankan
./15-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- `json.Marshal` dan `json.MarshalIndent`
- `json.Unmarshal` ke `struct` dan `map[string]interface{}`
- `json.Decoder`/`json.Encoder` (streaming)
- Error handling untuk JSON tidak valid

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 15: JSON HANDLING ===

--- Test 1: Marshal struct & MarshalIndent ---
{
  "name": "Gopher",
  "age": 5,
  "tags": [
    "go",
    "dev"
  ]
}
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/15-json-handling.md`
- 🌐 Dokumentasi: https://pkg.go.dev/encoding/json
