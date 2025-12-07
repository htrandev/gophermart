package application

type AuthorizationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ContextUserId определяет ключ для передачи userId через контекст.
type ContextUserId struct{}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
