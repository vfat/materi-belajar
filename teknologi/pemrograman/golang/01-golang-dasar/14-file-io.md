---
topik: File I/O
urutan: 14 dari 18
posisi: lanjutan
sebelumnya: Package & Module
---

> 🔗 **Lanjutan dari:** Package & Module  
> ← Kembali ke: `13-package-module.md`

# File I/O

## Tujuan Belajar

- Membaca file (text dan binary)
- Menulis file (text dan binary)
- Manipulasi path dan direktori
- Menggunakan buffered I/O
- Error handling untuk file operations

---

## Baca File

### os.ReadFile (Cara Paling Simpel)

Baca seluruh file sekaligus:

```go
import (
    "fmt"
    "os"
)

func main() {
    data, err := os.ReadFile("test.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(string(data))
}
```

### os.Open + bufio.Reader

Baca file baris per baris:

```go
import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    file, err := os.Open("test.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    lineNum := 1
    for scanner.Scan() {
        fmt.Printf("%d: %s\n", lineNum, scanner.Text())
        lineNum++
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error scanner:", err)
    }
}
```

### Baca dengan bufio.Reader (Manual)

```go
import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    file, _ := os.Open("test.txt")
    defer file.Close()

    reader := bufio.NewReader(file)
    for {
        line, err := reader.ReadString('\n')
        if err != nil && len(line) == 0 {
            break
        }
        fmt.Print(line)
    }
}
```

### Baca File Chunk (Bagian-bagian)

```go
import (
    "fmt"
    "os"
)

func main() {
    file, err := os.Open("large.bin")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    buffer := make([]byte, 1024)  // 1KB buffer
    for {
        n, err := file.Read(buffer)
        if n == 0 {
            break
        }
        fmt.Printf("Dibaca %d bytes\n", n)
        if err != nil {
            break
        }
    }
}
```

---

## Tulis File

### os.WriteFile (Cara Paling Simpel)

Tulis seluruh file sekaligus:

```go
import (
    "os"
)

func main() {
    content := "Hello, World!\nBaris kedua"

    err := os.WriteFile("output.txt", []byte(content), 0644)
    if err != nil {
        panic(err)
    }
}
```

### os.Create + bufio.Writer

```go
import (
    "bufio"
    "fmt"
    "os"
)

func main() {
    file, err := os.Create("output.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    writer := bufio.NewWriter(file)

    fmt.Fprintln(writer, "Baris pertama")
    fmt.Fprintln(writer, "Baris kedua")
    fmt.Fprintf(writer, "Angka: %d\n", 42)

    writer.Flush()  // Penting! Flush ke disk
}
```

### Append ke File

```go
import (
    "os"
)

func main() {
    // Buka dengan flag O_APPEND
    file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    defer file.Close()

    _, err = file.WriteString("Entry baru\n")
    if err != nil {
        panic(err)
    }
}
```

---

## File Info dan Permissions

### os.Stat

Dapatkan informasi file:

```go
import (
    "fmt"
    "os"
)

func main() {
    info, err := os.Stat("test.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Nama:", info.Name())
    fmt.Println("Ukuran:", info.Size(), "bytes")
    fmt.Println("Mode:", info.Mode())
    fmt.Println("Modifikasi:", info.ModTime())
    fmt.Println("Apakah direktori:", info.IsDir())
}
```

### File Permissions (os.FileMode)

```go
import "os"

func main() {
    info, _ := os.Stat("test.txt")

    mode := info.Mode()

    // Cek tipe
    fmt.Println("Is regular:", mode.IsRegular())
    fmt.Println("Is dir:", mode.IsDir())

    // Cek permission
    fmt.Println("Owner read:", mode&00400 != 0)
    fmt.Println("Owner write:", mode&00200 != 0)
    fmt.Println("Owner exec:", mode&00100 != 0)
}
```

---

## Direktori

### os.Mkdir dan os.MkdirAll

```go
import (
    "os"
)

func main() {
    // Buat satu direktori
    err := os.Mkdir("mydir", 0755)
    if err != nil {
        fmt.Println("Error:", err)
    }

    // Buat direktori beserta parent-nya
    err = os.MkdirAll("path/to/deep/dir", 0755)
}
```

### os.Remove dan os.RemoveAll

```go
func main() {
    // Hapus file
    os.Remove("file.txt")

    // Hapus direktori (kosong)
    os.Remove("emptydir")

    // Hapus direktori beserta isinya
    os.RemoveAll("somedir")
}
```

### Baca Direktori

```go
import (
    "fmt"
    "os"
)

func main() {
    entries, err := os.ReadDir(".")
    if err != nil {
        panic(err)
    }

    for _, entry := range entries {
        fmt.Printf("%s\t", entry.Name())
        if entry.IsDir() {
            fmt.Println("[DIR]")
        } else {
            info, _ := entry.Info()
            fmt.Printf("[FILE] %d bytes\n", info.Size())
        }
    }
}
```

### Walk Directory (Rekursif)

```go
import (
    "fmt"
    "path/filepath"
)

func main() {
    root := "/path/to/dir"

    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            fmt.Printf("[DIR]  %s\n", path)
        } else {
            fmt.Printf("[FILE] %s (%d bytes)\n", path, info.Size())
        }
        return nil
    })
}
```

---

## Path Manipulation

### path/filepath

