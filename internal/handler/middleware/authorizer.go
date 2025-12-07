package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/htrandev/gophermart/internal/application"
	"go.uber.org/zap"
)

// Authorizer реализует интерфейс авторизации пользователя.
//
//go:generate mockgen -source=authorizer.go -destination=mocks/mocks.go
type Authorizer interface {
	GetIDFromToken(token string) (string, error)
}

type Auth struct {
	logger *zap.Logger
	auth   Authorizer
}

func NewAuth(auth Authorizer, l *zap.Logger) *Auth {
	return &Auth{
		logger: l,
		auth:   auth,
	}
}

func (a *Auth) Authorize() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := getTokenFromHeader(r.Header.Get("Authorization"))

			id, err := a.auth.GetIDFromToken(token)
			if err != nil {
				a.logger.Error("cant get id from token", zap.Error(err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			uid, err := uuid.Parse(id)
			if err != nil {
				a.logger.Error("cant parse user id", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), application.ContextUserID{}, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getTokenFromHeader(header string) string {
	return strings.TrimPrefix(header, "Bearer ")
}
