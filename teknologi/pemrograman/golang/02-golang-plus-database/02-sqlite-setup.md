---
topik: Golang + SQLite Setup & CRUD
urutan: 2 dari 20
posisi: setelah fondasi database
tag:
  - sqlite
  - crud
  - database/sql
  - go-sqlite3
prerequisites:
  - Fondasi Database di Go (01-db-fundamentals)
level: pemula-menengah
---

> 🚀 **Materi #02** — Setup SQLite dengan Go dan operasi CRUD lengkap.

# Golang + SQLite: Setup & CRUD

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami karakteristik SQLite dan kapan tepat用它
- Setup koneksi SQLite dengan driver `mattn/go-sqlite3`
- Mengimplementasikan Create, Read, Update, Delete (CRUD) operations
- Menggunakan transactions untuk operasi yang butuh atomicity
- Membuat repository pattern untuk akses data

---

## 1. SQLite: Karakteristik & Kapan Pakai

### 1.1 Apa Itu SQLite?

SQLite adalah **serverless, self-contained, zero-configuration** database engine. Karakteristik utamanya:

| Aspek | Penjelasan |
|-------|------------|
| **Serverless** | Tidak perlu proses server terpisah, database adalah satu file |
| **Zero-config** | Tidak ada instalasi atau konfigurasi |
| **ACID** | Fully transactional, even after crashes |
| **File-based** | Seluruh database ada di satu file (`*.db` atau `*.sqlite`) |
| **Cross-platform** | Satu file bisa di-copy ke sistem operasi lain |

### 1.2 Kapan Pakai SQLite?

```
✅ GOOD untuk SQLite:
├── Development lokal (no setup)
├── Prototyping & MVP
├── Aplikasi desktop (Electron, mobile apps)
├── Testing (easy reset dengan delete file)
├── Small-to-medium traffic websites (< 100K hits/day)
└── Cache layer atau local storage

❌ BAD untuk SQLite:
├── High-concurrency (multiple writers simultan)
├── Very large datasets (> 1TB)
├── Multi-user network access
├── Complex client-server architecture
└── Horizontal scaling requirement
```

### 1.3 SQLite di Ekosistem Go

```
┌────────────────────────────────────────────┐
│          DATABASE DRIVER                   │
├────────────────────────────────────────────┤
│  mattn/go-sqlite3 (paling populer)         │
│  → Wrapper around SQLite C library          │
│  → Requires CGO (gcc/clang)                │
│                                            │
│  modernc.org/sqlite (pure Go alternative)  │
│  → Pure Go, no CGO needed                  │
│  → Slightly slower tapi easier deploy     │
└────────────────────────────────────────────┘
```

---

## 2. Setup Project

### 2.1 Inisialisasi Project

```bash
mkdir -p ~/projects/golang-sqlite && cd ~/projects/golang-sqlite
go mod init golang-sqlite
```

### 2.2 Install Driver

```bash
# mattn/go-sqlite3 (paling umum, perlu gcc)
go get github.com/mattn/go-sqlite3

# ATAU modernc.org/sqlite (pure Go, no CGO)
go get modernc.org/sqlite
```

> ⚠️ **Catatan CGO:** `mattn/go-sqlite3` membutuhkan CGO karena wrapping library C SQLite. Di Windows, butuh MinGW-w64. Di macOS/Linux, perlu gcc atau clang.

### 2.3 Verifikasi Install

```go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("SQLite connection OK!")
}
```

```bash
go run main.go
# Output: SQLite connection OK!
```

---

## 3. Koneksi SQLite

### 3.1 DSN SQLite Format

Format DSN SQLite: `file:path/to/database.db?options`

| Parameter | Default | Deskripsi |
|-----------|---------|------------|
| `file:app.db` | - | Path ke file database |
| `:memory:` | - | In-memory database (volatile) |
| `?mode=ro` | rw | Mode: ro (read-only), rw (read-write), rwc (read-write-create) |
| `?_foreign_keys=1` | 0 | Enable foreign key enforcement |
| `?_busy_timeout=5000` | 0 | Wait timeout dalam ms |
| `?_journal_mode=WAL` | DELETE | Journal mode: DELETE, WAL, MEMORY |

### 3.2 Connection Pattern

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

// Config menyimpan konfigurasi SQLite
type Config struct {
	DSN           string
	ForeignKeys   bool
	BusyTimeout   time.Duration
	JournalMode   string
	MaxOpenConns  int
	MaxIdleConns  int
	ConnMaxLifetime time.Duration
}

