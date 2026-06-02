# Manual Test: Channel

Folder ini berisi script Go untuk menguji konsep pada materi **17-channel.md**.

## Struktur File

```
17-channel/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 17-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd minilab/go/17-channel
go run main.go
```

### 2. Build & Jalankan
```bash
cd minilab/go/17-channel

# Compile
go build -o 17-test main.go

# Jalankan
./17-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Send/receive pada channel
- Close + range
- Buffered channel
- Fan-in menggunakan `select`
- Pipeline pattern (`generator` -> `sq`)
- Perilaku `nil` channel di `select`

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 17: CHANNEL ===

--- Test 1: Simple send/receive ---
received: 1
...
=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golang-dasar/17-channel.md`
- 🌐 Dokumentasi: https://pkg.go.dev/builtin#chan
