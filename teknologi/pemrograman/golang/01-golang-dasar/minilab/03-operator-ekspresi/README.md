# Manual Test: Operator & Ekspresi

Folder ini berisi script Go untuk menguji semua konsep dari materi **03-operator-ekspresi.md**.

## Struktur File

```
03-operator-ekspresi/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 03-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd 03-operator-ekspresi
go run main.go
```

### 2. Build & Jalankan
```bash
cd 03-operator-ekspresi

# Compile
go build -o 03-test main.go

# Jalankan
./03-test
```

## Apa yang Ditest?

File `main.go` mencakup test untuk:

- Operator aritmatika (`+`, `-`, `*`, `/`, `%`)
- Operator perbandingan (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Operator logika (`&&`, `||`, `!`)
- Operator penugasan (`+=`, `-=`, `*=`, `/=`, `%=`)
- Increment / Decrement (`++`, `--`)
- Operator bitwise (`&`, `|`, `^`, `&^`, `<<`, `>>`)
- Presedensi operator
- Contoh gabungan dan output yang diharapkan

## Expected Output

Program akan menampilkan serangkaian test case dan hasilnya, contoh awal:

```
=== MANUAL TEST MATERI 03: OPERATOR & EKSPRESI ===

--- Test 1: Operator Aritmatika ---
10 + 3 = 13
10 - 3 = 7
...

=== SEMUA TEST SELESAI ===
```

## Referensi

- 📖 Materi: `/belajar/golan-dasar/03-operator-ekspresi.md`
- 🌐 Dokumentasi: https://go.dev/doc/
