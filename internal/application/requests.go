package application

type AuthorizationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ContextUserID определяет ключ для передачи userId через контекст.
type ContextUserID struct{}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
