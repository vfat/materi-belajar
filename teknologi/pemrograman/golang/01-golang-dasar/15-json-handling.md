---
topik: JSON Handling
urutan: 15 dari 18
posisi: lanjutan
sebelumnya: File I/O
---

> 🔗 **Lanjutan dari:** File I/O  
> ← Kembali ke: `14-file-io.md`

# JSON Handling

## Tujuan Belajar

- Encode struct ke JSON
- Decode JSON ke struct
- Customize JSON field names dan behavior
- Handle nested dan array JSON
- Stream JSON (Encoder/Decoder)
- Error handling untuk JSON

---

## Package encoding/json

Go menyediakan package `encoding/json` untuk semua kebutuhan JSON:

```go
import "encoding/json"
```

---

## Encode Struct ke JSON

### Basic Encoding

```go
import (
    "encoding/json"
    "fmt"
)

type User struct {
    Name  string
    Email string
    Age   int
}

func main() {
    user := User{Name: "John", Email: "john@example.com", Age: 30}

    jsonData, err := json.Marshal(user)
    if err != nil {
        panic(err)
    }

    fmt.Println(string(jsonData))
    // Output: {"Name":"John","Email":"john@example.com","Age":30}
}
```

### Marshal vs MarshalIndent

```go
func main() {
    user := User{Name: "John", Email: "john@example.com", Age: 30}

    // Compact (tanpa spasi)
    compact, _ := json.Marshal(user)
    fmt.Println(string(compact))

    // Indented (dengan format)
    indent, _ := json.MarshalIndent(user, "", "  ")
    fmt.Println(string(indent))
}
```

Output MarshalIndent:
```json
{
  "Name": "John",
  "Email": "john@example.com",
  "Age": 30
}
```

---

## Decode JSON ke Struct

### Basic Decoding

```go
import (
    "encoding/json"
    "fmt"
)

type User struct {
    Name  string
    Email string
    Age   int
}

func main() {
    jsonStr := `{"Name":"John","Email":"john@example.com","Age":30}`

    var user User
    err := json.Unmarshal([]byte(jsonStr), &user)
    if err != nil {
        panic(err)
    }

    fmt.Printf("%s <%s>, umur %d\n", user.Name, user.Email, user.Age)
}
```

### Decode ke Map/Interface

```go
func main() {
    jsonStr := `{"name":"John","age":30,"active":true}`

    // Decode ke map[string]interface{}
    var data map[string]interface{}
    json.Unmarshal([]byte(jsonStr), &data)

    fmt.Println(data["name"])   // John
    fmt.Println(data["age"])    // 30
    fmt.Println(data["active"]) // true
}
```

---

## Custom Field Names

### json tag

```go
type User struct {
    FullName string `json:"full_name"`          // Ubah nama field
    Email    string `json:"email,omitempty"`     // Hapus jika kosong
    Password string `json:"-"`                  // Jangan di-encode
    Secret   string `json:"secret,omitempty"`   // Jangan decode
    Internal int    `json:"-"`                   // Skip sama sekali
}
```

### Tag Options

| Tag | Fungsi |
|-----|--------|
| `json:"name"` | Rename ke "name" |
| `json:"-"` | Skip field |
| `json:",omitempty"` | Skip jika zero value |
| `json:"name,omitempty"` | Rename + omitempty |

```go
func main() {
    type Product struct {
        ID       int     `json:"id"`
        Name     string  `json:"product_name"`
        Price    float64 `json:"price"`
        InStock  bool    `json:"in_stock,omitempty"`
        Password string  `json:"-"`
    }

    p := Product{ID: 1, Name: "Laptop", Price: 999.99}

    data, _ := json.MarshalIndent(p, "", "  ")
    fmt.Println(string(data))
}
```

Output:
```json
{
  "id": 1,
  "product_name": "Laptop",
  "price": 999.99
}
```

---

## Nested Struct

```go
type Address struct {
    Street  string
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address Address  // Nested struct
}

func main() {
    person := Person{
        Name: "John",
        Age:  30,
        Address: Address{
            Street:  "123 Main St",
            City:    "Jakarta",
            Country: "Indonesia",
        },
    }

    data, _ := json.MarshalIndent(person, "", "  ")
    fmt.Println(string(data))
}
```

