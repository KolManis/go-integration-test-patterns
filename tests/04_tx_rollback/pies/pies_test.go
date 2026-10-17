//go:build integration

// Подход 4: духовка получает транзакцию теста и видит только заказы этого теста,
// поэтому проверки «ровно 1 заказ» и «пустая очередь» снова работают — при параллельном запуске.
package pies

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithTxPool(m, func(p *pgxpool.Pool) { pool = p }))
}

func setup(t *testing.T) (*bakery.Repository, *bakery.Oven) {
	tx := testdb.Tx(t, pool)
	return bakery.NewRepository(tx), bakery.NewOven(tx, testdb.BakeTime)
}

func TestBakesQueuedOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, oven := setup(t)
	order, err := repo.Create(ctx, "медовик")
	require.NoError(t, err)

	n, err := oven.BakeQueued(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)

	got, err := repo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusBaked, got.Status)
}

func TestBurntPastry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, oven := setup(t)
	order, err := repo.Create(ctx, bakery.BurntPastry)
	require.NoError(t, err)

	_, err = oven.BakeQueued(ctx)
	require.NoError(t, err)

	got, err := repo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusBurnt, got.Status)
}

func TestEmptyQueue(t *testing.T) {
	t.Parallel()
	_, oven := setup(t)
	n, err := oven.BakeQueued(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 0, n)
}

func TestBakesWholeQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, oven := setup(t)
	for _, p := range []string{"медовик", "тирамису", bakery.BurntPastry} {
		_, err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	_, err := oven.BakeQueued(ctx)
	require.NoError(t, err)

	baked, err := repo.CountByStatus(ctx, bakery.StatusBaked)
	require.NoError(t, err)
	require.Equal(t, 2, baked)
}

// Ограничение: в проде духовка работает со своим пулом и коммитит. Если дать ей пул, а не транзакцию
// теста, она заказ не увидит — проверяем ровно это, чтобы ограничение было видно в отчёте.
func TestLimitationOvenOnOwnPool(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := bakery.NewRepository(testdb.Tx(t, pool))
	order, err := repo.Create(ctx, "медовик")
	require.NoError(t, err)

	_, err = bakery.NewOven(pool, 0).BakeQueued(ctx)
	require.NoError(t, err)

	got, err := repo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusQueued, got.Status, "духовка на своём пуле заказ теста не видит")
}
