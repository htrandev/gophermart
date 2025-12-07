package service

import (
	"errors"

	mock_service "github.com/htrandev/gophermart/internal/service/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var errTest = errors.New("test error")

type ServiceSuite struct {
	suite.Suite

	ctrl *gomock.Controller

	authorizer *mock_service.MockAuthorizer
	repository *mock_service.MockRepository
	client     *mock_service.MockClient

	service *Service
}

func (s *ServiceSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())

	s.authorizer = mock_service.NewMockAuthorizer(s.ctrl)
	s.repository = mock_service.NewMockRepository(s.ctrl)
	s.client = mock_service.NewMockClient(s.ctrl)

	s.service = New(&ServiceOptions{
		Authorizer: s.authorizer,
		Repository: s.repository,
		Client:     s.client,
	})
}
