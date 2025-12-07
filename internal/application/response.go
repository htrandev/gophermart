package application

import "time"

type Order struct {
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual"`
	CreatedAt time.Time `json:"uploaded_at"`
}

//easyjson:json
type OrdersResponse []Order
