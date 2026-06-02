---
topik: Channel
urutan: 17 dari 18
posisi: lanjutan
sebelumnya: Goroutine (Concurrency)
---

> 🔗 **Lanjutan dari:** Goroutine (Concurrency)  
> ← Kembali ke: `16-goroutine.md`

# Channel

## Tujuan Belajar

- Memahami konsep channel sebagai alat komunikasi antar goroutines
- Membuat dan menggunakan channel
- Channel buffering vs unbuffered
- Select statement untuk multiple channels
- Channel direction dan closed channel
- Channel patterns (fan-in, fan-out, pipeline)

---

## Apa itu Channel?

Channel adalah **pipa** untuk komunikasi antar goroutines. Mirip tube yang bisa digunakan untuk mengirim dan menerima data.

```
Goroutine A  ──────►  [CHANNEL]  ──────►  Goroutine B
   send                              receive
```

```go
// Membuat channel
ch := make(chan int)
```

---

## Channel Dasar

### Send dan Receive

```go
import "fmt"

func main() {
    // Buat channel
    ch := make(chan string)

    // Kirim data ke channel (dari goroutine)
    go func() {
        ch <- "Hello from goroutine!"
    }()

    // Terima data dari channel
    message := <-ch
    fmt.Println(message)
}
```

### Channel Arrow Direction

```go
ch := make(chan int)

// Send only (hanya bisa kirim)
go func(ch chan<- int) {
    ch <- 42
}(ch)

// Receive only (hanya bisa terima)
go func(ch <-chan int) {
    val := <-ch
    fmt.Println("Received:", val)
}(ch)
```

| Syntax | Tipe | Fungsi |
|--------|------|--------|
| `chan T` | bidirectional | Bisa kirim & terima |
| `chan<- T` | send only | Hanya kirim |
| `<-chan T` | receive only | Hanya terima |

---

## Unbuffered Channel

### Karakteristik

- Kapasitas 0, tidak ada buffer
- **Send akan blocking** sampai ada receive
- **Receive akan blocking** sampai ada send
- Kedua operasi harus ready bersamaan

```go
ch := make(chan int)  // Unbuffered

go func() {
    fmt.Println("Before send")
    ch <- 100  // Blocking di sini
    fmt.Println("After send")
}()

fmt.Println("Before receive")
val := <-ch  // Blocking di sini
fmt.Println("After receive:", val)
```

### Sinkronisasi Otomatis

```go
func main() {
    done := make(chan bool)

    go func() {
        fmt.Println("Task sedang dikerjakan...")
        time.Sleep(time.Second)
        done <- true  // Sinyal selesai
    }()

    <-done  // Tunggu sampai selesai
    fmt.Println("Task selesai!")
}
```

---

## Buffered Channel

### Karakteristik

- Memiliki kapasitas buffer
- **Send hanya block** jika buffer penuh
- **Receive hanya block** jika buffer kosong

```go
// Buffer size 3
ch := make(chan int, 3)

// Bisa kirim 3 data tanpa blocking
ch <- 1
ch <- 2
ch <- 3

// Channel penuh, send ke-4 akan block
// ch <- 4  // Ini akan block!

// Terima untuk mengosongkan buffer
fmt.Println(<-ch)  // 1
fmt.Println(<-ch)  // 2
```

### Buffer vs Unbuffered

```go
import (
    "fmt"
    "time"
)

func test(name string, ch chan bool) {
    time.Sleep(time.Second)
    <-ch  // Receive
    fmt.Println(name, "received signal")
}

func main() {
    // Unbuffered: perlu receiver ready
    // ch := make(chan bool)

    // Buffered: sender tidak perlu wait
    ch := make(chan bool, 1)

    go test("Task-1", ch)
    go test("Task-2", ch)

    ch <- true  // Send dengan buffer
    ch <- true

    time.Sleep(2 * time.Second)
}
```

---

## Channel Length dan Capacity

```go
ch := make(chan int, 5)

// Length: jumlah data di buffer
fmt.Println("Length:", len(ch))   // 0

// Capacity: kapasitas buffer
fmt.Println("Capacity:", cap(ch)) // 5

ch <- 1
ch <- 2
ch <- 3

fmt.Println("Length:", len(ch))   // 3
```

---

## Close Channel

### Menutup Channel

