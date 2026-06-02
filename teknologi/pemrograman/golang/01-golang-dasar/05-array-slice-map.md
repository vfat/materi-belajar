---
topik: Array, Slice & Map
urutan: 5 dari 6
posisi: lanjutan
sebelumnya: Control Flow (If/Else, Switch, For)
---

> 🔗 **Lanjutan dari:** Control Flow (If/Else, Switch, For)  
> ← Kembali ke: `04-control-flow.md`

# Array, Slice & Map

## Tujuan Belajar

- Memahami Array (ukuran tetap)
- Mengenal Slice (dinamis, lebih fleksibel)
- Menggunakan Map (key-value pairs)
- Operasi dasar: add, remove, search

---

## Array

Array adalah kumpulan data dengan **ukuran tetap**.

```
Array: [0] [1] [2] [3] [4]
       10   20   30   40   50
       ↑index
```

### Declarasi Array

```go
// Cara 1: dengan nilai
var angka [5]int = [5]int{1, 2, 3, 4, 5}

// Cara 2: tanpa ukuran (auto)
nama := [...]string{"Budi", "Siti", "Ani"}

// Cara 3: dengan ukuran, kosong (zero value)
var kosong [3]int  // [0, 0, 0]
```

### Akses & Modifikasi

```go
buah := [4]string{"Apel", "Jeruk", "Mangga", "Pisang"}

fmt.Println(buah[0])   // Apel
fmt.Println(buah[3])   // Pisang

buah[1] = "Nanas"      // Ubah elemen
fmt.Println(buah)      // [Apel Nanas Mangga Pisang]
```

### Panjang Array

```go
angka := [5]int{1, 2, 3, 4, 5}
fmt.Println(len(angka))  // 5
```

### Iterasi Array

```go
// Dengan index
for i := 0; i < len(angka); i++ {
    fmt.Println(i, angka[i])
}

// Dengan range
for index, value := range angka {
    fmt.Printf("%d: %d\n", index, value)
}
```

---

## Slice

Slice adalah "view" atau ** referensi ke Array**. Ukurannya **bisa berubah** (dinamis).

```
Array underlying: [0] [1] [2] [3] [4] [5] [6]
                   10   20   30   40   50   60   70

Slice A (index 1-4): [20, 30, 40, 50]
Slice B (index 3-6): [40, 50, 60, 70]
```

### Membuat Slice

```go
// Dari array
arr := [5]int{1, 2, 3, 4, 5}
slice := arr[1:4]           // index 1, 2, 3 → [2, 3, 4]

// Langsung dengan make
slice1 := make([]int, 5)       // []int{0, 0, 0, 0, 0}
slice2 := make([]int, 3, 10)   // len=3, cap=10

// Langsung dengan literal
buah := []string{"Apel", "Jeruk", "Mangga"}
```

### len vs cap

| Fungsi | Arti |
|--------|------|
| `len()` | Jumlah elemen saat ini |
| `cap()` | Kapasitas (maksimum tanpa alokasi ulang) |

```go
s := make([]int, 3, 10)
fmt.Println(len(s))  // 3
fmt.Println(cap(s))  // 10
```

### Append — Menambah Elemen

```go
angka := []int{1, 2, 3}
fmt.Println(angka)   // [1 2 3]

angka = append(angka, 4, 5)
fmt.Println(angka)   // [1 2 3 4 5]

// Append slice ke slice (gunakan ...)
more := []int{6, 7, 8}
angka = append(angka, more...)
fmt.Println(angka)   // [1 2 3 4 5 6 7 8]
```

### Copy — Menyalin Slice

```go
source := []int{1, 2, 3}
target := make([]int, len(source))

copied := copy(target, source)
fmt.Println(copied)   // 3
fmt.Println(target)   // [1 2 3]
```

### Sub-slice

```go
s := []int{1, 2, 3, 4, 5, 6}

// Dari index 1 sampai sebelum 4
fmt.Println(s[1:4])   // [2, 3, 4]

// Dari awal sampai sebelum 3
fmt.Println(s[:3])    // [1, 2, 3]

// Dari index 3 sampai akhir
fmt.Println(s[3:])    // [4, 5, 6]
```

### Remove Elemen dari Slice

```go
s := []int{1, 2, 3, 4, 5}

// Hapus elemen di index 2 (nilai 3)
index := 2
s = append(s[:index], s[index+1:]...)
fmt.Println(s)   // [1, 2, 4, 5]
```

### Insert Elemen

```go
s := []int{1, 2, 3}

// Sisipkan 99 di index 1
s = append(s[:1], append([]int{99}, s[1:]...)...)
fmt.Println(s)   // [1, 99, 2, 3]
```

---

## Map

Map adalah koleksi **key-value pairs**, mirip dictionary.

```
Map:
┌─────────────┬──────────┐
│    Key      │  Value   │
├─────────────┼──────────┤
│   "nama"    │  "Budi"  │
│   "umur"    │    25    │
│   "kota"    │  "Bandung"│
└─────────────┴──────────┘
```

### Membuat Map

```go
// Dengan make
umur := make(map[string]int)
umur["Budi"] = 25
umur["Siti"] = 22

// Langsung dengan literal
data := map[string]int{
    "Budi": 25,
    "Siti": 22,
}

// Map dengan string
alamat := map[string]string{
    "Budi": "Jakarta",
    "Siti": "Bandung",
}
```

### Akses, Tambah, Ubah

