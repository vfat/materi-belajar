---
topik: MySQL Transactions & Query Patterns
urutan: 5 dari 20
posisi: setelah MySQL connection & config
prerequisites:
  - Golang + MySQL: Connection & Config (04-mysql-setup)
level: menengah
---

> 🚀 **Materi #05** — Transaksi, query lanjutan, dan pattern optimasi query di MySQL dengan Go.

# MySQL: Transactions & Query Patterns

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami konsep transaksi dan isolation level di MySQL
- Mengimplementasikan `BeginTx`, `Commit`, dan `Rollback` dengan benar
- Menggunakan prepared statements untuk keamanan dan performa
- Menerapkan query lanjutan: JOIN, subquery, aggregate
- Mengoptimasi query dengan pagination dan indexing
- Mengimplementasikan row-level locking untuk menghindari race condition
- Menulis batch insert/update yang efisien

---

## 1. Persiapan Docker Compose

Gunakan `docker-compose.yml` yang sudah ada dari materi 04 di folder `04-mysql-setup`, atau buat ulang dengan konfigurasi yang sama:

```yaml
services:
  mysql:
    image: mysql:8.0
    container_name: golang_mysql_demo
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: golang_demo
      MYSQL_USER: demo
      MYSQL_PASSWORD: demo
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-proot"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:
    driver: local
```

Jalankan:

```bash
docker compose -f ../04-mysql-setup/docker-compose.yml up -d
```

---

## 2. Schema: Database Toko Online

Untuk materi ini kita gunakan skema **toko online** yang lebih kompleks:

```sql
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    FOREIGN KEY (customer_id) REFERENCES customers(id)
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

CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
```

---

## 3. Transaksi Dasar

### 3.1 Pattern Transaksi yang Aman

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
)

type OrderRepo struct {
    db *sql.DB
}

func (r *OrderRepo) CreateOrder(ctx context.Context, customerID int, items []OrderItem) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    // Pastikan rollback jika terjadi panic atau error
    defer func() {
        if tx != nil {
            _ = tx.Rollback()
        }
    }()

    // Insert order header
    res, err := tx.ExecContext(ctx,
        `INSERT INTO orders (customer_id) VALUES (?)`, customerID)
    if err != nil {
        return fmt.Errorf("insert order: %w", err)
    }

    orderID, _ := res.LastInsertId()

    // Insert order items
    var total float64
    for _, item := range items {
        // Lock row product untuk cek stock
        var stock int
        err := tx.QueryRowContext(ctx,
            `SELECT stock FROM products WHERE id = ? FOR UPDATE`, item.ProductID).Scan(&stock)
        if err != nil {
            return fmt.Errorf("get product %d: %w", item.ProductID, err)
        }
        if stock < item.Qty {
            return fmt.Errorf("insufficient stock for product %d: have %d, need %d",
                item.ProductID, stock, item.Qty)
        }

        // Kurangi stock
        _, err = tx.ExecContext(ctx,
            `UPDATE products SET stock = stock - ? WHERE id = ?`, item.Qty, item.ProductID)
        if err != nil {
            return fmt.Errorf("update stock: %w", err)
        }

        // Insert item
        _, err = tx.ExecContext(ctx,
            `INSERT INTO order_items (order_id, product_id, qty, price)
             VALUES (?, ?, ?, (SELECT price FROM products WHERE id = ?))`,
            orderID, item.ProductID, item.Qty, item.ProductID)
        if err != nil {
            return fmt.Errorf("insert order item: %w", err)
        }

        total += float64(item.Qty) * item.Price
    }

    // Update total pesanan
    _, err = tx.ExecContext(ctx,
        `UPDATE orders SET total = ? WHERE id = ?`, total, orderID)
    if err != nil {
        return fmt.Errorf("update order total: %w", err)
    }

    // Commit
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    tx = nil // tandai sudah commit, bypass defer Rollback
    return nil
}
```

### 3.2 Isolation Level

```go
// Read Committed (default MySQL)
tx, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})

