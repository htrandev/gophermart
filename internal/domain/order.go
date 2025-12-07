package domain

import (
	"time"

	"github.com/google/uuid"
)

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

func (o OrderStatus) String() string {
	return orderStatusString[o]
}

type Order struct {
	ID          uuid.UUID
	Number      string      `json:"number"`
	Status      OrderStatus `json:"status"`
	Accrual     float64     `json:"accrual"`
	CreatedAt   time.Time   `json:"uploaded_at"`
	ProcessedAt time.Time
	UserID      uuid.UUID
}

func (o Order) IsEqual(v Order) bool {
	return o.ID == v.ID &&
		o.Number == v.Number &&
		o.Status == v.Status &&
		o.Accrual == v.Accrual &&
		o.CreatedAt.Equal(v.CreatedAt) &&
		o.UserID == v.UserID
}

func (o Order) IsEmpty() bool {
	return o.IsEqual(Order{})
}

//easyjson:json
type Orders []Order
