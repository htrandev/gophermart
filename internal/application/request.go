package application

// AuthorizationRequest определяет формат запроса для регистрации и авторизации пользователя.
type AuthorizationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ContextUserID определяет ключ для передачи userId через контекст.
type ContextUserID struct{}

// WithdrawRequest определяет формат запроса списание средств.
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
