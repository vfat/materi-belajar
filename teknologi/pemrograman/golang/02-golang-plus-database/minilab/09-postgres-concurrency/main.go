package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	poolCfg.MaxConns = 20
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
	return nil, fmt.Errorf("postgres tidak tersedia: %w", lastErr)
}

// === Setup Schema ===

func setupSchema(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			owner      TEXT NOT NULL,
			balance    NUMERIC(15,2) NOT NULL DEFAULT 0,
			CHECK (balance >= 0)
		)`,
		`CREATE TABLE IF NOT EXISTS transfers (
			id         SERIAL PRIMARY KEY,
			from_id    UUID NOT NULL,
			to_id      UUID NOT NULL,
			amount     NUMERIC(15,2) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS products_v (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name       TEXT NOT NULL,
			stock      INT NOT NULL DEFAULT 0,
			version    INT NOT NULL DEFAULT 1,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CHECK (stock >= 0)
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id         SERIAL PRIMARY KEY,
			payload    JSONB NOT NULL,
			status     TEXT NOT NULL DEFAULT 'pending',
			attempts   INT NOT NULL DEFAULT 0,
			result     TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status, created_at)
		 WHERE status IN ('pending', 'failed')`,
	}
	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("setup: %w", err)
		}
	}
	return nil
}

// === Error Helpers ===

var ErrConflict = errors.New("optimistic conflict")
var ErrInsufficientBalance = errors.New("saldo tidak cukup")

func isDeadlock(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40P01"
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40001"
}

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// === Transaction Helpers ===

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

func withRetryTx(ctx context.Context, pool *pgxpool.Pool, maxAttempts int, fn func(pgx.Tx) error) error {
	for i := 0; i < maxAttempts; i++ {
		err := withTx(ctx, pool, fn)
		if err == nil {
			return nil
		}
		if isDeadlock(err) || isSerializationFailure(err) {
			jitter := time.Duration(rand.Intn(80)+20) * time.Millisecond
			log.Printf("   ⚠️  Retry %d/%d (code=%s, sleep=%v)", i+1, maxAttempts, pgErrCode(err), jitter)
			time.Sleep(jitter)
			continue
		}
		return err
	}
	return fmt.Errorf("gagal setelah %d retry", maxAttempts)
}

// === Accounts ===

type Account struct {
	ID      string
	Owner   string
	Balance float64
}

func createAccount(ctx context.Context, pool *pgxpool.Pool, owner string, balance float64) (*Account, error) {
	a := &Account{}
	err := pool.QueryRow(ctx,
		`INSERT INTO accounts (owner, balance) VALUES ($1, $2) RETURNING id, owner, balance`,
		owner, balance,
	).Scan(&a.ID, &a.Owner, &a.Balance)
	return a, err
}

func getBalance(ctx context.Context, pool *pgxpool.Pool, id string) (float64, error) {
	var balance float64
	err := pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, id).Scan(&balance)
	return balance, err
}

func transferBalance(ctx context.Context, pool *pgxpool.Pool, fromID, toID string, amount float64) error {
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		// Lock keduanya dengan urutan konsisten (cegah deadlock)
		ids := []string{fromID, toID}
		if fromID > toID {
			ids = []string{toID, fromID}
		}
		rows, err := tx.Query(ctx,
			`SELECT id, balance FROM accounts WHERE id = ANY($1) ORDER BY id FOR UPDATE`,
			ids,
		)
		if err != nil {
			return err
		}
		balances := map[string]float64{}
		for rows.Next() {
			var id string
			var bal float64
			rows.Scan(&id, &bal)
			balances[id] = bal
		}
		rows.Close()

		if balances[fromID]-amount < 0 {
			return ErrInsufficientBalance
		}

		if _, err := tx.Exec(ctx,
			`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID); err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO transfers (from_id, to_id, amount) VALUES ($1, $2, $3)`,
			fromID, toID, amount)
		return err
	})
}

// === Optimistic Locking ===

type ProductV struct {
	ID      string
	Name    string
	Stock   int
	Version int
}

func createProductV(ctx context.Context, pool *pgxpool.Pool, name string, stock int) (*ProductV, error) {
	p := &ProductV{}
	err := pool.QueryRow(ctx,
		`INSERT INTO products_v (name, stock) VALUES ($1, $2) RETURNING id, name, stock, version`,
		name, stock,
	).Scan(&p.ID, &p.Name, &p.Stock, &p.Version)
	return p, err
}

func getProductV(ctx context.Context, pool *pgxpool.Pool, id string) (*ProductV, error) {
	p := &ProductV{}
	err := pool.QueryRow(ctx,
		`SELECT id, name, stock, version FROM products_v WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Stock, &p.Version)
	return p, err
}

