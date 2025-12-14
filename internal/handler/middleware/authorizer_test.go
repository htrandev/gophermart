package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	mock_middleware "github.com/htrandev/gophermart/internal/handler/middleware/mocks"
	"github.com/htrandev/gophermart/pkg/logger"
)

type AutorizerSuite struct {
	suite.Suite

	authorizer *mock_middleware.MockAuthorizer
	logger     *zap.Logger
	auth       Auth
}

func (s *AutorizerSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())

	var err error
	s.logger, err = logger.NewZapLogger("debug")
	s.Require().NoError(err)

	s.authorizer = mock_middleware.NewMockAuthorizer(ctrl)
	s.auth = *NewAuth(s.authorizer, s.logger)
}

func TestAuthorizer(t *testing.T) {
	suite.Run(t, new(AutorizerSuite))
}

func (s *AutorizerSuite) TestAuthorize() {
	token := "token"
	userID := uuid.New()

	s.Run("valid", func() {
		s.authorizer.EXPECT().GetIDFromToken(token).Return(userID.String(), nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		dummyHandler := dummyHandler()
		wrapper := s.auth.Authorize()
		wrapper(dummyHandler).ServeHTTP(rec, req)
	})

	s.Run("authorizer error", func() {
		s.authorizer.EXPECT().GetIDFromToken(token).Return("", errTest).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		dummyHandler := dummyHandler()
		wrapper := s.auth.Authorize()
		wrapper(dummyHandler).ServeHTTP(rec, req)

		s.Require().Equal(http.StatusUnauthorized, rec.Code)
	})

	s.Run("invalid user id", func() {
		s.authorizer.EXPECT().GetIDFromToken(token).Return("test", nil).Times(1)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		dummyHandler := dummyHandler()
		wrapper := s.auth.Authorize()
		wrapper(dummyHandler).ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
