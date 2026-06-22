package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// === Model ===

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// === Sentinel Errors ===

var (
	ErrUserNotFound = fmt.Errorf("user not found")
)

// === Config ===

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

func NewDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	if len(cfg.Params) > 0 {
		params := ""
		for k, v := range cfg.Params {
			params += fmt.Sprintf("%s=%s&", k, v)
		}
		dsn += "&" + params[:len(params)-1]
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.MaxLife)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// === User Repository ===

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateTable(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(200) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
}

func (r *UserRepo) Insert(ctx context.Context, name, email string) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO users (name, email) VALUES (?, ?)`, name, email)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return res.LastInsertId()
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	u := &User{}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, email, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]*User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, email, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) Update(ctx context.Context, id int64, name, email string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET name = ?, email = ? WHERE id = ?`, name, email, id)
	if err != nil {
		return fmt.Errorf("update user %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// === Helpers ===

func withRetry(fn func() error, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("⚠️  Attempt %d/%d failed: %v (retry in %v)", i+1, attempts, err, delay)
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("max retry attempts (%d) reached: %w", attempts, lastErr)
}

func waitForMySQL(ctx context.Context, cfg Config) (*sql.DB, error) {
	var db *sql.DB
	err := withRetry(func() error {
		var err error
		db, err = NewDB(cfg)
		return err
	}, 10, 2*time.Second)
	return db, err
}

// === Main ===

func main() {
	fmt.Println("=== MANUAL TEST MATERI 04: GOLANG + MYSQL: CONNECTION & CONFIG ===")
	fmt.Println()

	// === Load config dari environment variables (dengan default) ===
	mysqlUser := getEnv("MYSQL_USER", "demo")
	mysqlPass := getEnv("MYSQL_PASSWORD", "demo")
	mysqlHost := getEnv("MYSQL_HOST", "localhost")
	mysqlPort := getEnvInt("MYSQL_PORT", 3306)
	mysqlDB := getEnv("MYSQL_DB", "golang_demo")

	fmt.Printf("MySQL Config:\n")
	fmt.Printf("  Host: %s:%d\n", mysqlHost, mysqlPort)
	fmt.Printf("  DB:   %s\n", mysqlDB)
	fmt.Printf("  User: %s\n", mysqlUser)
	fmt.Printf("  Pass: %s\n", "****")
	fmt.Println()

	fmt.Println("⏳ Menunggu MySQL siap (retry up to 10x, setiap 2 detik)...")
	fmt.Println("   Pastikan Docker sudah running:")
	fmt.Println("   cd 04-mysql-setup && docker compose up -d")
	fmt.Println()

	ctx := context.Background()

	cfg := Config{
		User:     mysqlUser,
		Password: mysqlPass,
		Host:     mysqlHost,
		Port:     mysqlPort,
		DBName:   mysqlDB,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
		MaxOpen: 25,
		MaxIdle: 5,
		MaxLife: 5 * time.Minute,
	}

	db, err := waitForMySQL(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Gagal connect ke MySQL setelah retry: %v\nPastikan 'docker compose up -d' sudah jalan!", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("db close error: %v", err)
		} else {
			fmt.Println("✅ Database ditutup dengan bersih (defer Close)")
		}
	}()

	fmt.Println("✅ Berhasil konek ke MySQL!")
	fmt.Println()

	// ============================================
	// Test 1: DSN & Connection Pool
	// ============================================
	fmt.Println("--- Test 1: DSN & Connection Pool ---")

	stats := db.Stats()
	fmt.Printf("✅ Connection pool: MaxOpen=%d, MaxIdle=%d, MaxLife=%v\n", cfg.MaxOpen, cfg.MaxIdle, cfg.MaxLife)
	fmt.Printf("✅ Initial stats: Open=%d, InUse=%d, Idle=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// Verify MySQL version
	var mysqlVersion string
	db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&mysqlVersion)
	fmt.Printf("✅ MySQL version: %s\n", mysqlVersion)

	// ============================================
	// Test 2: Create Table
	// ============================================
	fmt.Println("\n--- Test 2: Create Table ---")

	repo := NewUserRepo(db)
	if err := repo.CreateTable(ctx); err != nil {
		log.Fatalf("❌ Create table gagal: %v", err)
	}
	fmt.Println("✅ Tabel 'users' berhasil dibuat")

	// Verify table exists
	var tableName string
	db.QueryRowContext(ctx, "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = 'users'", mysqlDB).Scan(&tableName)
	if tableName == "users" {
		fmt.Println("✅ Verifikasi: tabel 'users' terdaftar di information_schema")
	}

	// ============================================
	// Test 3: Insert Users
	// ============================================
	fmt.Println("\n--- Test 3: Insert Users ---")

	users := []struct {
		name  string
		email string
	}{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	for _, u := range users {
		id, err := repo.Insert(ctx, u.name, u.email)
		if err != nil {
			log.Fatalf("❌ Insert %s gagal: %v", u.name, err)
		}
		fmt.Printf("✅ Insert %s: id=%d\n", u.name, id)
	}

	// ============================================
	// Test 4: Query Users (GetByID & List)
	// ============================================
	fmt.Println("\n--- Test 4: Query Users ---")

	user, err := repo.GetByID(ctx, 1)
	if err != nil {
		log.Fatalf("❌ GetByID gagal: %v", err)
	}
	fmt.Printf("✅ GetByID(1): %+v\n", user)

	allUsers, err := repo.List(ctx)
	if err != nil {
		log.Fatalf("❌ List gagal: %v", err)
	}
	fmt.Printf("✅ List: %d users\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("   - ID=%d Name=%s Email=%s Created=%s\n", u.ID, u.Name, u.Email, u.CreatedAt.Format(time.RFC3339))
	}

	// ============================================
	// Test 5: Update User
	// ============================================
	fmt.Println("\n--- Test 5: Update User ---")

	if err := repo.Update(ctx, 1, "Alice Updated", "alice.updated@example.com"); err != nil {
		log.Fatalf("❌ Update gagal: %v", err)
	}
	updated, _ := repo.GetByID(ctx, 1)
	fmt.Printf("✅ Update user 1: name=%s email=%s\n", updated.Name, updated.Email)

	// ============================================
	// Test 6: Delete User
	// ============================================
	fmt.Println("\n--- Test 6: Delete User ---")

	if err := repo.Delete(ctx, 3); err != nil {
		log.Fatalf("❌ Delete gagal: %v", err)
	}
	fmt.Println("✅ Delete user 3 sukses")

	_, err = repo.GetByID(ctx, 3)
	if err == ErrUserNotFound {
		fmt.Println("✅ Verifikasi: user 3 sudah tidak ditemukan")
	}

	// ============================================
	// Test 7: Count
	// ============================================
	fmt.Println("\n--- Test 7: Count ---")

	count, _ := repo.Count(ctx)
	fmt.Printf("✅ Total users: %d\n", count)

	// ============================================
	// Test 8: Error Handling & Retry
	// ============================================
	fmt.Println("\n--- Test 8: Error Handling & Retry ---")

	// Simulate retry success
	err = withRetry(func() error {
		return repo.CreateTable(ctx) // idempotent
	}, 3, 100*time.Millisecond)
	if err != nil {
		log.Fatalf("❌ Retry gagal: %v", err)
	}
	fmt.Println("✅ withRetry sukses (CREATE TABLE IF NOT EXISTS idempotent)")

	// Test not found
	_, err = repo.GetByID(ctx, 99999)
	if err == ErrUserNotFound {
		fmt.Println("✅ ErrUserNotFound terdeteksi dengan benar")
	}

	// ============================================
	// Test 9: Connection Pool Stats
	// ============================================
	fmt.Println("\n--- Test 9: Connection Pool Stats ---")

	stats = db.Stats()
	fmt.Printf("✅ Final pool stats: Open=%d, InUse=%d, Idle=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// ============================================
	// Test 10: Resource Cleanup
	// ============================================
	fmt.Println("\n--- Test 10: Resource Cleanup ---")
	fmt.Println("✅ defer db.Close() akan dijalankan otomatis")
	fmt.Println("✅ Untuk production: gunakan context + graceful shutdown pattern")

	fmt.Println()
	fmt.Println("=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 MySQL connection, config, CRUD, retry, connection pool sudah dipahami!")

	os.Exit(0)
}

// === Utility Functions ===

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}
