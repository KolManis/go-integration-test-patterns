//go:build integration

// Подход 5: те же тесты духовки, что в подходе 4 (транзакция на тест), но на фреймворке Testo.
// Testo не изолирует данные сам — изоляция по-прежнему от транзакции. Он даёт сьют с хуками,
// где параллельность включается один раз в BeforeEach, а не строкой в каждом тесте.
package cookies

import (
	"os"
	"testing"

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

func TestOven(t *testing.T) {
	testo.RunSuite(t, new(Suite))
}

// BeforeEach делает каждый тест сьюта параллельным.
func (*Suite) BeforeEach(t T) {
	t.Parallel()
}

// Состояние теста не храним в полях сьюта: тесты параллельные и делят один экземпляр Suite.
func setup(t T) (*bakery.Repository, *bakery.Oven) {
	tx := testdb.Tx(t, pool)
	return bakery.NewRepository(tx), bakery.NewOven(tx, testdb.BakeTime)
}

func (*Suite) TestBakesQueuedOrder(t T) {
	repo, oven := setup(t)
	order, err := repo.Create(t.Context(), "медовик")
	require.NoError(t, err)

	n, err := oven.BakeQueued(t.Context())
	require.NoError(t, err)
	require.EqualValues(t, 1, n)

	got, err := repo.Get(t.Context(), order.ID)
	require.NoError(t, err)
	require.Equal(t, bakery.StatusBaked, got.Status)
}

func (*Suite) TestEmptyQueue(t T) {
	_, oven := setup(t)
	n, err := oven.BakeQueued(t.Context())
	require.NoError(t, err)
	require.EqualValues(t, 0, n)
}

// CasesPastry — параметры для TestBakeResult: Testo запустит тест по разу на каждую выпечку.
func (*Suite) CasesPastry() []string {
	return []string{"медовик", "тирамису", "эклер", bakery.BurntPastry}
}

func (*Suite) TestBakeResult(t T, p struct{ Pastry string }) {
	repo, oven := setup(t)
	order, err := repo.Create(t.Context(), p.Pastry)
	require.NoError(t, err)

	_, err = oven.BakeQueued(t.Context())
	require.NoError(t, err)

	want := bakery.StatusBaked
	if p.Pastry == bakery.BurntPastry {
		want = bakery.StatusBurnt
	}
	got, err := repo.Get(t.Context(), order.ID)
	require.NoError(t, err)
	require.Equal(t, want, got.Status)
}
