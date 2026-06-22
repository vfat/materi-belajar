package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"example.com/minilab/02-golang-plus-database/06-mysql-migration/internal/db"
)

func main() {
	var (
		up      = flag.Bool("up", false, "Run all pending migrations")
		down    = flag.Bool("down", false, "Rollback one migration")
		step    = flag.Int("step", 0, "Step N migrations (+ up / - down)")
		version = flag.Bool("version", false, "Show current migration version")
		force   = flag.Int("force", 0, "Force set version (dirty recovery)")
		dbName  = flag.String("db", "", "Database name (overrides env)")
		env     = flag.String("env", "development", "Environment: development/staging/production")
	)
	flag.Parse()

	// Tentukan prefix environment
	var prefix string
	switch *env {
	case "production":
		prefix = "PROD_"
	case "staging":
		prefix = "STAGING_"
	default:
		prefix = "DEV_"
	}

	cfg := db.ConfigFromEnv(prefix)
	if *dbName != "" {
		cfg.DBName = *dbName
	}

	// Pastikan database ada
	if err := db.EnsureDatabase(cfg); err != nil {
		log.Fatalf("ensure database: %v", err)
	}

	// Koneksi ke database
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer database.Close()

	// Inisialisasi migrator
	migrator, err := db.NewMigrator(database, cfg.DBName)
	if err != nil {
		log.Fatalf("init migrator: %v", err)
	}

	switch {
	case *up:
		if err := migrator.Up(); err != nil {
			log.Fatalf("up: %v", err)
		}
		fmt.Println("✅ Migrations applied (up)")

	case *down:
		if err := migrator.Down(); err != nil {
			log.Fatalf("down: %v", err)
		}
		fmt.Println("✅ Last migration rolled back (down)")

	case *step != 0:
		if err := migrator.Step(*step); err != nil {
			log.Fatalf("step: %v", err)
		}
		fmt.Printf("✅ Stepped %d migrations\n", *step)

	case *force != 0:
		if err := migrator.Force(*force); err != nil {
			log.Fatalf("force: %v", err)
		}
		fmt.Printf("✅ Version forced to %d\n", *force)

	case *version:
		v, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("version: %v", err)
		}
		state := "clean"
		if dirty {
			state = "DIRTY"
		}
		fmt.Printf("📋 Version: %d (%s)\n", v, state)

	default:
		flag.Usage()
		os.Exit(1)
	}
}
