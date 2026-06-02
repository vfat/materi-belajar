---
topik: Time & Date
urutan: 12 dari 18
posisi: lanjutan
sebelumnya: String & Formatting
---

> 🔗 **Lanjutan dari:** String & Formatting  
> ← Kembali ke: `11-string-formatting.md`

# Time & Date

## Tujuan Belajar

- Membuat dan memanipulasi waktu dengan package `time`
- Format waktu dengan reference time
- Parse string ke waktu
- Menghitung durasi dan selisih waktu
- Timer dan ticker untuk scheduling

---

## Package time

Go menyediakan package `time` yang lengkap untuk semua kebutuhan waktu.

```go
package main

import "time"
```

---

## Current Time

### time.Now()

```go
func main() {
    now := time.Now()
    fmt.Println(now)         // 2026-05-28 18:30:45.123456 +0700 WIB
    fmt.Println(now.Date())  // 2026 May 28
    fmt.Println(now.Clock()) // 18 30 45
    fmt.Println(now.Year())  // 2026
    fmt.Println(now.Month()) // May
    fmt.Println(now.Day())   // 28
    fmt.Println(now.Hour())  // 18
    fmt.Println(now.Minute()) // 30
    fmt.Println(now.Second()) // 45
}
```

### time.Today() dan time.Now().UTC()

```go
func main() {
    today := time.Now().AddDate(0, 0, 0).Truncate(24 * time.Hour)
    fmt.Println("Today:", today)

    utc := time.Now().UTC()
    fmt.Println("UTC:", utc)
}
```

---

## Creating Time

### time.Date()

```go
func main() {
    // Buat waktu spesifik
    t := time.Date(2026, 12, 25, 10, 30, 0, 0, time.Local)
    fmt.Println(t)  // 2026-12-25 10:30:00 +0700 WIB
}
```

### time.Parse()

Parse string ke time.Time:

```go
func main() {
    // Format standar
    t1, _ := time.Parse("2006-01-02", "2026-05-28")
    fmt.Println(t1)  // 2026-05-28 00:00:00 +0700 WIB

    // Format dengan waktu
    t2, _ := time.Parse("2006-01-02 15:04:05", "2026-05-28 14:30:00")
    fmt.Println(t2)  // 2026-05-28 14:30:00 +0700 WIB
}
```

> ⚠️ **Penting:** Angka "2006-01-02 15:04:05" adalah **reference time** di Go. Jangan diganti!

### Reference Time Layout

```
Jan 2  15:04:05 2006 MST  =  01/02 03:04:05PM 2006 -0700
```

| Komponen | Layout | Contoh |
|----------|--------|--------|
| Year | `2006` | 2026 |
| Month | `01` (numeric) | 05 |
| Month | `Jan` | May |
| Day | `02` | 28 |
| Hour | `15` (24h) | 14 |
| Hour | `03` (12h) | 2 |
| Minute | `04` | 30 |
| Second | `05` | 45 |
| PM/AM | `PM` | PM |

### Parse dengan Location

```go
func main() {
    // Parse UTC
    t1, _ := time.ParseInLocation("2006-01-02 15:04:05", 
        "2026-05-28 14:30:00", time.UTC)
    fmt.Println(t1)

    // Parse Local
    loc, _ := time.LoadLocation("Asia/Jakarta")
    t2, _ := time.ParseInLocation("2006-01-02 15:04:05",
        "2026-05-28 14:30:00", loc)
    fmt.Println(t2)
}
```

---

## Formatting Time

### time.Time.Format()

```go
func main() {
    t := time.Date(2026, 5, 28, 14, 30, 45, 0, time.Local)

    fmt.Println(t.Format("2006-01-02"))              // 2026-05-28
    fmt.Println(t.Format("15:04:05"))                // 14:30:45
    fmt.Println(t.Format("2006-01-02 15:04:05"))     // 2026-05-28 14:30:45
    fmt.Println(t.Format("Mon, 02 Jan 2006"))        // Thu, 28 May 2026
    fmt.Println(t.Format("2 January 2006"))          // 28 May 2026
    fmt.Println(t.Format(time.RFC3339))              // 2026-05-28T14:30:45+07:00
}
```

### Common Layouts

