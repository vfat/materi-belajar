---
topik: Error Handling
urutan: 10 dari 18
posisi: lanjutan
sebelumnya: Interface
---

> 🔗 **Lanjutan dari:** Interface  
> ← Kembali ke: `09-interface.md`

# Error Handling

## Tujuan Belajar

- Memahami konsep error di Go
- Membuat dan mengembalikan error kustom
- Menangani error dengan if check
- Menggunakan multiple return values untuk error
- Pattern error handling yang idiomatic

---

## Filosofi Error di Go

Go tidak menggunakan exception. Error di Go adalah **nilai biasa** (`error` interface) yang dikembalikan sebagai return value. Ini membuat error handling **eksplisit** dan **terkendali**.

```go
package main

import "fmt"

// Interface error bawaan Go
type error interface {
    Error() string
}
```

---

## Error Bawaan

### errors.New()

Cara paling sederhana membuat error:

```go
import (
    "errors"
    "fmt"
)

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("Error:", err)  // Error: division by zero
        return
    }
    fmt.Println("Hasil:", result)
}
```

### fmt.Errorf()

Error dengan format string (mirip fmt.Sprintf):

```go
import (
    "errors"
    "fmt"
)

func getUser(id int) (string, error) {
    if id < 0 {
        return "", fmt.Errorf("invalid user id: %d", id)
    }
    if id == 0 {
        return "", errors.New("user not found")
    }
    return "User" + fmt.Sprintf("%d", id), nil
}

func main() {
    name, err := getUser(-1)
    if err != nil {
        fmt.Println("Error:", err)  // Error: invalid user id: -1
        return
    }
    fmt.Println("Nama:", name)
}
```

---

## Pattern Handle Error

Pattern yang paling umum di Go:

```go
result, err := functionThatMightFail()
if err != nil {
    // Handle error
    // bisa: return, log, wrap, atau ignore
    return err
}
// Lanjutkan dengan result
```

### Contoh Lengkap: Baca File

```go
import (
    "fmt"
    "os"
)

func readConfig(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", fmt.Errorf("gagal baca config: %w", err)
    }
    return string(data), nil
}

func main() {
    content, err := readConfig("config.txt")
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
    fmt.Println("Config:", content)
}
```

---

## Sentinel Error

Error yang didefinisikan di level package sebagai variabel publik:

```go
import (
    "errors"
    "fmt"
)

// Sentinel errors (konstanta error)
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized access")
    ErrInvalidInput = errors.New("invalid input")
)

func getData(key string) (string, error) {
    if key == "" {
        return "", ErrInvalidInput
    }
    if key != "valid" {
        return "", ErrNotFound
    }
    return "data for " + key, nil
}

func main() {
    data, err := getData("")
    if err != nil {
        if errors.Is(err, ErrInvalidInput) {
            fmt.Println("Input tidak boleh kosong!")
            return
        }
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Data:", data)
}
```

---

## Custom Error Type

Buat tipe error sendiri dengan field tambahan:

```go
import (
    "errors"
    "fmt"
    "time"
)

// Custom error type
type ValidationError struct {
    Field   string
    Message string
    Time    time.Time
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on '%s': %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 {
        return ValidationError{
            Field:   "age",
            Message: "must be positive",
            Time:    time.Now(),
        }
    }
    if age > 150 {
        return ValidationError{
            Field:   "age",
            Message: "too old",
            Time:    time.Now(),
        }
    }
    return nil
}

func main() {
    err := validateAge(-5)
    if err != nil {
        // Type assertion
        if ve, ok := err.(ValidationError); ok {
            fmt.Printf("Field: %s, Pesan: %s, Waktu: %v\n", 
                ve.Field, ve.Message, ve.Time)
        }
    }
}
```

---

## Error Wrapping

Wrapping error сохраняет informasi error asli (chained errors):

```go
import (
    "errors"
    "fmt"
)

func inner() error {
    return errors.New("inner error")
}

func middle() error {
    if err := inner(); err != nil {
        return fmt.Errorf("middle failed: %w", err)
    }
    return nil
}

func outer() error {
    if err := middle(); err != nil {
        return fmt.Errorf("outer failed: %w", err)
    }
    return nil
}

func main() {
    err := outer()
    if err != nil {
        fmt.Println("Error:", err)
        // Output: Error: outer failed: middle failed: inner error

        // Cek error spesifik dengan errors.Is()
        if errors.Is(err, errors.New("inner error")) {
            fmt.Println("Ini error dari inner!")
        }
    }
}
```

---

## errors.Is() dan errors.As()

