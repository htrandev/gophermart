package handler

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
)

func (h *Handler) GetBalance(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "GetBalance")

	uid, ok := ctx.Value(application.ContextUserID{}).(uuid.UUID)
	if !ok {
		h.opts.Logger.Error("get user id from request", method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if uid == uuid.Nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.opts.Service.GetBalance(ctx, uid)
	if err != nil {
		h.opts.Logger.Error("get balance", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := easyjson.Marshal(balance)
	if err != nil {
		h.opts.Logger.Error("marshal response", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}

func (h *Handler) Withdraw(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "withdraw")

	uid, ok := ctx.Value(application.ContextUserID{}).(uuid.UUID)
	if !ok {
		h.opts.Logger.Error("get user id from request", method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if uid == uuid.Nil {
		h.opts.Logger.Error("uid is nil", method)
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	var f WithdrawForm
	if err := f.BuildAndValidate(r, uid); err != nil {
		var orderNumErr *domain.InvalidOrderNumberError
		if errors.As(err, &orderNumErr) {
			h.opts.Logger.Error("invalid order number", zap.Error(err), method)
			rw.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		h.opts.Logger.Error("validate request", method, zap.Error(err))
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.opts.Service.Withdraw(ctx, f.Req); err != nil {
		if errors.Is(err, domain.ErrNotEnoughPoints) {
			h.opts.Logger.Error("not enought points", zap.Int("needs", f.Req.Sum), method, zap.Error(err))
			rw.WriteHeader(http.StatusPaymentRequired)
			return
		}

		h.opts.Logger.Error("withdraw", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Withdrawals(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "GetOrders")

	uid, ok := ctx.Value(application.ContextUserID{}).(uuid.UUID)
	if !ok {
		h.opts.Logger.Error("get user id from request", method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if uid == uuid.Nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.opts.Service.GetWithdrawals(ctx, uid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.opts.Logger.Error("no withdrawals", zap.String("user", uid.String()), method)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.opts.Logger.Error("get withdrawals", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := buildWithdrawalsResponse(withdrawals)
	if err != nil {
		h.opts.Logger.Error("build response", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}

func buildWithdrawalsResponse(withdrawals domain.Withdrawals) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(withdrawals)
	if err != nil {
		return nil, fmt.Errorf("buildManyBody: can't marshal metrics: %w", err)
	}
	_, err = gz.Write(p)
	if err != nil {
		return nil, fmt.Errorf("buildManyBody: can't write: %w", err)
	}

	gz.Close()
	return buf.Bytes(), nil
}
