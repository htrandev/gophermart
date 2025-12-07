package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotUniqueLogin            = errors.New("login is not unique")
	ErrIncorrectPassword         = errors.New("incorrect password")
	ErrOrderAlreadyAddToUser     = errors.New("order has been already add to this user")
	ErrOrderCreatedByAnotherUser = errors.New("order has been added by another user")
	ErrNotFound                  = errors.New("not found")
	ErrNotEnoughPoints           = errors.New("user doesn't have enough points")
)

type InvalidOrderNumberError struct {
	Number string
}

func (e *InvalidOrderNumberError) Error() string {
	return fmt.Sprintf("invalid order number: %s", e.Number)
}

func NewInvalidOrderError(number string) error {
	return &InvalidOrderNumberError{
		Number: number,
	}
}
