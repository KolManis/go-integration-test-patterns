package bakery

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DB — то, что умеют и пул соединений, и транзакция. Благодаря этому репозиторий
// можно запустить и на пуле (как в приложении), и внутри транзакции теста.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct{ db DB }

func NewRepository(db DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, pastry string) (Order, error) {
	if pastry == "" {
		return Order{}, ErrEmptyPastry
	}
	o := Order{ID: uuid.New(), Pastry: pastry, Status: StatusQueued}
	err := r.db.QueryRow(ctx,
		`INSERT INTO orders (id, pastry, status) VALUES ($1, $2, $3) RETURNING created_at`,
		o.ID, o.Pastry, o.Status,
	).Scan(&o.CreatedAt)
	if err != nil {
		return Order{}, fmt.Errorf("create order: %w", err)
	}
	return o, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Order, error) {
	var o Order
	err := r.db.QueryRow(ctx,
		`SELECT id, pastry, status, created_at, baked_at FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.Pastry, &o.Status, &o.CreatedAt, &o.BakedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("get order: %w", err)
	}
	return o, nil
}

// CountByStatus — типичная «глобальная» проверка: сколько заказов в статусе во всей таблице.
func (r *Repository) CountByStatus(ctx context.Context, status Status) (int, error) {
	var n int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM orders WHERE status = $1`, status).Scan(&n); err != nil {
		return 0, fmt.Errorf("count orders: %w", err)
	}
	return n, nil
}
