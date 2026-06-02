---
topik: Function (Fungsi)
urutan: 6 dari 6
posisi: akhir
sebelumnya: Array, Slice & Map
---

> 🔗 **Lanjutan dari:** Array, Slice & Map  
> ← Kembali ke: `05-array-slice-map.md`

# Function (Fungsi)

## Tujuan Belajar

- Membuat dan memanggil fungsi
- Parameter dan return value
- Multiple return values & named return
- Variadic function
- Function as value (first-class citizen)
- Closure & Anonymous function
- Error handling dengan fungsi

---

## Apa Itu Function?

Function adalah blok kode yang bisa **dipanggil ulang**. Bayangkan seperti mesin:

```
Input          Function           Output
  5       →    kuadrat()    →     25
  3       →    kuadrat()    →     9
```

---

## Membuat Function

### Fungsi Dasar

```go
// Tanpa parameter, tanpa return
func greet() {
    fmt.Println("Halo, selamat datang!")
}

// Panggil fungsi
greet()  // Output: Halo, selamat datang!
```

### Fungsi dengan Parameter

```go
// Satu parameter
func greetName(nama string) {
    fmt.Println("Halo,", nama)
}

greetName("Budi")  // Output: Halo, Budi

// Multiple parameter
func tambah(a int, b int) {
    fmt.Println(a + b)
}

tambah(3, 5)  // Output: 8
```

### Fungsi dengan Return Value

```go
func kuadrat(n int) int {
    return n * n
}

hasil := kuadrat(5)
fmt.Println(hasil)  // 25
```

### Multiple Return Values

```go
func bagi(a, b int) (int, int) {
    hasil := a / b
    sisa := a % b
    return hasil, sisa
}

q, r := bagi(10, 3)
fmt.Printf("Hasil: %d, Sisa: %d\n", q, r)
// Output: Hasil: 3, Sisa: 1
```

### Return dengan Nama

Go允许命名返回值，类似于声明变量：

```go
func hitung(a, b int) (hasil int) {
    hasil = a + b
    return  // auto return hasil
}
```

---

## Multiple Return dengan Error

Pola umum di Go: fungsi mengembalikan nilai dan error:

```go
func bagi(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("tidak bisa bagi 0")
    }
    return a / b, nil
}

// Penggunaan
hasil, err := bagi(10, 0)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println("Hasil:", hasil)
}
```

---

## Variadic Function

Jumlah parameter fleksibel:

```go
// Fungsi penjumlahan fleksibel
func jumlah(numbers ...int) int {
    total := 0
    for _, n := range numbers {
        total += n
    }
    return total
}

fmt.Println(jumlah(1, 2))           // 3
fmt.Println(jumlah(1, 2, 3, 4, 5))   // 15
fmt.Println(jumlah())                // 0
```

### Spread Operator

```go
angka := []int{1, 2, 3, 4, 5}
fmt.Println(jumlah(angka...))  // 15  (spread slice jadi params)
```

---

## Function as Value

Di Go, fungsi adalah **first-class citizen** — bisa disimpan ke variabel:

```go
// Simpan fungsi ke variabel
kuadrat := func(n int) int {
    return n * n
}

fmt.Println(kuadrat(5))  // 25

// Langsung panggil
result := func(a, b int) int {
    return a + b
}(3, 5)

fmt.Println(result)  // 8
```

### Function sebagai Parameter

```go
// Function sebagai tipe parameter
func apply(n int, f func(int) int) int {
    return f(n)
}

kuadrat := func(n int) int { return n * n }
pangkat3 := func(n int) int { return n * n * n }

fmt.Println(apply(5, kuadrat))   // 25
fmt.Println(apply(5, pangkat3))  // 125
```

### Function sebagai Return Value

