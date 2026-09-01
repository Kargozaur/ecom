package dbresp

import (
	"time"
	"uuid"
)

type FetchOrder struct {
	CreatedAt  time.Time `json:"created_at"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	TotalPrice float64   `json:"total_price"`
	OrderID    uuid.UUID `json:"order_id"`
}

type Item struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Orders struct {
	CreatedAt  time.Time
	Status     string
	TotalPrice float64
	OrderID    uuid.UUID
}

type CreateOrderResponse struct {
	ID         string
	Status     string
	TotalPrice float64
}
