---
topik: Interface
urutan: 9 dari 18
posisi: lanjutan
sebelumnya: Pointer
---

> 🔗 **Lanjutan dari:** Pointer  
> ← Kembali ke: `08-pointer.md`

# Interface

## Tujuan Belajar

- Memahami konsep interface sebagai kontrak
- Mendefinisikan interface dengan method signature
- Mengimplementasikan interface secara implicit
- Menggunakan interface kosong `any` (interface{})

---

## Apa Itu Interface?

Interface adalah **kontrak** yang mendefinisikan sekumpulan method. Tipe data apa pun yang mengimplementasikan semua method di interface tersebut secara otomatis "memenuhi" interface itu.

```
tanpa interface:
┌──────────┐    ┌──────────┐    ┌──────────┐
│  Duck    │    │  Cat     │    │  Dog     │
│  swim()  │    │  swim()  │    │  swim()  │
└──────────┘    └──────────┘    └──────────┘

dengan interface:
       ┌─────────────────┐
       │  Swimmer        │
       │  swim()         │
       └────────┬────────┘
                │
    ┌───────────┼───────────┐
    ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│  Duck    │ │  Cat     │ │  Dog     │
└──────────┘ └──────────┘ └──────────┘
```

---

## Mendefinisikan Interface

```go
package main

import "fmt"

// Definisikan interface
type Speaker interface {
    Speak() string
}
```

Setelah interface didefinisikan, tipe data apa pun yang punya method `Speak() string` secara otomatis mengimplementasikan interface `Speaker`.

---

## Mengimplementasikan Interface

### Contoh 1: Struct dengan Method

```go
import "fmt"

type Speaker interface {
    Speak() string
}

type Dog struct {
    name string
}

type Cat struct {
    name string
}

// Dog mengimplementasikan Speaker
func (d Dog) Speak() string {
    return "Woof!"
}

// Cat juga mengimplementasikan Speaker
func (c Cat) Speak() string {
    return "Meow!"
}

func main() {
    var s Speaker  // variabel bertipe interface

    d := Dog{name: "Buddy"}
    c := Cat{name: "Whiskers"}

    s = d
    fmt.Println(s.Speak())  // Woof!

    s = c
    fmt.Println(s.Speak())  // Meow!
}
```

### Contoh 2: Polymorphism dengan Interface

```go
func saySomething(s Speaker) {
    fmt.Println(s.Speak())
}

func main() {
    dog := Dog{name: "Buddy"}
    cat := Cat{name: "Whiskers"}

    saySomething(dog)  // Woof!
    saySomething(cat)  // Meow!
}
```

---

## Interface dengan Multiple Methods

Interface bisa mendefinisikan lebih dari satu method.

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Rectangle struct {
    width, height float64
}

func (r Rectangle) Area() float64 {
    return r.width * r.height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.width + r.height)
}

func printShapeInfo(s Shape) {
    fmt.Printf("Luas: %.2f, Keliling: %.2f\n", s.Area(), s.Perimeter())
}

func main() {
    rect := Rectangle{width: 10, height: 5}
    printShapeInfo(rect)
}
```

---

## Empty Interface (interface{})

Empty interface adalah interface tanpa method — **semua tipe data** mengimplementasinya!

```go
func main() {
    var empty interface{}

    empty = 42
    fmt.Println(empty)  // 42

    empty = "Hello"
    fmt.Println(empty)  // Hello

    empty = []int{1, 2, 3}
    fmt.Println(empty)  // [1 2 3]
}
```

### Fungsi dengan Parameter Empty Interface

```go
func printAll(items ...interface{}) {
    for _, item := range items {
        fmt.Println(item)
    }
}

func main() {
    printAll(1, "dua", 3.0, true)
}
```

### Type Assertion

Gunakan type assertion untuk mendapatkan tipe sebenarnya dari interface.

```go
func main() {
    var data interface{} = "Hello"

    // Type assertion
    str, ok := data.(string)
    if ok {
        fmt.Println("Ini string:", str)
    }

    // Contoh gagal
    num, ok := data.(int)
    if !ok {
        fmt.Println("Bukan int")
    }
}
```

---

## Type Switch dengan Interface

```go
func identifyType(v interface{}) {
    switch val := v.(type) {
    case int:
        fmt.Printf("Integer: %d\n", val)
    case string:
        fmt.Printf("String: %s\n", val)
    case float64:
        fmt.Printf("Float: %.2f\n", val)
    case bool:
        fmt.Printf("Boolean: %t\n", val)
    default:
        fmt.Printf("Tipe lain: %T\n", val)
    }
}

