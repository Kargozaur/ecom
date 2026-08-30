package dbresp

import (
	"time"
	"uuid"
)

type FetchOrder struct {
	OrderID    uuid.UUID `json:"order_id"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Items      []Item    `json:"items"`
}

type Item struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Orders struct {
	OrderID    uuid.UUID
	TotalPrice float64
	Status     string
	CreatedAt  time.Time
}

type CreateOrderResponse struct {
	ID         string
	TotalPrice float64
	Status     string
}
