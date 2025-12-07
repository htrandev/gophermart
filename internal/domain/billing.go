package domain

import (
	"time"

	"github.com/google/uuid"
)

type WithdrawRequest struct {
	UserID uuid.UUID
	Order  string
	Sum    float64
}

type Withdraw struct {
	Order     string    `json:"order"`
	Sum       float64   `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

//easyjson:json
type Withdrawals []Withdraw
