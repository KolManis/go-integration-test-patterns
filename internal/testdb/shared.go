package testdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/go-integration-test-patterns/internal/bakery"
)

// TxDatabase — отдельная БД подхода 4. В общей базе подхода 3 остаются закоммиченные заказы,
// а транзакционные тесты рассчитывают, что закоммиченных данных нет вовсе.
const TxDatabase = "bakery_tx"

// RunWithSharedPool — TestMain для подхода с уникальными данными: одно подключение к общей БД
// на весь пакет, без очистки. setup получает репозиторий и духовку на этом подключении.
func RunWithSharedPool(m *testing.M, setup func(*bakery.Repository, *bakery.Oven)) int {
	return run(m, DSN(), func(pool *pgxpool.Pool) {
		setup(bakery.NewRepository(pool), bakery.NewOven(pool, BakeTime))
	})
}

// RunWithTxPool — TestMain подхода 4: пул к БД TxDatabase (создаётся при первом запуске),
// тесты открывают на нём транзакции через Tx.
func RunWithTxPool(m *testing.M, setup func(*pgxpool.Pool)) int {
	dsn, err := ensureDatabase(TxDatabase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return run(m, dsn, setup)
}

func run(m *testing.M, dsn string, setup func(*pgxpool.Pool)) int {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
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
	setup(pool)
	return m.Run()
}

// ensureDatabase создаёт БД, если её нет, и возвращает её адрес. Пакеты, стартующие одновременно,
// могут попытаться создать её оба — ошибку «уже существует» пропускаем.
func ensureDatabase(name string) (string, error) {
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, DSN())
	if err != nil {
		return "", fmt.Errorf("connect admin: %w", err)
	}
	defer admin.Close()
	var pgErr *pgconn.PgError
	if _, err := admin.Exec(ctx, `CREATE DATABASE `+name); err != nil &&
		!(errors.As(err, &pgErr) && (pgErr.Code == "42P04" || pgErr.Code == "23505")) {
		return "", fmt.Errorf("create database %s: %w", name, err)
	}
	u, err := url.Parse(DSN())
	if err != nil {
		return "", fmt.Errorf("parse dsn: %w", err)
	}
	u.Path = "/" + name
	return u.String(), nil
}
