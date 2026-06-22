package db

import (
	"fmt"
	"os"
	"strconv"
)

// Config menyimpan konfigurasi koneksi MySQL.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Params   map[string]string
	MaxOpen  int
	MaxIdle  int
}

// ConfigFromEnv membaca konfigurasi dari environment variables.
// prefix digunakan untuk membedakan environment: DEV_, STAGING_, PROD_.
func ConfigFromEnv(prefix string) Config {
	cfg := Config{
		Host:   getEnv(prefix+"DB_HOST", "localhost"),
		Port:   getEnvInt(prefix+"DB_PORT", 3306),
		User:   getEnv(prefix+"DB_USER", "demo"),
		DBName: getEnv(prefix+"DB_NAME", "golang_demo"),
		Params: map[string]string{
			"charset":   "utf8mb4",
			"parseTime": "true",
			"loc":       "Asia/Jakarta",
		},
		MaxOpen: getEnvInt(prefix+"DB_MAX_OPEN", 25),
		MaxIdle: getEnvInt(prefix+"DB_MAX_IDLE", 5),
	}

	// Password dari file (docker secret) lebih aman
	if pwdFile := os.Getenv(prefix + "DB_PASSWORD_FILE"); pwdFile != "" {
		if data, err := os.ReadFile(pwdFile); err == nil {
			cfg.Password = string(data)
		}
	} else {
		cfg.Password = getEnv(prefix+"DB_PASSWORD", "demo")
	}

	return cfg
}

// DSN menghasilkan Data Source Name untuk koneksi MySQL.
func (c Config) DSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?",
		c.User, c.Password, c.Host, c.Port, c.DBName)
	for k, v := range c.Params {
		dsn += fmt.Sprintf("%s=%s&", k, v)
	}
	return dsn[:len(dsn)-1]
}

// MigrationDSN menghasilkan DSN tanpa database (untuk create database).
func (c Config) MigrationDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?",
		c.User, c.Password, c.Host, c.Port)
	for k, v := range c.Params {
		dsn += fmt.Sprintf("%s=%s&", k, v)
	}
	return dsn[:len(dsn)-1]
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
