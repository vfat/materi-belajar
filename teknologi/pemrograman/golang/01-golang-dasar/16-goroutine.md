---
topik: Goroutine (Concurrency)
urutan: 16 dari 18
posisi: lanjutan
sebelumnya: JSON Handling
---

> 🔗 **Lanjutan dari:** JSON Handling  
> ← Kembali ke: `15-json-handling.md`

# Goroutine (Concurrency)

## Tujuan Belajar

- Memahami perbedaan concurrency vs parallelism
- Membuat goroutine dengan keyword `go`
- Memahami anonymous goroutine
- Mengenal race condition dan cara mengatasinya
- Menggunakan `sync.Mutex` dan `sync.WaitGroup`

---

## Concurrency vs Parallelism

### Konsep Dasar

| Istilah | Penjelasan |
|---------|------------|
| **Concurrency** | Mampu mengelola banyak tugas secara bersamaan (tidak harus simultan) |
| **Parallelism** | Menjalankan banyak tugas secara bersamaan di waktu yang sama (butuh multi-core) |

```go
// Concurrency: satu core, switching antar task
// Parallelism: banyak core, task berjalan bersamaan
```

### Contoh Perbandingan

```go
// Sequential (satu per satu)
func sequential() {
    doTask1()
    doTask2()
    doTask3()
}

// Concurrent (bergantian, bisa paralel jika multi-core)
func concurrent() {
    go doTask1()
    go doTask2()
    go doTask3()
}
```

---

## Membuat Goroutine

### Basic Goroutine

```go
import (
    "fmt"
    "time"
)

func sayHello() {
    fmt.Println("Hello!")
}

func main() {
    // Jalankan sebagai goroutine
    go sayHello()

    // Tunggu sebentar agar goroutine selesai
    time.Sleep(time.Second)
    fmt.Println("Main function done")
}
```

### Tanpa Sleep (Langsung Keluar)

```go
func main() {
    go fmt.Println("This won't print")

    // main() selesai → program terminate
    // goroutine belum dijalankan
    fmt.Println("Main done")
}
// Output: Main done
// Goroutine "This won't print" tidak muncul!
```

### Menggunakan fmt.Scanln

```go
func main() {
    go func() {
        fmt.Println("Goroutine running...")
    }()

    fmt.Scanln()  // Tunggu input user
    fmt.Println("Main done")
}
```

---

## Anonymous Goroutine

### Dengan Function Literal

```go
func main() {
    // Anonymous goroutine dengan closure
    go func() {
        for i := 1; i <= 3; i++ {
            fmt.Printf("Counting: %d\n", i)
            time.Sleep(100 * time.Millisecond)
        }
    }()

    time.Sleep(time.Second)
    fmt.Println("Done")
}
```

### Dengango Keyword + Function Call

```go
func main() {
    message := "Hello from goroutine!"

    go func(msg string) {
        fmt.Println(msg)
    }(message)  // Pass argument

    time.Sleep(time.Second)
}
```

---

## sync.WaitGroup

### Masalah: Kapan Main Selesai?

```go
import (
    "fmt"
    "time"
)

func task(name string) {
    for i := 1; i <= 3; i++ {
        fmt.Printf("[%s] step %d\n", name, i)
        time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    go task("A")
    go task("B")
    // Main tidak tahu kapan goroutines selesai!
    // Program bisa terminate sebelum goroutines selesai
}
```

### Solusi: WaitGroup

```go
import (
    "fmt"
    "sync"
    "time"
)

var wg sync.WaitGroup

func task(name string) {
    defer wg.Done()  // Panggil saat function selesai
    for i := 1; i <= 3; i++ {
        fmt.Printf("[%s] step %d\n", name, i)
        time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    wg.Add(2)  // Ada 2 goroutines

    go task("A")
    go task("B")

    wg.Wait()  // Tunggu sampai semua Done
    fmt.Println("All tasks completed!")
}
```

