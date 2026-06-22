package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// === Config & Pool ===

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func DefaultConfig() Config {
	return Config{
		Host:     getEnv("PG_HOST", "localhost"),
		Port:     getEnvInt("PG_PORT", 5432),
		User:     getEnv("PG_USER", "demo"),
		Password: getEnv("PG_PASSWORD", "demo"),
		DBName:   getEnv("PG_DB", "golang_demo"),
		SSLMode:  getEnv("PG_SSLMODE", "disable"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
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

// === Setup Schema ===

func setupSchema(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		// Tabel categories (untuk recursive CTE)
		`CREATE TABLE IF NOT EXISTS categories (
			id        SERIAL PRIMARY KEY,
			name      TEXT NOT NULL,
			parent_id INT REFERENCES categories(id)
		)`,
		// Tabel products (untuk window functions)
		`CREATE TABLE IF NOT EXISTS products (
			id          SERIAL PRIMARY KEY,
			name        TEXT NOT NULL,
			category    TEXT NOT NULL,
			price       NUMERIC(12,2) NOT NULL,
			metadata    JSONB,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// Tabel employees (untuk window functions lanjutan)
		`CREATE TABLE IF NOT EXISTS employees (
			id         SERIAL PRIMARY KEY,
			name       TEXT NOT NULL,
			department TEXT NOT NULL,
			salary     NUMERIC(12,2) NOT NULL
		)`,
		// Tabel price_history (untuk LAG/LEAD)
		`CREATE TABLE IF NOT EXISTS price_history (
			id          SERIAL PRIMARY KEY,
			product_id  INT NOT NULL,
			price       NUMERIC(12,2) NOT NULL,
			recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// Tabel orders (untuk CTE dengan agregasi)
		`CREATE TABLE IF NOT EXISTS orders (
			id         SERIAL PRIMARY KEY,
			product_id INT NOT NULL,
			quantity   INT NOT NULL,
			amount     NUMERIC(12,2) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		// Tabel articles (untuk Full Text Search)
		`CREATE TABLE IF NOT EXISTS articles (
			id         SERIAL PRIMARY KEY,
			title      TEXT NOT NULL,
			body       TEXT NOT NULL,
			author     TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			search_vec TSVECTOR GENERATED ALWAYS AS (
				setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
				setweight(to_tsvector('english', coalesce(body,  '')), 'B') ||
				setweight(to_tsvector('english', coalesce(author,'')), 'C')
			) STORED
		)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_search ON articles USING GIN (search_vec)`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("setup schema: %w\nQuery: %s", err, q[:min(80, len(q))])
		}
	}
	return nil
}

func seedData(ctx context.Context, pool *pgxpool.Pool) error {
	// Categories (hierarki 3 level)
	catQueries := []string{
		`INSERT INTO categories (name, parent_id) VALUES ('Elektronik', NULL) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Buku', NULL) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Laptop', (SELECT id FROM categories WHERE name='Elektronik')) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Aksesoris', (SELECT id FROM categories WHERE name='Elektronik')) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Mouse', (SELECT id FROM categories WHERE name='Aksesoris')) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Keyboard', (SELECT id FROM categories WHERE name='Aksesoris')) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Fiksi', (SELECT id FROM categories WHERE name='Buku')) ON CONFLICT DO NOTHING`,
		`INSERT INTO categories (name, parent_id) VALUES ('Non-Fiksi', (SELECT id FROM categories WHERE name='Buku')) ON CONFLICT DO NOTHING`,
	}
	for _, q := range catQueries {
		pool.Exec(ctx, q)
	}

	// Products
	type prod struct{ name, cat string; price float64 }
	prods := []prod{
		{"Laptop Pro X1", "Elektronik", 15000000},
		{"Laptop Air M2", "Elektronik", 20000000},
		{"Laptop Gaming G5", "Elektronik", 18000000},
		{"Mouse Wireless A", "Elektronik", 350000},
		{"Mouse Ergonomic B", "Elektronik", 450000},
		{"Keyboard Mech C", "Elektronik", 800000},
		{"Go Programming", "Buku", 120000},
		{"Clean Code", "Buku", 150000},
		{"The Pragmatic Programmer", "Buku", 130000},
		{"Database Design", "Buku", 140000},
	}
	for _, p := range prods {
		pool.Exec(ctx,
			`INSERT INTO products (name, category, price) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			p.name, p.cat, p.price)
	}

	// Employees
	type emp struct{ name, dept string; salary float64 }
	emps := []emp{
		{"Alice", "Engineering", 12000000},
		{"Bob", "Engineering", 15000000},
		{"Charlie", "Engineering", 11000000},
		{"Dave", "Marketing", 9000000},
		{"Eve", "Marketing", 10000000},
		{"Frank", "HR", 8000000},
		{"Grace", "HR", 8500000},
	}
	for _, e := range emps {
		pool.Exec(ctx,
			`INSERT INTO employees (name, department, salary) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			e.name, e.dept, e.salary)
	}

	// Price history untuk product_id=1
	basePrice := 15000000.0
	for i := 0; i < 7; i++ {
		change := (rand.Float64() - 0.4) * 500000
		basePrice += change
		pool.Exec(ctx,
			`INSERT INTO price_history (product_id, price, recorded_at)
			 VALUES (1, $1, NOW() - ($2 || ' days')::interval)`,
			basePrice, 6-i)
	}

	// Orders
	for i := 1; i <= 10; i++ {
		prodID := (i % 5) + 1
		qty := rand.Intn(5) + 1
		price := []float64{15000000, 20000000, 18000000, 350000, 450000}[prodID-1]
		pool.Exec(ctx,
			`INSERT INTO orders (product_id, quantity, amount) VALUES ($1, $2, $3)`,
			prodID, qty, price*float64(qty))
	}

	// Articles (untuk FTS)
	type article struct{ title, body, author string }
	arts := []article{
		{
			"Getting Started with Go and PostgreSQL",
			"Go is a statically typed compiled language. PostgreSQL is a powerful open-source relational database. Together they form a robust backend stack. This article covers connection pooling, CRUD operations, and error handling in Go with PostgreSQL using the pgx driver.",
			"Alice Developer",
		},
		{
			"Advanced PostgreSQL Query Techniques",
			"PostgreSQL offers powerful features like CTE, window functions, and full text search. Common Table Expressions allow you to write modular SQL queries. Window functions compute values across rows without collapsing them like GROUP BY does.",
			"Bob Engineer",
		},
		{
			"Full Text Search in PostgreSQL",
			"Full text search enables fast searching across large text fields. PostgreSQL uses tsvector for documents and tsquery for search queries. The GIN index dramatically speeds up full text search performance. Using websearch_to_tsquery you can support Google-style search syntax.",
			"Charlie Researcher",
		},
		{
			"Golang Concurrency Patterns",
			"Go's goroutines and channels make concurrent programming elegant. Worker pools, fan-out fan-in patterns, and context cancellation are essential patterns for production Go services. This article also covers database connection pooling strategies.",
			"Dave Gopher",
		},
		{
			"Database Performance Optimization",
			"Index selection is critical for query performance. PostgreSQL supports B-tree, GIN, GiST, and BRIN indexes. EXPLAIN ANALYZE helps identify slow queries and missing indexes. Partial indexes reduce index size by only indexing rows matching a condition.",
			"Eve DBA",
		},
	}
	for _, a := range arts {
		pool.Exec(ctx,
			`INSERT INTO articles (title, body, author) VALUES ($1, $2, $3)`,
			a.title, a.body, a.author)
	}

	return nil
}

// === CTE Structs & Functions ===

type CategoryNode struct {
	ID       int
	Name     string
	ParentID *int
	Depth    int
	Path     string
}

func GetCategoryTree(ctx context.Context, pool *pgxpool.Pool) ([]*CategoryNode, error) {
	rows, err := pool.Query(ctx, `
		WITH RECURSIVE category_tree AS (
			SELECT id, name, parent_id, 0 AS depth, name::TEXT AS path
			FROM categories
			WHERE parent_id IS NULL
			UNION ALL
			SELECT c.id, c.name, c.parent_id, ct.depth + 1,
			       ct.path || ' > ' || c.name
			FROM categories c
			JOIN category_tree ct ON ct.id = c.parent_id
		)
		SELECT id, name, parent_id, depth, path
		FROM category_tree
		ORDER BY path
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*CategoryNode
	for rows.Next() {
		n := &CategoryNode{}
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.Depth, &n.Path); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

type TopSeller struct {
	ProductID    int
	TotalSold    int64
	TotalRevenue float64
	Rank         int
}

func GetTopSellers(ctx context.Context, pool *pgxpool.Pool) ([]*TopSeller, error) {
	rows, err := pool.Query(ctx, `
		WITH sales_summary AS (
			SELECT
				product_id,
				SUM(quantity)::BIGINT AS total_sold,
				SUM(amount)           AS total_revenue
			FROM orders
			GROUP BY product_id
		),
		ranked AS (
			SELECT
				product_id,
				total_sold,
				total_revenue,
				ROW_NUMBER() OVER (ORDER BY total_revenue DESC) AS rank
			FROM sales_summary
		)
		SELECT product_id, total_sold, total_revenue, rank
		FROM ranked
		WHERE rank <= 5
		ORDER BY rank
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*TopSeller
	for rows.Next() {
		ts := &TopSeller{}
		if err := rows.Scan(&ts.ProductID, &ts.TotalSold, &ts.TotalRevenue, &ts.Rank); err != nil {
			return nil, err
		}
		results = append(results, ts)
	}
	return results, rows.Err()
}

// === Window Functions ===

type ProductRank struct {
	Name      string
	Category  string
	Price     float64
	RowNum    int
	Rank      int
	DenseRank int
}

func GetProductRankByCategory(ctx context.Context, pool *pgxpool.Pool) ([]*ProductRank, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			name, category, price,
			ROW_NUMBER() OVER (PARTITION BY category ORDER BY price DESC) AS row_num,
			RANK()       OVER (PARTITION BY category ORDER BY price DESC) AS rank,
			DENSE_RANK() OVER (PARTITION BY category ORDER BY price DESC) AS dense_rank
		FROM products
		ORDER BY category, price DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*ProductRank
	for rows.Next() {
		pr := &ProductRank{}
		if err := rows.Scan(&pr.Name, &pr.Category, &pr.Price,
			&pr.RowNum, &pr.Rank, &pr.DenseRank); err != nil {
			return nil, err
		}
		results = append(results, pr)
	}
	return results, rows.Err()
}

type PriceHistoryRow struct {
	RecordedAt time.Time
	Price      float64
	PrevPrice  *float64
	NextPrice  *float64
	Change     *float64
}

func GetPriceHistory(ctx context.Context, pool *pgxpool.Pool, productID int) ([]*PriceHistoryRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			recorded_at,
			price::FLOAT8,
			LAG(price::FLOAT8)  OVER (ORDER BY recorded_at) AS prev_price,
			LEAD(price::FLOAT8) OVER (ORDER BY recorded_at) AS next_price,
			price::FLOAT8 - LAG(price::FLOAT8) OVER (ORDER BY recorded_at) AS change
		FROM price_history
		WHERE product_id = $1
		ORDER BY recorded_at
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*PriceHistoryRow
	for rows.Next() {
		r := &PriceHistoryRow{}
		if err := rows.Scan(&r.RecordedAt, &r.Price, &r.PrevPrice, &r.NextPrice, &r.Change); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

type RunningTotalRow struct {
	RecordedAt   time.Time
	Amount       float64
	RunningTotal float64
}

func GetRunningTotal(ctx context.Context, pool *pgxpool.Pool) ([]*RunningTotalRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			DATE_TRUNC('day', created_at) AS day,
			SUM(amount)::FLOAT8 AS daily_amount,
			SUM(SUM(amount)::FLOAT8) OVER (
				ORDER BY DATE_TRUNC('day', created_at)
				ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
			) AS running_total
		FROM orders
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY day
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*RunningTotalRow
	for rows.Next() {
		r := &RunningTotalRow{}
		if err := rows.Scan(&r.RecordedAt, &r.Amount, &r.RunningTotal); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

type EmployeeQuartile struct {
	Name     string
	Dept     string
	Salary   float64
	Quartile int
}

func GetSalaryQuartiles(ctx context.Context, pool *pgxpool.Pool) ([]*EmployeeQuartile, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			name,
			department,
			salary::FLOAT8,
			NTILE(4) OVER (ORDER BY salary) AS quartile
		FROM employees
		ORDER BY salary
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*EmployeeQuartile
	for rows.Next() {
		eq := &EmployeeQuartile{}
		if err := rows.Scan(&eq.Name, &eq.Dept, &eq.Salary, &eq.Quartile); err != nil {
			return nil, err
		}
		results = append(results, eq)
	}
	return results, rows.Err()
}

// === Full Text Search ===

type Article struct {
	ID        int
	Title     string
	Author    string
	CreatedAt time.Time
	Rank      float32
}

func SearchArticles(ctx context.Context, pool *pgxpool.Pool, query string, limit int) ([]*Article, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			id, title, author, created_at,
			ts_rank(search_vec, websearch_to_tsquery('english', $1)) AS rank
		FROM articles
		WHERE search_vec @@ websearch_to_tsquery('english', $1)
		ORDER BY rank DESC
		LIMIT $2
	`, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*Article
	for rows.Next() {
		a := &Article{}
		if err := rows.Scan(&a.ID, &a.Title, &a.Author, &a.CreatedAt, &a.Rank); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

func SearchWithHighlight(ctx context.Context, pool *pgxpool.Pool, query string) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT ts_headline(
			'english',
			body,
			websearch_to_tsquery('english', $1),
			'MaxWords=25, MinWords=10, StartSel=[, StopSel=]'
		)
		FROM articles
		WHERE search_vec @@ websearch_to_tsquery('english', $1)
		ORDER BY ts_rank(search_vec, websearch_to_tsquery('english', $1)) DESC
		LIMIT 3
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		snippets = append(snippets, s)
	}
	return snippets, rows.Err()
}

// === Bulk Insert dengan CopyFrom ===

type ProductRow struct {
	Name     string
	Category string
	Price    float64
	Metadata map[string]any
}

func BulkInsertProducts(ctx context.Context, pool *pgxpool.Pool, products []ProductRow) (int64, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	rows := make([][]any, len(products))
	for i, p := range products {
		metaJSON, _ := json.Marshal(p.Metadata)
		rows[i] = []any{p.Name, p.Category, p.Price, metaJSON}
	}

	n, err := conn.Conn().CopyFrom(
		ctx,
		pgx.Identifier{"products"},
		[]string{"name", "category", "price", "metadata"},
		pgx.CopyFromRows(rows),
	)
	return n, err
}

// === Batch Insert dengan SendBatch ===

func BatchInsertEmployees(ctx context.Context, pool *pgxpool.Pool, emps []struct{ Name, Dept string; Salary float64 }) error {
	batch := &pgx.Batch{}
	for _, e := range emps {
		batch.Queue(
			`INSERT INTO employees (name, department, salary) VALUES ($1, $2, $3) RETURNING id`,
			e.Name, e.Dept, e.Salary,
		)
	}

	results := pool.SendBatch(ctx, batch)
	defer results.Close()

	for _, e := range emps {
		var id int
		if err := results.QueryRow().Scan(&id); err != nil {
			return fmt.Errorf("batch insert %s: %w", e.Name, err)
		}
		fmt.Printf("   Batch inserted %s: id=%d\n", e.Name, id)
	}
	return results.Close()
}

// === EXPLAIN ANALYZE ===

func ExplainQuery(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]string, error) {
	explainSQL := "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) " + query
	rows, err := pool.Query(ctx, explainSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		rows.Scan(&line)
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

// === Utility ===

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// === Main ===

func main() {
	fmt.Println("=== MANUAL TEST MATERI 08: POSTGRESQL ADVANCED QUERIES ===")
	fmt.Println()

	ctx := context.Background()
	cfg := DefaultConfig()

	fmt.Println("⏳ Menunggu PostgreSQL siap...")
	fmt.Println("   Pastikan: cd 08-postgres-advanced && docker compose up -d")
	fmt.Println()

	pool, err := WaitForPostgres(ctx, cfg, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Gagal connect: %v", err)
	}
	defer func() {
		pool.Close()
		fmt.Println("\n✅ Pool ditutup (defer pool.Close())")
	}()
	fmt.Println("✅ Berhasil konek ke PostgreSQL!")
	fmt.Println()

	// Setup
	fmt.Println("⚙️  Setup schema & seed data...")
	if err := setupSchema(ctx, pool); err != nil {
		log.Fatalf("❌ Setup schema: %v", err)
	}
	if err := seedData(ctx, pool); err != nil {
		log.Fatalf("❌ Seed data: %v", err)
	}
	fmt.Println("✅ Schema & data siap")
	fmt.Println()

	// ============================================
	// Test 1: CTE Dasar — Top Sellers
	// ============================================
	fmt.Println("--- Test 1: CTE Dasar — Top Sellers ---")

	topSellers, err := GetTopSellers(ctx, pool)
	if err != nil {
		log.Fatalf("❌ GetTopSellers: %v", err)
	}
	fmt.Printf("✅ Top %d sellers (via CTE WITH ... AS):\n", len(topSellers))
	for _, ts := range topSellers {
		fmt.Printf("   Rank %d | product_id=%d | sold=%d | revenue=%.0f\n",
			ts.Rank, ts.ProductID, ts.TotalSold, ts.TotalRevenue)
	}

	// ============================================
	// Test 2: Recursive CTE — Category Tree
	// ============================================
	fmt.Println("\n--- Test 2: Recursive CTE — Category Tree ---")

	tree, err := GetCategoryTree(ctx, pool)
	if err != nil {
		log.Fatalf("❌ GetCategoryTree: %v", err)
	}
	fmt.Printf("✅ Category tree (%d nodes, via WITH RECURSIVE):\n", len(tree))
	for _, node := range tree {
		indent := strings.Repeat("  ", node.Depth)
		fmt.Printf("   %s├─ [%d] %s (depth=%d)\n", indent, node.ID, node.Name, node.Depth)
	}

	// ============================================
	// Test 3: Window Functions — ROW_NUMBER, RANK, DENSE_RANK
	// ============================================
	fmt.Println("\n--- Test 3: Window Functions — ROW_NUMBER, RANK, DENSE_RANK ---")

	ranks, err := GetProductRankByCategory(ctx, pool)
	if err != nil {
		log.Fatalf("❌ GetProductRankByCategory: %v", err)
	}
	fmt.Printf("✅ Product rank per category (%d rows):\n", len(ranks))
	prevCat := ""
	for _, r := range ranks {
		if r.Category != prevCat {
			fmt.Printf("   [%s]\n", r.Category)
			prevCat = r.Category
		}
		fmt.Printf("     %-30s price=%-12.0f row_num=%d rank=%d dense=%d\n",
			r.Name, r.Price, r.RowNum, r.Rank, r.DenseRank)
	}

	// ============================================
	// Test 4: Window Functions — LAG & LEAD (price history)
	// ============================================
	fmt.Println("\n--- Test 4: Window Functions — LAG & LEAD (Price History) ---")

	history, err := GetPriceHistory(ctx, pool, 1)
	if err != nil {
		log.Fatalf("❌ GetPriceHistory: %v", err)
	}
	fmt.Printf("✅ Price history product_id=1 (%d rows, via LAG/LEAD):\n", len(history))
	for _, h := range history {
		change := "-"
		if h.Change != nil {
			if *h.Change >= 0 {
				change = fmt.Sprintf("+%.0f", *h.Change)
			} else {
				change = fmt.Sprintf("%.0f", *h.Change)
			}
		}
		fmt.Printf("   %s | price=%.0f | change=%s\n",
			h.RecordedAt.Format("2006-01-02"), h.Price, change)
	}

	// ============================================
	// Test 5: Window Functions — Running Total
	// ============================================
	fmt.Println("\n--- Test 5: Window Functions — Running Total ---")

	totals, err := GetRunningTotal(ctx, pool)
	if err != nil {
		log.Fatalf("❌ GetRunningTotal: %v", err)
	}
	fmt.Printf("✅ Running total orders (%d baris, via SUM OVER ROWS UNBOUNDED PRECEDING):\n", len(totals))
	for _, t := range totals {
		fmt.Printf("   %s | daily=%.0f | running_total=%.0f\n",
			t.RecordedAt.Format("2006-01-02"), t.Amount, t.RunningTotal)
	}

	// ============================================
	// Test 6: Window Functions — NTILE (quartile gaji)
	// ============================================
	fmt.Println("\n--- Test 6: Window Functions — NTILE (Salary Quartile) ---")

	quartiles, err := GetSalaryQuartiles(ctx, pool)
	if err != nil {
		log.Fatalf("❌ GetSalaryQuartiles: %v", err)
	}
	fmt.Printf("✅ Salary quartiles (%d employees, via NTILE(4)):\n", len(quartiles))
	for _, q := range quartiles {
		bar := strings.Repeat("█", q.Quartile)
		fmt.Printf("   Q%d %s | %-10s | %-20s | salary=%.0f\n",
			q.Quartile, bar, q.Dept, q.Name, q.Salary)
	}

	// ============================================
	// Test 7: Full Text Search
	// ============================================
	fmt.Println("\n--- Test 7: Full Text Search ---")

	queries := []string{"postgresql", "golang concurrency", "index performance", "full text"}
	for _, q := range queries {
		results, err := SearchArticles(ctx, pool, q, 3)
		if err != nil {
			log.Fatalf("❌ SearchArticles '%s': %v", q, err)
		}
		fmt.Printf("✅ FTS '%s': %d hasil\n", q, len(results))
		for _, a := range results {
			fmt.Printf("   [%.4f] %s — %s\n", a.Rank, a.Title, a.Author)
		}
	}

	// ============================================
	// Test 8: FTS Highlight dengan ts_headline
	// ============================================
	fmt.Println("\n--- Test 8: FTS Highlight (ts_headline) ---")

	snippets, err := SearchWithHighlight(ctx, pool, "postgresql index")
	if err != nil {
		log.Fatalf("❌ SearchWithHighlight: %v", err)
	}
	fmt.Printf("✅ Highlight 'postgresql index' (%d snippet):\n", len(snippets))
	for i, s := range snippets {
		fmt.Printf("   [%d] %s\n", i+1, s)
	}

	// ============================================
	// Test 9: Bulk Insert dengan CopyFrom
	// ============================================
	fmt.Println("\n--- Test 9: Bulk Insert — pgx.CopyFrom ---")

	bulkProducts := make([]ProductRow, 100)
	for i := range bulkProducts {
		bulkProducts[i] = ProductRow{
			Name:     fmt.Sprintf("Bulk Product %03d", i+1),
			Category: []string{"Elektronik", "Buku", "Lainnya"}[i%3],
			Price:    float64((i+1)*10000 + rand.Intn(50000)),
			Metadata: map[string]any{"bulk": true, "batch": i / 10},
		}
	}

	start := time.Now()
	n, err := BulkInsertProducts(ctx, pool, bulkProducts)
	elapsed := time.Since(start)
	if err != nil {
		log.Fatalf("❌ BulkInsertProducts: %v", err)
	}
	fmt.Printf("✅ CopyFrom: inserted %d rows in %v (%.0f rows/sec)\n",
		n, elapsed, float64(n)/elapsed.Seconds())

	// Verifikasi
	var totalProds int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&totalProds)
	fmt.Printf("✅ Total products sekarang: %d\n", totalProds)

	// ============================================
	// Test 10: Batch Insert dengan SendBatch
	// ============================================
	fmt.Println("\n--- Test 10: Batch Insert — pgx.SendBatch ---")

	batchEmps := []struct {
		Name, Dept string
		Salary     float64
	}{
		{"Zara", "Engineering", 13500000},
		{"Yudi", "Marketing", 9500000},
		{"Xena", "HR", 8800000},
	}

	fmt.Println("✅ SendBatch insert 3 employees:")
	if err := BatchInsertEmployees(ctx, pool, batchEmps); err != nil {
		log.Fatalf("❌ BatchInsertEmployees: %v", err)
	}

	var totalEmps int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM employees").Scan(&totalEmps)
	fmt.Printf("✅ Total employees: %d\n", totalEmps)

	// ============================================
	// Test 11: EXPLAIN ANALYZE
	// ============================================
	fmt.Println("\n--- Test 11: EXPLAIN ANALYZE ---")

	planLines, err := ExplainQuery(ctx, pool,
		"SELECT * FROM articles WHERE search_vec @@ websearch_to_tsquery('english', $1)",
		"postgresql")
	if err != nil {
		log.Fatalf("❌ ExplainQuery: %v", err)
	}
	fmt.Println("✅ EXPLAIN ANALYZE (FTS query):")
	for _, line := range planLines {
		if strings.Contains(line, "Index") || strings.Contains(line, "Seq Scan") ||
			strings.Contains(line, "cost=") || strings.Contains(line, "actual") {
			fmt.Printf("   %s\n", line)
		}
	}
	if len(planLines) > 0 {
		// Tunjukkan apakah GIN index terpakai
		fullPlan := strings.Join(planLines, "\n")
		if strings.Contains(fullPlan, "Bitmap Index Scan") || strings.Contains(fullPlan, "Index Scan") {
			fmt.Println("   ✅ GIN index digunakan untuk FTS query!")
		} else {
			fmt.Println("   ℹ️  Seq Scan (data terlalu kecil untuk index)")
		}
	}

	// ============================================
	// Selesai
	// ============================================
	fmt.Println()
	fmt.Println("=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 Advanced Queries: CTE, Recursive CTE, Window Functions,")
	fmt.Println("   Full Text Search, CopyFrom, SendBatch, EXPLAIN ANALYZE — sudah dipahami!")
}