| Layout | Contoh Output |
|--------|--------------|
| `time.RFC3339` | 2026-05-28T14:30:45+07:00 |
| `time.RFC1123` | Thu, 28 May 2026 14:30:45 WIB |
| `time.ANSIC` | Thu May 28 14:30:45 2026 |
| `time.UnixDate` | Thu May 28 14:30:45 WIB 2026 |

---

## Duration (Durasi)

### time.Duration

```go
func main() {
    // Berbagai cara buat Duration
    d1 := 5 * time.Second
    d2 := time.Duration(3000) * time.Millisecond
    d3, _ := time.ParseDuration("2h30m")

    fmt.Println(d1)  // 5s
    fmt.Println(d2)  // 3s
    fmt.Println(d3)  // 2h30m0s

    // Konversi
    fmt.Println(d3.Hours())    // 2.5
    fmt.Println(d3.Minutes()) // 150
    fmt.Println(d3.Seconds()) // 9000
}
```

### Duration Formatting

```go
func main() {
    d, _ := time.ParseDuration("1h25m30s")

    fmt.Printf("%d jam %d menit %d detik\n",
        int(d.Hours()),
        int(d.Minutes())%60,
        int(d.Seconds())%60)
    // Output: 1 jam 25 menit 30 detik
}
```

---

## Arithmetic (Operasi Waktu)

### Add dan Subtract

```go
func main() {
    now := time.Now()

    // Tambah waktu
    tomorrow := now.AddDate(0, 0, 1)
    fmt.Println("Besok:", tomorrow.Format("2006-01-02"))

    nextWeek := now.Add(7 * 24 * time.Hour)
    fmt.Println("Minggu depan:", nextWeek.Format("2006-01-02"))

    // Kurang waktu
    yesterday := now.Add(-24 * time.Hour)
    fmt.Println("Kemarin:", yesterday.Format("2006-01-02"))
}
```

### Sub (Selisih Waktu)

```go
func main() {
    t1 := time.Date(2026, 5, 28, 10, 0, 0, 0, time.Local)
    t2 := time.Date(2026, 5, 28, 15, 30, 0, 0, time.Local)

    diff := t2.Sub(t1)
    fmt.Println("Selisih:", diff)        // 5h30m0s
    fmt.Println("Dalam jam:", diff.Hours())   // 5.5
    fmt.Println("Dalam detik:", diff.Seconds()) // 19800
}
```

### Compare

```go
func main() {
    t1 := time.Now()
    t2 := time.Now().Add(1 * time.Hour)
    t3 := t1

    fmt.Println(t1.Before(t2))  // true
    fmt.Println(t2.After(t1))  // true
    fmt.Println(t1.Equal(t3))  // true
}
```

---

## Timer dan Ticker

### Timer (Jalankan sekali)

```go
import (
    "fmt"
    "time"
)

func main() {
    fmt.Println("Mulai...")
    
    timer := time.NewTimer(2 * time.Second)
    <-timer.C  // Tunggu sampai timer selesai

    fmt.Println("Timer selesai!")
}

// Atau dengan After (lebih simple)
func main() {
    fmt.Println("Mulai...")
    <-time.After(2 * time.Second)
    fmt.Println("2 detik berlalu!")
}
```

### Ticker (Jalankan berulang)

```go
func main() {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    count := 0
    for range ticker.C {
        count++
        fmt.Printf("Tick %d\n", count)
        if count >= 5 {
            break
        }
    }
    fmt.Println("Ticker selesai")
}
```

### Sleep

```go
func main() {
    fmt.Println("Awal")
    time.Sleep(1 * time.Second)
    fmt.Println("1 detik kemudian")

    // Sleep dengan fmt
    fmt.Print("Loading")
    for i := 0; i < 3; i++ {
        time.Sleep(500 * time.Millisecond)
        fmt.Print(".")
    }
    fmt.Println(" Done!")
}
```

---

## Unix Timestamp

### Konversi ke Unix

```go
func main() {
    now := time.Now()

    // Unix seconds
    unixSec := now.Unix()
    fmt.Println("Unix seconds:", unixSec)

    // Unix milliseconds
    unixMs := now.UnixMilli()
    fmt.Println("Unix milliseconds:", unixMs)

    // Unix nanoseconds
    unixNs := now.UnixNano()
    fmt.Println("Unix nanoseconds:", unixNs)
}
```