Output:
```
[A] step 1
[B] step 1
[A] step 2
[B] step 2
[A] step 3
[B] step 3
All tasks completed!
```

---

## Race Condition

### Apa itu Race Condition?

Ketika 2+ goroutines mengakses data yang sama secara bersamaan dan minimal satu mengakses untuk menulis.

```go
var counter int = 0

func increment() {
    counter++  // Race condition di sini!
}

func main() {
    for i := 0; i < 1000; i++ {
        go increment()
    }
    time.Sleep(time.Second)
    fmt.Println("Counter:", counter)
    // Hasil tidak konsisten! (bukan 1000)
}
```

### Deteksi Race Condition

Jalankan dengan flag `-race`:

```bash
go run -race main.go
```

Output jika ada race:
```
WARNING: DATA RACE
Read at 0x00...
Write at 0x00...
```

---

## sync.Mutex

### Lock untuk Perlindungan Data

```go
import (
    "fmt"
    "sync"
    "time"
)

var (
    counter int
    mutex   sync.Mutex
)

func increment() {
    mutex.Lock()         // Kunci akses
    counter++
    mutex.Unlock()       // Lepas kunci
}

func main() {
    for i := 0; i < 1000; i++ {
        go increment()
    }

    time.Sleep(time.Second)

    mutex.Lock()
    fmt.Println("Counter:", counter)
    mutex.Unlock()
    // Hasil: 1000 (konsisten)
}
```

### Alternative: sync.Mutex dengan Method

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}
```

---

## sync.RWMutex

### Read-Write Lock

```go
import (
    "fmt"
    "sync"
    "time"
)

var (
    data     string
    rwMutex  sync.RWMutex
)

func read() {
    rwMutex.RLock()
    defer rwMutex.RUnlock()
    fmt.Println("Read:", data)
    time.Sleep(10 * time.Millisecond)
}

func write(newData string) {
    rwMutex.Lock()
    defer rwMutex.Unlock()
    fmt.Println("Writing:", newData)
    data = newData
    time.Sleep(10 * time.Millisecond)
}

func main() {
    // Multiple readers bisa bersamaan
    for i := 0; i < 5; i++ {
        go read()
    }
    go write("Hello World")

    time.Sleep(100 * time.Millisecond)
}
```

| Method | Fungsi |
|--------|--------|
| `RLock()` | Lock untuk baca (banyak yang bisa baca bersamaan) |
| `RUnlock()` | Unlock baca |
| `Lock()` | Lock untuk tulis (hanya 1 yang bisa tulis) |
| `Unlock()` | Unlock tulis |

---

## sync.Once

### Jalankan Kode Sekali Saja

```go
import (
    "fmt"
    "sync"
)

var (
    once     sync.Once
    instance *Database
)

type Database struct {
    connection string
}

func getDatabase() *Database {
    once.Do(func() {
        fmt.Println("Creating database connection...")
        instance = &Database{connection: "connected"}
    })
    return instance
}

func main() {
    // Connection hanya dibuat sekali meskipun dipanggil berkali-kali
    db1 := getDatabase()
    db2 := getDatabase()
    db3 := getDatabase()

    fmt.Printf("Same instance? %v\n", db1 == db2 && db2 == db3)
    // Output: Same instance? true
    // "Creating database connection..." hanya muncul 1 kali
}

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/16-goroutine/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/16-goroutine
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/16-goroutine
go build -o 16-test main.go
./16-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Menjalankan goroutine sederhana
- Unbuffered vs buffered channels
- Worker pool dengan `sync.WaitGroup`
- `select` dengan timeout
- Race condition demo dan penggunaan `sync.Mutex`

---

## ➡️ Selanjutnya

**[Channel]**  
→ Lanjut ke: `17-channel.md`
```

---

## sync.Map

### Thread-Safe Map

```go
import (
    "fmt"
    "sync"
)

var safeMap sync.Map