### errors.Is() — Cek Error Spesifik

```go
import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("not found")

func findUser(id int) error {
    if id != 1 {
        return ErrNotFound
    }
    return nil
}

func main() {
    err := findUser(99)
    if err != nil {
        // Cek apakah error ini adalah ErrNotFound
        if errors.Is(err, ErrNotFound) {
            fmt.Println("User tidak ditemukan")
        }
    }
}
```

### errors.As() — Ambil Error dengan Tipe Tertentu

```go
import (
    "errors"
    "fmt"
)

type MyError struct {
    Code    int
    Message string
}

func (e MyError) Error() string {
    return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

func doSomething() error {
    return MyError{Code: 404, Message: "not found"}
}

func main() {
    err := doSomething()
    if err != nil {
        var myErr MyError
        if errors.As(err, &myErr) {
            fmt.Printf("MyError: code=%d, msg=%s\n", myErr.Code, myErr.Message)
        }
    }
}
```

---

## Panic dan Recover

### panic — Stop Execution

`panic` menghentikan eksekusi normal, biasanya untuk error yang tidak bisa di-handle:

```go
func main() {
    fmt.Println("Awal")

    panic(" sesuatu yang salah!")

    fmt.Println("Tidak akan pernah muncul")
}

// Output:
// Awal
// panic: sesuatu yang salah!
```

### recover — Tangkap Panic

`recover` menangkap panic agar program tidak crash:

```go
import "fmt"

func safeCall() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()

    fmt.Println("Sebelum panic")
    panic("Terjadi error!")
    fmt.Println("Tidak akan muncul")
}

func main() {
    safeCall()
    fmt.Println("Program tetap jalan setelah recover")
}
```

### Panic dalam Function

```go
import "fmt"

func divideSafe(a, b float64) (result float64) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Panic tertangkap:", r)
            result = 0
        }
    }()

    if b == 0 {
        panic("division by zero")
    }
    return a / b
}

func main() {
    fmt.Println("Hasil:", divideSafe(10, 2))  // 5
    fmt.Println("Hasil:", divideSafe(10, 0))  // 0 (recovered)
}
```

> ⚠️ **Gunakan panic/recover sparingly!** Untuk flow normal, gunakan error return values.

---

## Defer untuk Cleanup

`defer` menjamin kode dijalankan bahkan ada error:

```go
import (
    "fmt"
    "os"
)

func readFile(path string) {
    file, err := os.Open(path)
    if err != nil {
        fmt.Println("Error buka file:", err)
        return
    }

    // defer menutup file sebelum function selesai
    defer file.Close()

    // Baca file...
    data := make([]byte, 100)
    n, _ := file.Read(data)
    fmt.Println(string(data[:n]))
}

func main() {
    readFile("config.txt")
}
```

### Multiple Defer

```go
func example() {
    defer fmt.Println("1")  // Dijalankan ketiga (terakhir masuk, pertama keluar)
    defer fmt.Println("2")  // Dijalankan kedua
    defer fmt.Println("3")  // Dijalankan pertama

    fmt.Println("utama")
}
// Output:
// utama
// 3
// 2
// 1
```

---

## Latihan

### Latihan 1: Age Validator

Buat fungsi `ValidateAge(age int) error`:
- Return error `"age cannot be negative"` jika age < 0
- Return error `"age cannot exceed 150"` jika age > 150
- Return nil jika valid

### Latihan 2: Stack Error

Buat 3 fungsi (inner, middle, outer) dengan error wrapping. Di outer, gunakan `errors.Is()` untuk cek error spesifik.

### Latihan 3: Custom Error dengan Context

Buat custom error type `DatabaseError` dengan fields:
- `Query string`
- `ErrorMessage string`

Implementasikan method `Error()`. Buat fungsi yang mock koneksi database dan return error ini.

### Latihan 4: Safe Division

Buat fungsi `SafeDivide(a, b float64) (float64, error)` yang:
- Menggunakan panic jika b == 0
- Menggunakan recover untuk menangkap panic
- Return hasil bagi atau 0 jika terjadi panic

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/10-error-handling/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/10-error-handling
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/10-error-handling
go build -o 10-test main.go
./10-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Pembuatan dan wrapping error (`fmt.Errorf`)
- Multiple-return functions mengembalikan error
- Sentinel error dan `errors.Is`
- Custom error type dan `errors.As`
- Panic & recover pattern
- Error dari operasi OS (`os.Open`)
- Error wrapping/unwrapping

---

## ➡️ Selanjutnya

**[String & Formatting]**  
→ Lanjut ke: `11-string-formatting.md`
