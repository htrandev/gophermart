package handler_test

import (
	"errors"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/handler"
	mock_handler "github.com/htrandev/gophermart/internal/handler/mocks"
	"github.com/htrandev/gophermart/pkg/logger"
)

var errTest = errors.New("test error")

type HandlerSuite struct {
	suite.Suite

	logger  *zap.Logger
	service *mock_handler.MockService
	handler *handler.Handler
}

func (s *HandlerSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())

	var err error
	s.logger, err = logger.NewZapLogger("debug")
	s.Require().NoError(err)

	s.service = mock_handler.NewMockService(ctrl)

	s.handler = handler.New(&handler.HandlerOptions{
		Logger:  s.logger,
		Service: s.service,
	})
}