```go
ch := make(chan int)

go func() {
    for i := 1; i <= 5; i++ {
        ch <- i
    }
    close(ch)  // Tutup channel
}()

// Terima data sampai channel ditutup
for val := range ch {
    fmt.Println("Received:", val)
}

fmt.Println("Channel closed!")
```

### Deteksi Channel Tertutup

```go
ch := make(chan int)
close(ch)

// Receive dari channel tertutup
val, ok := <-ch
fmt.Println("Value:", val, "Open:", ok)
// Output: Value: 0 Open: false
```

| `ok` value | Arti |
|------------|------|
| `true` | Channel terbuka, ada data |
| `false` | Channel tertutup |

---

## Select Statement

### Multiple Channel Operations

```go
ch1 := make(chan string)
ch2 := make(chan string)

go func() {
    time.Sleep(time.Second)
    ch1 <- "From channel 1"
}()

go func() {
    time.Sleep(500 * time.Millisecond)
    ch2 <- "From channel 2"
}()

for i := 0; i < 2; i++ {
    select {
    case msg1 := <-ch1:
        fmt.Println(msg1)
    case msg2 := <-ch2:
        fmt.Println(msg2)
    }
}
```

### Select dengan Timeout

```go
ch := make(chan string)

go func() {
    time.Sleep(2 * time.Second)
    ch <- "Delayed response"
}()

select {
case msg := <-ch:
    fmt.Println(msg)
case <-time.After(1 * time.Second):
    fmt.Println("Timeout! No response received")
}
```

### Select Default Case

```go
ch := make(chan int, 1)
ch <- 1

select {
case msg := <-ch:
    fmt.Println("Received:", msg)
default:
    fmt.Println("No message, continuing...")
}
```

---

## Channel Patterns

### 1. Fan-Out (Distribusi Kerja)

```go
import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(200 * time.Millisecond)
    }
}

func main() {
    jobs := make(chan int, 10)
    var wg sync.WaitGroup

    // Start 3 workers
    for w := 1; w <= 3; w++ {
        wg.Add(1)
        go worker(w, jobs, &wg)
    }

    // Send jobs
    for j := 1; j <= 9; j++ {
        jobs <- j
    }
    close(jobs)

    wg.Wait()
    fmt.Println("All done!")
}
```

### 2. Fan-In (Combine Results)

```go
func merge(ch1, ch2 <-chan int) <-chan int {
    result := make(chan int)

    go func() {
        defer close(result)
        for val := range ch1 {
            result <- val
        }
    }()

    go func() {
        defer close(result)
        for val := range ch2 {
            result <- val
        }
    }()

    return result
}
```

