---
topik: Control Flow (If/Else, Switch, For)
urutan: 4 dari 6
posisi: lanjutan
sebelumnya: Operator & Ekspresi
---

> 🔗 **Lanjutan dari:** Operator & Ekspresi  
> ← Kembali ke: `03-operator-ekspresi.md`

# Control Flow (If/Else, Switch, For)

## Tujuan Belajar

- Menggunakan percabangan `if`, `else if`, `else`
- Memahami `switch` dan `select`
- Mengenal perulangan `for` di Go

---

## Apa Itu Control Flow?

Control flow menentukan **arah eksekusi program**:
- Pilih jalur mana yang dijalankan (percabangan)
- Ulangi blok kode tertentu (perulangan)

```
    Mulai
      ↓
   [Kondisi]
   /       \
 Benar     Salah
  ↓          ↓
 [Aksi A]  [Aksi B]
      \       /
        ↓
      Selesai
```

---

## Percabangan If/Else

### If Dasar

```go
umur := 18

if umur >= 18 {
    fmt.Println("Kamu dewasa")
}
```

### If dengan Else

```go
nilai := 55

if nilai >= 75 {
    fmt.Println("Lulus!")
} else {
    fmt.Println("Tidak lulus")
}
```

### If, Else If, Else

```go
nilai := 85

if nilai >= 90 {
    fmt.Println("Grade: A")
} else if nilai >= 80 {
    fmt.Println("Grade: B")
} else if nilai >= 70 {
    fmt.Println("Grade: C")
} else if nilai >= 60 {
    fmt.Println("Grade: D")
} else {
    fmt.Println("Grade: E")
}
```

### If dengan Inisialisasi

Go允许在if语句中初始化变量，作用域仅在if块内：

```go
// Cek panjang nama
if nama := "Budi"; len(nama) > 5 {
    fmt.Println("Nama panjang:", nama)
} else {
    fmt.Println("Nama pendek")
}

// Parsing dan cek error sekaligus
if angka, err := strconv.Atoi("123"); err == nil {
    fmt.Println("Angka:", angka)
} else {
    fmt.Println("Error:", err)
}
```

---

## Operator Kondisi di If

| Operator | Arti |
|----------|------|
| `&&` | AND |
| `\|\|` | OR |
| `!` | NOT |

```go
umur := 20
punyaKTP := true

// AND - semua syarat harus true
if umur >= 17 && punyaKTP {
    fmt.Println("Bisa buat SIM")
}

// OR - salah satu true
if umur < 13 || umur > 65 {
    fmt.Println("Dapat diskon")
}
```

---

## Switch

### Switch Dasar

```go
hari := 3

switch hari {
case 1:
    fmt.Println("Senin")
case 2:
    fmt.Println("Selasa")
case 3:
    fmt.Println("Rabu")
case 4:
    fmt.Println("Kamis")
case 5:
    fmt.Println("Jumat")
case 6:
    fmt.Println("Sabtu")
case 7:
    fmt.Println("Minggu")
default:
    fmt.Println("Tidak valid")
}
```

### Multiple Values per Case

```go
huruf := "a"

switch huruf {
case "a", "e", "i", "o", "u":  // salah satu vokal
    fmt.Println("Huruf vokal")
case "b", "c", "d", "f", "g":
    fmt.Println("Huruf konsonan")
default:
    fmt.Println("Karakter lain")
}
```

### Switch Tanpa Ekspresi

Mirip if-else chain, tapi lebih rapi:

```go
nilai := 85

switch {
case nilai >= 90:
    fmt.Println("A")
case nilai >= 80:
    fmt.Println("B")
case nilai >= 70:
    fmt.Println("C")
default:
    fmt.Println("D")
}
```

### Fallthrough

Secara default, Go **tidak jatuh ke case berikutnya**. Gunakan `fallthrough` untuk memaksa:

```go
nilai := 55

switch {
case nilai >= 70:
    fmt.Println("Cukup")
    fallthrough
case nilai >= 50:
    fmt.Println("Lulus")  // akan execute juga
default:
    fmt.Println("Tidak lulus")
}
```

### Switch dengan Inisialisasi

```go
switch nama := "Siti"; nama {
case "Budi", "Siti":
    fmt.Println("Nama known")
default:
    fmt.Println("Nama unknown")
}
```

---

## Perulangan For

Di Go, **`for` adalah satu-satunya struktur perulangan** (tapi sangat fleksibel).

### For Standar

```go
for i := 0; i < 5; i++ {
    fmt.Println("Iterasi ke:", i)
}
```

### For Tanpa Init & Post

