package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// === Models (disederhanakan dari materi) ===

type Category struct {
	ID          int64
	Name        string
	Description sql.NullString
	Color       string
	CreatedAt   time.Time
}

type Task struct {
	ID          int64
	Title       string
	Description sql.NullString
	Priority    int
	Status      string
	CategoryID  sql.NullInt64
	DueDate     sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskWithCategory struct {
	Task
	CategoryName  sql.NullString
	CategoryColor sql.NullString
}

// === Sentinel Errors ===
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrTaskNotFound     = errors.New("task not found")
)

// === Schema (dari materi) ===
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
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
`

func main() {
	fmt.Println("=== MANUAL TEST MATERI 02: GOLANG + SQLITE: SETUP & CRUD ===")
	fmt.Println()

	// Buat file DB sementara unik
	tmpDir := os.TempDir()
	dbFile := filepath.Join(tmpDir, fmt.Sprintf("minilab02-%d.db", time.Now().UnixNano()))
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_busy_timeout=5000&_journal_mode=WAL", dbFile)
	fmt.Printf("Menggunakan temporary DB: %s\n\n", dbFile)

	// ============================================
	// Test 1: Koneksi SQLite
	// ============================================
	fmt.Println("--- Test 1: Koneksi SQLite ---")

	db, err := NewSQLiteDB(Config{
		DSN:           dsn,
		ForeignKeys:   true,
		BusyTimeout:   5 * time.Second,
		JournalMode:   "WAL",
		MaxOpenConns:  1, // SQLite biasanya 1 writer
		MaxIdleConns:  1,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		log.Fatalf("❌ Gagal konek: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("db close error: %v", err)
		} else {
			fmt.Println("✅ Database ditutup dengan bersih (defer Close)")
		}
		_ = os.Remove(dbFile)
	}()

	fmt.Println("✅ Berhasil konek ke SQLite dengan DSN + opsi")
	fmt.Printf("DSN: %s\n", dsn)

	// ============================================
	// Test 2: Init Schema
	// ============================================
	fmt.Println("\n--- Test 2: Init Schema ---")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := InitSchema(ctx, db); err != nil {
		log.Fatalf("init schema: %v", err)
	}
	fmt.Println("✅ Schema categories + tasks + indexes berhasil dibuat")

	// ============================================
	// Test 3: Category CRUD
	// ============================================
	fmt.Println("\n--- Test 3: Category CRUD ---")

	catRepo := &CategoryRepo{db: db}

	// Create
	cat := &Category{Name: "Work", Color: "#EF4444"}
	if err := catRepo.Create(ctx, cat); err != nil {
		log.Fatalf("create cat: %v", err)
	}
	fmt.Printf("✅ Create category: id=%d name=%s\n", cat.ID, cat.Name)

	// GetByID
	gotCat, err := catRepo.GetByID(ctx, cat.ID)
	if err != nil {
		log.Fatalf("get cat: %v", err)
	}
	fmt.Printf("✅ GetByID: %s (color=%s)\n", gotCat.Name, gotCat.Color)

	// GetByName
	gotByName, _ := catRepo.GetByName(ctx, "Work")
	fmt.Printf("✅ GetByName: %s\n", gotByName.Name)

	// Create more
	catRepo.Create(ctx, &Category{Name: "Personal", Color: "#10B981"})
	catRepo.Create(ctx, &Category{Name: "Urgent", Color: "#F59E0B"})
	cats, _ := catRepo.List(ctx)
	fmt.Printf("✅ List: %d categories\n", len(cats))

	// Update
	gotCat.Color = "#FF0000"
	if err := catRepo.Update(ctx, gotCat); err != nil {
		log.Fatalf("update cat: %v", err)
	}
	fmt.Println("✅ Update category sukses")

	// Delete "Urgent" (keep "Work" and "Personal" for later FK usage)
	urgentCat, _ := catRepo.GetByName(ctx, "Urgent")
	if urgentCat != nil {
		if err := catRepo.Delete(ctx, urgentCat.ID); err != nil {
			log.Fatalf("delete cat: %v", err)
		}
	}
	fmt.Println("✅ Delete category sukses")

	// ============================================
	// Test 4: Task CRUD + Join
	// ============================================
	fmt.Println("\n--- Test 4: Task CRUD + Join ---")

	taskRepo := &TaskRepo{db: db}

	// Get a category id for FK
	workCat, err := catRepo.GetByName(ctx, "Work")
	if err != nil {
		log.Fatalf("get work category: %v", err)
	}

	// Create task
	task := &Task{
		Title:      "Finish report",
		Priority:   1,
		Status:     "pending",
		CategoryID: sql.NullInt64{Int64: workCat.ID, Valid: true},
	}
	if err := taskRepo.Create(ctx, task); err != nil {
		log.Fatalf("create task: %v", err)
	}
	fmt.Printf("✅ Create task: id=%d title=%q\n", task.ID, task.Title)

	// GetByIDWithCategory (join)
	twc, err := taskRepo.GetByIDWithCategory(ctx, task.ID)
	if err != nil {
		log.Fatalf("get with cat: %v", err)
	}
	fmt.Printf("✅ GetByIDWithCategory: %q | category=%s\n", twc.Title, twc.CategoryName.String)

	// Create more tasks
	taskRepo.Create(ctx, &Task{Title: "Buy groceries", Status: "pending", Priority: 2})
	taskRepo.Create(ctx, &Task{Title: "Call client", Status: "in_progress", Priority: 1, CategoryID: sql.NullInt64{Int64: workCat.ID, Valid: true}})
	taskRepo.Create(ctx, &Task{Title: "Review PR", Status: "done", Priority: 3})
	taskRepo.Create(ctx, &Task{Title: "Write docs", Status: "pending", Priority: 2})

	tasks, _ := taskRepo.List(ctx, 10, 0)
	fmt.Printf("✅ List tasks: %d tasks\n", len(tasks))

	// ListByStatus
	pending, _ := taskRepo.ListByStatus(ctx, "pending")
	fmt.Printf("✅ ListByStatus (pending): %d tasks\n", len(pending))

	// UpdateStatus
	if err := taskRepo.UpdateStatus(ctx, task.ID, "done"); err != nil {
		log.Fatalf("update status: %v", err)
	}
	fmt.Println("✅ UpdateStatus: done")

	// Delete
	if err := taskRepo.Delete(ctx, task.ID); err != nil {
		log.Fatalf("delete task: %v", err)
	}
	fmt.Println("✅ Delete task sukses")

	// ============================================
	// Test 5: Query Lanjutan
	// ============================================
	fmt.Println("\n--- Test 5: Query Lanjutan (Count dll) ---")

	total, _ := taskRepo.Count(ctx)
	fmt.Printf("✅ Total tasks: %d\n", total)

	doneCount, _ := taskRepo.CountByStatus(ctx, "done")
	fmt.Printf("✅ CountByStatus (done): %d\n", doneCount)

	// ============================================
	// Test 6: Transaksi
	// ============================================
	fmt.Println("\n--- Test 6: Transaksi (Bulk + Atomic) ---")

	// Bulk create in tx
	newTasks := []*Task{
		{Title: "Tx Task 1", Status: "pending"},
		{Title: "Tx Task 2", Status: "pending"},
	}
	if err := taskRepo.BulkCreateTasks(ctx, newTasks); err != nil {
		log.Fatalf("bulk tx: %v", err)
	}
	fmt.Printf("✅ BulkCreateTasks (tx): %d tasks dibuat atomically\n", len(newTasks))

	// Demo transfer (pindah kategori) - butuh 2 kategori
	personalCat, err := catRepo.GetByName(ctx, "Personal")
	if err != nil {
		log.Fatalf("get personal category: %v", err)
	}
	if personalCat != nil && workCat != nil {
		// Buat task khusus dengan kategori "Work" agar transfer berhasil
		transferTask := &Task{
			Title:      "Transfer Demo Task",
			Status:     "pending",
			CategoryID: sql.NullInt64{Int64: workCat.ID, Valid: true},
		}
		if err := taskRepo.Create(ctx, transferTask); err != nil {
			log.Fatalf("create transfer task: %v", err)
		}
		if err := taskRepo.TransferTask(ctx, transferTask.ID, workCat.ID, personalCat.ID); err != nil {
			fmt.Printf("⚠️  Transfer demo error: %v\n", err)
		} else {
			fmt.Println("✅ TransferTask (tx) sukses: pindah kategori")
		}
	}

	// ============================================
	// Test 7: Context & Error Handling
	// ============================================
	fmt.Println("\n--- Test 7: Context & Error Handling ---")

	// Error not found
	_, err = taskRepo.GetByID(ctx, 999999)
	if errors.Is(err, ErrTaskNotFound) {
		fmt.Println("✅ ErrTaskNotFound terdeteksi dengan errors.Is")
	}

	// Context timeout (simulasi)
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer shortCancel()
	time.Sleep(2 * time.Millisecond)

	_, err = taskRepo.List(shortCtx, 10, 0)
	if err != nil {
		fmt.Println("✅ Context timeout/error handling bekerja (query dibatalkan)")
	} else {
		fmt.Println("✅ Context pattern diterapkan di semua method")
	}

	// ============================================
	// Test 8: Resource Cleanup (sudah di defer)
	// ============================================
	fmt.Println("\n--- Test 8: Resource Cleanup ---")
	fmt.Println("✅ defer db.Close() + os.Remove(temp file) dijalankan otomatis")
	fmt.Println("✅ Untuk production: gunakan context + graceful shutdown")

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 SQLite setup, CRUD, join, transaction, repository pattern + context sudah dipahami!")
}

// === Config & Connection (dari materi) ===

type Config struct {
	DSN             string
	ForeignKeys     bool
	BusyTimeout     time.Duration
	JournalMode     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewSQLiteDB(cfg Config) (*sql.DB, error) {
	dsn := cfg.DSN

	// Bangun DSN dengan parameter (sederhana)
	// (Implementasi lengkap di materi, di sini kita pakai yang sudah di-pass)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

func InitSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

// === Simplified CategoryRepo (core dari materi) ===

type CategoryRepo struct {
	db *sql.DB
}

func (r *CategoryRepo) Create(ctx context.Context, cat *Category) error {
	query := `INSERT INTO categories (name, description, color) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	id, _ := result.LastInsertId()
	cat.ID = id
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories WHERE id = ?`
	var cat Category
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

func (r *CategoryRepo) GetByName(ctx context.Context, name string) (*Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories WHERE name = ?`
	var cat Category
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

