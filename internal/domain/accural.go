package domain

// AccrualStatus статус расчета начислений в системе расчета баллов лояльности.
type AccrualStatus string

const (
	AccrualStatusUnknown    AccrualStatus = "UNKNOWN"
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

// String преобразует внутренний статус в строку.
func (a AccrualStatus) String() string {
	return string(a)
}

// Accrual определяет формат ответа от систумы расчета баллов лояльности.
type Accrual struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual float64       `json:"accrual"`
}
