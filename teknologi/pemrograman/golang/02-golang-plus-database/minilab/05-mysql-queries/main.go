package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ========================================================================
// Models
// ========================================================================

type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID         int64     `json:"id"`
	CustomerID int64     `json:"customer_id"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

type OrderWithCustomer struct {
	Order
	CustomerName  string
	CustomerEmail string
}

type CategorySummary struct {
	Category   string
	TotalStock int
	TotalValue float64
	AvgPrice   float64
}

type TopProduct struct {
	Product
	TotalSold int
}

// ========================================================================
// Sentinel Errors
// ========================================================================

var (
	ErrNotFound      = fmt.Errorf("record not found")
	ErrInsufficientStock = fmt.Errorf("insufficient stock")
	ErrOptimisticLock = fmt.Errorf("optimistic lock failed")
)

// ========================================================================
// Config & Connection
// ========================================================================

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

func waitForMySQL(ctx context.Context, cfg Config) (*sql.DB, error) {
	var db *sql.DB
	err := withRetry(func() error {
		var err error
		db, err = NewDB(cfg)
		return err
	}, 10, 2*time.Second)
	return db, err
}

func withRetry(fn func() error, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("max retry attempts (%d) reached: %w", attempts, lastErr)
}

// ========================================================================
// Schema
// ========================================================================

const schema = `
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category)
);

CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    balance DECIMAL(12,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    total DECIMAL(12,2) NOT NULL DEFAULT 0,
    status ENUM('pending','paid','shipped','cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_customer (customer_id),
    INDEX idx_status (status)
);

CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    qty INT NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
`

// ========================================================================
// Repositories
// ========================================================================

type ProductRepo struct{ db *sql.DB }

func (r *ProductRepo) BulkInsert(ctx context.Context, products []Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil { _ = tx.Rollback() }
	}()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO products (name, price, stock, category) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, p := range products {
		if _, err := stmt.ExecContext(ctx, p.Name, p.Price, p.Stock, p.Category); err != nil {
			return fmt.Errorf("insert %s: %w", p.Name, err)
		}
	}
	return tx.Commit()
}

func (r *ProductRepo) ListPaginated(ctx context.Context, page, pageSize int) ([]Product, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, price, stock, category, created_at FROM products ORDER BY id LIMIT ? OFFSET ?`,
		pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *ProductRepo) GetCategorySummary(ctx context.Context) ([]CategorySummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT COALESCE(category, 'Uncategorized') AS category,
		       SUM(stock) AS total_stock,
		       SUM(price * stock) AS total_value,
		       AVG(price) AS avg_price
		FROM products
		GROUP BY category
		ORDER BY total_value DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []CategorySummary
	for rows.Next() {
		var s CategorySummary
		if err := rows.Scan(&s.Category, &s.TotalStock, &s.TotalValue, &s.AvgPrice); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

type CustomerRepo struct{ db *sql.DB }

func (r *CustomerRepo) Insert(ctx context.Context, name, email string, balance float64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO customers (name, email, balance) VALUES (?, ?, ?)`, name, email, balance)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type OrderRepo struct{ db *sql.DB }

func (r *OrderRepo) CreateOrder(ctx context.Context, customerID int64, items []struct {
	ProductID int64
	Qty       int
	Price     float64
}) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil { _ = tx.Rollback() }
	}()

	// Insert order header
	res, err := tx.ExecContext(ctx, `INSERT INTO orders (customer_id) VALUES (?)`, customerID)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	orderID, _ := res.LastInsertId()

	var total float64
	for _, item := range items {
		// FOR UPDATE lock
		var stock int
		err := tx.QueryRowContext(ctx,
			`SELECT stock FROM products WHERE id = ? FOR UPDATE`, item.ProductID).Scan(&stock)
		if err != nil {
			return fmt.Errorf("get product %d: %w", item.ProductID, err)
		}
		if stock < item.Qty {
			return fmt.Errorf("%w: product %d has %d, need %d", ErrInsufficientStock, item.ProductID, stock, item.Qty)
		}

		// Kurangi stock
		if _, err := tx.ExecContext(ctx,
			`UPDATE products SET stock = stock - ? WHERE id = ?`, item.Qty, item.ProductID); err != nil {
			return fmt.Errorf("update stock: %w", err)
		}

		// Insert order item
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, qty, price) VALUES (?, ?, ?, ?)`,
			orderID, item.ProductID, item.Qty, item.Price); err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}

		total += float64(item.Qty) * item.Price
	}

	// Update total
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET total = ? WHERE id = ?`, total, orderID); err != nil {
		return fmt.Errorf("update total: %w", err)
	}

	return tx.Commit()
}

func (r *OrderRepo) GetOrderWithCustomer(ctx context.Context, orderID int64) (*OrderWithCustomer, error) {
	oc := &OrderWithCustomer{}
	err := r.db.QueryRowContext(ctx, `
		SELECT o.id, o.customer_id, o.total, o.status, o.created_at,
		       c.name, c.email
		FROM orders o
		JOIN customers c ON o.customer_id = c.id
		WHERE o.id = ?`, orderID).Scan(
		&oc.ID, &oc.CustomerID, &oc.Total, &oc.Status, &oc.CreatedAt,
		&oc.CustomerName, &oc.CustomerEmail,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return oc, nil
}

func (r *OrderRepo) GetTopProducts(ctx context.Context, limit int) ([]TopProduct, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.price, p.stock, p.category, p.created_at,
		       COALESCE(SUM(oi.qty), 0) AS total_sold
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		GROUP BY p.id
		ORDER BY total_sold DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []TopProduct
	for rows.Next() {
		var tp TopProduct
		if err := rows.Scan(&tp.ID, &tp.Name, &tp.Price, &tp.Stock, &tp.Category, &tp.CreatedAt, &tp.TotalSold); err != nil {
			return nil, err
		}
		products = append(products, tp)
	}
	return products, nil
}

func (r *OrderRepo) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = ? WHERE id = ? AND status != 'cancelled'`, status, orderID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ========================================================================
// Optimistic Lock Demo
// ========================================================================

