---
topik: Package & Module
urutan: 13 dari 18
posisi: lanjutan
sebelumnya: Time & Date
---

> 🔗 **Lanjutan dari:** Time & Date  
> ← Kembali ke: `12-time-date.md`

# Package & Module

## Tujuan Belajar

- Memahami konsep package di Go
- Membuat dan mengorganisir module
- Menggunakan `go.mod` dan `go.sum`
- Import package internal dan eksternal
- Memahami visibility (exported vs unexported)
- Pattern struktur project Go

---

## Apa itu Package?

Package adalah cara Go **mengorganisir kode**. Semua file `.go` dalam direktori yang sama принадлежат ke package yang sama.

```go
package main  // File ini termasuk package main

import "fmt"

func main() {
    fmt.Println("Hello dari package main!")
}
```

---

## Jenis Package

### Executable Package (`main`)

Package yang menghasilkan **file executable** (binary). Harus memiliki fungsi `main()`:

```
project/
├── go.mod
└── main.go          → package main
```

```go
// main.go
package main

import "fmt"

func main() {
    fmt.Println("Ini akan jadi binary")
}
```

### Reusable Package (Library)

Package biasa yang bisa di-import oleh package lain:

```
calculator/
├── add.go
└── subtract.go
```

```go
// add.go
package calculator

// Exported (bisa diakses dari luar)
func Add(a, b int) int {
    return a + b
}

// unexported (hanya bisa diakses dalam package)
func addPrivate(a, b int) int {
    return a + b
}
```

---

## Module

Module adalah kumpulan package yang/versioned bersama. Dibuat dengan `go.mod`.

### Membuat Module Baru

```bash
# 1. Buat direktori
mkdir myproject
cd myproject

# 2. Initialize module
go mod init github.com/username/myproject

# 3. Buat file...
```

### go.mod

```go
// go.mod
module github.com/username/myproject

go 1.21

require (
    github.com/some/package v1.2.3
)
```

### go.sum

File auto-generated yang menyimpan checksum semua dependency:

```
github.com/some/package v1.2.3 h1:abcdef123456...
github.com/some/package v1.2.3/go.mod h1:xyz789...
```

---

## Import dan Penggunaan

### Import Standar Library

```go
import (
    "fmt"
    "strings"
    "time"
)
```

### Import Eksternal Package

```go
import (
    "github.com/gin-gonic/gin"    // Web framework
    "github.com/go-redis/redis/v8" // Redis client
    "gorm.io/gorm"                 // ORM
)
```

### Download Dependency

```bash
go mod tidy    # Otomatis download & update go.mod/go.sum
go get [package]  # Download package tertentu
```

---

## Visibility (Export Rules)

Aturan sederhana:
- **Huruf BESAR** → Exported (bisa diakses dari luar package)
- **huruf kecil** → unexported (hanya dalam package)

```go
package calculator

// Exported - bisa diakses dari package lain
func Add(a, b int) int {
    return a + b
}

// unexported - hanya dalam package ini
func formatResult(a, b int) string {
    return fmt.Sprintf("%d + %d = %d", a, b, a+b)
}
```

### Contoh Akses

```go
// File lain (package lain):
import "github.com/user/calculator"

result := calculator.Add(5, 3)    // ✅ Bisa
calculator.formatResult(5, 3)    // ❌ Error: not exported
```

---

## Inisialisasi Package

Kode `init()` dijalankan saat package pertama kali di-import:

```go
package database

import "fmt"

var DB *sql.DB

func init() {
    fmt.Println("Database package initialized")
    // Setup koneksi, dll
}

func Connect() (*sql.DB, error) {
    return DB, nil
}
```

### Multiple init()

Jika ada banyak file dengan `init()`, mereka dijalankan berdasarkan urutan alphabetically:

```go
// file1.go
package mypackage

func init() {
    fmt.Println("Init dari file1")
}

// file2.go
package mypackage

func init() {
    fmt.Println("Init dari file2")
}
```

---

## Package Alias

Gunakan alias jika nama package bentrok atau untuk shorthand:

```go
import (
    "fmt"
    myfmt "mypackage/format"  // Alias
    "net/http"
)

func main() {
    fmt.Println("normal")
    myfmt.PrintHello()  // Pakai alias
}
```

### Dot Import

Tidak direkomendasikan, tapi bisa untuk testing:

```go
import (
    . "fmt"  // Semua identifier langsung accessible
)

func main() {
    Println("Tanpa fmt.")  // Bisa langsung Println
}
```

### Blank Identifier

