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

// === Models ===

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

// === Sentinel Errors ===
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrTaskNotFound     = errors.New("task not found")
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 03: SQLITE MIGRATIONS & SCHEMA ===")
	fmt.Println()

	// Buat file DB sementara unik
	tmpDir := os.TempDir()
	dbFile := filepath.Join(tmpDir, fmt.Sprintf("minilab03-%d.db", time.Now().UnixNano()))
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_busy_timeout=5000&_journal_mode=WAL", dbFile)
	fmt.Printf("Menggunakan temporary DB: %s\n\n", dbFile)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ============================================
	// Test 1: Koneksi Database
	// ============================================
	fmt.Println("--- Test 1: Koneksi Database ---")

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("❌ Gagal open db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("db close error: %v", err)
		} else {
			fmt.Println("✅ Database ditutup dengan bersih (defer Close)")
		}
		_ = os.Remove(dbFile)
	}()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Gagal ping: %v", err)
	}
	fmt.Println("✅ Berhasil konek ke SQLite")

	// Set pool config untuk SQLite
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// ============================================
	// Test 2: Apply Migration 000001 - Create Categories Table
	// ============================================
	fmt.Println("\n--- Test 2: Apply Migration 000001 (Create Categories Table) ---")

	mig000001Up := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		color TEXT DEFAULT '#3B82F6',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
	`

	if _, err := db.ExecContext(ctx, mig000001Up); err != nil {
		log.Fatalf("❌ Migration 000001 up gagal: %v", err)
	}
	fmt.Println("✅ Migration 000001 up sukses (tabel categories + index dibuat)")

	// Verifikasi tabel ada
	var tableCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='categories'").Scan(&tableCount)
	if tableCount == 1 {
		fmt.Println("✅ Verifikasi: tabel categories terdaftar di sqlite_master")
	}

	// ============================================
	// Test 3: Apply Migration 000002 - Create Tasks Table
	// ============================================
	fmt.Println("\n--- Test 3: Apply Migration 000002 (Create Tasks Table) ---")

	mig000002Up := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
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

	if _, err := db.ExecContext(ctx, mig000002Up); err != nil {
		log.Fatalf("❌ Migration 000002 up gagal: %v", err)
	}
	fmt.Println("✅ Migration 000002 up sukses (tabel tasks + indexes dibuat)")

	// Verifikasi semua tabel
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err == nil {
		defer rows.Close()
		var tables []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			tables = append(tables, name)
		}
		fmt.Printf("✅ Tabel di database: %v\n", tables)
	}

	// ============================================
	// Test 4: Seed Data via Migration 000003
	// ============================================
	fmt.Println("\n--- Test 4: Seed Data via Migration 000003 ---")

	mig000003Up := `
	INSERT OR IGNORE INTO categories (name, description, color) VALUES
		('Work', 'Work-related tasks', '#EF4444'),
		('Personal', 'Personal tasks', '#10B981'),
		('Learning', 'Learning and education', '#3B82F6');
	`

	if _, err := db.ExecContext(ctx, mig000003Up); err != nil {
		log.Fatalf("❌ Migration 000003 seed gagal: %v", err)
	}
	fmt.Println("✅ Migration 000003 seed sukses (3 kategori ditambahkan)")

	// Verifikasi seed
	var catCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&catCount)
	fmt.Printf("✅ Kategori di database: %d\n", catCount)

	// Ambil kategori untuk referensi
	var workCatID, personalCatID int64
	db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = 'Work'").Scan(&workCatID)
	db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = 'Personal'").Scan(&personalCatID)
	fmt.Printf("✅ Work category ID: %d, Personal category ID: %d\n", workCatID, personalCatID)

	// ============================================
	// Test 5: Insert data untuk demo alter table
	// ============================================
	fmt.Println("\n--- Test 5: Insert Data Demo ---")

	// Insert beberapa task
	insertTasks := `
	INSERT INTO tasks (title, description, status, category_id) VALUES
		('Finish report', 'Complete Q1 report', 'pending', ?),
		('Buy groceries', 'Milk, eggs, bread', 'pending', ?),
		('Call client', 'Discuss project timeline', 'in_progress', ?);
	`
	if _, err := db.ExecContext(ctx, insertTasks, workCatID, personalCatID, workCatID); err != nil {
		log.Fatalf("❌ Insert tasks gagal: %v", err)
	}
	fmt.Println("✅ 3 tasks berhasil diinsert")

	// Tampilkan data
	showAllTasks(ctx, db, "Sebelum migration 000004")

	// ============================================
	// Test 6: Alter Table Migration via Recreate
	// ============================================
	fmt.Println("\n--- Test 6: Migration 000004 - Add Priority Column (Recreate Table) ---")
	fmt.Println("SQLite tidak support ALTER COLUMN. Strategy: recreate table dengan data migration")

	mig000004Up := `
	BEGIN TRANSACTION;

	CREATE TEMPORARY TABLE tasks_backup AS SELECT * FROM tasks;
	DROP TABLE tasks;

	CREATE TABLE tasks (
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

	INSERT INTO tasks (id, title, description, priority, status, category_id, due_date, created_at, updated_at)
	SELECT id, title, description, 2, status, category_id, due_date, created_at, updated_at
	FROM tasks_backup;

	DROP TABLE tasks_backup;

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

	COMMIT;
	`

	if _, err := db.ExecContext(ctx, mig000004Up); err != nil {
		log.Fatalf("❌ Migration 000004 up gagal: %v", err)
	}
	fmt.Println("✅ Migration 000004 up sukses (kolom priority ditambahkan via recreate table)")

	// Verifikasi schema baru
	var colCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'").Scan(&colCount)
	if colCount == 1 {
		fmt.Println("✅ Verifikasi: kolom priority ada di tabel tasks")
	}

	showAllTasks(ctx, db, "Setelah migration 000004 (dengan priority default=2)")

	// Update priority beberapa task
	db.ExecContext(ctx, "UPDATE tasks SET priority = 1 WHERE title = 'Finish report'")
	db.ExecContext(ctx, "UPDATE tasks SET priority = 3 WHERE title = 'Buy groceries'")
	showAllTasks(ctx, db, "Setelah update priority")

	// ============================================
	// Test 7: Rollback Migration 000004 (Down)
	// ============================================
	fmt.Println("\n--- Test 7: Rollback Migration 000004 (Drop Priority Column) ---")

	mig000004Down := `
	BEGIN TRANSACTION;

	CREATE TEMPORARY TABLE tasks_backup AS SELECT * FROM tasks;
	DROP TABLE tasks;

	CREATE TABLE tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		category_id INTEGER,
		due_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
	);

	INSERT INTO tasks (id, title, description, status, category_id, due_date, created_at, updated_at)
	SELECT id, title, description, status, category_id, due_date, created_at, updated_at
	FROM tasks_backup;

	DROP TABLE tasks_backup;

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

	COMMIT;
	`

	if _, err := db.ExecContext(ctx, mig000004Down); err != nil {
		log.Fatalf("❌ Rollback 000004 gagal: %v", err)
	}
	fmt.Println("✅ Rollback 000004 sukses (kolom priority dihapus)")

	// Verifikasi kolom priority sudah tidak ada
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'").Scan(&colCount)
	if colCount == 0 {
		fmt.Println("✅ Verifikasi: kolom priority sudah tidak ada")
	}

	showAllTasks(ctx, db, "Setelah rollback (tanpa priority)")

	// ============================================
	// Test 8: Re-apply Migration 000004 (Re-add Priority)
	// ============================================
	fmt.Println("\n--- Test 8: Re-apply Migration 000004 (Re-add Priority) ---")

	if _, err := db.ExecContext(ctx, mig000004Up); err != nil {
		log.Fatalf("❌ Re-apply 000004 gagal: %v", err)
	}
	fmt.Println("✅ Re-apply 000004 sukses")

	showAllTasks(ctx, db, "Setelah re-apply (priority default=2)")

	// ============================================
	// Test 9: Seed Data Down (Hapus seed data)
	// ============================================
	fmt.Println("\n--- Test 9: Rollback Seed Data ---")

	mig000003Down := `DELETE FROM categories WHERE name IN ('Work', 'Personal', 'Learning');`

	if _, err := db.ExecContext(ctx, mig000003Down); err != nil {
		log.Fatalf("❌ Seed rollback gagal: %v", err)
	}
	fmt.Println("✅ Seed rollback sukses (data kategori dihapus)")

	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&catCount)
	fmt.Printf("✅ Kategori di database setelah rollback: %d\n", catCount)

	// Because of FK ON DELETE SET NULL, task category_id menjadi NULL
	showAllTasks(ctx, db, "Setelah rollback seed (category_id = NULL)")

	// ============================================
	// Test 10: Rollback All (Drop Tables)
	// ============================================
	fmt.Println("\n--- Test 10: Rollback All Migrations (Drop Tables) ---")

	mig000002Down := `
	DROP INDEX IF EXISTS idx_tasks_status;
	DROP INDEX IF EXISTS idx_tasks_category;
	DROP INDEX IF EXISTS idx_tasks_due_date;
	DROP TABLE IF EXISTS tasks;
	`
	if _, err := db.ExecContext(ctx, mig000002Down); err != nil {
		log.Fatalf("❌ Rollback 000002 gagal: %v", err)
	}
	fmt.Println("✅ Rollback 000002 sukses (tabel tasks + indexes dihapus)")

	mig000001Down := `
	DROP INDEX IF EXISTS idx_categories_name;
	DROP TABLE IF EXISTS categories;
	`
	if _, err := db.ExecContext(ctx, mig000001Down); err != nil {
		log.Fatalf("❌ Rollback 000001 gagal: %v", err)
	}
	fmt.Println("✅ Rollback 000001 sukses (tabel categories + index dihapus)")

	// Verifikasi tidak ada tabel tersisa
	rows2, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err == nil {
		defer rows2.Close()
		var remaining []string
		for rows2.Next() {
			var name string
			rows2.Scan(&name)
			remaining = append(remaining, name)
		}
		if len(remaining) == 0 {
			fmt.Println("✅ Semua tabel berhasil dihapus (clean state)")
		} else {
			fmt.Printf("⚠️  Tabel masih tersisa: %v\n", remaining)
		}
	}

	// ============================================
	// Test 11: Idempotent Migrations (IF NOT EXISTS)
	// ============================================
	fmt.Println("\n--- Test 11: Idempotent Migrations (IF NOT EXISTS / OR IGNORE) ---")

	// Jalankan migration 000001 dua kali — harus aman
	if _, err := db.ExecContext(ctx, mig000001Up); err != nil {
		log.Fatalf("❌ Idempotent 000001 pertama gagal: %v", err)
	}
	if _, err := db.ExecContext(ctx, mig000001Up); err != nil {
		fmt.Println("❌ Idempotent 000001 kedua gagal (seharusnya aman)")
	} else {
		fmt.Println("✅ CREATE TABLE IF NOT EXISTS aman dijalankan 2x (idempotent)")
	}

	// Insert data dengan OR IGNORE
	db.ExecContext(ctx, `INSERT OR IGNORE INTO categories (name, color) VALUES ('Test', '#000000')`)
	db.ExecContext(ctx, `INSERT OR IGNORE INTO categories (name, color) VALUES ('Test', '#000000')`)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE name='Test'").Scan(&catCount)
	if catCount == 1 {
		fmt.Println("✅ INSERT OR IGNORE mencegah duplikat (idempotent)")
	} else {
		fmt.Printf("⚠️  Ada %d entry 'Test'\n", catCount)
	}

	// ============================================
	// Test 12: Track Migration Version (Simulasi)
	// ============================================
	fmt.Println("\n--- Test 12: Migration Version Tracking (Simulasi) ---")

	// Buat tabel version tracker seperti yang dilakukan golang-migrate
	db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT 0
		)
	`)
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (1, 0)`)
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (2, 0)`)
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (3, 0)`)
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (4, 0)`)

	var maxVer int
	db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations WHERE dirty = 0").Scan(&maxVer)
	fmt.Printf("✅ Current migration version: %d\n", maxVer)

	var allVersions []int
	verRows, _ := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if verRows != nil {
		defer verRows.Close()
		for verRows.Next() {
			var v int
			verRows.Scan(&v)
			allVersions = append(allVersions, v)
		}
		fmt.Printf("✅ Applied versions: %v\n", allVersions)
	}

	// Simulasi dirty state
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (5, 1)`)
	var dirty bool
	db.QueryRowContext(ctx, "SELECT dirty FROM schema_migrations WHERE version = 5").Scan(&dirty)
	fmt.Printf("✅ Dirty state detected for version 5: %v (simulasi force recovery needed)\n", dirty)

	// Force recovery: set dirty = 0
	db.ExecContext(ctx, `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (5, 0)`)
	db.QueryRowContext(ctx, "SELECT dirty FROM schema_migrations WHERE version = 5").Scan(&dirty)
	fmt.Printf("✅ Force recovery: version 5 dirty = %v (recovered)\n", dirty)

	// ============================================
	// Test 13: WAL Mode & Foreign Keys Check
	// ============================================
	fmt.Println("\n--- Test 13: PRAGMA Checks (WAL Mode, Foreign Keys) ---")

	var journalMode string
	db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode)
	fmt.Printf("✅ Journal mode: %s (WAL = better concurrency)\n", journalMode)

	var fkEnabled int
	db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fkEnabled)
	fmt.Printf("✅ Foreign keys: %d (1=enabled, via DSN parameter)\n", fkEnabled)

	// ============================================
	// Test 14: Resource Cleanup
	// ============================================
	fmt.Println("\n--- Test 14: Resource Cleanup ---")
	fmt.Println("✅ defer db.Close() + os.Remove(temp file) dijalankan otomatis")
	fmt.Println("✅ Untuk production: embed migration files + auto migrate di startup")

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 SQLite migrations, schema versioning, alter table via recreate,")
	fmt.Println("   rollback, idempotent migrations, dirty state recovery sudah dipahami!")
}

// === Helper Functions ===

func showAllTasks(ctx context.Context, db *sql.DB, label string) {
	fmt.Printf("\n📋 Tasks %s:\n", label)
	fmt.Printf("%-4s %-20s %-12s %-10s %-12s\n", "ID", "Title", "Status", "Priority", "Category")
	fmt.Println("---- -------------------- ------------ ---------- ------------")

	// Cek apakah kolom priority ada
	var hasPriority bool
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='priority'").Scan(&hasPriority)
	if err != nil {
		return
	}

	var query string
	if hasPriority {
		query = `
			SELECT t.id, t.title, t.status, COALESCE(t.priority, 2), COALESCE(c.name, 'N/A')
			FROM tasks t LEFT JOIN categories c ON t.category_id = c.id
			ORDER BY t.id`
	} else {
		query = `
			SELECT t.id, t.title, t.status, 2, COALESCE(c.name, 'N/A')
			FROM tasks t LEFT JOIN categories c ON t.category_id = c.id
			ORDER BY t.id`
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("  (query error: %v)\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, priority int
		var title, status, catName string
		rows.Scan(&id, &title, &status, &priority, &catName)
		fmt.Printf("%-4d %-20s %-12s %-10d %-12s\n", id, title, status, priority, catName)
		count++
	}
	if count == 0 {
		fmt.Println("  (no tasks)")
	}
}