```go
func operator(op string) func(int, int) int {
    switch op {
    case "+":
        return func(a, b int) int { return a + b }
    case "-":
        return func(a, b int) int { return a - b }
    case "*":
        return func(a, b int) int { return a * b }
    default:
        return func(a, b int) int { return 0 }
    }
}

tambah := operator("+")
fmt.Println(tambah(5, 3))   // 8

kali := operator("*")
fmt.Println(kali(5, 3))     // 15
```

---

## Closure

Closure adalah fungsi yang **mengingat** variabel di sekitarnya:

```go
// Counter dengan closure
func buatCounter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

counter := buatCounter()

fmt.Println(counter())  // 1
fmt.Println(counter())  // 2
fmt.Println(counter())  // 3

// Counter lain (independen)
counter2 := buatCounter()
fmt.Println(counter2())  // 1
fmt.Println(counter())  // 4 (counter pertama tetap jalan)
```

### Contoh: Filter dengan Closure

```go
func filter(numbers []int, predicate func(int) bool) []int {
    result := []int{}
    for _, n := range numbers {
        if predicate(n) {
            result = append(result, n)
        }
    }
    return result
}

angka := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// Filter bilangan genap
genap := filter(angka, func(n int) bool {
    return n%2 == 0
})
fmt.Println("Genap:", genap)  // [2, 4, 6, 8, 10]

// Filter bilangan > 5
lebihDari5 := filter(angka, func(n int) bool {
    return n > 5
})
fmt.Println("> 5:", lebihDari5)  // [6, 7, 8, 9, 10]
```

---

## Defer

`defer` menunda eksekusi sampai fungsi selesai:

```go
func readFile(name string) {
    defer fmt.Println("File ditutup")  // eksekusi terakhir
    
    fmt.Println("Membuka file:", name)
    fmt.Println("Membaca isi file...")
    
    return  // defer tetap jalan
    
    fmt.Println("Tidak dieksekusi")
}

readFile("data.txt")
```

**Output:**
```
Membuka file: data.txt
Membaca isi file...
File ditutup
```

### Defer untuk Cleanup

```go
func prosesData() {
    koneksi := connectDB()
    defer koneksi.Close()  // pasti ditutup
    
    // ... proses query ...
    
    if error {
        return  // koneksi tetap ditutup
    }
    
    // ... proses lagi ...
    
    return  // koneksi ditutup
}
```

### Multiple Defer

Defer dieksekusi **LIFO** (Last In, First Out):

```go
func demo() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}

demo()
```

**Output:**
```
3
2
1
```

---

## Panic & Recover

### Panic — Stop Eksekusi

```go
func demoPanic() {
    fmt.Println("A")
    panic("Something went wrong!")
    fmt.Println("B")  // tidak dieksekusi
}
```

### Recover — Tangkap Panic

```go
func safeCall() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered:", r)
        }
    }()
    
    fmt.Println("Aman sampai sini")
    panic("Ups!")
    fmt.Println("Tidak dieksekusi")
}

safeCall()
fmt.Println("Tetap jalan setelah recover")
```

**Output:**
```
Aman sampai sini
Recovered: Ups!
Tetap jalan setelah recover
```

---

## Method (Fungsi dengan Receiver)

Method adalah fungsi yang **terikat ke type**:

```go
// Struct
type Persegi struct {
    Sisi float64
}

// Method dengan receiver
func (p Persegi) Luas() float64 {
    return p.Sisi * p.Sisi
}

func (p Persegi) Keliling() float64 {
    return 4 * p.Sisi
}

// Usage
persegi := Persegi{Sisi: 5}
fmt.Println("Luas:", persegi.Luas())      // 25
fmt.Println("Keliling:", persegi.Keliling()) // 20
```

### Pointer Receiver vs Value Receiver

