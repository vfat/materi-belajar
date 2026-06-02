---
topik: Operator & Ekspresi
urutan: 3 dari 3
posisi: akhir
sebelumnya: Variabel & Tipe Data
---

> 🔗 **Lanjutan dari:** Variabel & Tipe Data  
> ← Kembali ke: `02-variabel-tipe-data.md`

# Operator & Ekspresi

## Tujuan Belajar

- Operator aritmatika (+, -, *, /, %)
- Operator perbandingan (==, !=, <, >, dll)
- Operator logika (&&, ||, !)
- Operator penugasan (=, +=, -=, dll)

---

## Apa Itu Operator & Ekspresi?

- **Operator** = simbol yang melakukan operasi
- **Ekspresi** = kombinasi nilai, variabel, dan operator yang menghasilkan nilai

```go
a := 10
b := 5
hasil := a + b   // a + b adalah ekspresi, + adalah operator
```

---

## Operator Aritmatika

### Daftar Operator

| Operator | Arti | Contoh | Hasil |
|----------|------|--------|-------|
| `+` | Penjumlahan | `5 + 3` | `8` |
| `-` | Pengurangan | `5 - 3` | `2` |
| `*` | Perkalian | `5 * 3` | `15` |
| `/` | Pembagian | `5 / 3` | `1` (integer) |
| `%` | Modulus (sisa bagi) | `5 % 3` | `2` |

### Contoh Kode

```go
package main

import "fmt"

func main() {
    a, b := 10, 3
    
    fmt.Println("a + b =", a + b)  // 13
    fmt.Println("a - b =", a - b)  // 7
    fmt.Println("a * b =", a * b)  // 30
    fmt.Println("a / b =", a / b)  // 3 (pembagian integer)
    fmt.Println("a % b =", a % b)  // 1 (sisa bagi)
}
```

### Catatan Penting

```go
// Pembagian integer → hasil integer (dibulatkan)
fmt.Println(5 / 2)   // 2, bukan 2.5

// Untuk desimal, gunakan float
fmt.Println(5.0 / 2.0)  // 2.5

// Modulus negatif
fmt.Println(-5 % 3)   // -2
```

---

## Operator Perbandingan

Menghasilkan nilai **boolean** (`true` atau `false`)

| Operator | Arti | Contoh | Hasil |
|----------|------|--------|-------|
| `==` | Sama dengan | `5 == 5` | `true` |
| `!=` | Tidak sama | `5 != 3` | `true` |
| `>` | Lebih besar | `5 > 3` | `true` |
| `<` | Lebih kecil | `5 < 3` | `false` |
| `>=` | Lebih besar atau sama | `5 >= 5` | `true` |
| `<=` | Lebih kecil atau sama | `5 <= 3` | `false` |

### Contoh Kode

```go
package main

import "fmt"

func main() {
    x := 10
    y := 5
    
    fmt.Println("x == y:", x == y)  // false
    fmt.Println("x != y:", x != y)  // true
    fmt.Println("x > y:", x > y)    // true
    fmt.Println("x < y:", x < y)    // false
    fmt.Println("x >= y:", x >= y)  // true
    fmt.Println("x <= y:", x <= y)  // false
}
```

---

## Operator Logika

Mengoperasikan nilai **boolean**

| Operator | Arti | Deskripsi |
|----------|------|-----------|
| `&&` | AND | `true` jika **kedua** operand true |
| `\|\|` | OR | `true` jika **salah satu** operand true |
| `!` | NOT | Membalik boolean |

### Tabel Kebenaran

#### AND (&&)

| A | B | A && B |
|---|---|--------|
| true | true | **true** |
| true | false | false |
| false | true | false |
| false | false | false |

#### OR (||)

| A | B | A \|\| B |
|---|---|----------|
| true | true | true |
| true | false | **true** |
| false | true | **true** |
| false | false | false |

#### NOT (!)

| A | !A |
|---|----|
| true | false |
| false | true |

### Contoh Kode

```go
package main

import "fmt"

func main() {
    umur := 20
    punyaKTP := true
    punyaUang := false
    
    // AND: semua syarat harus terpenuhi
    bisaBeli := umur >= 17 && punyaKTP
    fmt.Println("Bisa beli tiket?", bisaBeli)  // true
    
    // OR: salah satu syarat terpenuhi
    bisaMasuk := punyaKTP || punyaUang
    fmt.Println("Bisa masuk?", bisaMasuk)  // true
    
    // NOT: membalik nilai
    bukanAnak := !false
    fmt.Println("Bukan anak?", bukanAnak)  // true
}
```

---

## Operator Penugasan (Assignment)

Memberikan / mengubah nilai variabel

| Operator | Contoh | Arti |
|----------|--------|------|
| `=` | `x = 5` | Assign nilai |
| `+=` | `x += 3` | Tambah lalu assign |
| `-=` | `x -= 3` | Kurang lalu assign |
| `*=` | `x *= 3` | Kali lalu assign |
| `/=` | `x /= 3` | Bagi lalu assign |
| `%=` | `x %= 3` | Modulus lalu assign |

### Contoh Kode

```go
package main

import "fmt"

func main() {
    x := 10
    
    x += 5    // x = 10 + 5 = 15
    fmt.Println("Setelah +5:", x)
    
    x -= 3    // x = 15 - 3 = 12
    fmt.Println("Setelah -3:", x)
    
    x *= 2    // x = 12 * 2 = 24
    fmt.Println("Setelah *2:", x)
    
    x /= 4    // x = 24 / 4 = 6
    fmt.Println("Setelah /4:", x)
    
    x %= 5    // x = 6 % 5 = 1
    fmt.Println("Setelah %5:", x)
}
```

