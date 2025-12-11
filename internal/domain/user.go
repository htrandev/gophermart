package domain

import "github.com/google/uuid"

// User определяет пользователя.
type User struct {
	ID           uuid.UUID
	Login        string
	Password     string
	HashPassword string
	Balance      Balance
}

// Balance определяет баланс пользователя.
type Balance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
