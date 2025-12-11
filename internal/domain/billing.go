package domain

import (
	"time"

	"github.com/google/uuid"
)

// WithdrawRequest определяет формат запроса на списание баллов лояльности.
type WithdrawRequest struct {
	UserID uuid.UUID
	Order  string
	Sum    float64
}

// Withdraw определяет офрмат ответа для запроса на получение информации о выводе средств.
type Withdraw struct {
	Order     string    `json:"order"`
	Sum       float64   `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

//easyjson:json
type Withdrawals []Withdraw
