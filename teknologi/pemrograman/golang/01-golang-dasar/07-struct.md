---
topik: Struct
urutan: 7 dari 18
posisi: lanjutan
sebelumnya: Function (Fungsi)
---

> 🔗 **Lanjutan dari:** Function (Fungsi)  
> ← Kembali ke: `06-function.md`

# Struct

## Tujuan Belajar

- Memahami apa itu struct dan kapan menggunakannya
- Membuat struct dengan field dan tipe data yang berbeda
- Membuat method yang terikat pada struct
- Menggunakan struct dalam fungsi sebagai parameter

---

## Apa Itu Struct?

Struct (structure) adalah **tipe data komposit** yang mengelompokkan beberapa field dengan tipe data berbeda ke dalam satu unit. Bayangkan struct seperti **denah/kontrak** — kita mendefinisikan dulu bentuknya, baru bisa membuat "objek" berdasarkan denah itu.

```go
## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/07-struct/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/07-struct
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/07-struct
go build -o 07-test main.go
./07-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Deklarasi struct dasar
- Inisialisasi literal & pointer ke struct
- Anonymous struct
- Value vs Pointer receiver pada method
- Embedded struct (composition)
- JSON marshal/unmarshal sederhana
- Perbandingan struct dan zero-value

---

## Latihan

### Latihan 1: Data Mahasiswa

Buat struct `Mahasiswa` dengan field:
- `nim` (string)
- `nama` (string)
- `jurusan` (string)
- `ipk` (float64)

Buat method `Status()` yang mengembalikan:
- `"Cemerlang"` jika IPK >= 3.5
- `"Baik"` jika IPK >= 2.5
- `"Perlu Perbaikan"` jika IPK < 2.5

### Latihan 2: Kalkulator Bangun Datar

Buat struct `Rectangle` dan `Circle` dengan method untuk menghitung:
- Luas
- Keliling

Gunakan pointer receiver untuk method `Scale(s)` yang mengalikan dimensi.

---

## ➡️ Selanjutnya

**[Pointer]**  
→ Lanjut ke: `08-pointer.md`
type Student struct {
    name  string
    age   int
    grade float64
}

var s1 Student
s1.name = "Ani"
s1.age = 20
s1.grade = 85.5
```

### 2. Inisialisasi Singkat

```go
s2 := Student{
    name:  "Budi",
    age:   21,
    grade: 90.0,
}
```

### 3. Urutan Sesuai Field

```go
s3 := Student{"Caca", 19, 78.5}
```

### 4. Beberapa Field Boleh Kosong

```go
s4 := Student{name: "Dedi"}
// age = 0, grade = 0.0
```

---

## Struct dengan Tag (Tag)

Tag adalah metadata tambahan untuk field, sering digunakan untuk encoding/decoding (misalnya JSON).

```go
import "encoding/json"

type Product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

func main() {
    p := Product{1, "Laptop", 15000000}
    
    // Konversi ke JSON
    jsonData, _ := json.Marshal(p)
    fmt.Println(string(jsonData))
    // Output: {"id":1,"name":"Laptop","price":15000000}
}
```

---

## Method pada Struct

Di Go, function bisa "dibuat" menjadi method dengan menambahkan **receiver** sebelum nama fungsi.

### Value Receiver

```go
type Circle struct {
    radius float64
}

// Method dengan value receiver
func (c Circle) Area() float64 {
    return 3.14 * c.radius * c.radius
}

func (c Circle) Circumference() float64 {
    return 2 * 3.14 * c.radius
}

func main() {
    c := Circle{radius: 7}
    
    fmt.Printf("Luas: %.2f\n", c.Area())
    fmt.Printf("Keliling: %.2f\n", c.Circumference())
}
```

### Pointer Receiver

Gunakan pointer receiver jika method perlu **mengubah nilai field** dari struct.

```go
type Counter struct {
    value int
}

// Method dengan pointer receiver
func (c *Counter) Increment() {
    c.value++
}

func (c *Counter) Add(n int) {
    c.value += n
}

func (c Counter) GetValue() int {
    return c.value
}

func main() {
    counter := Counter{value: 0}
    
    counter.Increment()
    counter.Increment()
    counter.Add(5)
    
    fmt.Println(counter.GetValue()) // 7
}
```

> ⚠️ **Kapan pakai pointer receiver?**
> - Ketika method perlu mengubah nilai field
> - Ketika struct berukuran besar (efisiensi memori)

---

## Nested Struct (Struct Bertingkat)

Struct bisa memiliki field yang juga merupakan struct lain.

```go
type Address struct {
    street  string
    city    string
    zipCode string
}

type Employee struct {
    name    string
    position string
    address Address  // nested struct
}

func main() {
    emp := Employee{
        name:     "Rina",
        position: "Engineer",
        address: Address{
            street:  "Jl. Asia Afrika",
            city:    "Bandung",
            zipCode: "40111",
        },
    }

    fmt.Println(emp.name)
    fmt.Println(emp.address.city)    // Akses field nested
}
```

---

## Struct sebagai Parameter Fungsi

```go
type Rectangle struct {
    width  float64
    height float64
}

func Area(r Rectangle) float64 {
    return r.width * r.height
}

func Scale(r *Rectangle, factor float64) {
    r.width *= factor
    r.height *= factor
}

func main() {
    rect := Rectangle{width: 10, height: 5}
    
    fmt.Println("Luas:", Area(rect))  // 50
    
    Scale(&rect, 2)  // Skala 2x
    fmt.Println("Setelah skala:", Area(rect))  // 200
}
```

---

## Anonymous Struct

Struct tanpa nama, langsung dibuat instance-nya.

```go
func main() {
    // Anonymous struct
    p := struct {
        name string
        age  int
    }{
        name: "Anonymous",
        age:  22,
    }

    fmt.Println(p.name)  // Anonymous
}
```

---

## Latihan

### Latihan 1: Data Mahasiswa

Buat struct `Mahasiswa` dengan field:
- `nim` (string)
- `nama` (string)
- `jurusan` (string)
- `ipk` (float64)

Buat method `Status()` yang mengembalikan:
- `"Cemerlang"` jika IPK >= 3.5
- `"Baik"` jika IPK >= 2.5
- `"Perlu Perbaikan"` jika IPK < 2.5

### Latihan 2: Kalkulator Bangun Datar

Buat struct `Rectangle` dan `Circle` dengan method untuk menghitung:
- Luas
- Keliling

Gunakan pointer receiver untuk method `Scale(s)` yang mengalikan dimensi.

---

## ➡️ Selanjutnya

**[Pointer]**  
→ Lanjut ke: `08-pointer.md`