### 3. Pipeline

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func main() {
    // Pipeline: generate -> square -> print

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/17-channel/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/17-channel
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/17-channel
go build -o 17-test main.go
./17-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Send/receive pada channel
- Close + range
- Buffered channel
- Fan-in menggunakan `select`
- Pipeline pattern (`generator` -> `sq`)
- Perilaku `nil` channel di `select`

---

## ➡️ Selanjutnya

**[Unit Testing]**  
→ Lanjut ke: `18-unit-testing.md`
    c := generate(2, 3, 4, 5)
    out := square(c)

    for result := range out {
        fmt.Println(result)  // 4, 9, 16, 25
    }
}
```

---

## Context dengan Channel

### Context Cancellation

```go
import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, ch chan int) {
    for i := 0; ; i++ {
        select {
        case <-ctx.Done():
            fmt.Println("Worker cancelled")
            return
        case ch <- i:
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    ch := make(chan int)

    go worker(ctx, ch)

    for i := 0; i < 5; i++ {
        fmt.Println("Received:", <-ch)
    }

    cancel()  // Hentikan worker
    time.Sleep(100 * time.Millisecond)
}
```

---

## Channel vs Mutex

### Kapan Gunakan Channel?

```go
// ✅ Distribusi kerja / task
jobs := make(chan Job, 100)

// ✅ Mengirim data antar goroutines
result := make(chan Result)

// ✅ Signaling antar goroutines
done := make(chan struct{})

// ✅ Rate limiting
throttle := make(chan struct{}, maxConcurrent)
```

### Kapan Gunakan Mutex?

```go
// ✅ Proteksi state bersama
var (
    mu   sync.Mutex
    data map[string]int
)

// ✅ Counter sederhana
var counter int
```

| Penggunaan | Channel | Mutex |
|------------|---------|-------|
| Komunikasi data | ✅ | ❌ |
| Proteksi state | ❌ | ✅ |
| Sinkronisasi | ✅ | ✅ |
| Distribusi kerja | ✅ | ❌ |

---

## Channel Beste Practices

### Do's

```go
// ✅ Tutup channel dari sisi sender
ch := make(chan int)
go func() {
    defer close(ch)
    for i := 0; i < 5; i++ {
        ch <- i
    }
}()

// ✅ Gunakan range untuk terima data
for val := range ch {
    fmt.Println(val)
}

// ✅ Beri nama dengan jelas
jobQueue := make(chan Job, 100)
resultChan := make(chan Result, 100)
done := make(chan struct{})
```

### Don'ts

```go
// ❌ Jangan kirim ke channel yang sudah ditutup
// panic: send on closed channel

// ❌ Jangan close channel receive-only
// compile error

// ❌ Jangan forgot to close channel
// Memory leak jika tidak ditutup
```

---

## Contoh: Concurrent HTTP Requests

```go
import (
    "context"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

type Result struct {
    URL  string
    Body string
    Err  error
}

func fetchURLs(urls []string) []Result {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    results := make(chan Result, len(urls))

    for _, url := range urls {
        go func(u string) {
            req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                results <- Result{URL: u, Err: err}
                return
            }
            defer resp.Body.Close()
            body, _ := ioutil.ReadAll(resp.Body)
            results <- Result{URL: u, Body: string(body)}
        }(url)
    }

    var out []Result
    for i := 0; i < len(urls); i++ {
        out = append(out, <-results)
    }
    return out
}

func main() {
    urls := []string{
        "https://jsonplaceholder.typicode.com/posts/1",
        "https://jsonplaceholder.typicode.com/posts/2",
        "https://jsonplaceholder.typicode.com/posts/3",
    }

    results := fetchURLs(urls)
    for _, r := range results {
        if r.Err != nil {
            fmt.Printf("Error %s: %v\n", r.URL, r.Err)
        } else {
            fmt.Printf("Success %s: %d bytes\n", r.URL, len(r.Body))
        }
    }
}
```

---

## Contoh: Semaphore dengan Channel

```go
import (
    "fmt"
    "time"
)

// Semaphore dengan buffered channel
func runWithLimit(maxConcurrent int, tasks []func()) {
    semaphore := make(chan struct{}, maxConcurrent)
    done := make(chan struct{})

    for _, task := range tasks {
        go func(t func()) {
            semaphore <- struct{}{}  // Acquire
            t()
            <-semaphore  // Release
        }(task)
    }

    // Wait semua selesai
    for i := 0; i < maxConcurrent; i++ {
        semaphore <- struct{}{}
    }
    close(done)
}

func main() {
    tasks := []func(){
        func() { fmt.Println("Task 1"); time.Sleep(200*time.Millisecond) },
        func() { fmt.Println("Task 2"); time.Sleep(200*time.Millisecond) },
        func() { fmt.Println("Task 3"); time.Sleep(200*time.Millisecond) },
        func() { fmt.Println("Task 4"); time.Sleep(200*time.Millisecond) },
        func() { fmt.Println("Task 5"); time.Sleep(200*time.Millisecond) },
    }

    runWithLimit(2, tasks)
    fmt.Println("All tasks completed!")
}
```

---

## Latihan

### Latihan 1: Simple Pipeline
Buat pipeline: generate -> filter -> print. Generate angka 1-20, filter hanya genap, print hasilnya.

### Latihan 2: Merge Channels
Buat fungsi `mergeChannels(chs ...<-chan int) <-chan int` yang menggabungkan beberapa channel menjadi satu.

### Latihan 3: Timeout Channel
Buat fungsi `withTimeout(ch <-chan T, duration time.Duration) (T, bool)` yang return false jika timeout.

### Latihan 4: Producer-Consumer
Buat producer yang menghasilkan work dan consumer yang memproses. Gunakan buffered channel sebagai queue.

### Latihan 5: Rate Limiter
Implementasikan rate limiter yang membatasi maksimal N request per detik menggunakan channel.

---

## ➡️ Selanjutnya

**[Unit Testing]**  
→ Lanjut ke: `18-unit-testing.md`