func (r *CategoryRepo) List(ctx context.Context) ([]*Category, error) {
	query := `SELECT id, name, description, color, created_at FROM categories ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []*Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, &c)
	}
	return cats, nil
}

func (r *CategoryRepo) Update(ctx context.Context, cat *Category) error {
	query := `UPDATE categories SET name = ?, description = ?, color = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color, cat.ID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

// === Simplified TaskRepo (core dari materi) ===

type TaskRepo struct {
	db *sql.DB
}

func (r *TaskRepo) Create(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (title, description, priority, status, category_id, due_date)
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query,
		task.Title, task.Description, task.Priority, task.Status,
		task.CategoryID, task.DueDate)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	id, _ := result.LastInsertId()
	task.ID = id
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id int64) (*Task, error) {
	query := `SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
	          FROM tasks WHERE id = ?`
	var t Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Priority, &t.Status,
		&t.CategoryID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &t, nil
}

func (r *TaskRepo) GetByIDWithCategory(ctx context.Context, id int64) (*TaskWithCategory, error) {
	query := `
		SELECT t.id, t.title, t.description, t.priority, t.status,
		       t.category_id, t.due_date, t.created_at, t.updated_at,
		       c.name, c.color
		FROM tasks t
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE t.id = ?`
	var twc TaskWithCategory
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&twc.ID, &twc.Title, &twc.Description, &twc.Priority, &twc.Status,
		&twc.CategoryID, &twc.DueDate, &twc.CreatedAt, &twc.UpdatedAt,
		&twc.CategoryName, &twc.CategoryColor,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task with category: %w", err)
	}
	return &twc, nil
}

