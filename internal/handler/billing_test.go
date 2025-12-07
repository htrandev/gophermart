package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/htrandev/gophermart/internal/handler"
)

type BillingSuite struct {
	HandlerSuite
}

func (s *BillingSuite) SetupSuite() {
	s.HandlerSuite.SetupSuite()
}

func TestBilling(t *testing.T) {
	suite.Run(t, new(BillingSuite))
}

func (s *BillingSuite) TestGetBalance() {
	url := "/api/user/balance"
	userId := uuid.New()

	s.Run("valid", func() {
		s.service.EXPECT().GetBalance(gomock.Any(), userId).Return(domain.Balance{}, nil)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetBalance)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
		s.Require().NotEmpty(rec.Body.String())
	})

	s.Run("can't get user id from context", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetBalance)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("empty uid", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, uuid.Nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetBalance)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().GetBalance(gomock.Any(), userId).Return(domain.Balance{}, errTest)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.GetBalance)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}

func (s *BillingSuite) TestWithdraw() {
	url := "/api/user/balance/withdraw"

	userId := uuid.New()
	body := `{"order":"12345678903", "sum":123}`
	withdraw := domain.WithdrawRequest{Order: "12345678903", Sum: 123, UserId: userId}

	s.Run("valid", func() {
		s.service.EXPECT().Withdraw(gomock.Any(), withdraw).Return(nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
	})

	s.Run("can't get user id from context", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("empty uid", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, uuid.Nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})

	s.Run("can't read body", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, &handler.ErrorReader{})

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("invalid order number", func() {
		body := `{"order":"123", "sum":123}`

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	})

	s.Run("not enough points", func() {
		s.service.EXPECT().Withdraw(gomock.Any(), withdraw).Return(domain.ErrNotEnoughPoints).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusPaymentRequired, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().Withdraw(gomock.Any(), withdraw).Return(errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdraw)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}

func (s *BillingSuite) TestWithdrawals() {
	url := "/api/user/withdrawals"

	userId := uuid.New()

	s.Run("valid", func() {
		s.service.EXPECT().GetWithdrawals(gomock.Any(), userId).Return([]domain.Withdraw{}, nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdrawals)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
		s.Require().NotEmpty(rec.Body.String())
	})

	s.Run("can't get user id from context", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdrawals)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("empty uid", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, uuid.Nil)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdrawals)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})

	s.Run("no withdrawals", func() {
		s.service.EXPECT().GetWithdrawals(gomock.Any(), userId).Return([]domain.Withdraw{}, domain.ErrNotFound).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdrawals)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusNoContent, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().GetWithdrawals(gomock.Any(), userId).Return([]domain.Withdraw{}, errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		ctx := context.WithValue(context.Background(), application.ContextUserId{}, userId)

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Withdrawals)
		mux.ServeHTTP(rec, req.WithContext(ctx))

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}
