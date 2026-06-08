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

// === Sentinel Errors (dari materi) ===
var (
	ErrNotFound   = errors.New("record not found")
	ErrUserNotFound = errors.New("user not found")
)

// User struct (dari materi)
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 01: FONDASI DATABASE DI GO ===")
	fmt.Println()

	// Buat file DB sementara unik
	tmpDir := os.TempDir()
	dbFile := filepath.Join(tmpDir, fmt.Sprintf("minilab01-%d.db", time.Now().UnixNano()))
	dsn := fmt.Sprintf("file:%s?_foreign_keys=1&_busy_timeout=5000", dbFile)
	fmt.Printf("Menggunakan temporary DB: %s\n\n", dbFile)

	// ============================================
	// Test 1: Koneksi Database dengan Pattern Benar
	// ============================================
	fmt.Println("--- Test 1: Koneksi Database dengan Pattern Benar ---")

	db, err := NewDB(Config{
		Driver:      "sqlite3",
		DSN:         dsn,
		MaxOpen:     25,
		MaxIdle:     5,
		MaxLifetime: 5 * time.Minute,
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
		// Cleanup file
		_ = os.Remove(dbFile)
	}()

	fmt.Println("✅ Berhasil konek ke SQLite dengan pattern Config + New")
	fmt.Println("✅ PingContext sukses di dalam NewDB")

	// ============================================
	// Test 2: DSN Examples
	// ============================================
	fmt.Println("\n--- Test 2: DSN Examples ---")
	fmt.Println("SQLite DSN   :", `file:app.db?_foreign_keys=1&_busy_timeout=5000`)
	fmt.Println("PostgreSQL   :", `postgres://user:password@localhost:5432/mydb?sslmode=disable`)
	fmt.Println("MySQL        :", `user:password@tcp(localhost:3306)/mydb?parseTime=true`)
	fmt.Println("⚠️  Ingat: JANGAN hardcode DSN di production. Pakai os.Getenv!")

	// ============================================
	// Test 3: Context dalam Database Operations
	// ============================================
	fmt.Println("\n--- Test 3: Context dalam Database Operations ---")

	// Init schema
	if err := initSchema(context.Background(), db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Query dengan context + timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("query count: %v", err)
	}
	fmt.Printf("✅ QueryRowContext sukses (count users = %d)\n", count)

	// Demo timeout (simulasi query lambat tidak mungkin di sqlite, tapi pattern)
	fmt.Println("✅ Context + timeout pattern sudah diterapkan di semua query")

	// ============================================
	// Test 4: Connection Pooling
	// ============================================
	fmt.Println("\n--- Test 4: Connection Pooling ---")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	// ConnMaxIdleTime hanya Go 1.15+
	// db.SetConnMaxIdleTime(1 * time.Minute)

	stats := db.Stats()
	fmt.Printf("✅ Pool stats saat ini: Open=%d InUse=%d Idle=%d (configured MaxOpen=25, MaxIdle=10, Lifetime=5m)\n",
		stats.OpenConnections, stats.InUse, stats.Idle)
	fmt.Println("✅ Pool dikonfigurasi sebelum dipakai (best practice)")

	// ============================================
	// Test 5: Error Handling
	// ============================================
	fmt.Println("\n--- Test 5: Error Handling & Sentinel Errors ---")

	_, err = getUserByID(ctx, db, 999999) // tidak ada
	if errors.Is(err, ErrUserNotFound) {
		fmt.Println("✅ errors.Is(err, ErrUserNotFound) berhasil terdeteksi")
	} else if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("✅ sql.ErrNoRows terdeteksi (fallback)")
	} else {
		fmt.Printf("err: %v\n", err)
	}

	// ============================================
	// Test 6: Transaksi
	// ============================================
	fmt.Println("\n--- Test 6: Transaksi (BeginTx + Rollback/Commit) ---")

	if err := demoTransaction(ctx, db); err != nil {
		fmt.Printf("❌ tx error: %v\n", err)
	} else {
		fmt.Println("✅ Transaksi sukses (commit)")
	}

	// ============================================
	// Test 7: CRUD Lengkap
	// ============================================
	fmt.Println("\n--- Test 7: CRUD Lengkap (dengan Context) ---")

	// Create
	user := &User{Name: "Alice Demo", Email: "alice@demo.test"}
	if err := createUser(ctx, db, user); err != nil {
		log.Fatalf("create: %v", err)
	}
	fmt.Printf("✅ Create: user id=%d name=%s\n", user.ID, user.Name)

	// GetByID
	got, err := getUserByID(ctx, db, user.ID)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	fmt.Printf("✅ GetByID: %+v\n", got)

	// List
	users, err := listUsers(ctx, db, 10, 0)
	if err != nil {
		log.Fatalf("list: %v", err)
	}
	fmt.Printf("✅ List: %d users\n", len(users))

	// Update
	got.Name = "Alice Updated"
	got.Email = "alice-updated@demo.test"
	if err := updateUser(ctx, db, got); err != nil {
		log.Fatalf("update: %v", err)
	}
	fmt.Println("✅ Update sukses")

	// Delete
	if err := deleteUser(ctx, db, got.ID); err != nil {
		log.Fatalf("delete: %v", err)
	}
	fmt.Println("✅ Delete sukses")

	// Verify delete
	_, err = getUserByID(ctx, db, got.ID)
	if errors.Is(err, ErrUserNotFound) {
		fmt.Println("✅ Get after delete → ErrUserNotFound (expected)")
	}

	// ============================================
	// Test 8: Resource Cleanup (sudah di defer di atas)
	// ============================================
	fmt.Println("\n--- Test 8: Resource Cleanup ---")
	fmt.Println("✅ defer db.Close() akan dijalankan otomatis")
	fmt.Println("✅ Untuk server: gunakan srv.Shutdown(ctx) sebelum db.Close()")

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 Fondasi database/sql + context + pool + error + tx sudah dipahami!")
}

// === Helper: Config & New (dari materi 2.2) ===

type Config struct {
	Driver      string
	DSN         string
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
}

func NewDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Verifikasi koneksi (penting!)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

// === Schema ===

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

// === CRUD dengan Context & Error Handling (dari materi) ===

func createUser(ctx context.Context, db *sql.DB, user *User) error {
	query := `INSERT INTO users (name, email) VALUES (?, ?) RETURNING id`
	err := db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func getUserByID(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	query := `SELECT id, name, email FROM users WHERE id = ?`
	var user User
	err := db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return &user, nil
}

func listUsers(ctx context.Context, db *sql.DB, limit, offset int) ([]*User, error) {
	query := `SELECT id, name, email FROM users ORDER BY id LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return users, nil
}

func updateUser(ctx context.Context, db *sql.DB, user *User) error {
	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	res, err := db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func deleteUser(ctx context.Context, db *sql.DB, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	res, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// === Transaksi Demo (dari materi 6.2) ===

func demoTransaction(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Selalu rollback jika panic atau error sebelum commit
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Insert dummy
	_, err = tx.ExecContext(ctx,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		"TxUser", "tx@example.com")
	if err != nil {
		return fmt.Errorf("insert in tx: %w", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	tx = nil // hindari rollback di defer
	return nil
}
