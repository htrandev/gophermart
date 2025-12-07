package domain

import (
	"time"

	"github.com/google/uuid"
)

type WithdrawRequest struct {
	UserID uuid.UUID
	Order  string
	Sum    int
}

type Withdraw struct {
	Order     string    `json:"order"`
	Sum       int       `json:"sum"`
	CreatedAt time.Time `json:"processed_at"`
}

//easyjson:json
type Withdrawals []Withdraw
