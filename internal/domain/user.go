package domain

import "github.com/google/uuid"

type User struct {
	Id           uuid.UUID
	Login        string
	Password     string
	HashPassword string
	Balance      Balance
}

type Balance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
