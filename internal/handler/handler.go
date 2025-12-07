package handler

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
)

// Serviceописывает интерфейс работы с сервисом.
//
//go:generate mockgen -source=handler.go -destination=mocks/mocks.go
type Service interface {
	// user
	Register(ctx context.Context, user application.AuthorizationRequest) (token string, err error)
	Login(ctx context.Context, user application.AuthorizationRequest) (token string, err error)

	// order
	AddOrder(ctx context.Context, order domain.Order) error
	GetOrders(ctx context.Context, userId uuid.UUID) ([]domain.Order, error)

	// balance
	GetBalance(ctx context.Context, userId uuid.UUID) (domain.Balance, error)
	Withdraw(ctx context.Context, withdraw domain.WithdrawRequest) (domain error)
	GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]domain.Withdraw, error)
}

type HandlerOptions struct {
	Logger  *zap.Logger
	Service Service
}

func defaultHandlerOptions() *HandlerOptions {
	return &HandlerOptions{
		Logger: zap.NewNop(),
	}
}

type Handler struct {
	opts *HandlerOptions
}

func New(opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = defaultHandlerOptions()
	}
	opts.Logger.With(zap.String("scope", "handler"))
	return &Handler{opts: opts}
}
