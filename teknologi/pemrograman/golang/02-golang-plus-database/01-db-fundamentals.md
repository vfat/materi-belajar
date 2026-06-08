---
topik: Fondasi Database di Go
urutan: 1 dari 20
posisi: awal
tag:
  - database
  - sql
  - connection-pooling
  - context
  - database/sql
prerequisites:
  - Golang Dasar (variabel, function, struct, error handling)
level: pemula-menengah
---

> 🚀 **Materi #01** — Fondasi Database di Go. Topik ini wajib dipahami sebelum bekerja dengan database manapun.

# Fondasi Database di Go

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami arsitektur `database/sql` dan peran driver database
- Mampu melakukan koneksi ke database dengan pattern yang benar
- Menggunakan `context.Context` untuk timeout, cancellation, dan tracing
- Mengerti connection pooling dan cara mengoptimalkannya
- Mengimplementasikan graceful shutdown dengan clean resource handling

---

## 1. Arsitektur `database/sql`

### 1.1 Kenapa Pakai `database/sql`?

Go menyediakan package [`database/sql`](https://pkg.go.dev/database/sql) sebagai **abstraksi umum** untuk akses database relasional. Benefit-nya:

| Aspek | Penjelasan |
|-------|------------|
| **Generic API** | Satu API untuk SQLite, MySQL, PostgreSQL, dll |
| **Connection Pooling** | Built-in, tinggal konfigurasi |
| **Resource Management** | `Close()`, `Prepare()`, transaction handling |
| **Safety** | Query parameterization otomatis |

### 1.2 Peran Driver

Driver adalah implementasi spesifik untuk tiap database. Lo **tidak pernah** mengimport driver secara langsung di business logic—lo cukup import `database/sql` + driver, lalu driver **register dirinya** ke `database/sql`.

```
┌─────────────────────────────────────────┐
│           Aplikasi Go Lo               │
│                                         │
│   import "database/sql"                │
│   db.Query("SELECT ...")                │
│                                         │
│         ▲                              │
│         │ pakai                         │
│         │                              │
│   ┌─────┴──────┐                       │
│   │database/sql│                       │
│   └─────┬──────┘                       │
│         │ panggilan generic             │
│         ▼                              │
│   ┌─────┴──────┐                       │
│   │  Driver    │ ← mattn/go-sqlite3    │
│   │  (impl)    │   jackc/pgx           │
│   └───────────┘   go-sql-driver/mysql  │
└─────────────────────────────────────────┘
```

### 1.3 Driver yang Umum Dipakai

| Database | Driver | Import Path |
|----------|--------|-------------|
| SQLite | `mattn/go-sqlite3` | `github.com/mattn/go-sqlite3` |
| PostgreSQL | `jackc/pgx/v5` | `github.com/jackc/pgx/v5` |
| MySQL | `go-sql-driver/mysql` | `github.com/go-sql-driver/mysql` |

> 💡 **Tips:** Untuk PostgreSQL, `pgx` sering preferred karena lebih performant dan fitur lengkap. Untuk simplicity, SQLite + `go-sqlite3` enak buat development lokal.

---

## 2. Koneksi Database

### 2.1 DSN (Data Source Name)

DSN adalah string yang mengandung informasi koneksi ke database. Format-nya beda-beda tiap database:

```go
// SQLite
"file:app.db?_foreign_keys=1&_busy_timeout=5000"

// PostgreSQL (pgx)
"postgres://user:password@localhost:5432/mydb?sslmode=disable"

// MySQL
"user:password@tcp(localhost:3306)/mydb?parseTime=true"
```

> ⚠️ **Jangan hardcode DSN langsung di kode!** Selalu pakai environment variables.

### 2.2 Pattern Koneksi yang Benar

Berikut pattern koneksi yang **production-ready**:

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config menyimpan konfigurasi database
type Config struct {
	Driver   string
	DSN      string
	MaxOpen  int  // max koneksi simultan
	MaxIdle  int  // max koneksi idle di pool
	MaxLifetime time.Duration
}

// New creates a new database connection
func New(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Konfigurasi connection pool
	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Verifikasi koneksi
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
```

### 2.3 Environment-based Configuration

```go
// NewFromEnv creates database connection from environment variables
func NewFromEnv() (*sql.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite3"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "file:app.db"
	}

	return New(Config{
		Driver:      driver,
		DSN:         dsn,
		MaxOpen:     25,
		MaxIdle:     5,
		MaxLifetime: 5 * time.Minute,
	})
}
```

---

## 3. Context dalam Database Operations

### 3.1 Kenapa Context Penting?

[`context.Context`](https://pkg.go.dev/context) adalah mechanism Go untuk:

- **Timeout** — batasi waktu eksekusi query
- **Cancellation** — batalkan operasi yang tidak perlu lagi
- **Tracing** — propagate request ID ke database layer

> 🔑 **Rule:** **Selalu** gunakan context di setiap `Query()`, `Exec()`, `Prepare()`, dll. Jangan pernah pakai `context.Background()` untuk long-running operations—gunakan timeout context yang sesuai.

### 3.2 Contoh Penggunaan Context

```go
// ❌ BAD: Tanpa context, tidak bisa dibatalkan
func GetUserBad(db *sql.DB, id int64) (*User, error) {
	row := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id)
	// tidak ada timeout, tidak bisa dibatalkan
}

