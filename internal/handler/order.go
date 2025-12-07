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

func (h *Handler) AddOrder(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := zap.String("method", "AddOrder")

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

	var f AddOrderForm
	if err := f.BuildAndValidate(r); err != nil {
		var orderNumErr *domain.InvalidOrderNumberError
		if errors.As(err, &orderNumErr) {
			h.opts.Logger.Error("invalid order number", zap.Error(err), method)
			rw.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		h.opts.Logger.Error("invalid request", zap.Error(err), method)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	f.Req.UserID = uid
	err := h.opts.Service.AddOrder(ctx, f.Req)
	if err != nil {
		if errors.Is(err, domain.ErrOrderAlreadyAddToUser) {
			h.opts.Logger.Error("order has been already added to user", method)
			rw.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, domain.ErrOrderCreatedByAnotherUser) {
			h.opts.Logger.Error("order has been already added by another user", method)
			rw.WriteHeader(http.StatusConflict)
			return
		}

		h.opts.Logger.Error("add order", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(rw http.ResponseWriter, r *http.Request) {
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

	orders, err := h.opts.Service.GetOrders(ctx, uid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.opts.Logger.Error("no orders", zap.String("user", uid.String()), method)
			rw.WriteHeader(http.StatusNoContent)
			return
		}

		h.opts.Logger.Error("get orders", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := buildOrdersResponse(orders)
	if err != nil {
		h.opts.Logger.Error("build response", method, zap.Error(err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}

func buildOrdersResponse(orders domain.Orders) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	p, err := easyjson.Marshal(orders)
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