// Repeatable Read (default InnoDB)
tx, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})

// Serializable (paling ketat, performa rendah)
tx, _ := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
```

| Isolation Level | Dirty Read | Non-repeatable Read | Phantom Read |
|----------------|-----------|---------------------|--------------|
| Read Uncommitted | ✅ Mungkin | ✅ Mungkin | ✅ Mungkin |
| Read Committed | ❌ Aman | ✅ Mungkin | ✅ Mungkin |
| Repeatable Read | ❌ Aman | ❌ Aman | ✅ Mungkin |
| Serializable | ❌ Aman | ❌ Aman | ❌ Aman |

---

## 4. Query Patterns

### 4.1 JOIN Query

```go
type OrderWithCustomer struct {
    Order
    CustomerName  string
    CustomerEmail string
}

func (r *OrderRepo) GetOrderWithDetails(ctx context.Context, orderID int64) (*OrderWithCustomer, error) {
    query := `
        SELECT o.id, o.customer_id, o.total, o.status, o.created_at,
               c.name, c.email
        FROM orders o
        JOIN customers c ON o.customer_id = c.id
        WHERE o.id = ?`

    var oc OrderWithCustomer
    err := r.db.QueryRowContext(ctx, query, orderID).Scan(
        &oc.ID, &oc.CustomerID, &oc.Total, &oc.Status, &oc.CreatedAt,
        &oc.CustomerName, &oc.CustomerEmail,
    )
    if err != nil {
        return nil, fmt.Errorf("get order with customer: %w", err)
    }
    return &oc, nil
}
```

### 4.2 Aggregate Query

```go
type CategorySummary struct {
    Category    string  `json:"category"`
    TotalStock  int     `json:"total_stock"`
    TotalValue  float64 `json:"total_value"`
    AvgPrice    float64 `json:"avg_price"`
}

