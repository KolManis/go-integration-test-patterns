package testdb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/go-integration-test-patterns/internal/bakery"
)

// RunWithSharedPool — TestMain для подхода с уникальными данными: одно подключение к общей БД
// на весь пакет, без очистки. setup получает репозиторий и духовку на этом подключении.
func RunWithSharedPool(m *testing.M, setup func(*bakery.Repository, *bakery.Oven)) int {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		return 1
	}
	defer pool.Close()
	sql, err := os.ReadFile(migrationPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read migration: %v\n", err)
		return 1
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		return 1
	}
	setup(bakery.NewRepository(pool), bakery.NewOven(pool, BakeTime))
	return m.Run()
}