func updateStockOptimistic(ctx context.Context, pool *pgxpool.Pool, id string, delta, knownVersion int) error {
	res, err := pool.Exec(ctx, `
		UPDATE products_v
		SET stock = stock + $1,
		    version = version + 1,
		    updated_at = NOW()
		WHERE id = $2 AND version = $3
	`, delta, id, knownVersion)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrConflict
	}
	return nil
}

func updateStockWithRetry(ctx context.Context, pool *pgxpool.Pool, id string, delta, maxRetry int) (int, error) {
	for i := 0; i < maxRetry; i++ {
		p, err := getProductV(ctx, pool, id)
		if err != nil {
			return i, err
		}
		err = updateStockOptimistic(ctx, pool, id, delta, p.Version)
		if err == nil {
			return i + 1, nil
		}
		if !errors.Is(err, ErrConflict) {
			return i, err
		}
		time.Sleep(time.Duration(rand.Intn(30)+10) * time.Millisecond)
	}
	return maxRetry, fmt.Errorf("optimistic: gagal setelah %d retry", maxRetry)
}

// === Pessimistic Locking ===

func updateStockPessimistic(ctx context.Context, pool *pgxpool.Pool, id string, delta int) error {
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		var stock int
		err := tx.QueryRow(ctx,
			`SELECT stock FROM products_v WHERE id = $1 FOR UPDATE`, id,
		).Scan(&stock)
		if err != nil {
			return err
		}
		if stock+delta < 0 {
			return fmt.Errorf("stok tidak cukup (stock=%d, delta=%d)", stock, delta)
		}
		_, err = tx.Exec(ctx,
			`UPDATE products_v SET stock = stock + $1, updated_at = NOW() WHERE id = $2`,
			delta, id)
		return err
	})
}

// === Job Queue (SKIP LOCKED) ===

type Job struct {
	ID      int
	Payload json.RawMessage
	Status  string
}

func enqueueJobs(ctx context.Context, pool *pgxpool.Pool, payloads []map[string]any) error {
	for _, p := range payloads {
		b, _ := json.Marshal(p)
		if _, err := pool.Exec(ctx,
			`INSERT INTO jobs (payload) VALUES ($1)`, b); err != nil {
			return err
		}
	}
	return nil
}

func claimJob(ctx context.Context, pool *pgxpool.Pool) (*Job, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	job := &Job{}
	err = tx.QueryRow(ctx, `
		SELECT id, payload, status FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`).Scan(&job.ID, &job.Payload, &job.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`UPDATE jobs SET status='processing', attempts=attempts+1, updated_at=NOW() WHERE id=$1`,
		job.ID)
	if err != nil {
		return nil, err
	}
	return job, tx.Commit(ctx)
}

func completeJob(ctx context.Context, pool *pgxpool.Pool, id int, result string, failed bool) error {
	status := "done"
	if failed {
		status = "failed"
	}
	_, err := pool.Exec(ctx,
		`UPDATE jobs SET status=$1, result=$2, updated_at=NOW() WHERE id=$3`,
		status, result, id)
	return err
}

