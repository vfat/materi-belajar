package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/minilab/02-golang-plus-database/06-mysql-migration/internal/db"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Prefix environment
	prefix := ""
	switch env {
	case "production":
		prefix = "PROD_"
	case "staging":
		prefix = "STAGING_"
	default:
		prefix = "DEV_"
	}

	// === 1. Init Config ===
	fmt.Println("🔧 Environment:", env)
	cfg := db.ConfigFromEnv(prefix)
	fmt.Printf("📦 Database: %s@%s:%d/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.DBName)

	// === 2. Ensure Database Ada ===
	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("❌ Ensure database: %v", err)
	}
	fmt.Println("✅ Database ensured")

	// === 3. Koneksi dengan Retry ===
	database, err := db.WaitForMySQL(cfg, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Connect: %v", err)
	}
	defer database.Close()
	fmt.Println("✅ Connected to MySQL")

	// === 4. Auto-Migrate ===
	fmt.Println("📋 Running migrations...")
	if err := db.RunMigrations(database, cfg.DBName); err != nil {
		log.Fatalf("❌ Migrate: %v", err)
	}
	fmt.Println("✅ Migrations complete")

	// === 5. Query Data untuk Verifikasi ===
	fmt.Println("\n📊 Verifying seed data...")
	queryTables(database)
	queryCategories(database)
	queryProducts(database)

	// === 6. Demo Transaction dengan Migration ===
	fmt.Println("\n🔄 Demo: menambah produk baru...")
	addSampleProduct(database)

	fmt.Println("\n✅ All checks passed!")
}

func queryTables(db *sql.DB) {
	rows, err := db.QueryContext(context.Background(), "SHOW TABLES")
	if err != nil {
		log.Printf("⚠️  show tables: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("  Tables:")
	for rows.Next() {
		var table string
		rows.Scan(&table)
		fmt.Printf("    - %s\n", table)
	}
}

func queryCategories(db *sql.DB) {
	rows, err := db.QueryContext(context.Background(),
		"SELECT id, name, slug FROM categories ORDER BY id")
	if err != nil {
		log.Printf("⚠️  query categories: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("  Categories:")
	for rows.Next() {
		var id int
		var name, slug string
		if err := rows.Scan(&id, &name, &slug); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		fmt.Printf("    %d. %s (%s)\n", id, name, slug)
	}
}

func queryProducts(db *sql.DB) {
	rows, err := db.QueryContext(context.Background(), `
		SELECT p.id, p.name, p.price, p.stock, COALESCE(c.name, 'Uncategorized') AS category
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id`)
	if err != nil {
		log.Printf("⚠️  query products: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("  Products:")
	for rows.Next() {
		var id, stock int
		var name, category string
		var price float64
		if err := rows.Scan(&id, &name, &price, &stock, &category); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		fmt.Printf("    %d. %s | Rp%.0f | stock: %d | %s\n", id, name, price, stock, category)
	}
}

func addSampleProduct(db *sql.DB) {
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("⚠️  begin tx: %v", err)
		return
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Cek apakah produk sudah ada
	var count int
	tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE slug = 'mouse-gaming'").Scan(&count)
	if count > 0 {
		fmt.Println("  ⏭️  Produk 'mouse-gaming' sudah ada, skip")
		tx.Commit()
		tx = nil
		return
	}

	// Tambah kategori baru
	res, err := tx.ExecContext(ctx,
		`INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE name = VALUES(name)`,
		"Aksesoris", "aksesoris", "Aksesoris komputer dan gadget")
	if err != nil {
		log.Printf("⚠️  insert kategori: %v", err)
		return
	}
	catID, _ := res.LastInsertId()
	if catID == 0 {
		// Kategori sudah ada, ambil id-nya
		tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE slug = 'aksesoris'").Scan(&catID)
	}

	// Tambah produk
	_, err = tx.ExecContext(ctx,
		`INSERT INTO products (category_id, name, slug, price, stock) VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE name = VALUES(name)`,
		catID, "Mouse Gaming Pro", "mouse-gaming", 250000, 100)
	if err != nil {
		log.Printf("⚠️  insert produk: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("⚠️  commit: %v", err)
		return
	}
	tx = nil

	fmt.Println("  ✅ Produk 'Mouse Gaming Pro' berhasil ditambahkan!")
	time.Sleep(500 * time.Millisecond)

	// Tampilkan hasil
	queryProducts(db)
}
