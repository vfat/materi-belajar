---
topik: String & Formatting
urutan: 11 dari 18
posisi: lanjutan
sebelumnya: Error Handling
---

> 🔗 **Lanjutan dari:** Error Handling  
> ← Kembali ke: `10-error-handling.md`

# String & Formatting

## Tujuan Belajar

- Memahami string sebagai slice of bytes
- Melakukan operasi string dasar (len, concatenation, comparison)
- Menggunakan package `strings` untuk manipulasi
- Memformat output dengan `fmt`
- Konversi antara string dan tipe lain

---

## Dasar String di Go

String di Go adalah **slice of bytes** yang immutable (tidak bisa diubah setelah dibuat).

```go
package main

import "fmt"

func main() {
    // String literal
    s1 := "Hello, Go!"

    // Panjang string (dalam bytes)
    fmt.Println("Panjang:", len(s1))  // 9

    // Akses karakter (byte)
    fmt.Println("Karakter pertama:", s1[0])      // 72 (ASCII 'H')
    fmt.Println("Karakter pertama:", string(s1[0])) // "H"

    // String kosong
    empty := ""
    fmt.Println("Kosong:", len(empty) == 0)  // true
}
```

---

## Rune (Unicode Character)

Go mendukung full Unicode. Gunakan `rune` untuk karakter Unicode:

```go
import "fmt"

func main() {
    s := "Halo Dunia"

    // Iterasi dengan rune (Unicode)
    for i, r := range s {
        fmt.Printf("Index: %d, Rune: %c\n", i, r)
    }

    // Panjang dalam rune (bukan bytes)
    runes := []rune(s)
    fmt.Println("Jumlah karakter:", len(runes))  // 11

    // String Unicode
    emoji := "🎉 Selamat!"
    fmt.Println("Panjang bytes:", len(emoji))   // 17 (UTF-8 encoding)
    fmt.Println("Panjang runes:", len([]rune(emoji)))  // 10
}
```

---

## String Operations

### Concatenation (Gabungkan)

```go
func main() {
    // Cara 1: operator +
    s1 := "Hello" + " " + "World"
    fmt.Println(s1)  // Hello World

    // Cara 2: fmt.Sprintf
    name := "John"
    age := 30
    s2 := fmt.Sprintf("Nama: %s, Umur: %d", name, age)
    fmt.Println(s2)  // Nama: John, Umur: 30

    // Cara 3: strings.Join
    parts := []string{"Go", "is", "awesome"}
    s3 := strings.Join(parts, " ")
    fmt.Println(s3)  // Go is awesome

    // Builder (efisien untuk banyak concatenation)
    var sb strings.Builder
    for i := 0; i < 3; i++ {
        sb.WriteString("Go ")
    }
    fmt.Println(sb.String())  // Go Go Go
}
```

### Comparison

```go
import "fmt"
import "strings"

func main() {
    s1, s2 := "hello", "hello"
    s3 := "world"

    // == (case-sensitive)
    fmt.Println(s1 == s2)  // true
    fmt.Println(s1 == s3)  // false

    // strings.EqualFold (case-insensitive)
    fmt.Println(strings.EqualFold("Hello", "hello"))  // true
    fmt.Println(strings.EqualFold("Go", "go"))        // true
}
```

### Comparison Operators

```go
func main() {
    fmt.Println("apple" < "banana")  // true (alphabetical)
    fmt.Println("a" > "z")          // false

    // strings.Compare
    fmt.Println(strings.Compare("abc", "abd"))  // -1 (abc < abd)
    fmt.Println(strings.Compare("abc", "abc"))  // 0 (sama)
    fmt.Println(strings.Compare("abd", "abc"))  // 1 (abd > abc)
}
```

---

## Package strings

### strings.Contains

```go
import "strings"

func main() {
    text := "The quick brown fox"

    fmt.Println(strings.Contains(text, "quick"))    // true
    fmt.Println(strings.Contains(text, "cat"))      // false
    fmt.Println(strings.Contains("", "cat"))        // false
    fmt.Println(strings.Contains("cat", ""))        // true
}
```

### strings.HasPrefix / HasSuffix

```go
func main() {
    url := "https://example.com"

    fmt.Println(strings.HasPrefix(url, "https://"))  // true
    fmt.Println(strings.HasSuffix(url, ".com"))      // true
}
```

### strings.ToUpper / ToLower / Title

