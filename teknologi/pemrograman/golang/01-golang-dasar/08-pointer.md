---
topik: Pointer
urutan: 8 dari 18
posisi: lanjutan
sebelumnya: Struct
---

> 🔗 **Lanjutan dari:** Struct  
> ← Kembali ke: `07-struct.md`

# Pointer

## Tujuan Belajar

- Memahami konsep memory address dan pointer
- Menggunakan operator `&` (address of) dan `*` (dereference)
- Memahami perbedaan value types vs reference types
- Menggunakan pointer dalam function dan method

---

## Apa Itu Pointer?

**Pointer** adalah variabel yang menyimpan **alamat memori** dari variabel lain, bukan nilainya langsung.

```
Variabel Normal:
┌─────────┐
│    42   │  ← menyimpan nilai
└─────────┘

Pointer:
┌─────────┐         ┌─────────┐
│ 0xc000  │────────→│   42    │  ← menyimpan alamat memori
└─────────┘         └─────────┘
```

---

## Operator Dasar Pointer

| Operator | Nama | Fungsi |
|----------|------|--------|
| `&` | Address Of | Mengambil alamat memori sebuah variabel |
| `*` | Dereference | Mengakses nilai di alamat memori yang ditunjuk |

### Contoh Dasar

```go
package main

import "fmt"

func main() {
    // Variabel biasa
    x := 10
    fmt.Println("Nilai x:", x)        // 10
    fmt.Println("Alamat x:", &x)     // contoh: 0xc000014088

    // Pointer
    var p *int = &x  // p menyimpan alamat x
    fmt.Println("Nilai p (alamat):", p)   // 0xc000014088
    fmt.Println("Nilai *p (dereference):", *p) // 10

    // Mengubah nilai melalui pointer
    *p = 25
    fmt.Println("x sekarang:", x)  // 25 (x ikut berubah!)
}
```

---

## Membuat Pointer dengan `new`

Fungsi `new(T)` mengalokasikan memori untuk tipe `T` dan mengembalikan pointer ke tipe tersebut.

```go
func main() {
    // Membuat pointer ke int
    p := new(int)
    fmt.Println(*p)  // 0 (nilai default int)

    *p = 100
    fmt.Println(*p)  // 100

    // Pointer ke string
    name := new(string)
    *name = "Go"
    fmt.Println(*name)  // Go
}
```

---

## Value Types vs Reference Types

Ini penting! Go membedakan tipe data berdasarkan cara penyimpanan nilainya di memori:

### Value Types
Disimpan langsung di **stack** (salinan setiap kali dikirim ke fungsi).

```go
type Point struct {
    X, Y int
}

func main() {
    p1 := Point{X: 1, Y: 2}
    p2 := p1  // Salinan! p2 adalah copy dari p1

    p2.X = 100
    fmt.Println(p1.X)  // 1 (TIDAK berubah)
    fmt.Println(p2.X)  // 100
}
```

### Reference Types
Yang disimpan adalah **alamat/pointer** ke data sebenarnya di **heap**.

```go
func main() {
    slice1 := []int{1, 2, 3}
    slice2 := slice1  // Bukan salinan, tapi referensi ke array yang sama

    slice2[0] = 100
    fmt.Println(slice1[0])  // 100 (BERUBAH!)
}
```

### Tabel Perbedaan

| Value Types | Reference Types |
|-------------|-----------------|
| int, float, bool, string | slice, map, channel |
| array, struct | function |
| | pointer |

---

## Pointer sebagai Parameter Fungsi

Gunakan pointer di fungsi agar fungsi bisa **mengubah nilai asli** variabel.

### Tanpa Pointer (Tidak Berubah)

```go
func double(n int) {
    n = n * 2
    fmt.Println("Dalam fungsi:", n)  // 20
}

func main() {
    num := 10
    double(num)
    fmt.Println("Setelah fungsi:", num)  // 10 (TIDAK berubah)
}
```

### Dengan Pointer (Berubah!)

```go
func doublePtr(n *int) {
    *n = *n * 2
    fmt.Println("Dalam fungsi:", *n)  // 20
}

func main() {
    num := 10
    doublePtr(&num)
    fmt.Println("Setelah fungsi:", num)  // 20 (BERUBAH!)
}
```

