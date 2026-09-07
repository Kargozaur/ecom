package orderstructs

type Item struct {
	ItemID   string  `json:"item_id"`
	Name     string  `json:"name"`
	Price    float32 `json:"price"`
	Quantity int     `json:"quantity"`
}

type OrderResponse struct {
	OrderID   string  `json:"order_id"`
	Items     []Item  `json:"items"`
	Total     float32 `json:"total"`
	CreatedAt string  `json:"created_at"`
}

type OrdersResponse struct {
	Orders []Order `json:"orders"`
}

type Order struct {
	OrderID   string  `json:"order_id"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	Total     float32 `json:"total"`
}
