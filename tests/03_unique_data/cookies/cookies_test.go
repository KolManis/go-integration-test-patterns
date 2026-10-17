//go:build integration

// Подход 3: одна общая БД без очистки, тесты параллельные. Главная сложность — воркер:
// духовка печёт ВСЕ заказы в очереди, в том числе чужие, поэтому проверять «сколько обработал» нельзя.
package cookies

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

var (
	repo *bakery.Repository
	oven *bakery.Oven
)

func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithSharedPool(m, func(r *bakery.Repository, o *bakery.Oven) { repo, oven = r, o }))
}

// Было (подход 1): require.Equal(1, n) — «духовка взяла ровно наш заказ».
// Стало: n не проверяем (духовка могла испечь и чужие), проверяем статус своего заказа.
func TestBakesQueuedOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	order, err := repo.Create(ctx, "медовик")
	require.NoError(t, err)

	_, err = oven.BakeQueued(ctx)
	require.NoError(t, err)

	got, err := repo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusBaked, got.Status)
	require.NotNil(t, got.BakedAt)
}

func TestBurntPastry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	order, err := repo.Create(ctx, bakery.BurntPastry)
	require.NoError(t, err)

	_, err = oven.BakeQueued(ctx)
	require.NoError(t, err)

	got, err := repo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusBurnt, got.Status)
}

// Было (подход 1): «пустая очередь — духовка ничего не печёт». В общей базе очередь пустой не бывает:
// такой сценарий на этом подходе не проверить. Остаётся проверить, что проход без своих заказов не падает.
func TestOvenRunsWithoutOwnOrders(t *testing.T) {
	t.Parallel()
	_, err := oven.BakeQueued(context.Background())
	require.NoError(t, err)
}

// Было (подход 1): CountByStatus(baked) == 2 по всей таблице. Стало: статус каждого своего заказа.
func TestBakesWholeQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := map[string]bakery.Status{
		"медовик":          bakery.StatusBaked,
		"тирамису":         bakery.StatusBaked,
		bakery.BurntPastry: bakery.StatusBurnt,
	}
	ids := map[string]bakery.Order{}
	for p := range want {
		o, err := repo.Create(ctx, p)
		require.NoError(t, err)
		ids[p] = o
	}

	_, err := oven.BakeQueued(ctx)
	require.NoError(t, err)

	for p, o := range ids {
		got, err := repo.Get(ctx, o.ID)
		require.NoError(t, err)
		require.Equal(t, want[p], got.Status, p)
	}
}