```go
func main() {
    text := "hello world"

    fmt.Println(strings.ToUpper(text))   // HELLO WORLD
    fmt.Println(strings.ToLower(text))   // hello world
    fmt.Println(strings.ToTitle(text))   // HELLO WORLD

    // Title membuat setiap kata kapital
    fmt.Println(strings.ToTitle("he said hello"))  // HE SAID HELLO
}
```

### strings.Trim / Split / Join

```go
func main() {
    // Trim whitespace
    s := "  Hello Go!  "
    fmt.Printf("[%s]\n", strings.TrimSpace(s))    // [Hello Go!]
    fmt.Printf("[%s]\n", strings.Trim(s, " "))   // [Hello Go!]

    // Split
    csv := "apple,banana,cherry"
    parts := strings.Split(csv, ",")
    fmt.Println(parts)  // [apple banana cherry]

    // SplitAfter (保留 delimiter)
    withDelim := strings.SplitAfter(csv, ",")
    fmt.Println(withDelim)  // [apple, banana, cherry]

    // Join
    words := []string{"Hello", "World"}
    fmt.Println(strings.Join(words, "-"))  // Hello-World
}
```

### strings.Replace / ReplaceAll

```go
func main() {
    text := "hello world, hello go"

    // Replace (ganti n kali)
    r1 := strings.Replace(text, "hello", "hi", 1)  // hi world, hello go
    fmt.Println(r1)

    // ReplaceAll (ganti semua)
    r2 := strings.ReplaceAll(text, "hello", "hi")  // hi world, hi go
    fmt.Println(r2)
}
```

### strings.Count / Repeat

```go
func main() {
    // Count occurrences
    fmt.Println(strings.Count("banana", "a"))    // 3
    fmt.Println(strings.Count("hello", "ll"))    // 1

    // Repeat
    fmt.Println(strings.Repeat("Go", 3))  // GoGoGo
}
```

### strings.Index / LastIndex

```go
func main() {
    text := "Hello World"

    fmt.Println(strings.Index(text, "o"))       // 4
    fmt.Println(strings.LastIndex(text, "o"))   // 7
    fmt.Println(strings.Index(text, "X"))       // -1 (tidak ditemukan)
}
```

### Fields dan Splits

```go
func main() {
    // Fields memecah berdasarkan whitespace
    text := "Go   is   awesome"
    fields := strings.Fields(text)
    fmt.Println(fields)  // [Go is awesome]

    // SplitN untuk limitar jumlah hasil
    path := "a/b/c/d"
    parts := strings.SplitN(path, "/", 2)
    fmt.Println(parts)  // [a b/c/d]
}
```

---

## Package strconv (String Conversion)

### Int ↔ String

```go
import "strconv"

func main() {
    // Int ke String
    i := 42
    s1 := strconv.Itoa(i)        // "42"
    s2 := strconv.FormatInt(int64(i), 10)  // "42"

    // String ke Int
    n1, _ := strconv.Atoi("123")     // 123
    n2, _ := strconv.ParseInt("255", 10, 64)  // 255

    fmt.Println(s1, n1, n2)
}
```

### Float ↔ String

```go
func main() {
    // Float ke String
    f := 3.14159
    s := strconv.FormatFloat(f, 'f', 2, 64)  // "3.14"
    fmt.Println(s)

    // String ke Float
    f2, _ := strconv.ParseFloat("2.718", 64)
    fmt.Println(f2)
}
```

### Bool ↔ String

```go
func main() {
    // Bool ke String
    fmt.Println(strconv.FormatBool(true))   // "true"
    fmt.Println(strconv.FormatBool(false))  // "false"

    // String ke Bool
    fmt.Println(strconv.ParseBool("true"))   // true
    fmt.Println(strconv.ParseBool("True"))   // false (case-sensitive!)
    fmt.Println(strconv.ParseBool("1"))       // true
    fmt.Println(strconv.ParseBool("false"))  // false
}
```

---

## Package fmt (Formatting)

### Print

```go
func main() {
    name := "Alice"
    age := 30

    // Print (tanpa newline)
    fmt.Print("Hello ")
    fmt.Print("World\n")

    // Println (dengan newline)
    fmt.Println("Nama:", name)
    fmt.Println("Umur:", age)

    // Printf (formatted)
    fmt.Printf("Nama: %s, Umur: %d\n", name, age)
    fmt.Printf("Pi: %f, Hex: %x\n", 3.14159, 255)
}
```

### Format Verbs