---

## Operator Increment & Decrement

| Operator | Arti | Contoh |
|----------|------|--------|
| `++` | Tambah 1 | `x++` sama dengan `x += 1` |
| `--` | Kurang 1 | `x--` sama dengan `x -= 1` |

```go
package main

import "fmt"

func main() {
    counter := 0
    
    counter++
    fmt.Println("Counter:", counter)  // 1
    
    counter++
    fmt.Println("Counter:", counter)  // 2
    
    counter--
    fmt.Println("Counter:", counter)  // 1
}
```

> ⚠️ **Catatan:** `++` dan `--` adalah **statement**, bukan ekspresi. Tidak bisa ditulis seperti `x = i++`

---

## Operator Bitwise

Bekerja pada level **bit** (biner)

| Operator | Arti | Contoh |
|----------|------|--------|
| `&` | AND bitwise | `5 & 3` = `1` |
| `\|` | OR bitwise | `5 \| 3` = `7` |
| `^` | XOR bitwise | `5 ^ 3` = `6` |
| `&^` | AND NOT | `5 &^ 3` = `4` |
| `<<` | Shift kiri | `1 << 2` = `4` |
| `>>` | Shift kanan | `4 >> 1` = `2` |

```go
package main

import "fmt"

func main() {
    a, b := 5, 3  // 5 = 101 (biner), 3 = 011 (biner)
    
    fmt.Println("a & b  =", a & b)   // 1   (001)
    fmt.Println("a | b  =", a | b)   // 7   (111)
    fmt.Println("a ^ b  =", a ^ b)   // 6   (110)
    fmt.Println("a &^ b =", a &^ b)  // 4   (100)
}
```

---

## Presedensi Operator

Urutan prioritas (tertinggi ke terendah):

```
1. ()           Parenthesis
2. * / %        Aritmatika tinggi
3. + -          Aritmatika rendah
4. < > <= >=    Perbandingan
5. == !=        Persamaan
6. &&           AND
7. ||           OR
8. = += -=      Assignment
```

```go
// Contoh: perkalian dilakukan lebih dulu
hasil := 2 + 3 * 4   // 2 + 12 = 14

// Tambah parentesis untuk mengubah urutan
hasil2 := (2 + 3) * 4  // 5 * 4 = 20
```

---

## Contoh Lengkap Gabungan

```go
package main

import "fmt"

func main() {
    // Data siswa
    var nama string = "Rina"
    var nilaiTeori int = 85
    var nilaiPraktik int = 90
    
    // Perhitungan
    totalNilai := nilaiTeori + nilaiPraktik
    rataRata := float64(totalNilai) / 2.0
    
    // Perbandingan
    lulusTeori := nilaiTeori >= 75
    lulusPraktik := nilaiPraktik >= 75
    lulusSemuanya := lulusTeori && lulusPraktik
    
    // Output
    fmt.Println("=== Hasil Belajar ===")
    fmt.Println("Nama:", nama)
    fmt.Printf("Nilai Teori: %d, Praktik: %d\n", nilaiTeori, nilaiPraktik)
    fmt.Println("Total:", totalNilai)
    fmt.Printf("Rata-rata: %.1f\n", rataRata)
    fmt.Println("Lulus?", lulusSemuanya)
}
```

**Output:**
```
=== Hasil Belajar ===
Nama: Rina
Nilai Teori: 85, Praktik: 90
Total: 175
Rata-rata: 87.5
Lulus? true
```

---

## Cheat Sheet

| Kategori | Operator |
|----------|----------|
| Aritmatika | `+`, `-`, `*`, `/`, `%` |
| Perbandingan | `==`, `!=`, `<`, `>`, `<=`, `>=` |
| Logika | `&&`, `\|\|`, `!` |
| Penugasan | `=`, `+=`, `-=`, `*=`, `/=`, `%=` |
| Increment | `++`, `--` |
| Bitwise | `&`, `\|`, `^`, `&^`, `<<`, `>>` |

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/03-operator-ekspresi/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/03-operator-ekspresi
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/03-operator-ekspresi
go build -o 03-test main.go
./03-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Operator aritmatika (`+`, `-`, `*`, `/`, `%`)
- Operator perbandingan (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Operator logika (`&&`, `||`, `!`)
- Operator penugasan (`+=`, `-=`, `*=`, `/=`, `%=`)
- Increment / Decrement (`++`, `--`)
- Operator bitwise (`&`, `|`, `^`, `&^`, `<<`, `>>`)
- Presedensi operator dan contoh gabungan

---

## Latihan

1. Buat kalkulator sederhana: minta input 2 angka, lalu print hasil dari semua operasi aritmatika
2. Buat program yang cek apakah seseorang bisa memilih班长 (umur >= 15 DAN punya hak pilih)
3. Apa hasil dari: `!(true && false) || (true || false)` ?

---

## ➡️ Selanjutnya

**Control Flow (If/Else, Switch, For)**  
→ Lanjut ke: `04-control-flow.md`

---

## 🏁 Selesai!

Selamat! Ini adalah **materi terakhir** dalam seri ini.  
Kamu telah menyelesaikan seluruh rangkaian materi belajar "Belajar Golang Dasar".

📋 Lihat ringkasan perjalanan belajarmu di: `context.md`

### Rangkuman 3 Materi:

1. ✅ **Pengenalan & Instalasi** — Apa itu Go, cara instal, program pertama
2. ✅ **Variabel & Tipe Data** — Declarasi, tipe data, konstanta
3. ✅ **Operator & Ekspresi** — Aritmatika, perbandingan, logika, penugasan
