package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus статус заказа.
type OrderStatus uint8

const (
	OrderStatusUnknown OrderStatus = iota

	OrderStatusNew
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

var orderStatusString = []string{
	"UNKNOWN",
	"NEW",
	"PROCESSING",
	"INVALID",
	"PROCESSED",
}

// ConvertStringToOrderStatus преобразует строку во внутренний статус заказа.
func ConvertStringToOrderStatus(value string) OrderStatus {
	switch value {
	case "NEW":
		return OrderStatusNew
	case "PROCESSING":
		return OrderStatusProcessing
	case "INVALID":
		return OrderStatusInvalid
	case "PROCESSED":
		return OrderStatusProcessed
	default:
		return OrderStatusUnknown
	}
}

// String преобразует внутренний статус в строку.
func (o OrderStatus) String() string {
	return orderStatusString[o]
}

// Order определяет заказ.
type Order struct {
	ID          uuid.UUID
	Number      string      `json:"number"`
	Status      OrderStatus `json:"status"`
	Accrual     float64     `json:"accrual"`
	CreatedAt   time.Time   `json:"uploaded_at"`
	ProcessedAt time.Time
	UserID      uuid.UUID
}

// Equal возвращает true, если заказ пользователя равен v.
func (o Order) IsEqual(v Order) bool {
	return o.ID == v.ID &&
		o.Number == v.Number &&
		o.Status == v.Status &&
		o.Accrual == v.Accrual &&
		o.CreatedAt.Equal(v.CreatedAt) &&
		o.UserID == v.UserID
}

// IsEmpty возвращает true, если заказ пользователя пуст.
func (o Order) IsEmpty() bool {
	return o.IsEqual(Order{})
}

//easyjson:json
type Orders []Order
