package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// === Sentinel Errors ===

var ErrNotFound = errors.New("record not found")

// === Config ===

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32
	MinConns int32
	MaxLife  time.Duration
}

func DefaultConfig() Config {
	return Config{
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnvInt("PG_PORT", 5432),
		User:     getEnv("PG_USER", "demo"),
		Password: getEnv("PG_PASSWORD", "demo"),
		DBName:   getEnv("PG_DB", "golang_demo"),
		SSLMode:  getEnv("PG_SSLMODE", "disable"),
		MaxConns: 10,
		MinConns: 2,
		MaxLife:  30 * time.Minute,
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// === Pool ===

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxLife

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func WaitForPostgres(ctx context.Context, cfg Config, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		pool, err := NewPool(ctx, cfg)
		if err == nil {
			return pool, nil
		}
		lastErr = err
		log.Printf("⚠️  Attempt %d/%d: %v (retry in %v)", i+1, attempts, err, delay)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("postgres tidak tersedia setelah %d percobaan: %w", attempts, lastErr)
}

// === Models ===

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductMeta struct {
	Tags       []string          `json:"tags,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Stock      int               `json:"stock,omitempty"`
}

type Product struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Price     float64     `json:"price"`
	Metadata  ProductMeta `json:"metadata"`
	CreatedAt time.Time   `json:"created_at"`
}

// === User Repo ===

type UserRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) *UserRepo { return &UserRepo{pool: pool} }

func (r *UserRepo) CreateTable(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name       VARCHAR(100) NOT NULL,
			email      VARCHAR(200) UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	return err
}

func (r *UserRepo) Insert(ctx context.Context, name, email string) (*User, error) {
	u := &User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (name, email) VALUES ($1, $2)
		 RETURNING id, name, email, created_at`,
		name, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, email, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]*User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, created_at FROM users ORDER BY created_at`)
	if err != nil {
		return nil, err
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
	return users, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, id, name, email string) error {
	res, err := r.pool.Exec(ctx,
		`UPDATE users SET name = $1, email = $2 WHERE id = $3`,
		name, email, id,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// === Product Repo (JSONB) ===

type ProductRepo struct{ pool *pgxpool.Pool }

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo { return &ProductRepo{pool: pool} }

func (r *ProductRepo) CreateTable(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name       VARCHAR(200) NOT NULL,
			price      NUMERIC(12,2) NOT NULL,
			metadata   JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_products_metadata ON products USING GIN (metadata);
	`)
	return err
}

func (r *ProductRepo) Insert(ctx context.Context, name string, price float64, meta ProductMeta) (*Product, error) {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	p := &Product{}
	var metaRaw []byte
	err = r.pool.QueryRow(ctx,
		`INSERT INTO products (name, price, metadata)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, price, metadata, created_at`,
		name, price, metaJSON,
	).Scan(&p.ID, &p.Name, &p.Price, &metaRaw, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}
	if err := json.Unmarshal(metaRaw, &p.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return p, nil
}

func (r *ProductRepo) List(ctx context.Context) ([]*Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, price, metadata, created_at FROM products ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p := &Product{}
		var metaRaw []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &metaRaw, &p.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(metaRaw, &p.Metadata)
		products = append(products, p)
	}
	return products, rows.Err()
}

// FindByTag mencari produk yang mengandung tag tertentu di JSONB (operator @>)
func (r *ProductRepo) FindByTag(ctx context.Context, tag string) ([]*Product, error) {
	filter, _ := json.Marshal(map[string][]string{"tags": {tag}})
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, price, metadata, created_at
		 FROM products
		 WHERE metadata @> $1::jsonb
		 ORDER BY name`,
		filter,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p := &Product{}
		var metaRaw []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &metaRaw, &p.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(metaRaw, &p.Metadata)
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetStockByID mengambil nilai field 'stock' dari metadata JSONB
func (r *ProductRepo) GetStockByID(ctx context.Context, id string) (int, error) {
	var stock int
	err := r.pool.QueryRow(ctx,
		`SELECT (metadata->>'stock')::int FROM products WHERE id = $1`, id,
	).Scan(&stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return stock, nil
}

// UpdateStock mengupdate nilai 'stock' dalam JSONB menggunakan jsonb_set
func (r *ProductRepo) UpdateStock(ctx context.Context, id string, newStock int) error {
	res, err := r.pool.Exec(ctx,
		`UPDATE products
		 SET metadata = jsonb_set(metadata, '{stock}', $1::text::jsonb)
		 WHERE id = $2`,
		newStock, id,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// KeyExists mengecek apakah key tertentu ada di metadata JSONB (operator ?)
func (r *ProductRepo) KeyExists(ctx context.Context, id, key string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT metadata ? $1 FROM products WHERE id = $2`, key, id,
	).Scan(&exists)
	return exists, err
}

