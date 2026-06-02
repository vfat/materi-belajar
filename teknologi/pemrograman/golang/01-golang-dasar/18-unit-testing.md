---
topik: Unit Testing
urutan: 18 dari 18
posisi: akhir
sebelumnya: Channel
---

> 🔗 **Lanjutan dari:** Channel  
> ← Kembali ke: `17-channel.md`

# Unit Testing

## Tujuan Belajar

- Memahami konsep testing dalam Go
- Menulis unit test dengan `testing` package
- Menggunakan fungsi `testing.T`
- Menulis table-driven tests
- Menjalankan benchmark tests
- Testing dengan subtests
- Mocking dan dependency injection
- Coverage test

---

## Package testing

Go sudah menyediakan package `testing` bawaan:

```go
import "testing"
```

### Aturan Penamaan

| File | File Test |
|------|-----------|
| `math.go` | `math_test.go` |
| `user.go` | `user_test.go` |
| `calculator.go` | `calculator_test.go` |

**Penting:** Nama file test harus berakhiran `_test.go`

---

## Test Function

### Basic Test

```go
// hello.go
package main

func SayHello(name string) string {
    return "Hello, " + name + "!"
}
```

```go
// hello_test.go
package main

import "testing"

func TestSayHello(t *testing.T) {
    result := SayHello("World")
    expected := "Hello, World!"

    if result != expected {
        t.Errorf("SayHello(\"World\") = %q; want %q", result, expected)
    }
}
```

### Menjalankan Test

```bash
# Jalankan semua test
go test

# Jalankan dengan verbose
go test -v

# Jalankan test tertentu
go test -v -run TestSayHello
```

Output:
```
=== RUN   TestSayHello
--- PASS: TestSayHello (0.00s)
PASS
```

---

## testing.T Methods

| Method | Fungsi |
|--------|--------|
| `t.Error(args)` | Log error tapi test tetap lanjut |
| `t.Errorf(format, args)` | Log formatted error |
| `t.Fatal(args)` | Log error dan stop test |
| `t.Fatalf(format, args)` | Log formatted error dan stop test |
| `t.Skip(args)` | Skip test ini |
| `t.Skipf(format, args)` | Skip dengan alasan formatted |
| `t.Log(args)` | Log info (biasanya muncul saat -v) |

### Error vs Fatal

```go
func TestWithFatal(t *testing.T) {
    result := someFunction()
    if result == "" {
        t.Fatal("Result cannot be empty")  // Stop di sini
    }
    // Code di bawah tidak dijalankan jika Fatal dipanggil
    t.Log("This won't print")
}
```

---

## Assert Helper

### Buat Sendiri

```go
// testutil_test.go
package main

import "testing"

func AssertEqual(t *testing.T, got, want interface{}) {
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

func AssertNil(t *testing.T, got interface{}) {
    if got != nil {
        t.Errorf("got %v, want nil", got)
    }
}

func AssertNotNil(t *testing.T, got interface{}) {
    if got == nil {
        t.Errorf("got nil, want non-nil value")
    }
}

// Usage
func TestSayHello(t *testing.T) {
    result := SayHello("World")
    AssertEqual(t, result, "Hello, World!")
}
```

---

## Table-Driven Tests

### Basic Table-Driven

```go
// math_test.go
package main

import "testing"

// Fungsi yang di-test
func Add(a, b int) int {
    return a + b
}

// Table-driven test
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive + positive", 1, 2, 3},
        {"positive + negative", 5, -3, 2},
        {"negative + negative", -1, -1, -2},
        {"zero + number", 0, 10, 10},
        {"zero + zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

### Output dengan Subtests

```bash
go test -v

=== RUN   TestAdd
=== RUN   TestAdd/positive_+_positive
=== RUN   TestAdd/positive_+_negative
=== RUN   TestAdd/negative_+_negative
=== RUN   TestAdd/zero_+_number
=== RUN   TestAdd/zero_+_zero
--- PASS: TestAdd (0.00s)
    --- PASS: TestAdd/positive_+_positive (0.00s)
    --- PASS: TestAdd/positive_+_negative (0.00s)
    --- PASS: TestAdd/negative_+_negative (0.00s)
    --- PASS: TestAdd/zero_+_number (0.00s)
    --- PASS: TestAdd/zero_+_zero (0.00s)
