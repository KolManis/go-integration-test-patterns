//go:build integration

// Подход 1: общая БД, очистка таблиц перед каждым тестом, запуск строго последовательно (-p=1).
package orders

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

type OrdersSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *bakery.Repository
}

func TestOrdersSuite(t *testing.T) { suite.Run(t, new(OrdersSuite)) }

func (s *OrdersSuite) SetupSuite() {
	s.pool = testdb.Connect(s.T(), testdb.DSN())
	testdb.Migrate(s.T(), s.pool)
	s.repo = bakery.NewRepository(s.pool)
}

func (s *OrdersSuite) SetupTest() { testdb.Truncate(s.T(), s.pool) }

func (s *OrdersSuite) TestCreateOrderIsQueued() {
	ctx := context.Background()
	created, err := s.repo.Create(ctx, "медовик")
	s.Require().NoError(err)

	got, err := s.repo.Get(ctx, created.ID)
	s.Require().NoError(err)
	s.Equal(bakery.StatusQueued, got.Status)
	s.Nil(got.BakedAt)
}

func (s *OrdersSuite) TestEmptyPastryRejected() {
	_, err := s.repo.Create(context.Background(), "")
	s.ErrorIs(err, bakery.ErrEmptyPastry)
}

func (s *OrdersSuite) TestUnknownOrderNotFound() {
	_, err := s.repo.Get(context.Background(), uuid.New())
	s.ErrorIs(err, bakery.ErrOrderNotFound)
}

// Глобальная проверка: работает только потому, что база перед тестом пустая.
func (s *OrdersSuite) TestQueueContainsExactlyNewOrders() {
	ctx := context.Background()
	for _, p := range []string{"медовик", "тирамису", "эклер"} {
		_, err := s.repo.Create(ctx, p)
		s.Require().NoError(err)
	}
	n, err := s.repo.CountByStatus(ctx, bakery.StatusQueued)
	s.Require().NoError(err)
	s.Equal(3, n)
}