// === Error Handling Helper ===

func classifyPgError(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Sprintf("Duplicate key: %s", pgErr.Detail)
		case "23503":
			return fmt.Sprintf("Foreign key violation: %s", pgErr.Detail)
		case "23502":
			return fmt.Sprintf("Not null violation: column '%s'", pgErr.ColumnName)
		default:
			return fmt.Sprintf("PG Error [%s]: %s", pgErr.Code, pgErr.Message)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "Record not found (ErrNoRows)"
	}
	return fmt.Sprintf("Unknown error: %v", err)
}

// === Utility Functions ===

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// === Main ===

func main() {
	fmt.Println("=== MANUAL TEST MATERI 07: GOLANG + POSTGRESQL: SETUP & JSONB ===")
	fmt.Println()

	cfg := DefaultConfig()
	fmt.Printf("PostgreSQL Config:\n")
	fmt.Printf("  Host:   %s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("  DB:     %s\n", cfg.DBName)
	fmt.Printf("  User:   %s\n", cfg.User)
	fmt.Printf("  SSL:    %s\n", cfg.SSLMode)
	fmt.Printf("  DSN:    postgres://%s:****@%s:%d/%s?sslmode=%s\n",
		cfg.User, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	fmt.Println()

	fmt.Println("⏳ Menunggu PostgreSQL siap (retry up to 10x, setiap 2 detik)...")
	fmt.Println("   Pastikan Docker sudah running:")
	fmt.Println("   cd 07-postgres-setup && docker compose up -d")
	fmt.Println()

	ctx := context.Background()

	pool, err := WaitForPostgres(ctx, cfg, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Gagal connect ke PostgreSQL: %v\nPastikan 'docker compose up -d' sudah jalan!", err)
	}
	defer func() {
		pool.Close()
		fmt.Println("✅ Pool koneksi ditutup (defer pool.Close())")
	}()

	fmt.Println("✅ Berhasil konek ke PostgreSQL!")
	fmt.Println()

	// ============================================
	// Test 1: Koneksi & Connection Pool
	// ============================================
	fmt.Println("--- Test 1: Koneksi & Connection Pool ---")

	var pgVersion string
	pool.QueryRow(ctx, "SELECT version()").Scan(&pgVersion)
	fmt.Printf("✅ PostgreSQL version: %s\n", pgVersion[:30]+"...")

	stats := pool.Stat()
	fmt.Printf("✅ Pool stats: Total=%d, InUse=%d, Idle=%d\n",
		stats.TotalConns(), stats.AcquiredConns(), stats.IdleConns())

	// Verifikasi UUID generation
	var newUUID string
	pool.QueryRow(ctx, "SELECT gen_random_uuid()").Scan(&newUUID)
	fmt.Printf("✅ gen_random_uuid(): %s\n", newUUID)

	// ============================================
	// Test 2: Create Table (users)
	// ============================================
	fmt.Println("\n--- Test 2: Create Table (users) ---")

	userRepo := NewUserRepo(pool)
	if err := userRepo.CreateTable(ctx); err != nil {
		log.Fatalf("❌ CreateTable users gagal: %v", err)
	}
	fmt.Println("✅ Tabel 'users' berhasil dibuat (UUID PK, TIMESTAMPTZ)")

	// Verifikasi kolom di information_schema
	var colCount int
	pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM information_schema.columns
		 WHERE table_name = 'users' AND table_schema = 'public'`,
	).Scan(&colCount)
	fmt.Printf("✅ Verifikasi: tabel 'users' memiliki %d kolom\n", colCount)

	// ============================================
	// Test 3: Insert Users (RETURNING clause)
	// ============================================
	fmt.Println("\n--- Test 3: Insert Users (dengan RETURNING) ---")

	testUsers := []struct{ name, email string }{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
	}

	insertedUsers := []*User{}
	for _, u := range testUsers {
		inserted, err := userRepo.Insert(ctx, u.name, u.email)
		if err != nil {
			log.Fatalf("❌ Insert %s gagal: %v", u.name, err)
		}
		insertedUsers = append(insertedUsers, inserted)
		fmt.Printf("✅ Insert %s: id=%s created_at=%s\n",
			inserted.Name, inserted.ID[:8]+"...", inserted.CreatedAt.Format("2006-01-02 15:04:05Z"))
	}
	fmt.Println("  💡 RETURNING clause langsung memberikan data baru tanpa query tambahan")

	// ============================================
	// Test 4: Query (GetByID & List)
	// ============================================
	fmt.Println("\n--- Test 4: Query Users (GetByID & List) ---")

	aliceID := insertedUsers[0].ID
	alice, err := userRepo.GetByID(ctx, aliceID)
	if err != nil {
		log.Fatalf("❌ GetByID gagal: %v", err)
	}
	fmt.Printf("✅ GetByID(%s...): Name=%s Email=%s\n", aliceID[:8], alice.Name, alice.Email)

	allUsers, err := userRepo.List(ctx)
	if err != nil {
		log.Fatalf("❌ List gagal: %v", err)
	}
	fmt.Printf("✅ List: %d users\n", len(allUsers))
	for _, u := range allUsers {
		fmt.Printf("   - %s | %-20s | %s\n", u.ID[:8]+"...", u.Name, u.Email)
	}

	// ============================================
	// Test 5: Update & Delete
	// ============================================
	fmt.Println("\n--- Test 5: Update & Delete ---")

	if err := userRepo.Update(ctx, aliceID, "Alice Updated", "alice.new@example.com"); err != nil {
		log.Fatalf("❌ Update gagal: %v", err)
	}
	updated, _ := userRepo.GetByID(ctx, aliceID)
	fmt.Printf("✅ Update user: name=%s email=%s\n", updated.Name, updated.Email)

	charlieID := insertedUsers[2].ID
	if err := userRepo.Delete(ctx, charlieID); err != nil {
		log.Fatalf("❌ Delete gagal: %v", err)
	}
	_, err = userRepo.GetByID(ctx, charlieID)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("✅ Delete & verifikasi: user Charlie sudah tidak ada")
	}

	count, _ := userRepo.Count(ctx)
	fmt.Printf("✅ Total users sekarang: %d\n", count)

	// ============================================
	// Test 6: JSONB — Create Table products
	// ============================================
	fmt.Println("\n--- Test 6: JSONB — Tabel products ---")

	prodRepo := NewProductRepo(pool)
	if err := prodRepo.CreateTable(ctx); err != nil {
		log.Fatalf("❌ CreateTable products gagal: %v", err)
	}
	fmt.Println("✅ Tabel 'products' berhasil dibuat (JSONB + GIN index)")

	// ============================================
	// Test 7: JSONB — Insert produk
	// ============================================
	fmt.Println("\n--- Test 7: JSONB — Insert Produk ---")

	testProducts := []struct {
		name  string
		price float64
		meta  ProductMeta
	}{
		{
			"Laptop Pro",
			15000000,
			ProductMeta{
				Tags:       []string{"elektronik", "sale"},
				Attributes: map[string]string{"brand": "TechBrand", "ram": "16GB", "storage": "512GB"},
				Stock:      10,
			},
		},
		{
			"Mouse Wireless",
			350000,
			ProductMeta{
				Tags:       []string{"elektronik", "aksesoris"},
				Attributes: map[string]string{"brand": "LogiTech", "color": "black"},
				Stock:      50,
			},
		},
		{
			"Buku Go Programming",
			120000,
			ProductMeta{
				Tags:       []string{"buku", "sale"},
				Attributes: map[string]string{"penulis": "A. Doe", "halaman": "350"},
				Stock:      25,
			},
		},
	}

	insertedProducts := []*Product{}
	for _, p := range testProducts {
		inserted, err := prodRepo.Insert(ctx, p.name, p.price, p.meta)
		if err != nil {
			log.Fatalf("❌ Insert product '%s' gagal: %v", p.name, err)
		}
		insertedProducts = append(insertedProducts, inserted)
		fmt.Printf("✅ Insert '%s' (id=%s...): price=%.0f tags=%v stock=%d\n",
			inserted.Name, inserted.ID[:8], inserted.Price,
			inserted.Metadata.Tags, inserted.Metadata.Stock)
	}

	// ============================================
	// Test 8: JSONB — FindByTag (operator @>)
	// ============================================
	fmt.Println("\n--- Test 8: JSONB — FindByTag (operator @>) ---")

	for _, tag := range []string{"sale", "elektronik", "buku"} {
		found, err := prodRepo.FindByTag(ctx, tag)
		if err != nil {
			log.Fatalf("❌ FindByTag '%s' gagal: %v", tag, err)
		}
		names := make([]string, len(found))
		for i, p := range found {
			names[i] = p.Name
		}
		fmt.Printf("✅ Tag '%s': %d produk → %v\n", tag, len(found), names)
	}

	// ============================================
	// Test 9: JSONB — GetStock (->>)
	// ============================================
	fmt.Println("\n--- Test 9: JSONB — GetStock (operator ->>) ---")

	laptopID := insertedProducts[0].ID
	stock, err := prodRepo.GetStockByID(ctx, laptopID)
	if err != nil {
		log.Fatalf("❌ GetStock gagal: %v", err)
	}
	fmt.Printf("✅ Stock 'Laptop Pro': %d (via metadata->>'stock')\n", stock)

	// ============================================
	// Test 10: JSONB — UpdateStock (jsonb_set)
	// ============================================
	fmt.Println("\n--- Test 10: JSONB — UpdateStock (jsonb_set) ---")

	if err := prodRepo.UpdateStock(ctx, laptopID, 8); err != nil {
		log.Fatalf("❌ UpdateStock gagal: %v", err)
	}
	newStock, _ := prodRepo.GetStockByID(ctx, laptopID)
	fmt.Printf("✅ Update stock 'Laptop Pro': %d → %d (via jsonb_set)\n", stock, newStock)

	// ============================================
	// Test 11: JSONB — KeyExists (operator ?)
	// ============================================
	fmt.Println("\n--- Test 11: JSONB — KeyExists (operator ?) ---")

	for _, key := range []string{"tags", "attributes", "nonexistent"} {
		exists, _ := prodRepo.KeyExists(ctx, laptopID, key)
		mark := "✅"
		if !exists {
			mark = "❌"
		}
		fmt.Printf("%s Key '%s' exists: %v\n", mark, key, exists)
	}

	// ============================================
	// Test 12: JSONB — List semua produk
	// ============================================
	fmt.Println("\n--- Test 12: JSONB — List Produk ---")

	allProducts, err := prodRepo.List(ctx)
	if err != nil {
		log.Fatalf("❌ List products gagal: %v", err)
	}
	fmt.Printf("✅ Total produk: %d\n", len(allProducts))
	for _, p := range allProducts {
		metaJSON, _ := json.Marshal(p.Metadata)
		fmt.Printf("   - %s | %-22s | price=%-12.0f | metadata=%s\n",
			p.ID[:8]+"...", p.Name, p.Price, metaJSON)
	}

	// ============================================
	// Test 13: Error Handling PostgreSQL
	// ============================================
	fmt.Println("\n--- Test 13: Error Handling PostgreSQL ---")

	// 23505 unique_violation — insert email duplikat
	_, err = userRepo.Insert(ctx, "Duplicate", "alice.new@example.com")
	if err != nil {
		fmt.Printf("✅ Duplicate email error → %s\n", classifyPgError(err))
	}

	// pgx.ErrNoRows — get user yang tidak ada
	_, err = userRepo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("✅ ErrNotFound terdeteksi untuk UUID yang tidak ada")
	}

	// Update user yang tidak ada
	err = userRepo.Update(ctx, "00000000-0000-0000-0000-000000000000", "X", "x@x.com")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("✅ Update non-existent user → ErrNotFound")
	}

	// ============================================
	// Test 14: Pool Stats
	// ============================================
	fmt.Println("\n--- Test 14: Connection Pool Stats ---")

	stats = pool.Stat()
	fmt.Printf("✅ Pool stats:\n")
	fmt.Printf("   Total connections:    %d\n", stats.TotalConns())
	fmt.Printf("   Acquired (in-use):    %d\n", stats.AcquiredConns())
	fmt.Printf("   Idle connections:     %d\n", stats.IdleConns())
	fmt.Printf("   Max connections:      %d\n", cfg.MaxConns)

	// ============================================
	// Selesai
	// ============================================
	fmt.Println()
	fmt.Println("=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 PostgreSQL: koneksi, UUID, TIMESTAMPTZ, CRUD, JSONB (@>, ->>, jsonb_set) sudah dipahami!")
}