```go
type Counter struct {
    Nilai int
}

// Value receiver - tidak ubah struct asli
func (c Counter) TambahVal(n int) {
    c.Nilai += n
}

// Pointer receiver - ubah struct asli
func (c *Counter) TambahPtr(n int) {
    c.Nilai += n
}

func main() {
    c := Counter{Nilai: 0}
    
    c.TambahVal(5)
    fmt.Println("Setelah Val:", c.Nilai)  // 0 (tidak berubah)
    
    c.TambahPtr(5)
    fmt.Println("Setelah Ptr:", c.Nilai)  // 5 (berubah)
}
```

---

## Contoh Lengkap

```go
package main

import (
    "errors"
    "fmt"
)

// Struct
type Matematika struct{}

// Method: Faktorial
func (m Matematika) Faktorial(n int) (int, error) {
    if n < 0 {
        return 0, errors.New("tidak bisa faktorial negatif")
    }
    if n == 0 || n == 1 {
        return 1, nil
    }
    hasil := 1
    for i := 2; i <= n; i++ {
        hasil *= i
    }
    return hasil, nil
}

// Fungsi higher-order: transformasi
func transform(numbers []int, fn func(int) int) []int {
    result := make([]int, len(numbers))
    for i, n := range numbers {
        result[i] = fn(n)
    }
    return result
}

func main() {
    m := Matematika{}
    
    // Faktorial
    for i := 0; i <= 7; i++ {
        f, err := m.Faktorial(i)
        if err != nil {
            fmt.Println("Error:", err)
        } else {
            fmt.Printf("%d! = %d\n", i, f)
        }
    }
    
    // Transform dengan anonymous function
    angka := []int{1, 2, 3, 4, 5}
    
    kuadrat := transform(angka, func(n int) int {
        return n * n
    })
    fmt.Println("Kuadrat:", kuadrat)  // [1, 4, 9, 16, 25]
    
    plus10 := transform(angka, func(n int) int {
        return n + 10
    })
    fmt.Println("Plus 10:", plus10)  // [11, 12, 13, 14, 15]
}
```

---

## Cheat Sheet

| Sintaks | Arti |
|---------|------|
| `func nama() { }` | Fungsi tanpa return |
| `func nama(a int) int` | Fungsi dengan param & return |
| `func nama() (int, error)` | Multiple return |
| `func nama(nums ...int)` | Variadic |
| `defer func() { }()` | Defer execution |
| `func() int { }()` | IIFE (langsung eksekusi) |
| `func (p *Person) Method()` | Method dengan pointer receiver |

---

## Latihan

1. Buat fungsi `max()` yang terima 2 angka, return yang lebih besar
2. Buat fungsi `reverse()` yang balik string: "hello" → "olleh"
3. Buat counter dengan closure yang bisa reset
4. Challenge: Buat kalkulator dengan higher-order function
   - `calculate(a, b, operation)` → operation adalah fungsi
   - operation bisa: add, subtract, multiply, divide

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/06-function/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/06-function
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/06-function
go build -o 06-test main.go
./06-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Deklarasi & pemanggilan fungsi
- Fungsi dengan return value
- Multiple return values
- Named return values
- Variadic functions
- `defer` dan urutannya
- Anonymous functions dan closures
- Rekursi (factorial)
- Fungsi sebagai nilai / parameter

---

## 🏁 Selesai!

Selamat! Ini adalah **materi terakhir** dalam seri ini.  
Kamu telah menyelesaikan seluruh rangkaian materi belajar "Belajar Golang Dasar".

### Rangkuman 6 Materi:

1. ✅ **Pengenalan & Instalasi** — Apa itu Go, cara instal, program pertama
2. ✅ **Variabel & Tipe Data** — Declarasi, tipe data, konstanta
3. ✅ **Operator & Ekspresi** — Aritmatika, perbandingan, logika, penugasan
4. ✅ **Control Flow** — If/else, switch, for, break, continue
5. ✅ **Array, Slice & Map** — Koleksi data di Go
6. ✅ **Function** — Fungsi, closure, method, defer, panic/recover

📋 Lihat ringkasan perjalanan belajarmu di: `context.md`