```go
skor := map[string]int{
    "Matematika": 90,
    "Bahasa": 85,
}

fmt.Println(skor["Matematika"])  // 90

skor["IPA"] = 88                 // Tambah
skor["Matematika"] = 95          // Ubah
```

### Cek Key Ada atau Tidak

```go
skor := map[string]int{"A": 90}

nilai, ada := skor["B"]          // ada = false
if ada {
    fmt.Println("Nilai:", nilai)
} else {
    fmt.Println("Key tidak ada")
}

// Atau dengan commas ok idiom
if _, ada := skor["A"]; ada {
    fmt.Println("A ada")
}
```

### Delete Key

```go
skor := map[string]int{"A": 90, "B": 85, "C": 70}

delete(skor, "B")               // Hapus key "B"
fmt.Println(skor)               // map[A:90 C:70]
```

### Iterasi Map

```go
skor := map[string]int{
    "Matematika": 90,
    "Bahasa": 85,
    "IPA": 88,
}

for key, value := range skor {
    fmt.Printf("%s: %d\n", key, value)
}
```

> ⚠️ **Catatan:** Urutan iterasi Map di Go **acak** (random). Gunakan `sort` jika butuh urutan.

### Urutkan Map

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    skor := map[string]int{
        "Matematika": 90,
        "Bahasa": 85,
        "IPA": 88,
    }

    // Ambil keys
    keys := make([]string, 0, len(skor))
    for key := range skor {
        keys = append(keys, key)
    }
    
    // Sort keys
    sort.Strings(keys)
    
    // Print sesuai urutan
    for _, key := range keys {
        fmt.Printf("%s: %d\n", key, skor[key])
    }
}
```

---

## Contoh Lengkap

```go
package main

import "fmt"

func main() {
    // === SLICE ===
    fmt.Println("=== Daftar Belanja ===")
    belanja := []string{"Telur", "Beras", "Minyak", "Gula"}
    
    // Tambah item
    belanja = append(belanja, "Kopi", "Teh")
    
    // Print dengan nomor
    for i, item := range belanja {
        fmt.Printf("%d. %s\n", i+1, item)
    }

    // === MAP ===
    fmt.Println("\n=== Data Siswa ===")
    siswa := map[string]map[string]int{
        "Budi":   {"nilai": 85, "absen": 20},
        "Siti":   {"nilai": 92, "absen": 22},
        "Rudi":   {"nilai": 78, "absen": 18},
    }

    for nama, data := range siswa {
        fmt.Printf("%s: Nilai=%d, Absen=%d\n", nama, data["nilai"], data["absen"])
    }

    // Hapus siswa
    delete(siswa, "Rudi")
    fmt.Println("\nSetelah hapus Rudi:", siswa)
}
```

**Output:**
```
=== Daftar Belanja ===
1. Telur
2. Beras
3. Minyak
4. Gula
5. Kopi
6. Teh

=== Data Siswa ===
Budi: Nilai=85, Absen=20
Siti: Nilai=92, Absen=22
Rudi: Nilai=78, Absen=18

Setelah hapus Rudi: map[Budi:map[nilai:85 absen:20] Siti:map[nilai:92 absen:22]]
```

---

## Perbandingan Array vs Slice vs Map

| Aspek | Array | Slice | Map |
|-------|-------|-------|-----|
| Ukuran | Tetap | Dinamis | Dinamis |
| Index | Angka | Angka | Key (string/int) |
| Zero Value | `[0, 0, 0]` | `nil` | `nil` |
| `len()` | ✅ | ✅ | ✅ |
| `cap()` | ✅ | ✅ | ❌ |
| `append()` | ❌ | ✅ | ❌ |
| `make()` | ❌ | ✅ | ✅ |

---

## Cheat Sheet

| Operasi | Array | Slice | Map |
|---------|-------|-------|-----|
| Buat | `[n]int{1,2}` | `[]int{1,2}` | `map[string]int{}` |
| Buat dengan make | - | `make([]int, 5)` | `make(map[string]int)` |
| Panjang | `len()` | `len()` | `len()` |
| Tambah elemen | - | `append()` | `m[key]=value` |
| Hapus | - | `append()` | `delete()` |
| Cek key | - | - | `val, ok := m[key]` |

---

## Manual Test

Semua konsep dalam materi ini sudah ditest di folder praktek:

📁 **Lokasi:** `/minilab/go/05-array-slice-map/`

### Cara Menjalankan Test

**1. Jalankan langsung (tanpa compile):**
```bash
cd minilab/go/05-array-slice-map
go run main.go
```

**2. Build & jalankan executable:**
```bash
cd minilab/go/05-array-slice-map
go build -o 05-test main.go
./05-test
```

### Apa yang Ditest?

File `main.go` berisi test untuk:

- Array (akses, len)
- Slice (make, append, len/cap, sub-slice, remove, insert, copy)
- Map (create, add, get, delete, iterate, sorted iteration)
- Mencari nilai max/min dari slice

---

## Latihan

1. Buat slice 5 nama, tambahkan 2 nama lagi, lalu hapus nama ke-3
2. Buat map untuk menyimpan 3 warna favorit (key: angka 1-3, value: nama warna)
3. Buat program kamus sederhana: kata Indonesia → Inggris
4. Challenge: Dari slice angka, cari nilai terbesar dan terkecil

---

## ➡️ Selanjutnya

**Function (Fungsi)**  
→ Lanjut ke: `06-function.md`
