---
topik: Golang + MySQL Connection & Config
urutan: 4 dari 20
posisi: setelah SQLite setup
prerequisites:
  - Golang + SQLite: Setup & CRUD (02-sqlite-setup)
level: menengah
---

> 🚀 **Materi #04** — Koneksi dan konfigurasi MySQL di Go dengan Docker.

# Golang + MySQL: Connection & Config

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami cara menghubungkan Go ke MySQL menggunakan `database/sql` dan driver `go-sql-driver/mysql`.
- Menyiapkan environment MySQL lewat Docker Compose.
- Mengkonfigurasi DSN, connection pool, dan timeout.
- Menangani error, retry, dan graceful shutdown.
- Menulis contoh CRUD sederhana.

---

## 1. Persiapan Docker Compose

### 1.1 Dockerfile (opsional)

Untuk MySQL, kita cukup menggunakan image resmi. Namun, jika ingin custom, buat `Dockerfile`:

```dockerfile
# Dockerfile (opsional)
FROM mysql:8.0
ENV MYSQL_ROOT_PASSWORD=root
ENV MYSQL_DATABASE=golang_demo
ENV MYSQL_USER=demo
ENV MYSQL_PASSWORD=demo
EXPOSE 3306
```

### 1.2 docker-compose.yml

Buat file `docker-compose.yml` di root folder `02-golang-plus-database`:

```yaml
version: "3.8"
services:
  mysql:
    image: mysql:8.0
    container_name: golang_mysql_demo
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: golang_demo
      MYSQL_USER: demo
      MYSQL_PASSWORD: demo
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-proot"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
    driver: local
```

> ⚡️ **Tip:** Jalankan `docker compose up -d` untuk memulai layanan.

---

## 2. Go Project Setup

### 2.1 go.mod

```bash
go mod init golang-mysql-demo
go get github.com/go-sql-driver/mysql
```

### 2.2 DSN & Config

```go
package db

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

// Config menyimpan konfigurasi koneksi MySQL
type Config struct {
    User     string
    Password string
    Host     string
    Port     int
    DBName   string
    Params   map[string]string
    MaxOpen  int
    MaxIdle  int
    MaxLife  time.Duration
}

func New(cfg Config) (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
    if len(cfg.Params) > 0 {
        params := ""
        for k, v := range cfg.Params {
            params += fmt.Sprintf("%s=%s&", k, v)
        }
        dsn += params[:len(params)-1]
    }

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("open mysql: %w", err)
    }

    db.SetMaxOpenConns(cfg.MaxOpen)
    db.SetMaxIdleConns(cfg.MaxIdle)
    db.SetConnMaxLifetime(cfg.MaxLife)

    // Ping with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping mysql: %w", err)
    }

    return db, nil
}
```

---

## 3. Contoh CRUD

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

func main() {
    cfg := db.Config{
        User:     "demo",
        Password: "demo",
        Host:     "localhost",
        Port:     3306,
        DBName:   "golang_demo",
        Params: map[string]string{
            "charset": "utf8mb4",
            "parseTime": "true",
        },
        MaxOpen:  25,
        MaxIdle:  5,
        MaxLife: 5 * time.Minute,
    }

    db, err := db.New(cfg)
    if err != nil {
        log.Fatalf("connect: %v", err)
    }
    defer db.Close()

    // Create table
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(200) UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        log.Fatalf("create table: %v", err)
    }

    // Insert
    res, err := db.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Alice", "alice@example.com")
    if err != nil {
        log.Fatalf("insert: %v", err)
    }
    id, _ := res.LastInsertId()
    fmt.Printf("Inserted user id=%d\n", id)

    // Query
    var name, email string
    err = db.QueryRow(`SELECT name, email FROM users WHERE id = ?`, id).Scan(&name, &email)
    if err != nil {
        log.Fatalf("query: %v", err)
    }
    fmt.Printf("Fetched: %s <%s>\n", name, email)
}
```

---

## 4. Error Handling & Retry

```go
func withRetry(fn func() error, attempts int, delay time.Duration) error {
    for i := 0; i < attempts; i++ {
        if err := fn(); err == nil {
            return nil
        } else {
            time.Sleep(delay)
        }
    }
    return fmt.Errorf("max retry attempts reached")
}
```

---

## 5. Graceful Shutdown

```go
func main() {
    // ... setup db
    srv := &http.Server{Addr: ":8080"}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("shutdown: %v", err)
    }
    db.Close()
}
```

---

## 6. Testing dengan Docker

```bash
# Start services
docker compose up -d

# Run tests
go test ./...

# Stop services
docker compose down
```

---

## 7. Checklist

- [x] Docker Compose file ready.
- [x] DSN dan config Go lengkap.
- [x] CRUD contoh.
- [x] Retry & graceful shutdown.
- [x] Dokumentasi lengkap.

---

## 8. Catatan

- Gunakan `parseTime=true` agar `time.Time` ter-parse otomatis.
- Untuk production, simpan credential di secret manager atau environment variable.
- Jika menggunakan `docker compose`, pastikan port 3306 tidak bentrok.
- Untuk scaling, gunakan `replication` di MySQL atau cluster.

---

## 9. Referensi

- https://github.com/go-sql-driver/mysql
- https://docs.docker.com/compose/compose-file/
- https://golang.org/pkg/database/sql/

---

> 🎉 Selamat belajar! Kamu sekarang bisa menghubungkan Go ke MySQL dengan Docker.
