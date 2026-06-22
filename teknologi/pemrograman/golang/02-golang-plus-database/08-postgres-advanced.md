---
topik: PostgreSQL Advanced Queries
urutan: 8 dari 20
posisi: lanjutan
sebelumnya: Golang + PostgreSQL Setup & JSONB
prerequisites:
  - Golang + PostgreSQL Setup & JSONB (07-postgres-setup)
level: menengah-lanjut
---

> 🔗 **Lanjutan dari:** Golang + PostgreSQL Setup & JSONB
> ← Kembali ke: `07-postgres-setup.md`

# PostgreSQL: Advanced Queries

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Menulis dan mengeksekusi **CTE (Common Table Expressions)** — termasuk recursive CTE
- Menggunakan **Window Functions** (`ROW_NUMBER`, `RANK`, `LAG`, `LEAD`, `SUM OVER`, dll)
- Mengimplementasikan **Full Text Search** di PostgreSQL dengan `tsvector` dan `tsquery`
- Menggunakan **Subquery** dan **Lateral Join** yang efisien
- Melakukan **Bulk Insert** dengan `pgx.CopyFrom` (protokol COPY PostgreSQL)
- Membaca dan menganalisis **EXPLAIN ANALYZE** output dari Go
- Menggunakan **Prepared Statements** dan **Batch queries** dengan pgx

---

## 1. CTE (Common Table Expressions)

CTE adalah *named subquery* yang ditulis di bagian `WITH` sebelum query utama. Meningkatkan readability dan memungkinkan reuse hasil subquery.

### 1.1 CTE Dasar

```go
// Contoh: ambil top-5 produk terlaris beserta kategorinya
const queryTopSellers = `
WITH sales_summary AS (
    SELECT
        product_id,
        SUM(quantity) AS total_sold,
        SUM(amount)   AS total_revenue
    FROM orders
    GROUP BY product_id
),
ranked AS (
    SELECT
        p.id,
        p.name,
        p.category,
        ss.total_sold,
        ss.total_revenue,
        ROW_NUMBER() OVER (ORDER BY ss.total_revenue DESC) AS rank
    FROM products p
    JOIN sales_summary ss ON ss.product_id = p.id
)
SELECT id, name, category, total_sold, total_revenue, rank
FROM ranked
WHERE rank <= 5
ORDER BY rank
`

type TopSeller struct {
    ID           string
    Name         string
    Category     string
    TotalSold    int64
    TotalRevenue float64
    Rank         int
}

func GetTopSellers(ctx context.Context, pool *pgxpool.Pool) ([]*TopSeller, error) {
    rows, err := pool.Query(ctx, queryTopSellers)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []*TopSeller
    for rows.Next() {
        ts := &TopSeller{}
        if err := rows.Scan(&ts.ID, &ts.Name, &ts.Category,
            &ts.TotalSold, &ts.TotalRevenue, &ts.Rank); err != nil {
            return nil, err
        }
        results = append(results, ts)
    }
    return results, rows.Err()
}
```

### 1.2 Recursive CTE

Digunakan untuk query hierarki/tree structure (kategori bersarang, organisasi, thread komentar).

```sql
-- Schema: kategori dengan parent
CREATE TABLE categories (
    id        SERIAL PRIMARY KEY,
    name      TEXT NOT NULL,
    parent_id INT REFERENCES categories(id)
);
```