func main() {
    // Store
    safeMap.Store("name", "John")
    safeMap.Store("age", 30)

    // Load
    value, ok := safeMap.Load("name")
    fmt.Println("Name:", value, "Found:", ok)

    // LoadOrStore
    val, loaded := safeMap.LoadOrStore("country", "Indonesia")
    fmt.Println("Country:", val, "Loaded:", loaded)

    // Delete
    safeMap.Delete("age")

    // Range
    safeMap.Store("x", 1)
    safeMap.Store("y", 2)
    safeMap.Range(func(key, value interface{}) bool {
        fmt.Printf("%s = %v\n", key, value)
        return true
    })
}
```

---

## Runtime GOMAXPROCS

### Jumlah CPU untuk Goroutines

```go
import (
    "fmt"
    "runtime"
)

func main() {
    // Default: jumlah CPU cores
    fmt.Println("NumCPU:", runtime.NumCPU())
    fmt.Println("NumGoroutine:", runtime.NumGoroutine())

    // Set maksimal CPU yang digunakan
    runtime.GOMAXPROCS(2)

    fmt.Println("NumCPU after GOMAXPROCS:", runtime.NumCPU())
}
```

| Function | Fungsi |
|----------|--------|
| `runtime.NumCPU()` | Jumlah CPU cores tersedia |
| `runtime.NumGoroutine()` | Jumlah goroutine aktif |
| `runtime.GOMAXPROCS(n)` | Set maksimal CPU untuk paralel |

---

## Context dengan Goroutine

### Cancellation Pattern

```go
import (
    "context"
    "fmt"
    "time"
)

func task(ctx context.Context, name string) {
    for i := 0; i < 5; i++ {
        select {
        case <-ctx.Done():
            fmt.Printf("[%s] cancelled\n", name)
            return
        default:
            fmt.Printf("[%s] working %d\n", name, i)
            time.Sleep(200 * time.Millisecond)
        }
    }
    fmt.Printf("[%s] done\n", name)
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    go task(ctx, "Task-1")
    go task(ctx, "Task-2")

    time.Sleep(time.Second)
    fmt.Println("Main finished")
}
```

---

## Best Practices

### Do's

```go
// ✅ Gunakan WaitGroup untuk sinkronisasi
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()
wg.Wait()

// ✅ Gunakan mutex untuk proteksi data bersama
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()

// ✅ Gunakan context untuk cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

### Don'ts

```go
// ❌ Jangan gunakan time.Sleep untuk sync
go doSomething()
time.Sleep(time.Second)  // TIDAK reliable!

// ❌ Jangan lock lebih dari perlu
// Lock hanya bagian yang perlu proteksi

// ❌ Jangan buat terlalu banyak goroutines
// Gunakan worker pool jika perlu banyak
```

---

## Contoh: Worker Pool

```go
import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(100 * time.Millisecond)
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // Start 3 workers
    var wg sync.WaitGroup
    for w := 1; w <= 3; w++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            worker(id, jobs, results)
        }(w)
    }

    // Send 9 jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)

    // Wait workers dan close results
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    for result := range results {
        fmt.Printf("Result: %d\n", result)
    }
}
```

---

## Latihan

### Latihan 1: Parallel Processing
Buat 3 fungsi yang masing-masing sleep 1 detik. Jalankan secara concurrent dan hitung total waktu. Apakah lebih cepat dari sequential?

### Latihan 2: Bank Account
Simulasikan bank account dengan multiple goroutines yang melakukan deposit dan withdraw. Gunakan mutex untuk proteksi dan pastikan balance tidak pernah negatif.

### Latihan 3: Producer-Consumer
Buat producer yang menghasilkan angka 1-10 dan consumer yang menampilkannya. Gunakan buffered channel untuk komunikasi.

### Latihan 4: Race Detector
Buat kode dengan race condition, jalankan dengan `go run -race` dan perbaiki dengan mutex.

### Latihan 5: Singleton Pattern
Implementasikan singleton pattern dengan `sync.Once` untuk database connection.

---

## ➡️ Selanjutnya

**[Channel]**  
→ Lanjut ke: `17-channel.md`
