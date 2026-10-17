package testdb

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tx открывает транзакцию на тест и откатывает её в конце: всё, что тест записал, исчезает.
// Код под тестом должен работать через эту транзакцию — другое подключение её данных не увидит.
func Tx(tb testing.TB, pool *pgxpool.Pool) pgx.Tx {
	tb.Helper()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		tb.Fatalf("begin: %v", err)
	}
	tb.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return tx
}
