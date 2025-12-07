package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/htrandev/gophermart/internal/handler"
)

type AuthorizationSuite struct {
	HandlerSuite
}

func (s *AuthorizationSuite) SetupSuite() {
	s.HandlerSuite.SetupSuite()
}

func TestAuthorization(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}

func (s *AuthorizationSuite) TestRegister() {
	url := "/api/user/register"

	token := "token"
	body := `{"login":"login","password":"password"}`
	user := application.AuthorizationRequest{Login: "login", Password: "password"}

	s.Run("valid", func() {
		s.service.EXPECT().Register(gomock.Any(), user).Return(token, nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
		s.Require().Equal(res.Header.Get("Authorization"), "Bearer "+token)
	})

	s.Run("invalid request", func() {
		body := `{"login":"login,"password":"password"}`

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("can't read body", func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, &handler.ErrorReader{})

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("empty login", func() {
		body := `{"login":"","password":"password"}`

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("not unique login", func() {
		s.service.EXPECT().Register(gomock.Any(), user).Return("", domain.ErrNotUniqueLogin).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusConflict, res.StatusCode)
	})

	s.Run("service error", func() {
		s.service.EXPECT().Register(gomock.Any(), user).Return("", errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Register)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}

func (s *AuthorizationSuite) TestLogin() {
	url := "/api/user/login"
	token := "token"
	body := `{"login":"login","password":"password"}`
	user := application.AuthorizationRequest{Login: "login", Password: "password"}

	s.Run("valid", func() {
		s.service.EXPECT().Login(gomock.Any(), user).Return(token, nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Login)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
		s.Require().Equal(res.Header.Get("Authorization"), "Bearer "+token)
	})

	s.Run("invalid request", func() {
		body := `{"login":"login,"password":"password"}`

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Login)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusBadRequest, res.StatusCode)
	})

	s.Run("incorrect password", func() {
		s.service.EXPECT().Login(gomock.Any(), user).Return("", domain.ErrIncorrectPassword).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Login)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusUnauthorized, res.StatusCode)
	})
	s.Run("service error", func() {
		s.service.EXPECT().Login(gomock.Any(), user).Return("", errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))

		mux := http.NewServeMux()
		mux.HandleFunc(url, s.handler.Login)
		mux.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		s.Require().Equal(http.StatusInternalServerError, res.StatusCode)
	})
}
