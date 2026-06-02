---
topik: Pengenalan & Instalasi
urutan: 1 dari 3
posisi: awal
---

> 🚀 **Ini adalah materi pertama** dalam seri "Belajar Golang Dasar".  
> Belum ada materi sebelumnya — mulailah dari sini!

# Pengenalan & Instalasi Go (Golang)

## Tujuan Belajar

- Memahami apa itu Go dan mengapa populer
- Menginstal Go di sistem operasi berbeda
- Menulis dan menjalankan program Go pertama

---

## Apa Itu Go (Golang)?

Go adalah bahasa pemrograman yang dikembangkan oleh **Google** pada tahun 2009. Dirancang oleh Robert Griesemer, Rob Pike, dan Ken Thompson — nama-nama besar di dunia pemrograman.

### Kenapa Go Populer?

| Kelebihan | Penjelasan |
|-----------|------------|
| **Cepat** | Kompilasi dan eksekusi sangat cepat |
| **Mudah dipelajari** | Sintaks sederhana, mirip bahasa C |
| **Goroutine** | Konkurensi native yang ringan |
| **Garbage Collection** | Manajemen memori otomatis |
| **Standar Library** | Library bawaan yang lengkap |
| **Cross-platform** | Bisa jalan di Windows, macOS, Linux |

### Siapa yang Pakai Go?

- **Google** — layanan internal
- **Docker** — containerization
- **Kubernetes** — container orchestration  
- **Dropbox** — cloud storage
- **Twitch** — streaming platform
- **Tokopedia / Gojek** — tech company Indonesia

---

## Instalasi Go

### Windows

1. Download installer dari https://go.dev/dl/
2. Jalankan file `.msi` dan ikut wizard instalasi
3. Buka Command Prompt / PowerShell
4. Verifikasi instalasi:

```powershell
go version
```

### macOS

**Menggunakan Homebrew:**

```bash
brew install go
```

**Manual:**
1. Download file `.pkg` dari https://go.dev/dl/
2. Jalankan installer
3. Verifikasi:

```bash
go version
```

### Linux

**Menggunakan package manager:**

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install golang-go

# Fedora / RHEL
sudo dnf install golang
```

**Verifikasi:**

```bash
go version
```

---

## Struktur Workspace Go

Go punya struktur direktori约定:

```
$HOME/
└── go/
    ├── bin/          ← executable hasil kompilasi
    ├── pkg/          ← package object
    └── src/          ← kode sumber
        └── github.com/
            └── username/
                └── project/
```

---

## Program Go Pertama

Buat file baru bernama `hello.go`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Halo dari Go!")
}
```

### Jalankan Program

**Opsi 1: Langsung jalan (interpret)**
```bash
go run hello.go
```

**Opsi 2: Compile dulu**
```bash
go build hello.go
./hello
```

### Penjelasan Kode

| Bagian | Arti |
|--------|------|
| `package main` | Setiap program Go harus punya package main |
| `import "fmt"` | Impor library format (untuk print) |
| `func main()` | Fungsi utama — titik awal program |
| `fmt.Println()` | Print baris baru ke terminal |

---

## Cheat Sheet

### Perintah Dasar Go

| Perintah | Fungsi | Contoh |
|----------|--------|--------|
| `go version` | Cek versi Go yang terinstal | `go version` → `go1.22 linux/amd64` |
| `go run file.go` | Jalankan program **tanpa** compile dulu — cocok untuk belajar & testing | `go run hello.go` |
| `go build file.go` | Compile kode jadi file executable (bisa dijalankan tanpa Go) | `go build hello.go` → hasilnya file `./hello` |
| `go fmt file.go` | Rapikan format kode secara otomatis mengikuti standar Go | `go fmt hello.go` |
| `go get package` | Download & install package dari internet ke proyek kamu | `go get github.com/gin-gonic/gin` |
| `go doc package` | Tampilkan dokumentasi resmi suatu package di terminal | `go doc fmt` |

### Kapan Pakai `go run` vs `go build`?

```
go run hello.go   → langsung jalan, tidak ada file output — untuk coba-coba
go build hello.go → hasilkan file ./hello — untuk distribusi atau deployment
./hello           → jalankan file hasil build
```

> 💡 **Tips:** Saat belajar, pakai `go run`. Saat sudah siap deploy, pakai `go build`.

---

## Latihan

1. ✅ Instal Go di komputermu
2. ✅ Jalankan `go version` untuk verifikasi
3. ✅ Buat program `hello.go` dan jalankan
4. ✅ Ubah teksnya jadi: "Perkenalkan, nama saya [nama kamu]!"

---

## ➡️ Selanjutnya

**Variabel & Tipe Data**  
→ Lanjut ke: `02-variabel-tipe-data.md`
