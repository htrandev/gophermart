package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/handler"
	"github.com/htrandev/gophermart/internal/handler/middleware"
)

type Authorizer interface {
	GetIDFromToken(token string) (string, error)
}

func New(auth Authorizer, handler *handler.Handler, l *zap.Logger) (*chi.Mux, error) {

	authMiddleware := middleware.NewAuth(auth, l)
	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		// авторизация
		r.With(
			middleware.MethodChecker(http.MethodPost),
			// middleware.Compress(),
		).Post("/register", handler.Register)
		r.With(
			middleware.MethodChecker(http.MethodPost),
			// middleware.Compress(),
		).Post("/login", handler.Login)

		// заказы
		r.With(
			middleware.MethodChecker(http.MethodPost),
			authMiddleware.Authorize(),
			// middleware.Compress(),
		).Post("/orders", handler.AddOrder)
		r.With(
			middleware.MethodChecker(http.MethodGet),
			authMiddleware.Authorize(),
			// middleware.Compress(),
		).Get("/orders", handler.GetOrders)

		// баланс
		r.With(
			middleware.MethodChecker(http.MethodGet),
			authMiddleware.Authorize(),
			// middleware.Compress(),
		).Get("/balance", handler.GetBalance)
		r.With(
			middleware.MethodChecker(http.MethodPost),
			authMiddleware.Authorize(),
			// middleware.Compress(),
		).Post("/balance/withdraw", handler.Withdraw)
		r.With(
			middleware.MethodChecker(http.MethodGet),
			authMiddleware.Authorize(),
			// middleware.Compress(),
		).Get("/withdrawals", handler.Withdrawals)
	})
	return r, nil
}
