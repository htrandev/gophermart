package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/htrandev/gophermart/internal/handler"
)

type OrderSuite struct {
	HandlerSuite
}

func (s *OrderSuite) SetupSuite() {
	s.HandlerSuite.SetupSuite()
}

func TestOrder(t *testing.T) {
	suite.Run(t, new(OrderSuite))
}

func (s *OrderSuite) TestAddOrder() {
	url := "/api/user/orders"

	orderNumber := "12345678903"
	userId := uuid.New()

	order := domain.Order{
		Number: orderNumber,
		UserId: userId,
	}

	s.Run("valid", func() {
		s.service.EXPECT().AddOrder(gomock.Any(), order).Return(nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusAccepted, res.StatusCode)
	})

	s.Run("can't get user from context", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("unauthorized", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, uuid.Nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})

	s.Run("invalid request", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, &handler.ErrorReader{})

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("empty order number", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(""))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("invalid order number", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader("411111111111111"))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("order already add to user", func() {
		s.service.EXPECT().AddOrder(gomock.Any(), order).Return(domain.ErrOrderAlreadyAddToUser).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
	})

	s.Run("order created by another user", func() {
		s.service.EXPECT().AddOrder(gomock.Any(), order).Return(domain.ErrOrderCreatedByAnotherUser).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusConflict, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().AddOrder(gomock.Any(), order).Return(errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(orderNumber))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.AddOrder)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}

func (s *OrderSuite) TestGetOrder() {
	url := "/api/user/orders"

	userId := uuid.New()
	orders := domain.Orders{
		{Number: "order_1", Status: domain.OrderStatusNew, Accrual: 0, CreatedAt: time.Now()},
		{Number: "order_2", Status: domain.OrderStatusProcessed, Accrual: 100, CreatedAt: time.Now()},
	}

	s.Run("valid", func() {
		s.service.EXPECT().GetOrders(gomock.Any(), userId).Return(orders, nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetOrders)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
		s.Require().NotEmpty(rec.Body.String())
	})

	s.Run("can't get user from context", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetOrders)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("unauthorized", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, uuid.Nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetOrders)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})

	s.Run("order not found", func() {
		s.service.EXPECT().GetOrders(gomock.Any(), userId).Return(nil, domain.ErrNotFound).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetOrders)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusNoContent, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().GetOrders(gomock.Any(), userId).Return(nil, errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetOrders)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}
