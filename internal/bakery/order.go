// Package bakery — учебный сервис: заказы выпечки и духовка, которая их печёт.
package bakery

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued Status = "queued" // заказ ждёт духовку
	StatusBaked  Status = "baked"  // испечён
	StatusBurnt  Status = "burnt"  // сгорел
)

// BurntPastry всегда сгорает — нужен, чтобы у духовки был второй исход.
const BurntPastry = "уголёк"

var (
	ErrEmptyPastry   = errors.New("pastry name is empty")
	ErrOrderNotFound = errors.New("order not found")
)

type Order struct {
	ID        uuid.UUID
	Pastry    string
	Status    Status
	CreatedAt time.Time
	BakedAt   *time.Time
}