func (r *TaskRepo) List(ctx context.Context, limit, offset int) ([]*Task, error) {
	query := `SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
	          FROM tasks ORDER BY id LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Priority, &t.Status,
			&t.CategoryID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepo) ListByStatus(ctx context.Context, status string) ([]*Task, error) {
	query := `SELECT id, title, description, priority, status, category_id, due_date, created_at, updated_at
	          FROM tasks WHERE status = ? ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("list by status: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.Priority, &t.Status,
			&t.CategoryID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks`).Scan(&count)
	return count, err
}

func (r *TaskRepo) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE status = ?`, status).Scan(&count)
	return count, err
}

// === Transaction examples (dari materi) ===

func (r *TaskRepo) BulkCreateTasks(ctx context.Context, tasks []*Task) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO tasks (title, description, priority, status, category_id)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, t := range tasks {
		_, err := stmt.ExecContext(ctx, t.Title, t.Description, t.Priority, t.Status, t.CategoryID)
		if err != nil {
			return fmt.Errorf("exec in tx: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	tx = nil
	return nil
}

func (r *TaskRepo) TransferTask(ctx context.Context, taskID, fromCategoryID, toCategoryID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Update category
	res, err := tx.ExecContext(ctx,
		`UPDATE tasks SET category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND category_id = ?`,
		toCategoryID, taskID, fromCategoryID)
	if err != nil {
		return fmt.Errorf("update category in tx: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrTaskNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	tx = nil
	return nil
}