```go
import (
    "fmt"
    "path/filepath"
)

func main() {
    // Join paths
    p := filepath.Join("dir1", "dir2", "file.txt")
    fmt.Println(p)  // dir1/dir2/file.txt

    // Ambil komponen
    fmt.Println(filepath.Dir(p))      // dir1/dir2
    fmt.Println(filepath.Base(p))     // file.txt
    fmt.Println(filepath.Ext(p))      // .txt

    // Cek exist
    fmt.Println(filepath.IsAbs("/abs/path"))  // true
    fmt.Println(filepath.IsAbs("rel/path"))    // false

    // Split
    dir, file := filepath.Split(p)
    fmt.Println(dir, file)
}
```

### path (Tanpa OS-specific)

```go
import (
    "fmt"
    "path"
)

func main() {
    // Selalu pakai forward slash
    p := path.Join("dir1", "dir2", "file.txt")
    fmt.Println(p)  // dir1/dir2/file.txt

    // Join dan clean
    fmt.Println(path.Join("a//b", "..", "c"))  // a/c
    fmt.Println(path.Clean("a/../b"))           // b
}
```

---

## Buffered I/O

### bufio.Reader

```go
import (
    "bufio"
    "os"
)

func main() {
    file, _ := os.Open("largefile.txt")
    defer file.Close()

    reader := bufio.NewReaderSize(file, 4096)  // 4KB buffer

    // Baca byte
    b, _ := reader.ReadByte()
    fmt.Println(string(b))

    // Baca n bytes
    buf := make([]byte, 10)
    n, _ := reader.Read(buf)
    fmt.Println(string(buf[:n]))

    // Peek (liat tanpa consume)
    peek, _ := reader.Peek(5)
    fmt.Println(string(peek))
}
```

### bufio.Writer

```go
import (
    "bufio"
    "os"
)

func main() {
    file, _ := os.Create("output.txt")
    defer file.Close()

    writer := bufio.NewWriterSize(file, 4096)

    writer.WriteString("Hello ")
    writer.WriteByte('W')
    writer.WriteRune('o')
    writer.Write([]byte("rld\n"))

    // Available buffer space
    fmt.Println("Available:", writer.Available())

    writer.Flush()
}
```

---

## Contoh Praktis

### Copy File

```go
import (
    "io"
    "os"
)

func copyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    dest, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dest.Close()

    _, err = io.Copy(dest, source)
    if err != nil {
        return err
    }

    return dest.Sync()  // Flush ke disk
}
```

### Word Counter

```go
import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func countWords(filename string) (int, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanWords)

    count := 0
    for scanner.Scan() {
        count++
    }

    return count, scanner.Err()
}

func main() {
    wordCount, err := countWords("article.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Total kata: %d\n", wordCount)
}
```

### Log File Writer

```go
import (
    "bufio"
    "fmt"
    "os"
    "time"
)

type Logger struct {
    file *os.File
    buf  *bufio.Writer
}

func NewLogger(filename string) *Logger {
    f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(err)
    }
    return &Logger{
        file: f,
        buf:  bufio.NewWriter(f),
    }
}

func (l *Logger) Log(level, message string) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    fmt.Fprintf(l.buf, "[%s] [%s] %s\n", timestamp, level, message)
}

func (l *Logger) Close() {
    l.buf.Flush()
    l.file.Close()
}

func main() {
    logger := NewLogger("app.log")
    defer logger.Close()

    logger.Log("INFO", "Application started")
    logger.Log("ERROR", "Something went wrong")
    logger.Log("INFO", "Application closed")
}
```

---

## Temporary Files

```go
import (
    "fmt"
    "os"
)

func main() {
    // Buat temporary file
    f, err := os.CreateTemp("", "myapp-*.txt")
    if err != nil {
        panic(err)
    }
    defer os.Remove(f.Name())  // Cleanup

    fmt.Println("Temp file:", f.Name())

    f.WriteString("Temporary content")
    f.Close()
}

func main() {
    // Buat temporary directory
    dir, err := os.MkdirTemp("", "myapp-")
    if err != nil {
        panic(err)
    }
    defer os.RemoveAll(dir)

    fmt.Println("Temp dir:", dir)
}
```

---

## File Locking (Unix)

```go
import (
    "fmt"
    "os"
    "syscall"
)

func lockFile(f *os.File) error {
    err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
    if err != nil {
        return fmt.Errorf("failed to lock: %w", err)
    }
    return nil
}

func unlockFile(f *os.File) error {
    return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
```

---

## Latihan

### Latihan 1: Line Counter

Buat program yang menghitung:
- Jumlah baris
- Jumlah kata
- Jumlah karakter

dalam file teks.

### Latihan 2: Find and Replace

Buat fungsi `ReplaceInFile(filename, old, new string)` yang membaca file, mengganti semua `old` dengan `new`, dan menulis kembali.

### Latihan 3: Directory Lister

Buat program yang list semua file dalam direktori dengan format:
```
Name          Type    Size    Modified
-------------------------------------------------
document.pdf  FILE    1024    2026-05-28
images/       DIR     -       2026-05-27
```

### Latihan 4: File Splitter

Buat fungsi yang memecah file besar menjadi beberapa file kecil (masing-masing n baris).

### Latihan 5: CSV Reader

Parse file CSV dengan header dan tampilkan sebagai table. Handle edge cases seperti:
- Quoted fields dengan comma
- Empty lines
- Different delimiters

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/14-file-io/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/14-file-io
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/14-file-io
go build -o 14-test main.go
./14-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Menulis file (`os.WriteFile`)
- Membaca file (`os.ReadFile`)
- Append file (`os.OpenFile` + `O_APPEND`)
- Mendapatkan info file (`os.Stat`)
- Menggunakan `os.CreateTemp`
- Menangani error ketika membuka file yang tidak ada
- Membersihkan file setelah test

---

## ➡️ Selanjutnya

**[JSON Handling]**  
→ Lanjut ke: `15-json-handling.md`
