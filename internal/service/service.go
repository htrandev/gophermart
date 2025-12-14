package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/domain"
)

// Authorizer описывает интерфейс работы с авторизацией и аутентификацией.
//
//go:generate mockgen -source=service.go -destination=mocks/mocks.go
type Authorizer interface {
	HashPassword(password string) (hahs string, err error)
	ValidatePassword(hash, password string) bool
	Token(id string) (token string, err error)
	GetIDFromToken(token string) (string, error)
}

// Repository описывает интерфейс работы с базой данных.
//
//go:generate mockgen -source=service.go -destination=mocks/mocks.go
type Repository interface {
	Register(ctx context.Context, user domain.User) (id string, err error)
	Login(ctx context.Context, login string) (domain.User, error)

	CreateOrder(ctx context.Context, order domain.Order) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, order domain.Order) error

	GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error)
	Withdraw(ctx context.Context, withdraw domain.WithdrawRequest) error
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]domain.Withdraw, error)
}

// Client описывает интерфейс работы с сервисов по подсчету баллов.
//
//go:generate mockgen -source=service.go -destination=mocks/mocks.go
type Client interface {
	GetAccrual(ctx context.Context, number string) (domain.ClientResponse, error)
}

const defaultPauseDuration = 5 * time.Second

type ServiceOptions struct {
	Authorizer Authorizer
	Repository Repository
	Client     Client
	Logger     *zap.Logger

	NumWorkers int
}

func validateOptions(opts *ServiceOptions) *ServiceOptions {
	if opts.NumWorkers <= 0 {
		opts.NumWorkers = 3
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	opts.Logger.With(zap.String("scope", "service"))

	return opts
}

type Service struct {
	opts *ServiceOptions

	wg     sync.WaitGroup
	queue  chan domain.Order
	termCh chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	pauseUntil time.Time
	pauseMu    sync.RWMutex
}

func New(opts *ServiceOptions) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Service{
		opts:   validateOptions(opts),
		wg:     sync.WaitGroup{},
		queue:  make(chan domain.Order, 100),
		termCh: make(chan struct{}),

		ctx:    ctx,
		cancel: cancel,
	}
	return s
}

func (s *Service) Close() {
	if s.closed() {
		return
	}
	s.cancel()
	close(s.queue)
	close(s.termCh)
	s.wg.Wait()
}

func (s *Service) closed() bool {
	select {
	case <-s.termCh:
		return true
	default:
		return false
	}
}

func (s *Service) Run(ctx context.Context) {
	s.opts.Logger.Info("start workers")

	for i := 0; i < s.opts.NumWorkers; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}
}

func (s *Service) worker(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-s.queue:
			if !ok {
				return
			}
			order, err := s.processOrderAccrual(ctx, order)
			if err != nil {
				s.opts.Logger.Error("can't process order",
					zap.String("method", "processOrderAccrual"),
					zap.Error(err),
				)
			}
			if order.IsEmpty() {
				continue
			}
			s.enqueue(order)
		}
	}
}

func (s *Service) enqueue(order domain.Order) {
	go func() {
		select {
		case <-s.ctx.Done():
			s.opts.Logger.Error("enqueue cancelled",
				zap.String("number", order.Number),
				zap.Error(s.ctx.Err()),
			)
		case <-s.termCh:
			return
		case s.queue <- order:
		}
	}()
}

func (s *Service) pause(duration time.Duration) {
	if duration <= 0 {
		duration = defaultPauseDuration
	}
	s.pauseMu.Lock()
	s.pauseUntil = time.Now().Add(duration)
	s.pauseMu.Unlock()

	s.opts.Logger.Warn("pause activated",
		zap.Duration("duration", duration),
		zap.Time("until", s.pauseUntil),
	)
}