Untuk执行 side effects dari package:

```go
import (
    _ "image/png"  // Register PNG decoder, tidak perlu akses identifier
)
```

---

## Struktur Project

### Single Module

```
myproject/
├── go.mod
├── main.go
└── helpers.go
```

### Multi-Package

```
myproject/
├── go.mod
├── cmd/
│   └── app/
│       └── main.go          → package main
├── internal/
│   ├── auth/
│   │   ├── auth.go
│   │   └── token.go
│   └── db/
│       └── database.go
├── pkg/
│   └── utils/
│       └── string.go
└── go.sum
```

### Module dengan Multiple Executable

```
multicmd/
├── go.mod
├── cmd/
│   ├── app1/
│   │   └── main.go
│   └── app2/
│       └── main.go
└── lib/
    └── lib.go
```

```bash
# Build specific module
go build -o app1 ./cmd/app1
go build -o app2 ./cmd/app2
```

---

## Internal Package

Package `internal` hanya bisa di-import oleh parent module:

```
mymodule/
├── go.mod
├── cmd/
│   └── main.go
└── internal/
    └── utils/     → Hanya bisa di-import oleh mymodule
```

---

## Replace Directive

Untuk development, bisa replace import path:

```go
// go.mod
module myproject

go 1.21

require (
    github.com/original/pkg v1.0.0
)

replace github.com/original/pkg => ../local/pkg
```

---

## Vendor

Vendoring menyimpan semua dependency lokal:

```bash
go mod vendor  # Buat direktori vendor/
```

```
project/
├── go.mod
├── go.sum
├── vendor/
│   └── github.com/
│       └── pkg/
│           └── ...
└── main.go
```

Build dengan vendor:

```bash
go build -mod=vendor
```

---

## Praktik Terbaik

### Do's ✅

```go
// 1. Nama package deskriptif
package userService

// 2. Satu package per direktori
// mypackage/
//     ├── file1.go    → package mypackage
//     └── file2.go    → package mypackage

// 3. Group import dengan blank line
import (
    "fmt"

    "github.com/pkg/a"
    "github.com/pkg/b"
)

// 4. Minimalisir scope
func main() {
    if err := doSomething(); err != nil {
        log.Fatal(err)
    }
}
```

### Don'ts ❌

```go
// 1. Jangan export everything
var internalState int  // Huruf kecil memang sengaja

// 2. Jangan circular import
// a.go imports b.go, b.go imports a.go = ❌

// 3. Jangan gunakan dot import di production
import . "fmt"  // ❌
```

---

## Konsep Package Patterns

### Package-Oriented Design

```go
package user

type User struct {
    ID    int
    Name  string
    Email string
}

// Constructor
func New(name, email string) *User {
    return &User{
        Name:  name,
        Email: email,
    }
}

// Methods
func (u *User) Validate() error {
    if u.Email == "" {
        return errors.New("email required")
    }
    return nil
}
```

### Dependency Injection via Package

```go
package database

type DB interface {
    Query(string) ([]string, error)
}

var db DB

func SetDB(d DB) {
    db = d
}

func GetAll() ([]string, error) {
    return db.Query("SELECT * FROM items")
}
```

---

## Latihan

### Latihan 1: Buat Package Calculator

Buat package `mathutil` dengan fungsi:
- `Multiply(a, b int) int`
- `Divide(a, b float64) (float64, error)`
- `Max(numbers ...int) int`
- `Min(numbers ...int) int`

Pastikan ada unit test.

### Latihan 2: Organize Project

Buat struktur project kecil dengan:
- `cmd/cli/main.go` - entry point
- `internal/strutil/` - package untuk string utilities
- `pkg/validator/` - package untuk validation

### Latihan 3: Import Alias

Ambil 2 package dengan nama fungsi yang sama, gunakan alias untuk keduanya, dan demo penggunaannya.

### Latihan 4: Init dengan Config

Buat package `config` dengan:
- Variabel global `Config`
- Fungsi `init()` yang load config dari environment
- Fungsi `Get()` untuk akses config

---

## ➡️ Selanjutnya

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/13-package-module/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/13-package-module
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/13-package-module
go build -o 13-test main.go
./13-test
```

### Apa yang Ditest?

File `main.go` dan subpackage `util` berisi test untuk:

- Struktur package lokal (`util`)
- Penggunaan subpackage dari `main`
- Contoh `go.mod` pada root module

---

## ➡️ Selanjutnya

**[File I/O]**  
→ Lanjut ke: `14-file-io.md`
