//go:build integration

// Подход 3: одна общая БД без очистки. Каждый тест создаёт свои данные и проверяет только их,
// поэтому тесты внутри пакета идут параллельно (t.Parallel), а пакеты — с любым -p.
package orders

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

var repo *bakery.Repository

// TestMain поднимает одно подключение на весь пакет — без TRUNCATE.
func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithSharedPool(m, func(r *bakery.Repository, _ *bakery.Oven) { repo = r }))
}

func TestCreateOrderIsQueued(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	created, err := repo.Create(ctx, "медовик")
	require.NoError(t, err)

	got, err := repo.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusQueued, got.Status)
	require.Nil(t, got.BakedAt)
}

func TestEmptyPastryRejected(t *testing.T) {
	t.Parallel()
	_, err := repo.Create(context.Background(), "")
	require.ErrorIs(t, err, bakery.ErrEmptyPastry)
}

func TestUnknownOrderNotFound(t *testing.T) {
	t.Parallel()
	_, err := repo.Get(context.Background(), uuid.New())
	require.ErrorIs(t, err, bakery.ErrOrderNotFound)
}

// Было (подход 1): «в очереди ровно 3 заказа» — CountByStatus по всей таблице.
// Так проверять нельзя: в общей базе лежат заказы других тестов. Проверяем каждый свой заказ.
func TestNewOrdersAreQueued(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, p := range []string{"медовик", "тирамису", "эклер"} {
		created, err := repo.Create(ctx, p)
		require.NoError(t, err)

		got, err := repo.Get(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, bakery.StatusQueued, got.Status, p)
	}
}
