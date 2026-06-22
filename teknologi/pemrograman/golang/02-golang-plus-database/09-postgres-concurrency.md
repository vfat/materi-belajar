---
topik: PostgreSQL Concurrency & Locking
urutan: 9 dari 20
posisi: lanjutan
sebelumnya: PostgreSQL Advanced Queries
prerequisites:
  - PostgreSQL Advanced Queries (08-postgres-advanced)
  - Golang + PostgreSQL Setup & JSONB (07-postgres-setup)
level: menengah-lanjut
---

> 🔗 **Lanjutan dari:** PostgreSQL Advanced Queries
> ← Kembali ke: `08-postgres-advanced.md`

# PostgreSQL: Concurrency & Locking

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Memahami **isolation levels** di PostgreSQL dan dampaknya terhadap concurrency
- Menggunakan **transaksi** (`BEGIN`/`COMMIT`/`ROLLBACK`) dengan `pgx`
- Mengimplementasikan **Optimistic Locking** dengan version/timestamp column
- Mengimplementasikan **Pessimistic Locking** dengan `SELECT FOR UPDATE` dan `SKIP LOCKED`
- Menangani **deadlock** dan retry logic
- Menggunakan **Advisory Locks** untuk koordinasi proses
- Membangun **job queue** sederhana dengan `SKIP LOCKED`
- Mendeteksi dan menghindari masalah concurrency umum: dirty read, phantom read, lost update

---

## 1. Isolation Levels

PostgreSQL mendukung 4 isolation level standar SQL, tapi secara internal mengimplementasikan 3:

| Level | Dirty Read | Non-Repeatable Read | Phantom Read | Lost Update |
|---|---|---|---|---|
| **Read Uncommitted** | ❌ (PG: tidak ada) | ✅ bisa | ✅ bisa | ✅ bisa |
| **Read Committed** (default) | ✅ aman | ✅ bisa | ✅ bisa | ✅ bisa |
| **Repeatable Read** | ✅ aman | ✅ aman | ✅ aman (PG) | ✅ aman |
| **Serializable** | ✅ aman | ✅ aman | ✅ aman | ✅ aman |

> 💡 PostgreSQL **Read Uncommitted** secara efektif berperilaku seperti **Read Committed**.
> **Repeatable Read** di PostgreSQL juga melindungi dari Phantom Reads (lebih kuat dari standar SQL).

### Set Isolation Level di Go

```go
tx, err := pool.BeginTx(ctx, pgx.TxOptions{
    IsoLevel: pgx.Serializable,  // atau ReadCommitted, RepeatableRead
})
```

---

## 2. Transaksi dengan pgx

### 2.1 Pola Dasar

```go
func TransferBalance(ctx context.Context, pool *pgxpool.Pool, fromID, toID string, amount float64) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx) // no-op jika sudah commit

    // Kurangi saldo pengirim
    var fromBalance float64
    err = tx.QueryRow(ctx,
        `UPDATE accounts SET balance = balance - $1 WHERE id = $2 RETURNING balance`,
        amount, fromID,
    ).Scan(&fromBalance)
    if err != nil {
        return fmt.Errorf("debit: %w", err)
    }
    if fromBalance < 0 {
        return fmt.Errorf("saldo tidak cukup")
    }

    // Tambah saldo penerima
    _, err = tx.Exec(ctx,
        `UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
        amount, toID,
    )
    if err != nil {
        return fmt.Errorf("credit: %w", err)
    }

    return tx.Commit(ctx)
}
```

### 2.2 Helper: `withTx`

```go
func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit(ctx)
}

