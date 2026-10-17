//go:build integration

// Подход 4: одна общая БД, у каждого теста своя транзакция, в конце ROLLBACK.
// Чужие незакоммиченные строки не видны, поэтому глобальные проверки снова работают и тесты идут параллельно.
package orders

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/example/go-integration-test-patterns/internal/bakery"
	"github.com/example/go-integration-test-patterns/internal/testdb"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithTxPool(m, func(p *pgxpool.Pool) { pool = p }))
}

func TestCreateOrderIsQueued(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := bakery.NewRepository(testdb.Tx(t, pool))

	created, err := repo.Create(ctx, "медовик")
	require.NoError(t, err)

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusQueued, got.Status)
}

func TestEmptyPastryRejected(t *testing.T) {
	t.Parallel()
	repo := bakery.NewRepository(testdb.Tx(t, pool))
	_, err := repo.Create(context.Background(), "")
	require.ErrorIs(t, err, bakery.ErrEmptyPastry)
}

func TestUnknownOrderNotFound(t *testing.T) {
	t.Parallel()
	repo := bakery.NewRepository(testdb.Tx(t, pool))
	_, err := repo.Get(context.Background(), uuid.New())
	require.ErrorIs(t, err, bakery.ErrOrderNotFound)
}

// Как в подходе 1: глобальная проверка по таблице. Работает, потому что в общей БД ничего не остаётся
// (все тесты откатываются), а чужие незакоммиченные заказы транзакции не видно.
func TestQueueCount(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := bakery.NewRepository(testdb.Tx(t, pool))
	for _, p := range []string{"медовик", "тирамису", "эклер"} {
		_, err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	n, err := repo.CountByStatus(ctx, bakery.StatusQueued)
	require.NoError(t, err)
	require.Equal(t, 3, n)
}

// Ограничение подхода: заказ, созданный в транзакции теста, не виден через другое подключение.
// Так же его не увидит HTTP-сервис, консьюмер или воркер, работающие со своим пулом.
func TestLimitationOtherConnectionDoesNotSeeData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	created, err := bakery.NewRepository(testdb.Tx(t, pool)).Create(ctx, "медовик")
	require.NoError(t, err)

	_, err = bakery.NewRepository(pool).Get(ctx, created.ID)
	require.ErrorIs(t, err, bakery.ErrOrderNotFound, "данные транзакции не видны снаружи")
}
