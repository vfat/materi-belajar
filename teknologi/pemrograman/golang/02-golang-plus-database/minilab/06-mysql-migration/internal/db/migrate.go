package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator membungkus golang-migrate untuk MySQL.
type Migrator struct {
	migrate *migrate.Migrate
}

// NewMigrator membuat instance Migrator baru.
func NewMigrator(db *sql.DB, dbName string) (*Migrator, error) {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

// Up menjalankan semua pending migrations.
func (m *Migrator) Up() error {
		if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

// Down me-rollback satu migration terakhir.
func (m *Migrator) Down() error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down: %w", err)
	}
	return nil
}

// Step menjalankan N migrations. Positif = up, negatif = down.
func (m *Migrator) Step(n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration step %d: %w", n, err)
	}
	return nil
}

// Version mengembalikan version dan dirty state saat ini.
func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force mengubah version tanpa menjalankan migration (recovery dirty state).
func (m *Migrator) Force(version int) error {
	return m.migrate.Force(version)
}

// Drop menghapus semua tabel (hati-hati!).
func (m *Migrator) Drop() error {
	return m.migrate.Drop()
}

// RunMigrations adalah one-liner untuk auto-migrate di startup.
func RunMigrations(db *sql.DB, dbName string) error {
	migrator, err := NewMigrator(db, dbName)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := migrator.Up(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err == nil {
		if dirty {
			log.Printf("⚠️  Migration version %d is DIRTY\n", version)
		} else {
			log.Printf("✅ Migration at version %d\n", version)
		}
	}

	return nil
}