func (r *ProductRepo) GetCategorySummary(ctx context.Context) ([]CategorySummary, error) {
    query := `
        SELECT
            COALESCE(category, 'Uncategorized') AS category,
            SUM(stock) AS total_stock,
            SUM(price * stock) AS total_value,
            AVG(price) AS avg_price
        FROM products
        GROUP BY category
        ORDER BY total_value DESC`

    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("category summary: %w", err)
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
```

### 4.3 Pagination

```go
func (r *ProductRepo) ListPaginated(ctx context.Context, page, pageSize int) ([]Product, int, error) {
    // Hitung total dulu
    var total int
    err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("count products: %w", err)
    }

    offset := (page - 1) * pageSize
    query := `
        SELECT id, name, price, stock, category, created_at
        FROM products
        ORDER BY id
        LIMIT ? OFFSET ?`

    rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("list products: %w", err)
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
```

### 4.4 Batch Insert

```go
func (r *ProductRepo) BulkInsert(ctx context.Context, products []Product) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() {
        if tx != nil {
            _ = tx.Rollback()
        }
    }()

    stmt, err := tx.PrepareContext(ctx,
        `INSERT INTO products (name, price, stock, category) VALUES (?, ?, ?, ?)`)
    if err != nil {
        return fmt.Errorf("prepare: %w", err)
    }
    defer stmt.Close()

    for _, p := range products {
        if _, err := stmt.ExecContext(ctx, p.Name, p.Price, p.Stock, p.Category); err != nil {
            return fmt.Errorf("insert product %s: %w", p.Name, err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    tx = nil
    return nil
}
```

### 4.5 Subquery

```go
// Produk dengan total penjualan di atas rata-rata
func (r *ProductRepo) TopSellingProducts(ctx context.Context, limit int) ([]Product, error) {
    query := `
        SELECT p.id, p.name, p.price, p.stock, p.category, p.created_at,
               COALESCE(SUM(oi.qty), 0) AS total_sold
        FROM products p
        LEFT JOIN order_items oi ON p.id = oi.product_id
        GROUP BY p.id
        ORDER BY total_sold DESC
        LIMIT ?`

    rows, err := r.db.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, fmt.Errorf("top selling: %w", err)
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        var totalSold int
        if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &totalSold); err != nil {
            return nil, err
        }
        products = append(products, p)
    }
    return products, nil
}
```

---

## 5. Row Locking

### 5.1 Pessimistic Locking (FOR UPDATE)

Gunakan `FOR UPDATE` untuk mengunci baris yang akan diupdate:

```go
func (r *OrderRepo) DeductStock(ctx context.Context, productID, qty int) error {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
    if err != nil {
        return err
    }
    defer func() {
        if tx != nil { _ = tx.Rollback() }
    }()

    // Kunci baris product
    var stock int
    err = tx.QueryRowContext(ctx,
        `SELECT stock FROM products WHERE id = ? FOR UPDATE`, productID).Scan(&stock)
    if err != nil {
        return err
    }
    if stock < qty {
        return fmt.Errorf("insufficient stock: %d < %d", stock, qty)
    }

    _, err = tx.ExecContext(ctx,
        `UPDATE products SET stock = stock - ? WHERE id = ?`, qty, productID)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### 5.2 Optimistic Locking (version column)

```go
type Product struct {
    ID      int64
    Name    string
    Price   float64
    Stock   int
    Version int // kolom version untuk optimistic lock
}

func (r *ProductRepo) UpdateStockOptimistic(ctx context.Context, productID, newStock, expectedVersion int) error {
    result, err := r.db.ExecContext(ctx,
        `UPDATE products SET stock = ?, version = version + 1
         WHERE id = ? AND version = ?`,
        newStock, productID, expectedVersion)
    if err != nil {
        return fmt.Errorf("update stock: %w", err)
    }
    n, _ := result.RowsAffected()
    if n == 0 {
        return fmt.Errorf("optimistic lock failed: product %d modified by another transaction", productID)
    }
    return nil
}
```

---

## 6. Error Handling dalam Transaksi

```go
func isRetryableError(err error) bool {
    // Deadlock
    if strings.Contains(err.Error(), "Deadlock found") {
        return true
    }
    // Lock wait timeout
    if strings.Contains(err.Error(), "Lock wait timeout") {
        return true
    }
    return false
}

func (r *OrderRepo) CreateOrderWithRetry(ctx context.Context, customerID int, items []OrderItem) error {
    for attempt := 0; attempt < 3; attempt++ {
        err := r.CreateOrder(ctx, customerID, items)
        if err == nil {
            return nil
        }
        if !isRetryableError(err) {
            return err // bukan error retryable, langsung return
        }
        // Exponential backoff
        time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
    }
    return fmt.Errorf("failed after 3 retries")
}
```

---

## 7. Optimasi Query

| Pattern | Cara | Efek |
|---------|------|------|
| **Prepared Statement** | `PrepareContext` + `ExecContext` | Query plan reuse, anti SQL injection |
| **Batch Insert** | 1 transaksi + 1 statement | 10x lebih cepat dari multiple round trips |
| **Pagination** | `LIMIT ? OFFSET ?` + `COUNT(*)` | Hindari load semua data |
| **Indexing** | Kolom di WHERE, JOIN, ORDER BY | Full table scan → index scan |
| **Locking Minimal** | `FOR UPDATE` hanya di baris perlu | Hindari deadlock |

---

## 8. Ringkasan

| Topik | Kunci Utama |
|-------|-------------|
| Transaksi | `BeginTx` → `defer Rollback` → `Commit` → `tx = nil` |
| Isolation Level | `ReadCommitted` / `RepeatableRead` / `Serializable` |
| JOIN | `QueryRowContext` / `QueryContext` dengan JOIN SQL |
| Pagination | `LIMIT ? OFFSET ?` + hitung total |
| Batch Insert | Satu transaksi + prepared statement |
| Row Locking | `FOR UPDATE` (pessimistic) / version column (optimistic) |
| Retry | Exponential backoff untuk deadlock |

> 📝 **Next:** Lanjut ke materi #06 untuk migrasi MySQL dengan environment sync.