Output:
```json
{
  "Name": "John",
  "Age": 30,
  "Address": {
    "Street": "123 Main St",
    "City": "Jakarta",
    "Country": "Indonesia"
  }
}
```

---

## Array/Slice

### Array dalam JSON

```go
type User struct {
    Name   string
    Hobbies []string  // Slice
}

func main() {
    user := User{
        Name:   "John",
        Hobbies: []string{"Coding", "Gaming", "Reading"},
    }

    data, _ := json.MarshalIndent(user, "", "  ")
    fmt.Println(string(data))
}
```

Output:
```json
{
  "Name": "John",
  "Hobbies": [
    "Coding",
    "Gaming",
    "Reading"
  ]
}
```

### Array of Structs

```go
type Product struct {
    Name  string
    Price float64
}

func main() {
    products := []Product{
        {Name: "Laptop", Price: 999.99},
        {Name: "Mouse", Price: 29.99},
        {Name: "Keyboard", Price: 79.99},
    }

    data, _ := json.MarshalIndent(products, "", "  ")
    fmt.Println(string(data))
}
```

---

## Map

### Map ke JSON

```go
func main() {
    data := map[string]interface{}{
        "name":  "John",
        "age":   30,
        "skills": []string{"Go", "Python", "JavaScript"},
    }

    jsonData, _ := json.MarshalIndent(data, "", "  ")
    fmt.Println(string(jsonData))
}
```

### JSON ke Map

```go
func main() {
    jsonStr := `{"name":"John","scores":{"math":90,"science":85}}`

    var result map[string]interface{}
    json.Unmarshal([]byte(jsonStr), &result)

    fmt.Println(result["name"])  // John

    // Akses nested map
    scores := result["scores"].(map[string]interface{})
    fmt.Println(scores["math"])  // 90
}
```

---

## Encoder dan Decoder

### Encoder (Stream ke Writer)

```go
import (
    "encoding/json"
    "os"
)

type Person struct {
    Name string
    Age  int
}

func main() {
    people := []Person{
        {"Alice", 25},
        {"Bob", 30},
    }

    file, _ := os.Create("people.json")
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")

    encoder.Encode(people)
}
```

### Decoder (Stream dari Reader)

```go
import (
    "encoding/json"
    "os"
)

func main() {
    file, _ := os.Open("people.json")
    defer file.Close()

    decoder := json.NewDecoder(file)

    var people []Person
    decoder.Decode(&people)

    for _, p := range people {
        fmt.Printf("%s, %d tahun\n", p.Name, p.Age)
    }
}
```

---

## omitempty dan Zero Values

### Tanpa omitempty

```go
type Config struct {
    Debug  bool
    Level  int
    Host   string
    Port   int
}

func main() {
    c := Config{}  // Semua zero values

    data, _ := json.MarshalIndent(c, "", "  ")
    fmt.Println(string(data))
}
```

Output:
```json
{
  "Debug": false,
  "Level": 0,
  "Host": "",
  "Port": 0
}
```

### Dengan omitempty

```go
type Config struct {
    Debug  bool   `json:"debug,omitempty"`
    Level  int    `json:"level,omitempty"`
    Host   string `json:"host,omitempty"`
    Port   int    `json:"port,omitempty"`
}

func main() {
    c := Config{}

    data, _ := json.MarshalIndent(c, "", "  ")
    fmt.Println(string(data))
}
```

Output:
```json
{}
```

---

## Anonymous Struct

```go
func main() {
    // Struct匿名 untuk data sekali pakai
    data := struct {
        Name    string   `json:"name"`
        Members []string `json:"members"`
    }{
        Name:    "Team Alpha",
        Members: []string{"Alice", "Bob"},
    }

    jsonData, _ := json.MarshalIndent(data, "", "  ")
    fmt.Println(string(jsonData))
}
```

---

## Custom Marshaler dan Unmarshaler

### json.Marshaler Interface

```go
import "encoding/json"

type Money struct {
    Amount   float64
    Currency string
}

// Custom marshal
func (m Money) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]interface{}{
        "amount":   m.Amount,
        "currency": m.Currency,
    })
}

func main() {
    m := Money{Amount: 99.99, Currency: "USD"}

    data, _ := json.MarshalIndent(m, "", "  ")
    fmt.Println(string(data))
}
```

