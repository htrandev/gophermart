package domain

type AccrualStatus string

const (
	AccrualStatusUnknown    AccrualStatus = "UNKNOWN"
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

func (a AccrualStatus) String() string {
	return string(a)
}

type Accrual struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual float64       `json:"accrual"`
}