### Contoh Nyata: Swap Dua Nilai

```go
func swap(a, b *int) {
    temp := *a
    *a = *b
    *b = temp
}

func main() {
    x, y := 5, 10
    fmt.Println("Sebelum:", x, y)  // 5, 10
    swap(&x, &y)
    fmt.Println("Sesudah:", x, y)  // 10, 5
}
```

---

## Pointer dalam Method

Kita sudah lihat ini di materi Struct — pointer receiver memungkinkan method mengubah nilai field.

```go
type Box struct {
    width, height, depth float64
}

func (b *Box) Scale(factor float64) {
    b.width *= factor
    b.height *= factor
    b.depth *= factor
}

func main() {
    box := Box{10, 5, 3}
    box.Scale(2)
    fmt.Printf("Ukuran box: %.1f x %.1f x %.1f\n", 
        box.width, box.height, box.depth)
    // Output: 20.0 x 10.0 x 6.0
}
```

---

## Nil Pointer

Pointer yang belum diarahkan ke apapun bernilai `nil`.

```go
func main() {
    var p *int  // Default nil
    fmt.Println(p == nil)  // true

    // Akses dereference nil pointer = PANIC!
    // fmt.Println(*p)  // ❌ Error: panic: runtime error: invalid memory address or nil pointer dereference

    // Selalu cek nil sebelum dereference
    if p != nil {
        fmt.Println(*p)
    }
}
```

### Return Pointer dari Fungsi

```go
func createPointer(value int) *int {
    p := new(int)
    *p = value
    return p  // Pointer aman dikembalikan (data di heap)
}

func main() {
    ptr := createPointer(42)
    fmt.Println(*ptr)  // 42
}
```

---

## Double Pointer

Pointer yang menunjuk ke pointer lain.

```go
func main() {
    x := 10
    p1 := &x      // *int
    p2 := &p1     // **int (double pointer)

    fmt.Println(x)    // 10
    fmt.Println(*p1)  // 10
    fmt.Println(**p2) // 10
}
```

---

## Kapan Gunakan Pointer?

| Gunakan Pointer | Gunakan Value |
|-----------------|--------------|
| Ingin mengubah nilai asli | Ingin aman dari perubahan tak sengaja |
| Struct berukuran besar | Tipe data kecil (int, float, bool, string) |
| Method perlu mengubah receiver | Method hanya membaca data |
| Menghindari copy data besar | Ingin predictably immutable |

---

## Latihan

### Latihan 1: Tukar Nilai

Buat fungsi `Swap(a, b *int)` yang menukar nilai dua variabel. Tunjukkan sebelum dan sesudah swap.

### Latihan 2: Counter dengan Pointer

Buat struct `Counter` dengan field `value int`. Tambahkan method:
- `Increment()` — tambah 1
- `Decrement()` — kurang 1
- `Reset()` — jadi 0

Gunakan pointer receiver.

### Latihan 3: Pointer vs Value

Jalankan kode ini dan jelaskan kenapa outputnya berbeda:

```go
type Data struct {
    value int
}

func modifyValue(d Data) {
    d.value = 100
}

func modifyPointer(d *Data) {
    d.value = 100
}

func main() {
    d1 := Data{value: 1}
    d2 := Data{value: 1}

    modifyValue(d1)
    modifyPointer(&d2)

    fmt.Println("d1.value:", d1.value) // ?
    fmt.Println("d2.value:", d2.value) // ?
}
```

---

## ➡️ Selanjutnya

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/08-pointer/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/08-pointer
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/08-pointer
go build -o 08-test main.go
./08-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Pointer dasar: `&` dan `*`
- Nil pointer (`nil`) dan cara pengecekan
- Membuat pointer dengan `new()`
- Passing pointer ke fungsi untuk modifikasi
- Pointer ke elemen array
- Pointer ke struct dan pengubahan field
- Perbandingan pointer

---

## ➡️ Selanjutnya

**[Interface]**  
→ Lanjut ke: `09-interface.md`
