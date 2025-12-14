package handler

import (
	"bytes"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/domain"
)

func (h *Handler) Register(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "register")

	var form AuthorizationForm
	if err := form.BuildAndValidate(r); err != nil {
		h.opts.Logger.Error("validate request", zap.Error(err), method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.opts.Service.Register(ctx, form.User)
	if err != nil {
		if errors.Is(err, domain.ErrNotUniqueLogin) {
			h.opts.Logger.Error("login is not unique",
				zap.String("login", form.User.Login),
				zap.Error(err),
				method,
			)
			rw.WriteHeader(http.StatusConflict)
			return
		}
		h.opts.Logger.Error("new user",
			zap.String("login", form.User.Login),
			zap.Error(err),
			method,
		)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Authorization", buildAuthorizationToken(token))
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "login")

	var form AuthorizationForm
	if err := form.BuildAndValidate(r); err != nil {
		h.opts.Logger.Error("validate request", zap.Error(err), method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.opts.Service.Login(ctx, form.User)
	if err != nil {
		if errors.Is(err, domain.ErrIncorrectPassword) {
			h.opts.Logger.Error("password is incorrect",
				zap.String("login", form.User.Login),
				zap.Error(err),
				method,
			)
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.opts.Logger.Error("login user",
			zap.String("login", form.User.Login),
			zap.Error(err),
			method,
		)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Authorization", buildAuthorizationToken(token))
	rw.WriteHeader(http.StatusOK)
}

func buildAuthorizationToken(token string) string {
	authorizationType := "Bearer"

	var buf bytes.Buffer

	buf.Grow(len(authorizationType) + len(token) + 1)
	buf.WriteString(authorizationType)
	buf.WriteString(" ")
	buf.WriteString(token)

	return buf.String()
}