// ✅ GOOD: Dengan context dan timeout
func GetUser(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	query := "SELECT id, name, email FROM users WHERE id = ?"
	
	row := db.QueryRowContext(ctx, query, id)
	
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("scan user %d: %w", id, err)
	}
	
	return &user, nil
}

// Usage dengan timeout
func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	user, err := GetUser(ctx, db, userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
```

### 3.3 Context Propagation dalam Request HTTP

```go
func handleUsers(w http.ResponseWriter, r *http.Request) {
	// Context sudah ada dari request, tinggal tambahkan timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users, err := ListUsers(ctx, db)
	if err != nil {
		log.Printf("list users: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}
```

---

## 4. Connection Pooling

### 4.1 Apa Itu Connection Pool?

Connection pool adalah **kumpulan koneksi database yang sudah siap pakai**. Tanpa pool, setiap query akan create-close koneksi → lambat & resource-intensive.

```
┌────────────┐     ┌──────────────────────────────────┐
│  App Go    │     │       Connection Pool            │
│            │     │  ┌────┐ ┌────┐ ┌────┐ ┌────┐    │
│ Query 1 ───┼────▶│  │ C1 │ │ C2 │ │ C3 │ │ C4 │    │
│            │     │  └────┘ └────┘ └────┘ └────┘    │
│ Query 2 ───┼────▶│        Available / In Use        │
│            │     └──────────────────────────────────┘
└────────────┘                   │
                                 ▼
                    ┌─────────────────────┐
                    │   Database Server   │
                    └─────────────────────┘
```

### 4.2 Konfigurasi Pool

| Parameter | Default | Penjelasan |
|-----------|---------|------------|
| `SetMaxOpenConns` | 0 (unlimited) | Max koneksi simultan ke database |
| `SetMaxIdleConns` | 2 | Max koneksi idle yang dipertahankan |
| `SetConnMaxLifetime` | 0 (forever) | Durasi maksimal satu koneksi |

```go
db, err := sql.Open("sqlite3", "file:app.db")
if err != nil {
	log.Fatal(err)
}

// Atur pool sebelum pakai
db.SetMaxOpenConns(25)              // max 25 koneksi simultan
db.SetMaxIdleConns(5)               // pertahankan 5 idle
db.SetConnMaxLifetime(5 * time.Minute) // recycling tiap 5 menit
```

### 4.3 Pooling Best Practices

```go
// Pool configuration yang reasonable untuk большинство aplikasi
const (
	MaxOpenConns    = 25
	MaxIdleConns    = 10
	ConnMaxLifetime = 5 * time.Minute
	ConnMaxIdleTime = 1 * time.Minute
)

// Config untuk production
db.SetMaxOpenConns(MaxOpenConns)
db.SetMaxIdleConns(MaxIdleConns)
db.SetConnMaxLifetime(ConnMaxLifetime)
// Note: ConnMaxIdleTime hanya ada di Go 1.15+
```

---

## 5. Graceful Shutdown & Resource Cleanup

### 5.1 Kenapa Penting?

Koneksi database adalah **resource sistem** (file descriptor, network socket). Kalau tidak di-cleanup dengan benar:

- File descriptor leak → eventually "too many open files"
- Koneksi database menggantung → resource exhaustion di server DB
- Data inconsistency kalau transaksi tidak di-commit/rollback

### 5.2 Pattern Graceful Shutdown

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup database
	db, err := sql.Open("sqlite3", "file:app.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Setup server
	srv := &http.Server{Addr: ":8080"}

	// Channel untuk signal
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Wait untuk signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
	case sig := <-sigCh:
		log.Printf("received %v, shutting down...", sig)

		// 1. Stop menerima request baru
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 2. Shutdown HTTP server (menunggu in-flight requests selesai)
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown: %v", err)
		}

		// 3. Tutup database connection
		//    Ping untuk memastikan tidak ada koneksi yang stuck
		if err := db.PingContext(ctx); err == nil {
			log.Println("db connections healthy, closing...")
		}

		// 4. Close database (menunggu semua koneksi idle selesai)
		if err := db.Close(); err != nil {
			log.Printf("db close: %v", err)
		}

		log.Println("graceful shutdown complete")
	}
}
```

### 5.3 Defer untuk Clean Resource

```go
func processData(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Pastikan rollback kalau ada error sebelum commit
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// ... do work ...

	// Commit kalau sukses
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	tx = nil // prevent rollback di defer
	return nil
}
```

---

## 6. Error Handling di Database

### 6.1 Pattern Error Handling

```go
import (
	"database/sql"
	"errors"
	"fmt"
)

// Define sentinel errors
var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicate    = errors.New("duplicate entry")
	ErrConstraint   = errors.New("constraint violation")
)

func GetUser(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	user, err := getUserByID(ctx, db, id)
	if err != nil {
		// Translate sql.ErrNoRows ke custom error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return user, nil
}

func getUserByID(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	query := `SELECT id, name, email FROM users WHERE id = ?`
	
	row := db.QueryRowContext(ctx, query, id)
	
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// Usage
func handler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUser(r.Context(), db, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(user)
}
```

### 6.2 Transaction Error Handling

```go
func Transfer(ctx context.Context, db *sql.DB, from, to int64, amount float64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	
	// Always rollback on panic/error
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Debit
	_, err = tx.ExecContext(ctx, 
		"UPDATE accounts SET balance = balance - ? WHERE id = ? AND balance >= ?",
		amount, from, amount)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("debit: %w", err)
	}

	// Credit
	_, err = tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + ? WHERE id = ?",
		amount, to)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("credit: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
```

---

## 7. Mini Lab: Full Working Example

Buat project sederhana dengan pattern connection + basic CRUD:

### 7.1 Setup Project

```bash
mkdir -p ~/projects/db-fundamentals && cd ~/projects/db-fundamentals
go mod init db-fundamentals
go get github.com/mattn/go-sqlite3
```

### 7.2 Project Structure

```
db-fundamentals/
├── main.go           # entry point + HTTP handlers
├── db/
│   ├── postgres.go   # postgres connection
│   ├── sqlite.go     # sqlite connection
│   └── user.go       # user CRUD operations
├── .env.example      # template environment
└── go.mod
```

### 7.3 Implementasi

**`db/postgres.go`** (jika pakai PostgreSQL):

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}
```

**`db/user.go`**:

```go
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (name, email) VALUES (?, ?) RETURNING id`
	
	err := s.db.QueryRowContext(ctx, query, user.Name, user.Email).
		Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	
	return nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, name, email FROM users WHERE id = ?`
	
	var user User
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	
	return &user, nil
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `SELECT id, name, email FROM users ORDER BY id LIMIT ? OFFSET ?`
	
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

func (s *UserService) Update(ctx context.Context, user *User) error {
	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	
	result, err := s.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
```

### 7.4 Main Entry Point

```go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"db-fundamentals/db"
)

func main() {
	// Setup database (SQLite for simplicity)
	database, err := sql.Open("sqlite3", "file:app.db?_foreign_keys=1")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	// Configure connection pool
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	// Initialize migrations
	if err := initSchema(context.Background(), database); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Setup user service
	userSvc := db.NewUserService(database)

	// HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", handleListUsers(userSvc))
	mux.HandleFunc("GET /users/{id}", handleGetUser(userSvc))
	mux.HandleFunc("POST /users", handleCreateUser(userSvc))
	mux.HandleFunc("PUT /users/{id}", handleUpdateUser(userSvc))
	mux.HandleFunc("DELETE /users/{id}", handleDeleteUser(userSvc))

	// Graceful shutdown server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Wait for interrupt signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	if err := database.Close(); err != nil {
		log.Printf("db close: %v", err)
	}

	log.Println("done")
}

func initSchema(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	);`
	_, err := db.ExecContext(ctx, schema)
	return err
}

// Handler functions
func handleListUsers(svc *db.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		users, err := svc.List(ctx, 100, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func handleGetUser(svc *db.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		id := r.PathValue("id")
		var userID int64
		if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		user, err := svc.GetByID(ctx, userID)
		if err != nil {
			if err == db.ErrUserNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleCreateUser(svc *db.UserService) http.HandlerFunc {
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		user := &db.User{Name: req.Name, Email: req.Email}
		if err := svc.Create(ctx, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

func handleUpdateUser(svc *db.UserService) http.HandlerFunc {
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		id := r.PathValue("id")
		var userID int64
		if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		user := &db.User{ID: userID, Name: req.Name, Email: req.Email}
		if err := svc.Update(ctx, user); err != nil {
			if err == db.ErrUserNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleDeleteUser(svc *db.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		id := r.PathValue("id")
		var userID int64
		if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := svc.Delete(ctx, userID); err != nil {
			if err == db.ErrUserNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
```

### 7.5 Test

```bash
# Run server
go run main.go

# Di terminal lain, test:
# Create user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# List users
curl http://localhost:8080/users

# Get user
curl http://localhost:8080/users/1

# Update user
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/users/1
```

---

## 8. Ringkasan

```
┌─────────────────────────────────────────────────────────┐
│              FONDASI DATABASE DI GO                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ✅ Pakai database/sql + driver terpisah                │
│  ✅ DSN dari environment variables (JANGAN hardcode)    │
│  ✅ SELALU pakai context di Query/Exec                  │
│  ✅ Konfigurasi connection pool yang tepat              │
│  ✅ Graceful shutdown: srv.Shutdown() → db.Close()     │
│  ✅ Error handling: sql.ErrNoRows → custom error        │
│  ✅ Transaction: BeginTx → defer Rollback → Commit      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Referensi

- [Go database/sql package](https://pkg.go.dev/database/sql)
- [Go database/sql tutorial](https://go.dev/doc/database/index)
- [Jackc/pgx PostgreSQL driver](https://github.com/jackc/pgx)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

---

## ➡️ Selanjutnya

**Golang + SQLite: Setup & CRUD**  
→ `02-sqlite-setup.md`

Mulai dari materi #02, kita akan praktik langsung dengan SQLite sebagai database pertama. SQLite cocok buat development lokal dan learning karena tidak perlu setup server terpisah.