### Konversi dari Unix

```go
func main() {
    unixSec := int64(1751134245)

    // Dari Unix seconds
    t1 := time.Unix(unixSec, 0)
    fmt.Println("Dari Unix:", t1.Format("2006-01-02 15:04:05"))

    // Dari Unix milliseconds
    t2 := time.UnixMilli(1751134245000)
    fmt.Println("Dari UnixMs:", t2.Format("2006-01-02 15:04:05"))
}
```

---

## Location dan Timezone

### Load Location

```go
func main() {
    // Dari nama timezone
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        panic(err)
    }

    now := time.Now().In(loc)
    fmt.Println("WIB:", now.Format("15:04:05 MST"))
}
```

### Timezone Umum

```go
func main() {
    locations := []string{
        "Asia/Jakarta",  // WIB
        "Asia/Makassar", // WITA
        "Asia/Jayapura", // WIT
        "America/New_York",
        "Europe/London",
        "Asia/Tokyo",
    }

    now := time.Now()
    for _, locName := range locations {
        loc, _ := time.LoadLocation(locName)
        t := now.In(loc)
        fmt.Printf("%-20s %s\n", locName, t.Format("15:04:05"))
    }
}
```

---

## Contoh Praktis

### Hitung Umur

```go
func calculateAge(birthDate time.Time) int {
    today := time.Now()
    age := today.Year() - birthDate.Year()

    // Kurangi 1 jika belum melewati birthday
    if today.YearDay() < birthDate.YearDay() {
        age--
    }
    return age
}

func main() {
    birthday, _ := time.Parse("2006-01-02", "2000-05-15")
    age := calculateAge(birthday)
    fmt.Printf("Umur: %d tahun\n", age)
}
```

### Countdown Timer

```go
func countdown(seconds int) {
    for i := seconds; i > 0; i-- {
        fmt.Printf("\rWaktu tersisa: %d detik", i)
        time.Sleep(1 * time.Second)
    }
    fmt.Println("\n⏰ Waktu habis!")
}

func main() {
    countdown(5)
}
```

### Rate Limiter Sederhana

```go
import "time"

type RateLimiter struct {
    lastCall time.Time
    interval time.Duration
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
    return &RateLimiter{
        interval: interval,
    }
}

func (r *RateLimiter) Allow() bool {
    now := time.Now()
    if now.Sub(r.lastCall) >= r.interval {
        r.lastCall = now
        return true
    }
    return false
}

func main() {
    limiter := NewRateLimiter(500 * time.Millisecond)

    for i := 0; i < 5; i++ {
        if limiter.Allow() {
            fmt.Printf("Request %d: OK\n", i+1)
        } else {
            fmt.Printf("Request %d: Ditolak\n", i+1)
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

---

## Latihan

### Latihan 1: Format Tanggal Indonesia

Buat fungsi `FormatIndonesian(t time.Time) string` yang menghasilkan format: "28 Mei 2026, 14:30:45"

### Latihan 2: Selisih Hari

Buat fungsi `DaysBetween(t1, t2 time.Time) int` yang menghitung jumlah hari antara dua tanggal (nilai absolut).

### Latihan 3: Workdays Calculator

Buat fungsi `WorkdaysBetween(start, end time.Time) int` yang menghitung jumlah hari kerja (Senin-Jumat) antara dua tanggal.

### Latihan 4: Timeout Channel

Implementasikan pattern timeout: jalankan function yang mengambil waktu, jika lebih dari 3 detik, abort dan print "Timeout!".

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/12-time-date/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/12-time-date
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/12-time-date
go build -o 12-test main.go
./12-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- `time.Now()` dan format RFC3339
- Parsing waktu dengan layout custom
- `time.ParseDuration` dan operasi add/sub
- Timezone / `time.LoadLocation`
- Unix timestamps (`Unix`, `UnixNano`)
- Common layouts (`ANSIC`, `RFC1123Z`)

---

## ➡️ Selanjutnya

**[Package & Module]**  
→ Lanjut ke: `13-package-module.md`