PASS
```

---

## Multiple Test Cases

```go
func TestStringFunctions(t *testing.T) {
    tests := []struct {
        name string
        fn   func() bool
    }{
        {"Contains", func() bool {
            return strings.Contains("Hello World", "World")
        }},
        {"HasPrefix", func() bool {
            return strings.HasPrefix("Hello", "Hel")
        }},
        {"ToUpper", func() bool {
            return strings.ToUpper("hello") == "HELLO"
        }},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if !tt.fn() {
                t.Error("test failed")
            }
        })
    }
}
```

---

## Error Testing

### Testing Error Cases

```go
// Fungsi yang di-test
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Test untuk error
func TestDivide(t *testing.T) {
    t.Run("division by zero", func(t *testing.T) {
        _, err := Divide(10, 0)
        if err == nil {
            t.Error("Divide(10, 0) expected error, got nil")
        }
    })

    t.Run("valid division", func(t *testing.T) {
        result, err := Divide(10, 2)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if result != 5 {
            t.Errorf("Divide(10, 2) = %f; want 5", result)
        }
    })
}
```

### Test dengan Error Types

```go
import "errors"

func TestErrorTypes(t *testing.T) {
    _, err := someFunction()

    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        t.Logf("Validation error: %v", validationErr)
    }
}
```

---

## Benchmark Tests

### Basic Benchmark

```go
// benchmark_test.go
package main

import "testing"

func Sum(numbers []int) int {
    sum := 0
    for _, n := range numbers {
        sum += n
    }
    return sum
}

func BenchmarkSum(b *testing.B) {
    numbers := make([]int, 1000)
    for i := range numbers {
        numbers[i] = i
    }

    b.ResetTimer()  // Reset timer setelah setup
    for i := 0; i < b.N; i++ {
        Sum(numbers)
    }
}
```

### Menjalankan Benchmark

```bash
# Jalankan benchmark
go test -bench=.

# Benchmark spesifik
go test -bench=BenchmarkSum

# Dengan memory stats
go test -bench=. -benchmem

# Dengan CPU profile
go test -bench=. -cpuprofile=cpu.prof
```

Output:
```
goos: linux
goarch: amd64
pkg: github.com/user/project
BenchmarkSum-8    1000000    200 ns/op
```

### Benchmark Comparison

```bash
# Run benchmark 3 times
go test -bench=. -count=3

# Compare with different inputs
go test -bench=. -benchtime=5s
```

---

## TestMain

### Setup dan Teardown Global

```go
// main_test.go
package main

import (
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    // Setup
    print("Setting up tests...\n")

    // Jalankan semua tests
    exitCode := m.Run()

    // Teardown
    print("Cleaning up...\n")

    os.Exit(exitCode)
}

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/18-unit-testing/`

### Cara Menjalankan Test

**1. Jalankan semua test (recursive):**
```bash
cd minilab/go/18-unit-testing
go test ./...
```

**2. Verbose / specific / bench / coverage:**
```bash
go test -v ./...
go test -run TestAdd ./...
go test -bench .
go test -cover ./...
go test -race ./...
```

### Apa yang Ditest?

Folder `mathutil` berisi contoh fungsi dan file test untuk:

- Unit tests (table-driven)
- Example tests
- Benchmarks
- Error case testing

---

## ➡️ Selesai

Seluruh materi selesai tercakup dalam roadmap.
```

---

## Setup dan Teardown per Test

```go
package main

import (
    "testing"
)

type TestDB struct {
    // database connection
}

func setupTestDB(t *testing.T) *TestDB {
    t.Helper()
    // Setup database connection
    db := &TestDB{}
    return db
}

func teardownTestDB(db *TestDB) {
    // Cleanup
    db.Close()
}

func TestUserCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(db)

    // Test code here
    t.Run("Create", func(t *testing.T) {
        // test create
    })

    t.Run("Read", func(t *testing.T) {
        // test read
    })
}
```

---

## Mocking

### Interface-Based Mocking

```go
// service.go
package main

type UserRepository interface {
    GetByID(id int) (*User, error)
    Save(user *User) error
}

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUserName(id int) (string, error) {
    user, err := s.repo.GetByID(id)
    if err != nil {
        return "", err
    }
    return user.Name, nil
}
```

### Mock Implementation

```go
// mock_test.go
package main

import (
    "errors"
    "testing"
)

type MockUserRepo struct {
    users map[int]*User
}

func NewMockUserRepo() *MockUserRepo {
    return &MockUserRepo{
        users: map[int]*User{
            1: {ID: 1, Name: "Alice"},
            2: {ID: 2, Name: "Bob"},
        },
    }
}

func (m *MockUserRepo) GetByID(id int) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, errors.New("user not found")
}

func (m *MockUserRepo) Save(user *User) error {
    m.users[user.ID] = user
    return nil
}

func TestGetUserName(t *testing.T) {
    mock := NewMockUserRepo()
    service := &UserService{repo: mock}

    t.Run("existing user", func(t *testing.T) {
        name, err := service.GetUserName(1)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
        if name != "Alice" {
            t.Errorf("GetUserName(1) = %q; want %q", name, "Alice")
        }
    })

    t.Run("non-existing user", func(t *testing.T) {
        _, err := service.GetUserName(999)
        if err == nil {
            t.Error("expected error for non-existing user")
        }
    })
}
```