```go
const queryCategoryTree = `
WITH RECURSIVE category_tree AS (
    -- Anchor: kategori root (tidak punya parent)
    SELECT id, name, parent_id, 0 AS depth, name::TEXT AS path
    FROM categories
    WHERE parent_id IS NULL

    UNION ALL

    -- Recursive: gabungkan dengan anak-anaknya
    SELECT c.id, c.name, c.parent_id, ct.depth + 1,
           ct.path || ' > ' || c.name
    FROM categories c
    JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT id, name, parent_id, depth, path
FROM category_tree
ORDER BY path
`

type CategoryNode struct {
    ID       int
    Name     string
    ParentID *int
    Depth    int
    Path     string
}

func GetCategoryTree(ctx context.Context, pool *pgxpool.Pool) ([]*CategoryNode, error) {
    rows, err := pool.Query(ctx, queryCategoryTree)
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
```

---

## 2. Window Functions

Window functions melakukan kalkulasi *across* sekumpulan baris yang berhubungan dengan baris saat ini — **tanpa mereduksi jumlah baris** (berbeda dengan `GROUP BY`).

### Syntax Umum

```sql
function_name(...) OVER (
    [PARTITION BY col1, col2]   -- bagi data jadi partisi
    [ORDER BY col3 DESC]        -- urutan dalam partisi
    [ROWS|RANGE frame_clause]   -- frame baris yang dihitung
)
```

### 2.1 ROW_NUMBER, RANK, DENSE_RANK

```go
const queryRankByCategory = `
SELECT
    id,
    name,
    category,
    price,
    ROW_NUMBER() OVER (PARTITION BY category ORDER BY price DESC) AS row_num,
    RANK()       OVER (PARTITION BY category ORDER BY price DESC) AS rank,
    DENSE_RANK() OVER (PARTITION BY category ORDER BY price DESC) AS dense_rank
FROM products
ORDER BY category, price DESC
`

// ROW_NUMBER: unik (1,2,3,4...), tidak ada lompatan
// RANK:       bisa sama (1,1,3...), ada lompatan jika ada tie
// DENSE_RANK: bisa sama (1,1,2...), tidak ada lompatan
```

### 2.2 LAG & LEAD — Akses Baris Sebelum/Sesudah

```go
const queryPriceChange = `
SELECT
    recorded_at,
    price,
    LAG(price)  OVER (ORDER BY recorded_at) AS prev_price,
    LEAD(price) OVER (ORDER BY recorded_at) AS next_price,
    price - LAG(price) OVER (ORDER BY recorded_at) AS change
FROM price_history
WHERE product_id = $1
ORDER BY recorded_at
`

type PriceHistory struct {
    RecordedAt time.Time
    Price      float64
    PrevPrice  *float64  // bisa NULL (baris pertama)
    NextPrice  *float64  // bisa NULL (baris terakhir)
    Change     *float64
}
```

### 2.3 Running Total & Moving Average

```go
const queryRunningTotal = `
SELECT
    recorded_at,
    amount,
    SUM(amount) OVER (ORDER BY recorded_at
                      ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
                     ) AS running_total,
    AVG(amount) OVER (ORDER BY recorded_at
                      ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
                     ) AS moving_avg_7day
FROM daily_sales
ORDER BY recorded_at
`
```

### 2.4 NTILE — Bagi Data Menjadi N Bucket

```go
const queryQuartile = `
SELECT
    id,
    name,
    salary,
    NTILE(4) OVER (ORDER BY salary) AS quartile
FROM employees
ORDER BY salary
`
// quartile 1 = 25% terbawah, quartile 4 = 25% teratas
```

### 2.5 FIRST_VALUE & LAST_VALUE

```go
const queryFirstLast = `
SELECT
    department,
    name,
    salary,
    FIRST_VALUE(name) OVER (PARTITION BY department ORDER BY salary DESC) AS top_earner,
    LAST_VALUE(salary) OVER (
        PARTITION BY department ORDER BY salary DESC
        ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING
    ) AS min_salary
FROM employees
`
```

---

## 3. Full Text Search (FTS)

PostgreSQL punya FTS bawaan yang kuat tanpa perlu extension pihak ketiga.

### 3.1 Konsep Dasar

| Konsep | Penjelasan |
|---|---|
| `tsvector` | Representasi dokumen yang sudah di-preprocess (tokenized, stemmed) |
| `tsquery` | Query pencarian (kata kunci yang di-preprocess) |
| `to_tsvector(lang, text)` | Konversi text ke tsvector |
| `to_tsquery(lang, query)` | Konversi query string ke tsquery |
| `plainto_tsquery` | Seperti `to_tsquery` tapi input lebih bebas |
| `websearch_to_tsquery` | Parse Google-style query (`"exact phrase"`, `-exclude`) |
| `@@` | Match operator: `tsvector @@ tsquery` |
| `ts_rank` | Skor relevance untuk sorting |

### 3.2 Schema FTS

```sql
CREATE TABLE articles (
    id         SERIAL PRIMARY KEY,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL,
    author     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Kolom tsvector untuk FTS (pre-computed, lebih cepat)
    search_vec TSVECTOR GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(body,  '')), 'B') ||
        setweight(to_tsvector('english', coalesce(author,'')), 'C')
    ) STORED
);

-- GIN index pada kolom tsvector
CREATE INDEX idx_articles_search ON articles USING GIN (search_vec);
```

> 💡 **`setweight`** memberi bobot berbeda pada field (A=tertinggi, D=terendah) — pengaruhi ranking.
> **`GENERATED ALWAYS AS ... STORED`** — kolom tsvector di-update otomatis saat data berubah (PG 12+).

### 3.3 FTS dari Go

```go
type Article struct {
    ID        int
    Title     string
    Body      string
    Author    string
    CreatedAt time.Time
    Rank      float32 // opsional, untuk sorting by relevance
}

func SearchArticles(ctx context.Context, pool *pgxpool.Pool, query string, limit int) ([]*Article, error) {
    rows, err := pool.Query(ctx, `
        SELECT
            id, title, body, author, created_at,
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

    var articles []*Article
    for rows.Next() {
        a := &Article{}
        if err := rows.Scan(&a.ID, &a.Title, &a.Body, &a.Author, &a.CreatedAt, &a.Rank); err != nil {
            return nil, err
        }
        articles = append(articles, a)
    }
    return articles, rows.Err()
}
```

### 3.4 Highlight Hasil FTS

```go
// ts_headline: highlight kata yang match dalam snippet
func SearchWithHighlight(ctx context.Context, pool *pgxpool.Pool, query string) ([]string, error) {
    rows, err := pool.Query(ctx, `
        SELECT ts_headline(
            'english',
            body,
            websearch_to_tsquery('english', $1),
            'MaxWords=35, MinWords=15, StartSel=<b>, StopSel=</b>'
        )
        FROM articles
        WHERE search_vec @@ websearch_to_tsquery('english', $1)
        ORDER BY ts_rank(search_vec, websearch_to_tsquery('english', $1)) DESC
        LIMIT 5
    `, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var snippets []string
    for rows.Next() {
        var snippet string
        rows.Scan(&snippet)
        snippets = append(snippets, snippet)
    }
    return snippets, rows.Err()
}
```

---

## 4. Lateral Join

`LATERAL` memungkinkan subquery di kanan JOIN untuk mereferensi kolom dari tabel di sebelah kiri — seperti "subquery yang di-loop per baris".

```go
// Contoh: untuk setiap kategori, ambil 3 produk termurah
const queryLateral = `
SELECT c.name AS category, p.name AS product, p.price
FROM categories c
CROSS JOIN LATERAL (
    SELECT name, price
    FROM products
    WHERE category_id = c.id
    ORDER BY price ASC
    LIMIT 3
) p
ORDER BY c.name, p.price
`
```

---

## 5. Bulk Insert dengan pgx.CopyFrom

Untuk insert ribuan baris sekaligus, gunakan **COPY protocol** — jauh lebih cepat dari INSERT individual.

```go
import "github.com/jackc/pgx/v5"

func BulkInsertProducts(ctx context.Context, pool *pgxpool.Pool, products []Product) (int64, error) {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return 0, err
    }
    defer conn.Release()

    rows := make([][]any, len(products))
    for i, p := range products {
        metaJSON, _ := json.Marshal(p.Metadata)
        rows[i] = []any{p.Name, p.Price, metaJSON}
    }

    n, err := conn.Conn().CopyFrom(
        ctx,
        pgx.Identifier{"products"},          // nama tabel
        []string{"name", "price", "metadata"}, // kolom
        pgx.CopyFromRows(rows),
    )
    return n, err
}
```

> 💡 `CopyFrom` menggunakan PostgreSQL COPY protocol yang bisa **10-100x lebih cepat** dari INSERT satu per satu untuk data besar.

---

## 6. Prepared Statements

Prepared statements di-compile sekali oleh server, eksekusi berikutnya lebih cepat (parse overhead hilang).

```go
// Dengan pgxpool, statement di-cache per koneksi otomatis.
// Gunakan nama statement eksplisit jika ingin kontrol penuh:
func PreparedInsert(ctx context.Context, pool *pgxpool.Pool) error {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()

    // Prepare statement
    _, err = conn.Conn().Prepare(ctx, "insert_user",
        `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`)
    if err != nil {
        return err
    }

    // Eksekusi berkali-kali
    names := []string{"Dave", "Eve", "Frank"}
    for i, name := range names {
        var id string
        err = conn.Conn().QueryRow(ctx, "insert_user",
            name, fmt.Sprintf("%s@example.com", strings.ToLower(name)),
        ).Scan(&id)
        if err != nil {
            return fmt.Errorf("insert %s: %w", names[i], err)
        }
        fmt.Printf("Inserted %s: id=%s\n", name, id)
    }
    return nil
}
```

---

## 7. Batch Queries dengan SendBatch

`pgxpool.SendBatch` mengirim banyak query dalam satu round-trip ke server — mengurangi network latency.

```go
func BatchInsert(ctx context.Context, pool *pgxpool.Pool, users []struct{ Name, Email string }) error {
    batch := &pgx.Batch{}
    for _, u := range users {
        batch.Queue(
            `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`,
            u.Name, u.Email,
        )
    }

    results := pool.SendBatch(ctx, batch)
    defer results.Close()

    for range users {
        var id string
        if err := results.QueryRow().Scan(&id); err != nil {
            return err
        }
        fmt.Printf("Inserted id=%s\n", id)
    }
    return results.Close()
}
```

---

## 8. EXPLAIN ANALYZE dari Go

```go
func ExplainQuery(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) (string, error) {
    explainQuery := "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) " + query
    var plan string
    err := pool.QueryRow(ctx, explainQuery, args...).Scan(&plan)
    // Untuk output multi-baris:
    // rows, _ := pool.Query(ctx, explainQuery, args...)
    // ...scan each row...
    return plan, err
}