func main() {
    identifyType(42)         // Integer: 42
    identifyType("Go")       // String: Go
    identifyType(3.14)       // Float: 3.14
    identifyType(true)       // Boolean: true
}
```

---

## Interface sebagai Tipe Return

```go
import "errors"

type算 Error interface {
    Error() string
}

type DivResult struct {
    quotient  float64
    remainder float64
}

func Divide(a, b float64) (DivResult, error) {
    if b == 0 {
        return DivResult{}, errors.New("division by zero")
    }
    return DivResult{
        quotient:  a / b,
        remainder: float64(int(a) % int(b)),
    }, nil
}

func main() {
    result, err := Divide(10, 3)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Hasil: %.2f, Sisa: %.0f\n", result.quotient, result.remainder)
}
```

---

## Interface Error (Bawaan Go)

Go punya interface bawaan `error` yang sangat penting:

```go
type error interface {
    Error() string
}
```

Banyak fungsi bawaan Go mengembalikan `error`:

```go
import (
    "fmt"
    "strconv"
)

func main() {
    // strconv.Atoi mengembalikan (int, error)
    num, err := strconv.Atoi("123")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Angka:", num)

    // Gagal konversi
    num2, err := strconv.Atoi("abc")
    if err != nil {
        fmt.Println("Error:", err)  // Error: strconv.Atoi: parsing "abc": invalid syntax
    }
}
```

---

## Interface dan Struct

Struct bisa menggunakan interface sebagai field type:

```go
import "fmt"

type Writer interface {
    Write(data string) error
}

type Logger struct {
    writer Writer  // interface sebagai field
}

func (l Logger) Log(message string) {
    l.writer.Write(message)
}

type ConsoleWriter struct{}

func (c ConsoleWriter) Write(data string) error {
    fmt.Println("Log:", data)
    return nil
}

func main() {
    writer := ConsoleWriter{}
    logger := Logger{writer: writer}

    logger.Log("Aplikasi dimulai")
    logger.Log("User login")
}
```

---

## Contoh Nyata: Interface Reader/Writer (I/O)

Go menggunakan interface untuk I/O standar:

```go
import (
    "fmt"
    "strings"
)

func main() {
    // strings.Reader mengimplementasikan io.Reader
    reader := strings.NewReader("Hello, Go!")

    // Baca byte per byte
    data := make([]byte, 5)
    for {
        n, err := reader.Read(data)
        if n > 0 {
            fmt.Printf("Dibaca: %s (%d bytes)\n", data[:n], n)
        }
        if err != nil {
            break
        }
    }
}
```

---

## Best Practice Interface

1. **Interface kecil** — buat interface dengan 1-3 method saja
2. **Nama dengan suffix `er`** — `Reader`, `Writer`, `Closer`
3. **Satu hal** — interface mendefinisikan satu kemampuan
4. **Accept interface, return struct** — lebih fleksibel

```go
// ✅ Good: interface kecil & spesifik
type ReadWriteCloser interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
}

// ❌ Bad: interface terlalu besar

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/09-interface/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/09-interface
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/09-interface
go build -o 09-test main.go
./09-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Interface dasar dan implicit implementation
- Empty interface (`interface{}` / `any`)
- Type assertion & type switch
- Perilaku `nil` pada interface
- Implementasi `fmt.Stringer`
- Slice of interfaces dan iterasi

---

## ➡️ Selanjutnya

**Pointer**  
→ Lanjut ke: `08-pointer.md`
type EverythingDoer interface {
    Read() string
    Write() string
    Delete() error
    Update() error
    Create() error
    // ... 20 method lagi
}
```

---

## Latihan

### Latihan 1: Interface Shape

Definisikan interface `Shape` dengan method `Area() float64` dan `Perimeter() float64`.

Implementasikan untuk:
- `Rectangle` (lebar, tinggi)
- `Circle` (radius)

Buat fungsi `PrintInfo(s Shape)` yang mencetak luas dan keliling.

### Latihan 2: Calculator Interface

Definisikan interface `Calculator` dengan method:
- `Add(a, b float64) float64`
- `Subtract(a, b float64) float64`
- `Multiply(a, b float64) float64`
- `Divide(a, b float64) (float64, error)`

Buat struct `SimpleCalc` yang mengimplementasikan interface ini.

### Latihan 3: Error Handling

Buat fungsi `ParseAge(input string) (int, error)` yang:
- Mengembalikan age integer jika input valid
- Mengembalikan error `"invalid input"` jika bukan angka
- Mengembalikan error `"age must be positive"` jika < 0
- Mengembalikan error `"age too old"` jika > 150

---

## ➡️ Selanjutnya

**[Error Handling]**  
→ Lanjut ke: `10-error-handling.md`
