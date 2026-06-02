---
topik: Variabel & Tipe Data
urutan: 2 dari 3
posisi: lanjutan
sebelumnya: Pengenalan & Instalasi
---

> 🔗 **Lanjutan dari:** Pengenalan & Instalasi  
> ← Kembali ke: `01-pengenalan-instalasi.md`

# Variabel & Tipe Data

## Tujuan Belajar

- Cara mendeklarasikan variabel di Go
- Memahami tipe data dasar (int, string, bool, dll)
- Perbedaan `var` dan short declaration `:=`

---

## Apa Itu Variabel?

Variabel adalah tempat **menyimpan data** di memori. Bayangkan seperti kotak berlabel:

```
┌─────────────┐
│     age     │  ← nama variabel
│     25      │  ← nilai (data)
└─────────────┘
```

---

## Cara Declarasikan Variabel

### 1. Dengan Kata Kunci `var`

```go
var nama string = "Budi"
var umur int = 25
var aktif bool = true
```

### 2. Tanpa Inisialisasi (Nilai Default)

```go
var nama string    // default: "" (string kosong)
var umur int       // default: 0
var aktif bool     // default: false
```

### 3. Short Declaration `:=`

```go
nama := "Budi"     // otomatis diketahui tipenya string
umur := 25         // otomatis diketahui tipenya int
aktif := true      // otomatis diketahui tipenya bool
```

> ⚠️ **Catatan:** `:=` hanya bisa di dalam fungsi. Untuk variabel level package, wajib pakai `var`.

---

## Aturan Nama Variabel

| Aturan | Contoh |
|--------|--------|
| ❌ Dimulai angka | `2nama` |
| ❌ Ada spasi | `nama awal` |
| ❌ Kata kunci Go | `var`, `func`, `if` |
| ✅ huruf, angka, underscore | `nama1`, `nama_depan` |
| ✅ PascalCase untuk export | `NamaDepan` |
| ✅ camelCase untuk internal | `namaDepan` |

---

## Tipe Data Dasar

### Numerik Integer

```go
// Bilangan bulat positif
var a uint8  = 255        // 0 - 255
var b uint16 = 65535      // 0 - 65535
var c uint32 = 4294967295 // 0 - 4 miliar
var d uint64 = 18446744073709551615

// Bilangan bulat (positif & negatif)
var e int8   = 127        // -128 - 127
var f int16  = 32767      // -32768 - 32767
var g int32  = 2147483647 // -2 miliar - 2 miliar
var h int64  = 9223372036854775807

// Alias praktis
var i int    // tergantung arsitektur (32/64 bit)
var j rune   // alias int32, untuk karakter Unicode
var k byte   // alias uint8
```

### Numerik Float (Desimal)

```go
var phi float32 = 3.14          // presisi 7 digit
var pi float64 = 3.141592653589 // presisi 15 digit

// Default float di Go adalah float64
uang := 99.99  // ini float64
```

### String

```go
var nama string = "Halo Dunia"
var kosong string = ""

// Multi-line (raw string)
alamat := `
Jl. Sudirman No.1
Jakarta Pusat
`

// Backtick (`) menjaga format persis
```

### Boolean

```go
var benar bool = true
var salah bool = false

// Hasil perbandingan
apakahLulus := 80 > 60  // true
```

---

## Zero Value (Nilai Default)

Setiap variabel di Go otomatis punya nilai awal:

| Tipe Data | Zero Value |
|-----------|------------|
| `int`, `float` | `0` |
| `string` | `""` (kosong) |
| `bool` | `false` |
| `pointer` | `nil` |

```go
var x int
var nama string
var aktif bool

fmt.Println(x)    // 0
fmt.Println(nama) // ""
fmt.Println(aktif)// false
```

---

## Konversi Tipe Data

Go **tidak otomatis konversi tipe** — harus eksplisit:

```go
var a int = 42
var b float64 = float64(a)    // int → float64
var c int = int(b)            // float64 → int (potong koma)

// Untuk string ke number
angka, err := strconv.Atoi("123")
if err != nil {
    fmt.Println("Error:", err)
}

// Number ke string
teks := strconv.Itoa(123)
```

---

## Contoh Lengkap

```go
package main

import "fmt"

func main() {
    // Declarasi berbagai cara
    var nama string = "Siti"
    umur := 22
    
    // Tipe data
    var tinggi float64 = 165.5
    var mahasiswa bool = true
    var pekerjaan string // zero value ""
    
    fmt.Println("Nama:", nama)
    fmt.Println("Umur:", umur)
    fmt.Println("Tinggi:", tinggi, "cm")
    fmt.Println("Mahasiswa:", mahasiswa)
    fmt.Println("Pekerjaan:", pekerjaan)
}
```

**Output:**
```
Nama: Siti
Umur: 22
Tinggi: 165.5 cm
Mahasiswa: true
Pekerjaan: 
```

---

## Konstanta

Nilai yang **tidak bisa diubah** setelah diassign:

```go
const PHI float64 = 3.14159
const NAMA = "Go Learning"
const UKURAN = 10

// Tidak bisa diubah!
// PHI = 3.14  // Error: cannot assign to PHI
```

---

## Cheat Sheet

| Sintaks | Arti |
|---------|------|
| `var x int` | Declarasi tanpa nilai (default 0) |
| `var x int = 10` | Declarasi dengan nilai |
| `x := 10` | Short declaration (di fungsi) |
| `const X = 10` | Konstanta tidak bisa berubah |
| `float64(x)` | Konversi ke float64 |

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/02-variabel-tipe-data/`

### Cara Menjalankan Test

**1. Langsung jalankan (tanpa compile):**
```bash
cd minilab/go/02-variabel-tipe-data
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/02-variabel-tipe-data
go build -o 02-test main.go
./02-test
```

### Apa yang Ditest?

Script `main.go` mencakup 11 test cases:
- ✅ Declarasi variabel dengan `var`
- ✅ Zero value (nilai default)
- ✅ Short declaration `:=`
- ✅ Tipe integer (int8, uint8, int, int64)
- ✅ Tipe float (float32, float64)
- ✅ Tipe string (biasa & multi-line)
- ✅ Tipe boolean
- ✅ Konversi tipe data
- ✅ Konstanta
- ✅ Contoh lengkap dari materi
- ✅ Jawaban latihan

---

## Latihan

1. Buat variabel untuk menyimpan: nama, umur, tinggi, dan status kuliah
2. Print semua variabel tersebut
3. Jelaskan perbedaan `var x int` dengan `x := 10`
4. Ubah nilai konstanta PHI dan amati error-nya

**💡 Tip:** Jalankan `go run main.go` di folder `/minilab/go/02-variabel-tipe-data/` untuk melihat jawaban semua latihan ini!

---

## ➡️ Selanjutnya

**Operator & Ekspresi**  
→ Lanjut ke: `03-operator-ekspresi.md`