---

## HTTP Handler Testing

### Testing HTTP Handlers

```go
// handler_test.go
package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealthHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/health", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(HealthHandler)

    handler.ServeHTTP(rr, req)

    // Check status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status: got %v want %v", status, http.StatusOK)
    }

    // Check response body
    var resp map[string]string
    json.Unmarshal(rr.Body.Bytes(), &resp)

    if resp["status"] != "ok" {
        t.Errorf("status = %q; want %q", resp["status"], "ok")
    }
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

---

## Test Coverage

### Melihat Coverage

```bash
# Coverage untuk semua test
go test -cover

# Coverage dengan detail per function
go test -coverprofile=coverage.out

# View coverage di browser
go tool cover -html=coverage.out

# Coverage per package
go test -coverprofile=coverage.out ./...
```

### Coverage Report

```
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
coverage: 85.7% of statements
```

### Generate HTML Coverage Report

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## Subtests dengan Shared Setup

```go
func TestDatabaseOperations(t *testing.T) {
    // Setup sekali
    db := setupTestDB(t)
    defer db.Close()

    t.Run("Create", func(t *testing.T) {
        // test create dengan db
    })

    t.Run("Read", func(t *testing.T) {
        // test read dengan db
    })

    t.Run("Update", func(t *testing.T) {
        // test update dengan db
    })

    t.Run("Delete", func(t *testing.T) {
        // test delete dengan db
    })
}
```

---

## Skip Tests

### Skip Based on Condition

```go
import (
    "testing"
    "runtime"
)

func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping test in short mode")
    }
    // Slow test code
}

func TestOSSpecific(t *testing.T) {
    if runtime.GOOS != "windows" {
        t.Skip("skipping test on non-Windows OS")
    }
}

func TestWithEnv(t *testing.T) {
    if os.Getenv("SKIP_INTEGRATION") == "true" {
        t.Skip("skipping integration test")
    }
}
```

### Conditional Skip with Build Tags

```go
// +build integration

package main

import "testing"

func TestIntegration(t *testing.T) {
    // Integration test
}
```

```bash
# Run dengan build tag
go test -tags=integration
```

---

## Example Tests

### Documented Examples

```go
// example_test.go
package main

import "fmt"

func ExampleAdd() {
    sum := Add(2, 3)
    fmt.Println(sum)
    // Output: 5
}
```

### Example dengan Multiple Outputs

```go
func ExampleSplit() {
    parts := split("a-b-c", "-")
    fmt.Println(parts[0])
    fmt.Println(parts[1])
    fmt.Println(parts[2])
    // Output:
    // a
    // b
    // c
}
```

---

## Best Practices

### Do's

```go
// ✅ Buat nama test yang jelas
func TestUserService_Create_Success(t *testing.T) {}

// ✅ Test normal dan error cases
func TestDivide(t *testing.T) {
    t.Run("success", ...)
    t.Run("division by zero", ...)
}

// ✅ Gunakan subtests untuk grouping
t.Run("valid input", func(t *testing.T) { ... })

// ✅ Cleanup resources
defer cleanup()

// ✅ Table-driven tests untuk multiple cases
```

### Don'ts

```go
// ❌ Jangan hardcode paths
// Gunakan relative paths atau test utilities

// ❌ Jangan skip tests tanpa alasan
t.Skip("This test is broken")  // Buat issue, jangan di-skip

// ❌ Jangan test internal implementation
// Test behavior, bukan implementation details

// ❌ Jangan ignore errors
result, err := DoSomething()  // Check err!
```

---

## CI/CD Integration

### Common Test Commands

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific packages
go test github.com/user/project/pkg/...

# Timeout
go test -timeout 30s ./...
```

### Makefile Example

```makefile
test:
    go test -v -race -cover ./...

test-short:
    go test -short ./...

bench:
    go test -bench=. -benchmem ./...

coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
```

---

## Latihan

### Latihan 1: String Calculator Test
Buat fungsi `Add(numbers string) int` yang menjumlahkan angka dalam string. Test kasus-kasus: empty string, single number, two numbers, many numbers, newline separator, custom delimiter.

### Latihan 2: HTTP API Test
Buat HTTP handler untuk CRUD user dan test dengan `httptest.NewRecorder()`.

### Latihan 3: Mock Repository
Buat interface `Storage` dengan implementasi mock untuk testing service layer.

### Latihan 4: Benchmark String Concatenation
Buat benchmark untuk membandingkan `+` concatenation vs `strings.Builder`.

### Latihan 5: Table-Driven Tests
Buat table-driven test untuk fungsi validation (email, phone, password strength).

---

## ➡️ Selanjutnya

**[Capstone Project — gotask CLI Task Manager]**  
→ Lanjut ke: `19-capstone-gotask.md`