```go
i := 0
for i < 5 {       // seperti while di bahasa lain
    fmt.Println(i)
    i++
}
```

### For Tanpa Kondisi (Infinite Loop)

```go
for {
    fmt.Println("Terus looping...")
    // butuh break untuk keluar
}
```

---

## Break & Continue

### Break — Keluar dari Perulangan

```go
for i := 0; i < 10; i++ {
    if i == 5 {
        break   // stop di i=5
    }
    fmt.Println(i)
}
// Output: 0, 1, 2, 3, 4
```

### Continue — Lanjut ke Iterasi Berikutnya

```go
for i := 0; i < 5; i++ {
    if i == 2 {
        continue   // skip i=2
    }
    fmt.Println(i)
}
// Output: 0, 1, 3, 4
```

---

## Nested Loop (Perulangan Bersarang)

```go
for i := 1; i <= 3; i++ {
    for j := 1; j <= 3; j++ {
        fmt.Printf("%d x %d = %d\n", i, j, i*j)
    }
    fmt.Println()
}
```

**Output:**
```
1 x 1 = 1
1 x 2 = 2
1 x 3 = 3

2 x 1 = 2
2 x 2 = 4
2 x 3 = 6

3 x 1 = 3
3 x 2 = 6
3 x 3 = 9
```

---

## Label untuk Break/Continue Bertingkat

```go
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if j == 1 {
            break outer   // keluar dari outer loop
        }
        fmt.Println(i, j)
    }
}
```

---

## Contoh Lengkap: Menu Sederhana

```go
package main

import "fmt"

func main() {
    var pilihan int

    fmt.Println("=== Kalkulator Sederhana ===")
    fmt.Println("1. Tambah")
    fmt.Println("2. Kurang")
    fmt.Println("3. Kali")
    fmt.Println("4. Bagi")
    fmt.Print("Pilih (1-4): ")
    fmt.Scan(&pilihan)

    var a, b float64
    fmt.Print("Angka 1: ")
    fmt.Scan(&a)
    fmt.Print("Angka 2: ")
    fmt.Scan(&b)

    var hasil float64

    switch pilihan {
    case 1:
        hasil = a + b
    case 2:
        hasil = a - b
    case 3:
        hasil = a * b
    case 4:
        if b == 0 {
            fmt.Println("Error: tidak bisa bagi 0")
            return
        }
        hasil = a / b
    default:
        fmt.Println("Pilihan tidak valid")
        return
    }

    fmt.Printf("Hasil: %.2f\n", hasil)
}
```

---

## For dengan Range

Iterasi untuk slice, map, array, string:

```go
// Array/Slice
buah := []string{"Apel", "Jeruk", "Mangga"}

for index, value := range buah {
    fmt.Printf("%d: %s\n", index, value)
}

// Map
umur := map[string]int{"Budi": 25, "Siti": 22}

for nama, umr := range umur {
    fmt.Printf("%s: %d tahun\n", nama, umr)
}

// String
for index, char := range "Halo" {
    fmt.Printf("%d: %c\n", index, char)
}

// Skip index dengan _
for _, value := range buah {
    fmt.Println(value)
}
```

---

## Cheat Sheet

| Sintaks | Arti |
|---------|------|
| `if kondisi { }` | Jika kondisi true |
| `if else { }` | Jika tidak, jalankan ini |
| `else if { }` | Kondisi lain |
| `switch var { case x: }` | Pilih berdasarkan nilai |
| `for i := 0; i < n; i++` | Loop standar |
| `for kondisi { }` | Loop seperti while |
| `for { }` | Infinite loop |
| `break` | Keluar dari loop |
| `continue` | Lanjut ke iterasi berikutnya |
| `range` | Iterasi koleksi |

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/04-control-flow/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/04-control-flow
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/04-control-flow
go build -o 04-test main.go
./04-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Percabangan `if`, `else if`, `else`
- `if` dengan inisialisasi dan pengecekan error
- `switch` (termasuk `fallthrough`)
- Berbagai bentuk `for` (standar, while-like)
- `break` dan `continue`
- Nested loop dan label `break`
- `for range` pada slice, map, dan string

---

## Latihan

1. Buat program cek grade nilai: A(90+), B(80+), C(70+), D(60+), E(<60)
2. Buat program kalkulator sederhana menggunakan switch
3. Buat program print bilangan genap dari 1-20 menggunakan continue
4. Challenge: Buat pola segitiga menggunakan nested loop

```
*
**
***
****
*****
```

---

## ➡️ Selanjutnya

**Array, Slice & Map**  
→ Lanjut ke: `05-array-slice-map.md`