// Gunakan saat development untuk cek apakah index terpakai:
// plan, _ := ExplainQuery(ctx, pool,
//     "SELECT * FROM articles WHERE search_vec @@ websearch_to_tsquery('english', $1)", "golang")
// fmt.Println(plan)
```

---

## 9. Tips Performa

| Teknik | Kapan Dipakai |
|---|---|
| **GIN index** | JSONB, array, tsvector |
| **Partial index** | `WHERE is_active = true` — index hanya baris tertentu |
| **Covering index** | `CREATE INDEX ... INCLUDE (col1, col2)` — hindari heap fetch |
| **CopyFrom** | Bulk insert ribuan baris |
| **SendBatch** | Banyak query kecil, kurangi round-trip |
| **Prepared statement** | Query yang sama dieksekusi berkali-kali |
| **Connection pool** | `pgxpool` — jangan buat koneksi baru per request |

### Partial Index (contoh)

```sql
-- Index hanya untuk artikel yang belum dihapus
CREATE INDEX idx_articles_active ON articles (created_at DESC)
WHERE deleted_at IS NULL;

-- Query yang memanfaatkan partial index:
SELECT * FROM articles WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 20;
```

---

## 10. Checklist

- [x] CTE dasar (multi-step query dengan `WITH`)
- [x] Recursive CTE (hierarki/tree)
- [x] Window functions: `ROW_NUMBER`, `RANK`, `LAG/LEAD`, running total, moving avg, `NTILE`
- [x] Full Text Search: `tsvector`, `tsquery`, `ts_rank`, `websearch_to_tsquery`, `ts_headline`
- [x] LATERAL JOIN (subquery yang referensi baris kiri)
- [x] Bulk insert dengan `pgx.CopyFrom`
- [x] Prepared statements
- [x] Batch queries dengan `SendBatch`
- [x] `EXPLAIN ANALYZE` dari Go

---

## 11. Catatan

- **CTE di PostgreSQL bersifat optimization fence** (pre-PG 12) — sejak PG 12, optimizer bisa inline CTE sederhana. Pakai `WITH ... AS MATERIALIZED` jika ingin paksa materialisasi.
- **Window functions** tidak bisa dipakai di `WHERE` clause — gunakan CTE atau subquery sebagai wrapper.
- **FTS bahasa Indonesia** — gunakan `'simple'` sebagai language config (tidak ada stemmer Indonesia bawaan), atau install extension `pg_trgm` untuk trigram search.
- **`CopyFrom`** membutuhkan koneksi dedicated (bukan dari pool secara langsung) — gunakan `pool.Acquire()` dulu.
- **`SendBatch`** — hasil harus dikonsumsi secara berurutan sesuai urutan query dalam batch.

---

## 12. Referensi

- https://www.postgresql.org/docs/current/queries-with.html — CTE & Recursive Queries
- https://www.postgresql.org/docs/current/functions-window.html — Window Functions
- https://www.postgresql.org/docs/current/textsearch.html — Full Text Search
- https://www.postgresql.org/docs/current/sql-explain.html — EXPLAIN ANALYZE
- https://pkg.go.dev/github.com/jackc/pgx/v5#hdr-Copy_Protocol — pgx CopyFrom
- https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#Pool.SendBatch — SendBatch

---

> ⏭️ **Selanjutnya:** `09-postgres-concurrency.md` — Concurrency & Locking di PostgreSQL