func countJobsByStatus(ctx context.Context, pool *pgxpool.Pool) (map[string]int, error) {
	rows, err := pool.Query(ctx, `SELECT status, COUNT(*) FROM jobs GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		counts[status] = count
	}
	return counts, rows.Err()
}

// === Advisory Locks ===

func tryAdvisoryLock(ctx context.Context, conn *pgxpool.Conn, lockID int64) (bool, error) {
	var acquired bool
	err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, lockID).Scan(&acquired)
	return acquired, err
}

func releaseAdvisoryLock(ctx context.Context, conn *pgxpool.Conn, lockID int64) {
	conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, lockID)
}

// === Savepoint ===

func withSavepoint(ctx context.Context, tx pgx.Tx, name string, fn func() error) error {
	if _, err := tx.Exec(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}
	if err := fn(); err != nil {
		tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+name)
		return err
	}
	_, err := tx.Exec(ctx, "RELEASE SAVEPOINT "+name)
	return err
}

// === Utilities ===

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
	fmt.Println("=== MANUAL TEST MATERI 09: POSTGRESQL CONCURRENCY & LOCKING ===")
	fmt.Println()

	ctx := context.Background()
	cfg := DefaultConfig()

	fmt.Println("⏳ Menunggu PostgreSQL siap...")
	fmt.Println("   Pastikan: cd 09-postgres-concurrency && docker compose up -d")
	fmt.Println()

	pool, err := WaitForPostgres(ctx, cfg, 10, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Gagal connect: %v", err)
	}
	defer func() {
		pool.Close()
		fmt.Println("\n✅ Pool ditutup")
	}()
	fmt.Println("✅ Berhasil konek ke PostgreSQL!")
	fmt.Println()

	if err := setupSchema(ctx, pool); err != nil {
		log.Fatalf("❌ Setup schema: %v", err)
	}
	fmt.Println("✅ Schema siap\n")

	// ============================================
	// Test 1: Transaksi Dasar — Transfer Balance
	// ============================================
	fmt.Println("--- Test 1: Transaksi Dasar — Transfer Balance ---")

	alice, _ := createAccount(ctx, pool, "Alice", 1000000)
	bob, _ := createAccount(ctx, pool, "Bob", 500000)
	fmt.Printf("✅ Akun dibuat: Alice=%.0f, Bob=%.0f\n", alice.Balance, bob.Balance)

	// Transfer normal
	if err := transferBalance(ctx, pool, alice.ID, bob.ID, 250000); err != nil {
		log.Fatalf("❌ Transfer: %v", err)
	}
	aliceBal, _ := getBalance(ctx, pool, alice.ID)
	bobBal, _ := getBalance(ctx, pool, bob.ID)
	fmt.Printf("✅ Transfer 250.000: Alice=%.0f → %.0f, Bob=%.0f → %.0f\n",
		alice.Balance, aliceBal, bob.Balance, bobBal)

	// Transfer melebihi saldo (harusnya gagal karena CHECK constraint)
	err = transferBalance(ctx, pool, alice.ID, bob.ID, 2000000)
	if err != nil {
		fmt.Printf("✅ Transfer melebihi saldo ditolak: %v\n", err)
	}

	aliceBal2, _ := getBalance(ctx, pool, alice.ID)
	fmt.Printf("✅ Saldo Alice tidak berubah setelah gagal: %.0f\n", aliceBal2)

	// ============================================
	// Test 2: Concurrent Transfers (Stress Test)
	// ============================================
	fmt.Println("\n--- Test 2: Concurrent Transfers (10 goroutine) ---")

	carol, _ := createAccount(ctx, pool, "Carol", 5000000)
	dave, _ := createAccount(ctx, pool, "Dave", 5000000)
	fmt.Printf("✅ Akun: Carol=%.0f, Dave=%.0f\n", carol.Balance, dave.Balance)

	var wg sync.WaitGroup
	successCount, failCount := 0, 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Goroutine ganjil: Carol→Dave, genap: Dave→Carol
			var err error
			if i%2 == 0 {
				err = withRetryTx(ctx, pool, 5, func(tx pgx.Tx) error {
					return transferInTx(ctx, tx, carol.ID, dave.ID, 100000)
				})
			} else {
				err = withRetryTx(ctx, pool, 5, func(tx pgx.Tx) error {
					return transferInTx(ctx, tx, dave.ID, carol.ID, 100000)
				})
			}
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failCount++
				log.Printf("   goroutine %d error: %v", i, err)
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	carolBal, _ := getBalance(ctx, pool, carol.ID)
	daveBal, _ := getBalance(ctx, pool, dave.ID)
	fmt.Printf("✅ 10 concurrent transfers: sukses=%d, gagal=%d\n", successCount, failCount)
	fmt.Printf("✅ Saldo akhir: Carol=%.0f, Dave=%.0f (total=%.0f, expected=10.000.000)\n",
		carolBal, daveBal, carolBal+daveBal)

	// ============================================
	// Test 3: Optimistic Locking
	// ============================================
	fmt.Println("\n--- Test 3: Optimistic Locking (version column) ---")

	laptop, _ := createProductV(ctx, pool, "Laptop Pro", 100)
	fmt.Printf("✅ Produk dibuat: %s stock=%d version=%d\n", laptop.Name, laptop.Stock, laptop.Version)

	// Simulasi conflict: baca versi yang sama dari 2 goroutine, lalu update
	p1, _ := getProductV(ctx, pool, laptop.ID)
	p2, _ := getProductV(ctx, pool, laptop.ID)

	// G1 update duluan (sukses)
	err1 := updateStockOptimistic(ctx, pool, laptop.ID, -5, p1.Version)
	// G2 coba update dengan versi lama (conflict!)
	err2 := updateStockOptimistic(ctx, pool, laptop.ID, -3, p2.Version)

	fmt.Printf("✅ Goroutine 1 update (version=%d): err=%v\n", p1.Version, err1)
	fmt.Printf("✅ Goroutine 2 update (version lama=%d): err=%v\n", p2.Version, err2)

	if errors.Is(err2, ErrConflict) {
		fmt.Println("✅ ErrConflict terdeteksi dengan benar!")
	}

	// Retry otomatis
	mouse, _ := createProductV(ctx, pool, "Mouse Wireless", 50)
	fmt.Printf("\n✅ Test retry optimistic (5 goroutine concurrent -1 stock):\n")

	var wg2 sync.WaitGroup
	retryTotal := 0
	var mu2 sync.Mutex
	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			attempts, err := updateStockWithRetry(ctx, pool, mouse.ID, -1, 10)
			mu2.Lock()
			retryTotal += attempts
			mu2.Unlock()
			if err != nil {
				log.Printf("   ❌ update failed: %v", err)
			}
		}()
	}
	wg2.Wait()

	final, _ := getProductV(ctx, pool, mouse.ID)
	fmt.Printf("   Stock: 50 → %d (expected: 45), total retry attempts: %d\n",
		final.Stock, retryTotal)

	// ============================================
	// Test 4: Pessimistic Locking (SELECT FOR UPDATE)
	// ============================================
	fmt.Println("\n--- Test 4: Pessimistic Locking (SELECT FOR UPDATE) ---")

	keyboard, _ := createProductV(ctx, pool, "Keyboard Mech", 30)
	fmt.Printf("✅ Produk: %s stock=%d\n", keyboard.Name, keyboard.Stock)

	var wg3 sync.WaitGroup
	pessimisticErrors := 0
	var mu3 sync.Mutex
	for i := 0; i < 5; i++ {
		wg3.Add(1)
		go func(i int) {
			defer wg3.Done()
			if err := updateStockPessimistic(ctx, pool, keyboard.ID, -2); err != nil {
				mu3.Lock()
				pessimisticErrors++
				mu3.Unlock()
			}
		}(i)
	}
	wg3.Wait()

	kbFinal, _ := getProductV(ctx, pool, keyboard.ID)
	fmt.Printf("✅ 5 concurrent -2 via FOR UPDATE: stock %d → %d, errors=%d\n",
		keyboard.Stock, kbFinal.Stock, pessimisticErrors)

	// ============================================
	// Test 5: Job Queue dengan SKIP LOCKED
	// ============================================
	fmt.Println("\n--- Test 5: Job Queue (FOR UPDATE SKIP LOCKED) ---")

	// Enqueue 10 jobs
	payloads := make([]map[string]any, 10)
	for i := range payloads {
		payloads[i] = map[string]any{"task": "send_email", "to": fmt.Sprintf("user%d@example.com", i+1)}
	}
	enqueueJobs(ctx, pool, payloads)

	counts, _ := countJobsByStatus(ctx, pool)
	fmt.Printf("✅ Enqueued %d jobs (status: %v)\n", counts["pending"], counts)

	// Jalankan 3 worker concurrent
	var wg4 sync.WaitGroup
	processed := 0
	var mu4 sync.Mutex
	for w := 0; w < 3; w++ {
		wg4.Add(1)
		go func(workerID int) {
			defer wg4.Done()
			for {
				job, err := claimJob(ctx, pool)
				if err != nil {
					log.Printf("worker %d claim error: %v", workerID, err)
					return
				}
				if job == nil {
					return // tidak ada job tersisa
				}
				// Simulasi kerja
				time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				result := fmt.Sprintf("processed by worker %d", workerID)
				completeJob(ctx, pool, job.ID, result, false)

				mu4.Lock()
				processed++
				mu4.Unlock()
			}
		}(w)
	}
	wg4.Wait()

	finalCounts, _ := countJobsByStatus(ctx, pool)
	fmt.Printf("✅ 3 concurrent workers memproses jobs:\n")
	fmt.Printf("   done=%d, processing=%d, pending=%d\n",
		finalCounts["done"], finalCounts["processing"], finalCounts["pending"])

	// ============================================
	// Test 6: Advisory Locks
	// ============================================
	fmt.Println("\n--- Test 6: Advisory Locks ---")

	const taskLockID = int64(12345678)

	conn1, _ := pool.Acquire(ctx)
	conn2, _ := pool.Acquire(ctx)

	acq1, _ := tryAdvisoryLock(ctx, conn1, taskLockID)
	acq2, _ := tryAdvisoryLock(ctx, conn2, taskLockID)

	fmt.Printf("✅ Koneksi 1 acquire lock %d: %v\n", taskLockID, acq1)
	fmt.Printf("✅ Koneksi 2 acquire lock %d (harus false): %v\n", taskLockID, acq2)

	if acq1 && !acq2 {
		fmt.Println("✅ Advisory lock bekerja: hanya 1 proses bisa dapat lock")
	}

	releaseAdvisoryLock(ctx, conn1, taskLockID)
	acq3, _ := tryAdvisoryLock(ctx, conn2, taskLockID)
	fmt.Printf("✅ Setelah release, koneksi 2 acquire: %v\n", acq3)
	releaseAdvisoryLock(ctx, conn2, taskLockID)

	conn1.Release()
	conn2.Release()

	// ============================================
	// Test 7: Savepoint — Partial Rollback
	// ============================================
	fmt.Println("\n--- Test 7: Savepoint — Partial Rollback ---")

	// Insert batch dengan duplikat — yang duplikat di-skip via savepoint
	names := []string{"Item-A", "Item-B", "Item-A", "Item-C", "Item-B"} // duplikat intentional

	// Buat tabel items jika belum ada
	pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	)`)

	tx, _ := pool.Begin(ctx)
	succeedSp, failedSp := 0, 0
	for i, name := range names {
		spName := fmt.Sprintf("sp_%d", i)
		err := withSavepoint(ctx, tx, spName, func() error {
			_, err := tx.Exec(ctx, `INSERT INTO items (name) VALUES ($1)`, name)
			return err
		})
		if err != nil {
			failedSp++
			fmt.Printf("   SAVEPOINT: '%s' gagal (duplikat) → rollback ke sp_%d\n", name, i)
		} else {
			succeedSp++
		}
	}
	tx.Commit(ctx)

	var itemCount int
	pool.QueryRow(ctx, `SELECT COUNT(*) FROM items`).Scan(&itemCount)
	fmt.Printf("✅ Savepoint: %d sukses, %d gagal-di-skip, total items=%d\n",
		succeedSp, failedSp, itemCount)

	// ============================================
	// Test 8: Isolation Level — Read Committed vs Repeatable Read
	// ============================================
	fmt.Println("\n--- Test 8: Isolation Level Demo ---")

	charlie2, _ := createAccount(ctx, pool, "Charlie", 1000000)

	// Baca dengan READ COMMITTED (default)
	var balRC float64
	txRC, _ := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	txRC.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, charlie2.ID).Scan(&balRC)

	// Update balance dari luar transaksi
	pool.Exec(ctx, `UPDATE accounts SET balance = balance + 100000 WHERE id = $1`, charlie2.ID)

	var balRC2 float64
	txRC.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, charlie2.ID).Scan(&balRC2)
	txRC.Commit(ctx)

	fmt.Printf("✅ READ COMMITTED: baca pertama=%.0f, baca kedua (setelah update luar)=%.0f\n",
		balRC, balRC2)
	fmt.Println("   💡 Non-repeatable read: nilai berubah dalam 1 transaksi (normal untuk RC)")

	// Reset
	pool.Exec(ctx, `UPDATE accounts SET balance = 1000000 WHERE id = $1`, charlie2.ID)

	// Baca dengan REPEATABLE READ
	var balRR float64
	txRR, _ := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	txRR.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, charlie2.ID).Scan(&balRR)

	pool.Exec(ctx, `UPDATE accounts SET balance = balance + 100000 WHERE id = $1`, charlie2.ID)

	var balRR2 float64
	txRR.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, charlie2.ID).Scan(&balRR2)
	txRR.Commit(ctx)

	fmt.Printf("✅ REPEATABLE READ: baca pertama=%.0f, baca kedua (setelah update luar)=%.0f\n",
		balRR, balRR2)
	fmt.Println("   💡 Repeatable read: nilai konsisten dalam 1 transaksi (snapshot isolation)")

	// ============================================
	// Selesai
	// ============================================
	fmt.Println()
	fmt.Println("=== SEMUA TEST SELESAI ===")
	fmt.Println("🎉 Concurrency: transaksi, optimistic lock, pessimistic lock,")
	fmt.Println("   SKIP LOCKED job queue, advisory lock, savepoint, isolation levels — dipahami!")
}

// transferInTx melakukan transfer dalam transaksi yang sudah ada (untuk withRetryTx)
func transferInTx(ctx context.Context, tx pgx.Tx, fromID, toID string, amount float64) error {
	ids := []string{fromID, toID}
	if fromID > toID {
		ids = []string{toID, fromID}
	}
	rows, err := tx.Query(ctx,
		`SELECT id, balance FROM accounts WHERE id = ANY($1) ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return err
	}
	balances := map[string]float64{}
	for rows.Next() {
		var id string
		var bal float64
		rows.Scan(&id, &bal)
		balances[id] = bal
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	if balances[fromID]-amount < 0 {
		return ErrInsufficientBalance
	}
	if _, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
	return err
}