// NewSQLiteDB creates new SQLite connection
func NewSQLiteDB(cfg Config) (*sql.DB, error) {
	dsn := cfg.DSN

	// Append query parameters
	if cfg.ForeignKeys {
		dsn += "?_foreign_keys=1"
	}
	if cfg.BusyTimeout > 0 {
		separator := "?"
		if cfg.ForeignKeys {
			separator = "&"
		}
		dsn += fmt.Sprintf("%s_busy_timeout=%d", separator, int(cfg.BusyTimeout.Seconds()*1000))
	}
	if cfg.JournalMode != "" {
		separator := "?"
		if cfg.ForeignKeys || cfg.BusyTimeout > 0 {
			separator = "&"
		}
		dsn += fmt.Sprintf("%s_journal_mode=%s", separator, cfg.JournalMode)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Konfigurasi pool (untuk SQLite biasanya 1 koneksi cukup)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verifikasi koneksi
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

// NewSQLiteFromEnv creates SQLite connection from environment
func NewSQLiteFromEnv() (*sql.DB, error) {
	dsn := os.Getenv("SQLITE_DSN")
	if dsn == "" {
		dsn = "file:app.db"
	}

	return NewSQLiteDB(Config{
		DSN:           dsn,
		ForeignKeys:   true,
		BusyTimeout:   5 * time.Second,
		JournalMode:   "WAL",
		MaxOpenConns:  1, // SQLite cuma bisa 1 writer
		MaxIdleConns:  1,
		ConnMaxLifetime: time.Hour,
	})
}
```

---

## 4. Schema & Table

### 4.1 Mendesain Schema Sederhana

Untuk materi ini, kita akan buat sistem **task management** sederhana:

```sql
-- Schema untuk task management
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT DEFAULT '#3B82F6',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 2, -- 1=high, 2=medium, 3=low
    status TEXT DEFAULT 'pending', -- pending, in_progress, done
    category_id INTEGER,
    due_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE SET NULL
);
```

### 4.2 Inisialisasi Schema di Go

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    color TEXT DEFAULT '#3B82F6',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 2,
    status TEXT DEFAULT 'pending',
    category_id INTEGER,
    due_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
`

// InitSchema creates tables if not exist
func InitSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}
```

---

## 5. CRUD Operations

### 5.1 Model Definitions

```go
package model

import (
	"database/sql"
	"time"
)

// Category merepresentasikan kategori task
type Category struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"-"`
	Color       string         `json:"color"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Task merepresentasikan task
type Task struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description sql.NullString `json:"-"`
	Priority    int            `json:"priority"`
	Status      string         `json:"status"`
	CategoryID  sql.NullInt64  `json:"-"`
	DueDate     sql.NullTime   `json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TaskWithCategory adalah Task yang sudah di-join dengan Category
type TaskWithCategory struct {
	Task
	CategoryName sql.NullString `json:"-"`
	CategoryColor sql.NullString `json:"-"`
}
```

### 5.2 Category Repository

```go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang-sqlite/model"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	query := `INSERT INTO categories (name, description, color) VALUES (?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	cat.ID = id
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*model.Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories WHERE id = ?`

	var cat model.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category: %w", err)
	}

	return &cat, nil
}

func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*model.Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories WHERE name = ?`

	var cat model.Category
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by name: %w", err)
	}

	return &cat, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]*model.Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var cat model.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Color, &cat.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat *model.Category) error {
	query := `UPDATE categories SET name = ?, description = ?, color = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color, cat.ID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
```

### 5.3 Task Repository

```go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang-sqlite/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (title, description, priority, status, category_id, due_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Priority, task.Status, task.CategoryID, task.DueDate,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	task.ID = id
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	query := `
		SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	var task model.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status,
		&task.CategoryID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &task, nil
}

func (r *TaskRepository) GetByIDWithCategory(ctx context.Context, id int64) (*model.TaskWithCategory, error) {
	query := `
		SELECT t.id, t.title, t.description, t.priority, t.status, t.category_id, t.due_date,
		       t.created_at, t.updated_at, c.name, c.color
		FROM tasks t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.id = ?
	`

	var task model.TaskWithCategory
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status,
		&task.CategoryID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
		&task.CategoryName, &task.CategoryColor,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task with category: %w", err)
	}

	return &task, nil
}

