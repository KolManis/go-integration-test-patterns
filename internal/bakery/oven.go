package bakery

import (
	"context"
	"fmt"
	"time"
)

// Oven — фоновый воркер. Как многие воркеры в реальных сервисах, за один проход он берёт
// ВСЕ заказы в очереди, а не конкретные: именно это и мешает параллельным тестам на общей БД.
type Oven struct {
	db       DB
	bakeTime time.Duration // имитация ожидания (сеть, внешняя система): на это уходит основное время тестов
}

func NewOven(db DB, bakeTime time.Duration) *Oven { return &Oven{db: db, bakeTime: bakeTime} }

// BakeQueued печёт все заказы в очереди и возвращает, сколько обработал.
func (o *Oven) BakeQueued(ctx context.Context) (int, error) {
	select {
	case <-time.After(o.bakeTime):
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	tag, err := o.db.Exec(ctx, `
		UPDATE orders
		SET status   = CASE WHEN pastry = $1 THEN $2 ELSE $3 END,
		    baked_at = now()
		WHERE status = $4`,
		BurntPastry, StatusBurnt, StatusBaked, StatusQueued,
	)
	if err != nil {
		return 0, fmt.Errorf("bake queued orders: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
