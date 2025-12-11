package application

import "time"

// Order определяет формат ответа для получения списка загруженных номеров заказов.
type Order struct {
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual"`
	CreatedAt time.Time `json:"uploaded_at"`
}

//easyjson:json
type OrdersResponse []Order
