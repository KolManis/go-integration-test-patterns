// Package testdb — общие помощники интеграционных тестов: подключение, миграции, очистка.
package testdb

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BakeTime — время выпекания в тестах: имитирует ожидание внешней системы.
const BakeTime = 300 * time.Millisecond

// DSN — адрес тестовой БД. По умолчанию — Postgres из docker-compose.yml.
func DSN() string {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://bakery:bakery@localhost:5433/bakery?sslmode=disable"
}

// Connect открывает пул к БД по адресу и закрывает его в конце теста (или сьюта).
func Connect(tb testing.TB, dsn string) *pgxpool.Pool {
	tb.Helper()
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		tb.Fatalf("connect to %s: %v", dsn, err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		tb.Fatalf("ping %s: %v", dsn, err)
	}
	tb.Cleanup(pool.Close)
	return pool
}

// Migrate применяет схему. Идемпотентно: CREATE ... IF NOT EXISTS.
func Migrate(tb testing.TB, pool *pgxpool.Pool) {
	tb.Helper()
	if _, err := pool.Exec(context.Background(), migrationSQL(tb)); err != nil {
		tb.Fatalf("apply migration: %v", err)
	}
}

func migrationSQL(tb testing.TB) string {
	tb.Helper()
	sql, err := os.ReadFile(filepath.Join(repoRoot(), "migrations", "001_init.sql"))
	if err != nil {
		tb.Fatalf("read migration: %v", err)
	}
	return string(sql)
}

// Truncate стирает все данные — так делают тесты на общей БД перед и после каждого теста.
func Truncate(tb testing.TB, pool *pgxpool.Pool) {
	tb.Helper()
	if _, err := pool.Exec(context.Background(), `TRUNCATE orders`); err != nil {
		tb.Fatalf("truncate: %v", err)
	}
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}
