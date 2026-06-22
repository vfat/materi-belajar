package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Connect membuka koneksi ke MySQL dengan Config yang diberikan.
func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// EnsureDatabase membuat database jika belum ada.
func EnsureDatabase(cfg Config) error {
	db, err := sql.Open("mysql", cfg.MigrationDSN())
	if err != nil {
		return fmt.Errorf("open mysql (no db): %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	))
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	return nil
}

// WaitForMySQL retry koneksi sampai MySQL siap.
func WaitForMySQL(cfg Config, attempts int, delay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var lastErr error

	for i := 0; i < attempts; i++ {
		db, lastErr = Connect(cfg)
		if lastErr == nil {
			return db, nil
		}
		fmt.Printf("⏳ Waiting for MySQL (attempt %d/%d): %v\n", i+1, attempts, lastErr)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("mysql not ready after %d attempts: %w", attempts, lastErr)
}
