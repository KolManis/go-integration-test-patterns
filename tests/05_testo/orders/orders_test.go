//go:build integration

// Подход 5: тесты заказов из подхода 4 на Testo.
package orders

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ozontech/testo"
	"github.com/stretchr/testify/require"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

type T = *testo.T

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(testdb.RunWithTxPool(m, func(p *pgxpool.Pool) { pool = p }))
}

type Suite struct{ testo.Suite[T] }

func TestOrders(t *testing.T) {
	testo.RunSuite(t, new(Suite))
}

func (*Suite) BeforeEach(t T) {
	t.Parallel()
}

func repo(t T) *bakery.Repository {
	return bakery.NewRepository(testdb.Tx(t, pool))
}

func (*Suite) TestCreateOrderIsQueued(t T) {
	r := repo(t)
	created, err := r.Create(t.Context(), "медовик")
	require.NoError(t, err)

	got, err := r.Get(t.Context(), created.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusQueued, got.Status)
}

func (*Suite) TestEmptyPastryRejected(t T) {
	_, err := repo(t).Create(t.Context(), "")
	require.ErrorIs(t, err, bakery.ErrEmptyPastry)
}

func (*Suite) TestUnknownOrderNotFound(t T) {
	_, err := repo(t).Get(t.Context(), uuid.New())
	require.ErrorIs(t, err, bakery.ErrOrderNotFound)
}

func (*Suite) TestQueueCount(t T) {
	r := repo(t)
	for _, p := range []string{"медовик", "тирамису", "эклер"} {
		_, err := r.Create(t.Context(), p)
		require.NoError(t, err)
	}
	n, err := r.CountByStatus(t.Context(), bakery.StatusQueued)
	require.NoError(t, err)
	require.Equal(t, 3, n)
}