// Penggunaan:
err = withTx(ctx, pool, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, `INSERT INTO logs (msg) VALUES ($1)`, "hello")
    return err
})
```

---

## 3. Masalah Concurrency: Lost Update

**Lost Update** terjadi saat dua transaksi membaca nilai yang sama lalu masing-masing melakukan update berdasarkan nilai tersebut — salah satu update hilang.

```
T1: READ balance=100  → SET balance=110 (tambah 10)
T2: READ balance=100  → SET balance=90  (kurangi 10)
Hasil: balance=90 — update T1 hilang!
```

Solusi: **Optimistic** atau **Pessimistic Locking**.

---

## 4. Optimistic Locking

Strategi: tambah kolom `version` (integer) atau `updated_at` (timestamp). Setiap UPDATE menyertakan versi yang diketahui — jika ada yang lebih dulu update, `RowsAffected() == 0`.

```sql
CREATE TABLE products (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    stock      INT NOT NULL DEFAULT 0,
    version    INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

```go
var ErrConflict = errors.New("conflict: data sudah diubah oleh proses lain")

func UpdateStockOptimistic(ctx context.Context, pool *pgxpool.Pool,
    id string, delta, knownVersion int) error {

    res, err := pool.Exec(ctx, `
        UPDATE products
        SET stock     = stock + $1,
            version   = version + 1,
            updated_at = NOW()
        WHERE id = $2
          AND version = $3
    `, delta, id, knownVersion)
    if err != nil {
        return err
    }
    if res.RowsAffected() == 0 {
        return ErrConflict
    }
    return nil
}

// Dengan retry:
func UpdateStockWithRetry(ctx context.Context, pool *pgxpool.Pool, id string, delta, maxRetry int) error {
    for i := 0; i < maxRetry; i++ {
        // Baca versi saat ini
        var version int
        err := pool.QueryRow(ctx,
            `SELECT version FROM products WHERE id = $1`, id,
        ).Scan(&version)
        if err != nil {
            return err
        }

        // Coba update
        err = UpdateStockOptimistic(ctx, pool, id, delta, version)
        if err == nil {
            return nil
        }
        if !errors.Is(err, ErrConflict) {
            return err
        }
        // Ada conflict — retry
        time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
    }
    return fmt.Errorf("gagal update setelah %d retry", maxRetry)
}
```

> **Kapan pakai Optimistic Locking?**
> - Konflik jarang terjadi (read-heavy workload)
> - Tidak ingin hold lock lama
> - Tidak masalah retry dari sisi client

---

## 5. Pessimistic Locking — SELECT FOR UPDATE

**SELECT FOR UPDATE** mengunci baris yang dibaca hingga transaksi selesai. Proses lain yang mencoba mengunci baris yang sama akan **menunggu (block)**.

```go
func UpdateStockPessimistic(ctx context.Context, pool *pgxpool.Pool, id string, delta int) error {
    return withTx(ctx, pool, func(tx pgx.Tx) error {
        // Kunci baris terlebih dahulu
        var stock int
        err := tx.QueryRow(ctx,
            `SELECT stock FROM products WHERE id = $1 FOR UPDATE`,
            id,
        ).Scan(&stock)
        if err != nil {
            return err
        }

        if stock+delta < 0 {
            return fmt.Errorf("stok tidak cukup (stock=%d, delta=%d)", stock, delta)
        }

        _, err = tx.Exec(ctx,
            `UPDATE products SET stock = stock + $1 WHERE id = $2`,
            delta, id,
        )
        return err
    })
}
```

### Variasi FOR UPDATE

| Syntax | Perilaku |
|---|---|
| `FOR UPDATE` | Lock baris, proses lain WAIT |
| `FOR UPDATE NOWAIT` | Lock baris, proses lain dapat error langsung |
| `FOR UPDATE SKIP LOCKED` | Skip baris yang sedang dikunci, ambil yang bebas |
| `FOR SHARE` | Lock shared (baca saja), proses lain boleh juga baca |
| `FOR NO KEY UPDATE` | Lock tanpa kunci foreign key reference |

---

## 6. SKIP LOCKED — Job Queue Pattern

`SKIP LOCKED` sangat berguna untuk membangun **work queue** — multiple worker mengambil job tanpa saling tunggu.

```sql
CREATE TABLE jobs (
    id         SERIAL PRIMARY KEY,
    payload    JSONB NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',  -- pending | processing | done | failed
    attempts   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status ON jobs (status, created_at)
WHERE status IN ('pending', 'failed');
```

```go
type Job struct {
    ID      int
    Payload json.RawMessage
    Status  string
}

// Worker mengambil 1 job secara atomik (tidak bisa diambil worker lain)
func ClaimJob(ctx context.Context, pool *pgxpool.Pool) (*Job, error) {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback(ctx)

    job := &Job{}
    err = tx.QueryRow(ctx, `
        SELECT id, payload, status
        FROM jobs
        WHERE status = 'pending'
        ORDER BY created_at
        FOR UPDATE SKIP LOCKED
        LIMIT 1
    `).Scan(&job.ID, &job.Payload, &job.Status)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil // tidak ada job
        }
        return nil, err
    }

    // Tandai sebagai processing
    _, err = tx.Exec(ctx,
        `UPDATE jobs SET status='processing', attempts=attempts+1, updated_at=NOW()
         WHERE id = $1`, job.ID)
    if err != nil {
        return nil, err
    }

    return job, tx.Commit(ctx)
}

func CompleteJob(ctx context.Context, pool *pgxpool.Pool, id int, failed bool) error {
    status := "done"
    if failed {
        status = "failed"
    }
    _, err := pool.Exec(ctx,
        `UPDATE jobs SET status=$1, updated_at=NOW() WHERE id=$2`,
        status, id,
    )
    return err
}
```

---

## 7. Deadlock — Deteksi & Handling

**Deadlock** terjadi saat dua transaksi saling menunggu satu sama lain:
- T1 memegang lock A, menunggu lock B
- T2 memegang lock B, menunggu lock A

PostgreSQL mendeteksi deadlock secara otomatis dan mem-**abort salah satu transaksi** dengan error `40P01`.

```go
import "github.com/jackc/pgx/v5/pgconn"

func isDeadlock(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "40P01"
}

func isSerializationFailure(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "40001"
}

// Retry otomatis saat terjadi deadlock atau serialization failure
func withRetryTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
    const maxAttempts = 5
    for i := 0; i < maxAttempts; i++ {
        err := withTx(ctx, pool, fn)
        if err == nil {
            return nil
        }
        if isDeadlock(err) || isSerializationFailure(err) {
            jitter := time.Duration(rand.Intn(100)) * time.Millisecond
            log.Printf("⚠️  Retry %d/%d (deadlock/serialization): %v", i+1, maxAttempts, err)
            time.Sleep(jitter)
            continue
        }
        return err
    }
    return fmt.Errorf("gagal setelah %d retry", maxAttempts)
}
```

### Mencegah Deadlock: Consistent Lock Order

```go
// SALAH — bisa deadlock jika T1(A→B) dan T2(B→A) berjalan bersamaan:
// T1: lock account A, lalu lock account B
// T2: lock account B, lalu lock account A

// BENAR — selalu lock dengan urutan yang sama (misal: ID terkecil dulu):
func safeTransfer(ctx context.Context, tx pgx.Tx, fromID, toID string, amount float64) error {
    // Pastikan urutan lock konsisten
    first, second := fromID, toID
    if fromID > toID {
        first, second = toID, fromID
    }

    // Lock keduanya sekaligus dalam satu query
    rows, err := tx.Query(ctx,
        `SELECT id, balance FROM accounts WHERE id = ANY($1) ORDER BY id FOR UPDATE`,
        []string{first, second},
    )
    // ... lanjutkan proses
    _ = rows
    return err
}
```

---

## 8. Advisory Locks

Advisory locks adalah lock yang dikelola aplikasi (bukan oleh row/table PostgreSQL). Berguna untuk koordinasi proses di luar transaksi database.

```go
// Session-level advisory lock (bertahan sampai koneksi ditutup atau eksplisit unlock)
func acquireAdvisoryLock(ctx context.Context, conn *pgxpool.Conn, lockID int64) (bool, error) {
    var acquired bool
    err := conn.QueryRow(ctx,
        `SELECT pg_try_advisory_lock($1)`, lockID,
    ).Scan(&acquired)
    return acquired, err
}

func releaseAdvisoryLock(ctx context.Context, conn *pgxpool.Conn, lockID int64) error {
    _, err := conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, lockID)
    return err
}

// Transaction-level advisory lock (auto-release saat COMMIT/ROLLBACK)
func acquireAdvisoryLockTx(ctx context.Context, tx pgx.Tx, lockID int64) error {
    _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID)
    return err
}

// Contoh: pastikan hanya 1 proses yang menjalankan scheduled task
func runExclusiveTask(ctx context.Context, pool *pgxpool.Pool, taskID int64) error {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()

    acquired, err := acquireAdvisoryLock(ctx, conn, taskID)
    if err != nil {
        return err
    }
    if !acquired {
        log.Printf("Task %d sedang dijalankan oleh proses lain, skip", taskID)
        return nil
    }
    defer releaseAdvisoryLock(ctx, conn, taskID)

    // Jalankan task...
    log.Printf("Menjalankan exclusive task %d", taskID)
    return nil
}
```

---

## 9. Savepoint

Savepoint memungkinkan **partial rollback** dalam satu transaksi — rollback sebagian operasi tanpa membatalkan seluruh transaksi.

```go
func withSavepoint(ctx context.Context, tx pgx.Tx, name string, fn func() error) error {
    if _, err := tx.Exec(ctx, "SAVEPOINT "+name); err != nil {
        return err
    }
    if err := fn(); err != nil {
        // Rollback ke savepoint (tidak abort seluruh tx)
        tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+name)
        return err
    }
    _, err := tx.Exec(ctx, "RELEASE SAVEPOINT "+name)
    return err
}

// Contoh: import batch, skip baris yang error
func importBatch(ctx context.Context, pool *pgxpool.Pool, items []string) (int, int) {
    tx, _ := pool.Begin(ctx)
    defer tx.Rollback(ctx)

    success, failed := 0, 0
    for i, item := range items {
        spName := fmt.Sprintf("sp_%d", i)
        err := withSavepoint(ctx, tx, spName, func() error {
            _, err := tx.Exec(ctx,
                `INSERT INTO items (name) VALUES ($1)`, item)
            return err
        })
        if err != nil {
            failed++
        } else {
            success++
        }
    }
    tx.Commit(ctx)
    return success, failed
}
```

---

## 10. Monitoring & pg_stat_activity

```go
// Lihat transaksi yang sedang berjalan dan menunggu lock
const queryLockInfo = `
SELECT
    pid,
    state,
    wait_event_type,
    wait_event,
    query_start,
    LEFT(query, 80) AS query_snippet
FROM pg_stat_activity
WHERE state != 'idle'
  AND pid != pg_backend_pid()
ORDER BY query_start
`

type ActiveQuery struct {
    PID           int
    State         string
    WaitEventType *string
    WaitEvent     *string
    QueryStart    time.Time
    QuerySnippet  string
}
```

---

## 11. Ringkasan: Kapan Pakai Apa?

| Skenario | Solusi |
|---|---|
| Transfer saldo / debit-kredit | Pessimistic: `SELECT FOR UPDATE` |
| Update counter yang jarang konflik | Optimistic: column `version` |
| Multiple worker ambil job dari queue | `FOR UPDATE SKIP LOCKED` |
| Pastikan hanya 1 proses jalankan cron | Advisory Lock |
| Operasi kompleks, butuh partial rollback | Savepoint |
| Baca data konsisten sepanjang satu request | `REPEATABLE READ` |
| Transaksi yang harus serially equivalent | `SERIALIZABLE` |

---

## 12. Checklist

- [x] Isolation levels dan dampaknya (Read Committed, Repeatable Read, Serializable)
- [x] `BEGIN`/`COMMIT`/`ROLLBACK` dengan pgx + helper `withTx`
- [x] Optimistic Locking dengan kolom `version`
- [x] Pessimistic Locking dengan `SELECT FOR UPDATE`
- [x] `FOR UPDATE SKIP LOCKED` untuk job queue pattern
- [x] Deadlock detection + retry (`40P01`), serialization failure (`40001`)
- [x] Consistent lock ordering untuk mencegah deadlock
- [x] Advisory Locks (`pg_advisory_lock`, `pg_advisory_xact_lock`)
- [x] Savepoint untuk partial rollback dalam transaksi
- [x] `pg_stat_activity` untuk monitoring

---

## 13. Catatan

- **`defer tx.Rollback(ctx)`** aman dipanggil bahkan setelah `tx.Commit()` — pgx mengabaikannya.
- **`FOR UPDATE` tanpa NOWAIT** bisa menyebabkan request menunggu lama jika ada lock contention — selalu set `statement_timeout` atau gunakan context dengan deadline.
- **Serializable** memiliki overhead paling tinggi karena PostgreSQL harus track dependency antar transaksi (SSI). Gunakan hanya jika benar-benar perlu.
- **Advisory Locks** tidak di-release otomatis saat transaksi commit/rollback (kecuali yang `xact_`). Pastikan selalu direlease.
- **Deadlock prevention**: selalu lock multiple rows dalam urutan yang konsisten (sort by primary key).

---

## 14. Referensi

- https://www.postgresql.org/docs/current/transaction-iso.html — Transaction Isolation
- https://www.postgresql.org/docs/current/explicit-locking.html — Locking (FOR UPDATE, Advisory)
- https://www.postgresql.org/docs/current/errcodes-appendix.html — Error codes (40P01, 40001)
- https://pkg.go.dev/github.com/jackc/pgx/v5#TxOptions — pgx TxOptions
- https://www.postgresql.org/docs/current/monitoring-stats.html — pg_stat_activity

---

> ⏭️ **Selanjutnya:** `10-postgres-tooling.md` — PostgreSQL Migrations & Backup