func (r *TaskRepository) List(ctx context.Context, limit, offset int) ([]*model.Task, error) {
	query := `
		SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
		FROM tasks ORDER BY created_at DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status,
			&task.CategoryID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) ListByStatus(ctx context.Context, status string) ([]*model.Task, error) {
	query := `
		SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
		FROM tasks WHERE status = ? ORDER BY priority ASC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("list tasks by status: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status,
			&task.CategoryID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) ListByCategory(ctx context.Context, categoryID int64) ([]*model.Task, error) {
	query := `
		SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
		FROM tasks WHERE category_id = ? ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list tasks by category: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Priority, &task.Status,
			&task.CategoryID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks
		SET title = ?, description = ?, priority = ?, status = ?, category_id = ?, due_date = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Priority, task.Status,
		task.CategoryID, task.DueDate, task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tasks`

	var count int64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}

	return count, nil
}

func (r *TaskRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE status = ?`

	var count int64
	if err := r.db.QueryRowContext(ctx, query, status).Scan(&count); err != nil {
		return 0, fmt.Errorf("count tasks by status: %w", err)
	}

	return count, nil
}
```

---

## 6. Transactions

### 6.1 Kapan Pakai Transaction?

```
GUNAKAN TRANSACTION ketika:
├── Multiple INSERT/UPDATE/DELETE yang harus atomic
├── Need rollback on partial failure
└── Performance: bulk operations lebih cepat

GUNAKAN SINGLE OPERATION ketika:
├── Single row operations
├── Read-only queries
└── Operations yang tidak perlu atomicity
```

### 6.2 Transaction Pattern di Go

```go
package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// TransferTask bergerakkan task dari satu kategori ke kategori lain
func (r *TaskRepository) TransferTask(ctx context.Context, taskID, fromCategoryID, toCategoryID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Ensure task exists in source category
	var count int64
	query := `SELECT COUNT(*) FROM tasks WHERE id = ? AND category_id = ?`
	if err := tx.QueryRowContext(ctx, query, taskID, fromCategoryID).Scan(&count); err != nil {
		tx.Rollback()
		return fmt.Errorf("check task: %w", err)
	}
	if count == 0 {
		tx.Rollback()
		return fmt.Errorf("task not found in source category")
	}

	// Update category
	updateQuery := `UPDATE tasks SET category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	if _, err := tx.ExecContext(ctx, updateQuery, toCategoryID, taskID); err != nil {
		tx.Rollback()
		return fmt.Errorf("update task category: %w", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// BulkCreateTasks membuat multiple tasks dalam satu transaction
func (r *TaskRepository) BulkCreateTasks(ctx context.Context, tasks []*model.Task) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	query := `INSERT INTO tasks (title, description, priority, status, category_id, due_date) VALUES (?, ?, ?, ?, ?, ?)`

	for _, task := range tasks {
		result, err := tx.ExecContext(ctx, query,
			task.Title, task.Description, task.Priority, task.Status, task.CategoryID, task.DueDate,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert task %s: %w", task.Title, err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("last insert id: %w", err)
		}
		task.ID = id
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// WithTransaction executes a function within a transaction
func WithTransaction(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
```

---

## 7. Full Working Example

### 7.1 Project Structure

```
golang-sqlite/
├── main.go
├── db/
│   └── sqlite.go
├── model/
│   └── models.go
├── repository/
│   ├── category.go
│   └── task.go
└── go.mod
```

### 7.2 Main Application

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
	"strconv"
	"syscall"
	"time"

	"golang-sqlite/db"
	"golang-sqlite/model"
	"golang-sqlite/repository"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Setup database
	database, err := db.NewSQLiteFromEnv()
	if err != nil {
		log.Fatalf("setup db: %v", err)
	}
	defer database.Close()

	// Initialize schema
	if err := db.InitSchema(context.Background(), database); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Seed initial data
	seedData(context.Background(), database)

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)

	// Setup routes
	mux := http.NewServeMux()

	// Task routes
	mux.HandleFunc("GET /tasks", withTimeout(listTasksHandler(taskRepo)))
	mux.HandleFunc("GET /tasks/{id}", withTimeout(getTaskHandler(taskRepo)))
	mux.HandleFunc("POST /tasks", withTimeout(createTaskHandler(taskRepo, categoryRepo)))
	mux.HandleFunc("PUT /tasks/{id}", withTimeout(updateTaskHandler(taskRepo)))
	mux.HandleFunc("DELETE /tasks/{id}", withTimeout(deleteTaskHandler(taskRepo)))
	mux.HandleFunc("PATCH /tasks/{id}/status", withTimeout(updateTaskStatusHandler(taskRepo)))

	// Category routes
	mux.HandleFunc("GET /categories", withTimeout(listCategoriesHandler(categoryRepo)))
	mux.HandleFunc("POST /categories", withTimeout(createCategoryHandler(categoryRepo)))

	// Stats route
	mux.HandleFunc("GET /stats", withTimeout(statsHandler(taskRepo)))

	// Start server
	srv := &http.Server{
		Addr:         getEnv("ADDR", ":8080"),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
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

// Helper functions
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func withTimeout(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		fn(w, r.WithContext(ctx))
	}
}

// Handlers
func listTasksHandler(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		status := r.URL.Query().Get("status")

		var tasks []*model.Task
		var err error

		if status != "" {
			tasks, err = repo.ListByStatus(r.Context(), status)
		} else {
			tasks, err = repo.List(r.Context(), limit, offset)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

func getTaskHandler(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

		task, err := repo.GetByIDWithCategory(r.Context(), id)
		if err != nil {
			if err == repository.ErrTaskNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func createTaskHandler(taskRepo *repository.TaskRepository, catRepo *repository.CategoryRepository) http.HandlerFunc {
	type Request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    int    `json:"priority"`
		CategoryID  *int64 `json:"category_id"`
		DueDate     string `json:"due_date"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}

		if req.Priority <= 0 {
			req.Priority = 2
		}

		task := &model.Task{
			Title:    req.Title,
			Status:   "pending",
			Priority: req.Priority,
		}

		if req.Description != "" {
			task.Description = sql.NullString{String: req.Description, Valid: true}
		}

		if req.CategoryID != nil {
			task.CategoryID = sql.NullInt64{Int64: *req.CategoryID, Valid: true}
		}

		if req.DueDate != "" {
			if t, err := time.Parse(time.RFC3339, req.DueDate); err == nil {
				task.DueDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		if err := taskRepo.Create(r.Context(), task); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

func updateTaskHandler(repo *repository.TaskRepository) http.HandlerFunc {
	type Request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    int    `json:"priority"`
		Status      string `json:"status"`
		CategoryID  *int64 `json:"category_id"`
		DueDate     string `json:"due_date"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

		task, err := repo.GetByID(r.Context(), id)
		if err != nil {
			if err == repository.ErrTaskNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Title != "" {
			task.Title = req.Title
		}
		if req.Description != "" {
			task.Description = sql.NullString{String: req.Description, Valid: true}
		}
		if req.Priority > 0 {
			task.Priority = req.Priority
		}
		if req.Status != "" {
			task.Status = req.Status
		}
		if req.CategoryID != nil {
			task.CategoryID = sql.NullInt64{Int64: *req.CategoryID, Valid: true}
		}
		if req.DueDate != "" {
			if t, err := time.Parse(time.RFC3339, req.DueDate); err == nil {
				task.DueDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		if err := repo.Update(r.Context(), task); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func deleteTaskHandler(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err := repo.Delete(r.Context(), id); err != nil {
			if err == repository.ErrTaskNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func updateTaskStatusHandler(repo *repository.TaskRepository) http.HandlerFunc {
	type Request struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		validStatuses := map[string]bool{"pending": true, "in_progress": true, "done": true}
		if !validStatuses[req.Status] {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}

		if err := repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
			if err == repository.ErrTaskNotFound {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func listCategoriesHandler(repo *repository.CategoryRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := repo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}
}

func createCategoryHandler(repo *repository.CategoryRepository) http.HandlerFunc {
	type Request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		cat := &model.Category{
			Name:  req.Name,
			Color: req.Color,
		}
		if cat.Color == "" {
			cat.Color = "#3B82F6"
		}
		if req.Description != "" {
			cat.Description = sql.NullString{String: req.Description, Valid: true}
		}

		if err := repo.Create(r.Context(), cat); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(cat)
	}
}

func statsHandler(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total, err := repo.Count(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pending, _ := repo.CountByStatus(r.Context(), "pending")
		inProgress, _ := repo.CountByStatus(r.Context(), "in_progress")
		done, _ := repo.CountByStatus(r.Context(), "done")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w, w).Encode(map[string]int64{
			"total":       total,
			"pending":     pending,
			"in_progress": inProgress,
			"done":        done,
		})
	}
}

func seedData(ctx context.Context, db *sql.DB) {
	// Check if data exists
	var count int64
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&count)
	if count > 0 {
		return
	}

	log.Println("seeding initial data...")

	categories := []struct {
		name        string
		description string
		color       string
	}{
		{"Work", "Work-related tasks", "#EF4444"},
		{"Personal", "Personal tasks", "#10B981"},
		{"Learning", "Learning goals", "#3B82F6"},
	}

	for _, c := range categories {
		db.ExecContext(ctx,
			"INSERT INTO categories (name, description, color) VALUES (?, ?, ?)",
			c.name, c.description, c.color,
		)
	}

	tasks := []struct {
		title       string
		description string
		priority    int
		status      string
	}{
		{"Complete project proposal", "Finish by end of week", 1, "in_progress"},
		{"Review pull requests", "Review and comment on open PRs", 2, "pending"},
		{"Update documentation", "Update README with new features", 3, "pending"},
		{"Read Go concurrency blog", "Learn more about channels", 2, "done"},
	}

	for _, t := range tasks {
		db.ExecContext(ctx,
			"INSERT INTO tasks (title, description, priority, status) VALUES (?, ?, ?, ?)",
			t.title, t.description, t.priority, t.status,
		)
	}

	log.Println("seed data complete")
}
```

---

## 8. Testing

### 8.1 Unit Test Repository

```go
package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"golang-sqlite/model"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=1")
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("open db: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		color TEXT DEFAULT '#3B82F6',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		priority INTEGER DEFAULT 2,
		status TEXT DEFAULT 'pending',
		category_id INTEGER,
		due_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("create schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func TestTaskRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	task := &model.Task{
		Title:    "Test Task",
		Priority: 1,
		Status:   "pending",
	}

	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	if task.ID == 0 {
		t.Error("expected task ID to be set")
	}

	// Verify
	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("title mismatch: got %q, want %q", got.Title, task.Title)
	}
}

func TestTaskRepository_CRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	// Create
	task := &model.Task{Title: "Original", Priority: 2, Status: "pending"}
	if err := repo.Create(context.Background(), task); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Read
	got, err := repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Original" {
		t.Errorf("title: got %q, want %q", got.Title, "Original")
	}

	// Update
	task.Title = "Updated"
	if err := repo.Update(context.Background(), task); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err = repo.GetByID(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("title after update: got %q, want %q", got.Title, "Updated")
	}

	// Delete
	if err := repo.Delete(context.Background(), task.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = repo.GetByID(context.Background(), task.ID)
	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskRepository_ListByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewTaskRepository(db)

	// Create tasks with different statuses
	tasks := []*model.Task{
		{Title: "Task 1", Status: "pending"},
		{Title: "Task 2", Status: "pending"},
		{Title: "Task 3", Status: "in_progress"},
		{Title: "Task 4", Status: "done"},
	}

	for _, task := range tasks {
		if err := repo.Create(context.Background(), task); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	// Test ListByStatus
	pending, err := repo.ListByStatus(context.Background(), "pending")
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("pending count: got %d, want 2", len(pending))
	}

	inProgress, err := repo.ListByStatus(context.Background(), "in_progress")
	if err != nil {
		t.Fatalf("list in_progress: %v", err)
	}
	if len(inProgress) != 1 {
		t.Errorf("in_progress count: got %d, want 1", len(inProgress))
	}
}
```

### 8.2 Run Tests

```bash
go test -v ./repository/...
# Output:
# === RUN   TestTaskRepository_Create
# --- PASS: TestTaskRepository_Create (0.00s)
# === RUN   TestTaskRepository_CRUD
# --- PASS: TestTaskRepository_CRUD (0.00s)
# === RUN   TestTaskRepository_ListByStatus
# --- PASS: TestTaskRepository_ListByStatus (0.00s)
# PASS
```

---

## 9. Ringkasan

```
┌─────────────────────────────────────────────────────────────┐
│              GOLANG + SQLite: SETUP & CRUD                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ✅ SQLite: serverless, file-based, zero-config            │
│  ✅ Driver: mattn/go-sqlite3 (paling populer)              │
│  ✅ DSN: file:app.db?_foreign_keys=1&_journal_mode=WAL      │
│  ✅ CRUD: Create, Read, Update, Delete pattern             │
│  ✅ Transactions: BeginTx, Commit, Rollback                │
│  ✅ Repository Pattern: isolate data access                  │
│  ✅ Testing: setup/teardown dengan temp file               │
│                                                             │
│  ⚠️  SQLite cuma support 1 writer simultan                 │
│  ⚠️  Selalu enable foreign_keys dengan ?_foreign_keys=1   │
│  ⚠️  WAL mode lebih baik untuk concurrency reads          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Referensi

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [modernc.org/sqlite](https://modernc.org/sqlite/) (pure Go alternative)
- [SQLite Go driver comparison](https://github.com/mattn/go-sqlite3/wiki/Alternatives)

---

## ➡️ Selanjutnya

**SQLite: Migrations & Schema**  
→ `03-sqlite-migration.md`

Di materi berikutnya, kita akan belajar cara management schema evolution dengan tools migrasi seperti `golang-migrate` atau `squirrel`, sehingga schema bisa di-version control dan di-apply secara konsisten di berbagai environment.
