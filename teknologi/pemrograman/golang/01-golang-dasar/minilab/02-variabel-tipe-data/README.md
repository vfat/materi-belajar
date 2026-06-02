# Manual Test: Variabel & Tipe Data

Folder ini berisi script Go untuk menguji semua konsep dari materi **02-variabel-tipe-data.md**.

## Struktur File

```
02-variabel-tipe-data/
├── main.go          ← Script testing
├── README.md        ← File ini
└── 02-test          ← Hasil compile (setelah build)
```

## Cara Menjalankan

### 1. Jalankan Langsung (tanpa compile)
```bash
cd 02-variabel-tipe-data
go run main.go
```

### 2. Build & Jalankan
```bash
cd 02-variabel-tipe-data

# Compile
go build -o 02-test main.go

# Jalankan
./02-test
```

## Apa yang Ditest?

File `main.go` berisi **11 test cases** untuk mendemonstrasikan:

| Test | Topik | Deskripsi |
|------|-------|-----------|
| 1 | Declarasi `var` | Membuat variabel dengan keyword `var` |
| 2 | Zero Value | Nilai default untuk setiap tipe data |
| 3 | Short Declaration | Menggunakan `:=` untuk deklarasi cepat |
| 4 | Tipe Integer | int8, uint8, int, int64, dll |
| 5 | Tipe Float | float32, float64, dan presisi |
| 6 | Tipe String | String biasa dan multi-line |
| 7 | Tipe Boolean | Nilai true/false dan perbandingan |
| 8 | Konversi Tipe | int↔float, string↔number |
| 9 | Konstanta | Nilai yang tidak bisa berubah |
| 10 | Contoh Lengkap | Profil mahasiswa (dari materi) |
| 11 | LATIHAN | Profil pribadi (jawaban contoh) |

## Expected Output

Ketika dijalankan, program akan menampilkan:

```
=== MANUAL TEST MATERI 02: VARIABEL & TIPE DATA ===

--- Test 1: Declarasi dengan var ---
nama: Budi (tipe: string)
umur: 25 (tipe: int)
aktif: true (tipe: bool)

... (dan 10 test lainnya)

=== SEMUA TEST SELESAI ===
```

## Catatan

- Semua output menggunakan `fmt.Printf()` dengan format specifier (`%v`, `%T`, `%.2f`, dll)
- Gunakan `%v` untuk nilai biasa
- Gunakan `%T` untuk menampilkan tipe data
- Gunakan `%.2f` untuk float dengan 2 digit desimal
- Gunakan `%d` untuk integer
- Gunakan `%s` untuk string

## Referensi

- 📖 Materi: `/belajar/golan-dasar/02-variabel-tipe-data.md`
- 🌐 Dokumentasi: https://go.dev/doc/