| Verb | Fungsi | Contoh |
|------|--------|--------|
| `%s` | String | `"hello"` |
| `%d` | Integer (decimal) | `42` |
| `%f` | Float | `3.14` |
| `%t` | Boolean | `true` |
| `%c` | Character | `'A'` |
| `%x` | Hexadecimal | `ff` |
| `%o` | Octal | `77` |
| `%b` | Binary | `1010` |
| `%p` | Pointer (address) | `0xc000...` |
| `%v` | Any value |varies|
| `%T` | Type | `string` |

### Width dan Precision

```go
func main() {
    // Width (lebar minimum)
    fmt.Printf("[%5d]\n", 42)    // [   42]
    fmt.Printf("[%-5d]\n", 42)   // [42   ]  (kiri)

    // Float precision
    fmt.Printf("[%8.2f]\n", 3.14)   // [    3.14]
    fmt.Printf("[%.2f]\n", 3.14159) // [3.14]

    // String width
    fmt.Printf("[%10s]\n", "Go")    // [        Go]
}
```

### Sprintf dan Sprint

```go
func main() {
    name := "Bob"
    score := 95.5

    // Sprintf - format ke string
    msg := fmt.Sprintf("%s scored %.1f", name, score)
    fmt.Println(msg)

    // Sprint - gabung tanpa newline
    line := fmt.Sprint("A", "B", "C")
    fmt.Println(line)

    // Sprintln - gabung dengan newline
    lines := fmt.Sprintln("X", "Y", "Z")
    fmt.Print(lines)

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/11-string-formatting/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/11-string-formatting
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/11-string-formatting
go build -o 11-test main.go
./11-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Basic `fmt` formatting (`%s`, `%d`, `%.2f`)
- `fmt.Sprintf` dan berbagai verb
- Konversi dengan `strconv`
- Formatting waktu (`time.Format`)
- Width, padding dan alignment
- Type/value printing dengan `%#v` dan `%T`

---

## ➡️ Selanjutnya

**[Time & Date]**
→ Lanjut ke: `12-time-date.md`
}
```

### Table Printing

```go
func main() {
    fmt.Println("=== Daftar Nilai ===")
    fmt.Println("No  Nama      Nilai")
    fmt.Println("--- --------- -----")

    for i, name := range []string{"Alice", "Bob", "Charlie"} {
        fmt.Printf("%-3d %-9s %5.1f\n", i+1, name, float64(80+i*5))
    }
}
```

---

## String Builder dan Buffer

Untuk concatenation banyak string (efisien):

```go
import "strings"
import "fmt"

func main() {
    // strings.Builder (paling efisien)
    var b1 strings.Builder
    for i := 0; i < 3; i++ {
        b1.WriteString("Go ")
    }
    fmt.Println(b1.String())  // Go Go Go

    // bytes.Buffer
    import "bytes"
    var b2 bytes.Buffer
    b2.WriteString("Hello")
    b2.WriteByte(' ')
    b2.Write([]byte("World"))
    fmt.Println(b2.String())  // Hello World
}
```

---

## Raw String Literal

Gunakan backtick untuk string literal (tanpa escape):

```go
func main() {
    // Regular string - perlu escape
    path := "C:\\Users\\Name\\Documents"
    fmt.Println(path)

    // Raw string - tidak perlu escape
    raw := `C:\Users\Name\Documents`
    fmt.Println(raw)

    // Berguna untuk JSON, SQL, regex
    json := `{"name": "John", "age": 30}`
    sql := `SELECT * FROM users WHERE id = 1`
    regex := `\d+\.\d+`
}
```

---

## Latihan

### Latihan 1: Palindrome Checker

Buat fungsi `IsPalindrome(s string) bool` yang mengecek apakah string adalah palindrome (dibaca sama dari depan dan belakang). Abaikan case dan spasi.

Contoh: `"Radar"` → true, `"Hello"` → false

### Latihan 2: Word Counter

Buat fungsi `WordCount(text string) map[string]int` yang menghitung frekuensi setiap kata. Hint: gunakan `strings.Fields` dan `strings.ToLower`.

### Latihan 3: CSV Parser

Parse string `"Nama,Umur,Kota"` menjadi slice `["Nama", "Umur", "Kota"]`. Gunakan `strings.Split`. Lalu gabungkan lagi dengan `strings.Join` menjadi string baru.

### Latihan 4: Format Table

Buat program yang menampilkan table daftar buku dengan format rapi menggunakan `fmt.Printf`:

```
Judul              Penulis      Tahun
------------------ ------------ ------
Go Programming     John Doe     2024
Learning Python    Jane Smith   2023
Clean Code         Bob Wilson   2022
```

---

## ➡️ Selanjutnya

**[Time & Date]**  
→ Lanjut ke: `12-time-date.md`