### json.Unmarshaler Interface

```go
import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
)

type Timestamp struct {
    time.Time
}

// Custom unmarshal dari format "DD-MM-YYYY"
func (t *Timestamp) UnmarshalJSON(data []byte) error {
    str := strings.Trim(string(data), `"`)
    parsed, err := time.Parse("02-01-2006", str)
    if err != nil {
        return err
    }
    t.Time = parsed
    return nil
}

func main() {
    jsonStr := `"28-05-2026"`

    var t Timestamp
    json.Unmarshal([]byte(jsonStr), &t)

    fmt.Println(t.Time.Format("2006-01-02"))
}
```

---

## Handle Unknown JSON Structure

### DisallowUnknownFields

```go
type Config struct {
    Name string
    Port int
}

func main() {
    jsonStr := `{"name":"test","unknown_field":123}`

    decoder := json.NewDecoder(strings.NewReader(jsonStr))
    decoder.DisallowUnknownFields()

    var cfg Config
    err := decoder.Decode(&cfg)
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

### Partial Decode

```go
import "github.com/buger/jsonparser"

func main() {
    jsonStr := `{"name":"John","age":30,"extra":"data"}`

    // Parse hanya field yang diperlukan
    name, _ := jsonparser.GetString([]byte(jsonStr), "name")
    age, _ := jsonparser.Int([]byte(jsonStr), "age")

    fmt.Println(name, age)
}
```

---

## Praktik Terbaik

### Valid JSON String

```go
func IsValidJSON(s string) bool {
    var js json.RawMessage
    return json.Unmarshal([]byte(s), &js) == nil
}
```

### Pretty Print dengan Error Handling

```go
import (
    "encoding/json"
    "fmt"
)

func PrettyPrint(v interface{}) {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Println(string(data))
}
```

### Clone/Deep Copy dengan JSON

```go
func CloneJSON(v interface{}) interface{} {
    data, _ := json.Marshal(v)
    var clone interface{}
    json.Unmarshal(data, &clone)
    return clone
}
```

---

## Contoh Lengkap: REST API Response

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// API Response structure
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func SuccessResponse(w http.ResponseWriter, data interface{}) {
    resp := APIResponse{
        Success: true,
        Data:    data,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
    resp := APIResponse{
        Success: false,
        Error:   message,
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(resp)
}

func main() {
    // Demo
    sampleData := map[string]interface{}{
        "users": []map[string]interface{}{
            {"name": "Alice", "age": 25},
            {"name": "Bob", "age": 30},
        },
        "total": 2,
    }

    // Simulate API response
    jsonData, _ := json.MarshalIndent(APIResponse{
        Success: true,
        Data:    sampleData,
    }, "", "  ")

    fmt.Println(string(jsonData))
}
```

---

## Latihan

### Latihan 1: User Profile

Definisikan struct untuk user profile dengan field:
- ID (int)
- Username (string)
- Email (string)
- CreatedAt (time.Time, format "2006-01-02T15:04:05Z")
- Tags ([]string)

Marshal dan print hasilnya.

### Latihan 2: Parse API Response

Ambil JSON dari API public (misal: https://jsonplaceholder.typicode.com/users/1) dan decode ke struct yang sesuai.

### Latihan 3: Merge JSON

Ambil dua JSON string, decode ke map, merge, lalu marshal ulang.

### Latihan 4: Config Loader

Buat fungsi `LoadConfig(filename string) (*Config, error)` yang membaca file JSON config dan decode ke struct Config.

### Latihan 5: Custom Time Format

Implementasikan custom marshaler untuk type Date yang output formatnya "DD Month YYYY" (contoh: "28 May 2026").

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/15-json-handling/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/15-json-handling
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/15-json-handling
go build -o 15-test main.go
./15-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- `json.Marshal` dan `json.MarshalIndent`
- `json.Unmarshal` ke `struct` dan `map[string]interface{}`
- `json.Decoder`/`json.Encoder` (streaming)
- Error handling untuk JSON tidak valid

---

## ➡️ Selanjutnya

**[Goroutine]**  
→ Lanjut ke: `16-goroutine.md`