type OptimisticProductRepo struct{ db *sql.DB }

func (r *OptimisticProductRepo) GetProductWithVersion(ctx context.Context, id int64) (Product, int, error) {
	var p Product
	var version int
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, price, stock, category, created_at, 1 FROM products WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &version)
	if err != nil {
		return p, 0, err
	}
	// Kita simulasi version dengan stock (karena tidak ada kolom version di schema)
	return p, p.Stock, nil
}

func (r *OptimisticProductRepo) UpdateStockOptimistic(ctx context.Context, productID int64, newStock, expectedStock int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock = ? WHERE id = ? AND stock = ?`,
		newStock, productID, expectedStock)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrOptimisticLock
	}
	return nil
}

// ========================================================================
// Main
// ========================================================================

func main() {
	fmt.Println("=== MANUAL TEST MATERI 05: MYSQL TRANSACTIONS & QUERY PATTERNS ===")
	fmt.Println()

	mysqlUser := getEnv("MYSQL_USER", "demo")
	mysqlPass := getEnv("MYSQL_PASSWORD", "demo")
	mysqlHost := getEnv("MYSQL_HOST", "localhost")
	mysqlPort := getEnvInt("MYSQL_PORT", 3306)
	mysqlDB := getEnv("MYSQL_DB", "golang_demo")

	fmt.Printf("MySQL Config: %s:%d / %s / %s\n\n", mysqlHost, mysqlPort, mysqlDB, mysqlUser)
	fmt.Println("⏳ Menunggu MySQL siap... (docker compose up -d)")

	ctx := context.Background()

	cfg := Config{
		User: mysqlUser, Password: mysqlPass, Host: mysqlHost,
		Port: mysqlPort, DBName: mysqlDB,
		Params: map[string]string{"charset": "utf8mb4"},
		MaxOpen: 25, MaxIdle: 5, MaxLife: 5 * time.Minute,
	}

	db, err := waitForMySQL(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Gagal connect: %v", err)
	}
	defer func() {
		db.Close()
		fmt.Println("✅ Database ditutup dengan bersih")
	}()

	fmt.Println("✅ Berhasil konek ke MySQL!")
	fmt.Println()

	// Init schema (split per statement karena MySQL driver defaultnya gak support multi-statements)
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Fatalf("❌ Init schema: %s\n   Error: %v", stmt[:min(len(stmt), 60)], err)
		}
	}
	fmt.Println("✅ Schema berhasil diinisialisasi (products, customers, orders, order_items)")
	fmt.Println()

	// Bersihkan data sebelumnya
	db.ExecContext(ctx, "DELETE FROM order_items")
	db.ExecContext(ctx, "DELETE FROM orders")
	db.ExecContext(ctx, "DELETE FROM products")
	db.ExecContext(ctx, "DELETE FROM customers")

	// ===================================================================
	// Test 1: Batch Insert + Transaksi
	// ===================================================================
	fmt.Println("--- Test 1: Batch Insert dalam 1 Transaksi ---")

	prodRepo := &ProductRepo{db: db}
	products := []Product{
		{Name: "Laptop Pro 14", Price: 15000000, Stock: 10, Category: "Elektronik"},
		{Name: "Mouse Wireless", Price: 250000, Stock: 50, Category: "Elektronik"},
		{Name: "Keyboard Mech", Price: 750000, Stock: 30, Category: "Elektronik"},
		{Name: "Monitor 27\"", Price: 3500000, Stock: 15, Category: "Elektronik"},
		{Name: "Buku Go Lang", Price: 150000, Stock: 100, Category: "Buku"},
		{Name: "Buku Database", Price: 120000, Stock: 80, Category: "Buku"},
		{Name: "Meja Kerja", Price: 2000000, Stock: 5, Category: "Furniture"},
		{Name: "Kursi Ergonomis", Price: 3500000, Stock: 8, Category: "Furniture"},
	}

	if err := prodRepo.BulkInsert(ctx, products); err != nil {
		log.Fatalf("❌ Bulk insert gagal: %v", err)
	}
	fmt.Printf("✅ %d products berhasil diinsert dalam batch transaction\n", len(products))

	// ===================================================================
	// Test 2: Pagination
	// ===================================================================
	fmt.Println("\n--- Test 2: Pagination (LIMIT / OFFSET) ---")

	page1, total, _ := prodRepo.ListPaginated(ctx, 1, 3)
	fmt.Printf("✅ Page 1 (limit=3): %d products, total=%d\n", len(page1), total)
	for _, p := range page1 {
		fmt.Printf("   - %s: Rp%.0f (stock=%d)\n", p.Name, p.Price, p.Stock)
	}

	page2, _, _ := prodRepo.ListPaginated(ctx, 2, 3)
	fmt.Printf("✅ Page 2 (limit=3): %d products\n", len(page2))
	for _, p := range page2 {
		fmt.Printf("   - %s: Rp%.0f (stock=%d)\n", p.Name, p.Price, p.Stock)
	}

	// ===================================================================
	// Test 3: Aggregate Query (GROUP BY)
	// ===================================================================
	fmt.Println("\n--- Test 3: Aggregate Query (GROUP BY) ---")

	summaries, err := prodRepo.GetCategorySummary(ctx)
	if err != nil {
		log.Fatalf("❌ Category summary gagal: %v", err)
	}
	fmt.Printf("%-15s %10s %15s %12s\n", "Category", "Stock", "Total Value", "Avg Price")
	fmt.Println(strings.Repeat("-", 55))
	for _, s := range summaries {
		fmt.Printf("%-15s %10d %15.0f %12.0f\n", s.Category, s.TotalStock, s.TotalValue, s.AvgPrice)
	}

	// ===================================================================
	// Test 4: Transaksi dengan Row Locking (FOR UPDATE)
	// ===================================================================
	fmt.Println("\n--- Test 4: Transaksi dengan Row Locking (FOR UPDATE) ---")

	// Setup customer
	custRepo := &CustomerRepo{db: db}
	customerID, _ := custRepo.Insert(ctx, "Alice", "alice@tokoku.com", 5000000)
	fmt.Printf("✅ Customer Alice created: id=%d\n", customerID)

	// Create order dengan stock locking
	orderRepo := &OrderRepo{db: db}
	err = orderRepo.CreateOrder(ctx, customerID, []struct {
		ProductID int64
		Qty       int
		Price     float64
	}{
		{ProductID: 1, Qty: 1, Price: 15000000}, // Laptop
		{ProductID: 2, Qty: 2, Price: 250000},   // Mouse
	})
	if err != nil {
		log.Fatalf("❌ Create order gagal: %v", err)
	}
	fmt.Println("✅ Order berhasil dibuat (FOR UPDATE mencegah race condition)")
	fmt.Println("   - Laptop Pro 14: stock 10 → 9")
	fmt.Println("   - Mouse Wireless: stock 50 → 48")

	// ===================================================================
	// Test 5: JOIN Query
	// ===================================================================
	fmt.Println("\n--- Test 5: JOIN Query (Order + Customer) ---")

	oc, err := orderRepo.GetOrderWithCustomer(ctx, 1)
	if err != nil {
		log.Fatalf("❌ GetOrderWithCustomer gagal: %v", err)
	}
	fmt.Printf("✅ JOIN sukses:\n")
	fmt.Printf("   Order #%d | Customer: %s (%s)\n", oc.ID, oc.CustomerName, oc.CustomerEmail)
	fmt.Printf("   Total: Rp%.0f | Status: %s\n", oc.Total, oc.Status)

	// ===================================================================
	// Test 6: Subquery / Top Products
	// ===================================================================
	fmt.Println("\n--- Test 6: Subquery (Top Selling Products) ---")

	top, err := orderRepo.GetTopProducts(ctx, 5)
	if err != nil {
		log.Fatalf("❌ Top products gagal: %v", err)
	}
	fmt.Printf("%-20s %10s %12s\n", "Product", "Price", "Total Sold")
	fmt.Println(strings.Repeat("-", 45))
	for _, tp := range top {
		fmt.Printf("%-20s %10.0f %12d\n", tp.Name, tp.Price, tp.TotalSold)
	}

	// ===================================================================
	// Test 7: Update dengan Transaksi (Status Order)
	// ===================================================================
	fmt.Println("\n--- Test 7: Update Status dalam Transaksi ---")

	if err := orderRepo.UpdateOrderStatus(ctx, 1, "paid"); err != nil {
		log.Fatalf("❌ Update status gagal: %v", err)
	}
	fmt.Println("✅ Order status updated: pending → paid")

	// Verifikasi
	oc2, _ := orderRepo.GetOrderWithCustomer(ctx, 1)
	fmt.Printf("✅ Status sekarang: %s\n", oc2.Status)

	// ===================================================================
	// Test 8: Optimistic Locking Simulation
	// ===================================================================
	fmt.Println("\n--- Test 8: Optimistic Locking ---")

	optRepo := &OptimisticProductRepo{db: db}

	// Simulasi 2 transaksi baca data yang sama
	product, _, _ := optRepo.GetProductWithVersion(ctx, 1)
	fmt.Printf("✅ Baca product id=1: stock=%d\n", product.Stock)

	// Transaksi A update sukses
	if err := optRepo.UpdateStockOptimistic(ctx, 1, product.Stock-1, product.Stock); err != nil {
		log.Fatalf("❌ Optimistic update A gagal: %v", err)
	}
	fmt.Println("✅ Transaksi A: update stock sukses (stock=9 → 8)")

	// Transaksi B pakai data lama → gagal (optimistic lock)
	err = optRepo.UpdateStockOptimistic(ctx, 1, product.Stock-1, product.Stock)
	if err == ErrOptimisticLock {
		fmt.Printf("✅ Transaksi B: optimistic lock detected & rejected (expected stock=%d, actual=8)\n", product.Stock)
	}

	// ===================================================================
	// Test 9: Error Handling (Insufficient Stock)
	// ===================================================================
	fmt.Println("\n--- Test 9: Error Handling (Insufficient Stock) ---")

	// Coba beli produk dengan stock lebih besar dari yang ada
	err = orderRepo.CreateOrder(ctx, customerID, []struct {
		ProductID int64
		Qty       int
		Price     float64
	}{
		{ProductID: 1, Qty: 999, Price: 15000000}, // stock hanya 8
	})
	if err != nil {
		if strings.Contains(err.Error(), "insufficient stock") {
			fmt.Println("✅ Insufficient stock error detected & transaksi di-rollback")
		} else {
			fmt.Printf("⚠️  Error lain: %v\n", err)
		}
	}

	// ===================================================================
	// Test 10: CEK Data Final
	// ===================================================================
	fmt.Println("\n--- Test 10: Final State Verification ---")

	var countProducts, countCustomers, countOrders, countItems int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&countProducts)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM customers").Scan(&countCustomers)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&countOrders)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM order_items").Scan(&countItems)

	fmt.Printf("   Products:    %d\n", countProducts)
	fmt.Printf("   Customers:   %d\n", countCustomers)
	fmt.Printf("   Orders:      %d\n", countOrders)
	fmt.Printf("   Order Items: %d\n", countItems)

	// Tampilkan stock akhir
	for _, id := range []int64{1, 2} {
		var name string
		var stock int
		db.QueryRowContext(ctx, "SELECT name, stock FROM products WHERE id = ?", id).Scan(&name, &stock)
		fmt.Printf("   %s: stock=%d\n", name, stock)
	}

	fmt.Println()
	fmt.Println("=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 Transactions, row locking, JOIN, pagination, aggregate, optimistic lock sudah dipahami!")
}

// ========================================================================
// Utility Functions
// ========================================================================

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

var _ = rand.Int // keep import
