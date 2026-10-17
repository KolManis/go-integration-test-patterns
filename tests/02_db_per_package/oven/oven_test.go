//go:build integration

// Подход 2: своя БД из шаблона на пакет — пакеты можно запускать параллельно.
package oven

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"

	"github.com/KolManis/go-integration-test-patterns/internal/bakery"
	"github.com/KolManis/go-integration-test-patterns/internal/testdb"
)

type OvenSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *bakery.Repository
	oven *bakery.Oven
}

func TestOvenSuite(t *testing.T) { suite.Run(t, new(OvenSuite)) }

func (s *OvenSuite) SetupSuite() {
	s.pool = testdb.NewDatabase(s.T())
	s.repo = bakery.NewRepository(s.pool)
	s.oven = bakery.NewOven(s.pool, testdb.BakeTime)
}

func (s *OvenSuite) SetupTest() { testdb.Truncate(s.T(), s.pool) }

func (s *OvenSuite) TestBakesQueuedOrder() {
	ctx := context.Background()
	order, err := s.repo.Create(ctx, "медовик")
	s.Require().NoError(err)

	n, err := s.oven.BakeQueued(ctx)
	s.Require().NoError(err)
	s.Equal(1, n, "духовка должна взять ровно один заказ — в базе только наш")

	got, err := s.repo.Get(ctx, order.ID)
	s.Require().NoError(err)
	s.Equal(bakery.StatusBaked, got.Status)
	s.NotNil(got.BakedAt)
}

func (s *OvenSuite) TestBurntPastry() {
	ctx := context.Background()
	order, err := s.repo.Create(ctx, bakery.BurntPastry)
	s.Require().NoError(err)

	_, err = s.oven.BakeQueued(ctx)
	s.Require().NoError(err)

	got, err := s.repo.Get(ctx, order.ID)
	s.Require().NoError(err)
	s.Equal(bakery.StatusBurnt, got.Status)
}

func (s *OvenSuite) TestEmptyQueue() {
	n, err := s.oven.BakeQueued(context.Background())
	s.Require().NoError(err)
	s.Zero(n, "в пустой базе печь нечего")
}

func (s *OvenSuite) TestBakesWholeQueue() {
	ctx := context.Background()
	for _, p := range []string{"медовик", "тирамису", bakery.BurntPastry} {
		_, err := s.repo.Create(ctx, p)
		s.Require().NoError(err)
	}
	n, err := s.oven.BakeQueued(ctx)
	s.Require().NoError(err)
	s.Equal(3, n)

	baked, err := s.repo.CountByStatus(ctx, bakery.StatusBaked)
	s.Require().NoError(err)
	s.Equal(2, baked)
}
