package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/htrandev/gophermart/internal/domain"
	mock_service "github.com/htrandev/gophermart/internal/service/mocks"
)

type BillingSuite struct {
	suite.Suite

	ctrl       *gomock.Controller
	repository *mock_service.MockRepository
	service    *Service
}

func (s *BillingSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.repository = mock_service.NewMockRepository(s.ctrl)

	s.service = New(&ServiceOptions{
		Repository: s.repository,
	})
}

func (s *BillingSuite) TearDownSuite() {
	s.service.Close()
}

func TestBilling(t *testing.T) {
	suite.Run(t, new(BillingSuite))
}

func (s *BillingSuite) TestGetBalance() {
	ctx := context.Background()

	id := uuid.New()

	dbBalance := domain.Balance{
		Balance:   1000,
		Withdrawn: 52,
	}

	s.Run("valid", func() {
		s.repository.EXPECT().GetBalance(gomock.Any(), id).Return(dbBalance, nil).Times(1)

		b, err := s.service.GetBalance(ctx, id)
		s.Require().NoError(err)
		s.Require().Equal(dbBalance, b)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().GetBalance(gomock.Any(), id).Return(domain.Balance{}, errTest).Times(1)

		_, err := s.service.GetBalance(ctx, id)
		s.Require().ErrorIs(err, errTest)
	})
}

func (s *BillingSuite) TestWithdraw() {
	ctx := context.Background()

	withdraw := domain.WithdrawRequest{UserID: uuid.New(), Order: "test_1", Sum: 3}

	s.Run("valid", func() {
		s.repository.EXPECT().Withdraw(gomock.Any(), withdraw).Return(nil).Times(1)

		err := s.service.Withdraw(ctx, withdraw)
		s.Require().NoError(err)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().Withdraw(gomock.Any(), withdraw).Return(errTest).Times(1)

		err := s.service.Withdraw(ctx, withdraw)
		s.Require().ErrorIs(err, errTest)
	})
}

func (s *BillingSuite) TestGetWithdrawals() {
	ctx := context.Background()

	userID := uuid.New()
	withdraws := []domain.Withdraw{
		{Order: "test_1", Sum: 1},
		{Order: "test_2", Sum: 2},
		{Order: "test_3", Sum: 3},
	}

	s.Run("valid", func() {
		s.repository.EXPECT().GetWithdrawals(gomock.Any(), userID).Return(withdraws, nil).Times(1)

		w, err := s.service.GetWithdrawals(ctx, userID)
		s.Require().NoError(err)
		s.Require().Equal(withdraws, w)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().GetWithdrawals(gomock.Any(), userID).Return(nil, errTest).Times(1)

		w, err := s.service.GetWithdrawals(ctx, userID)
		s.Require().ErrorIs(err, errTest)
		s.Require().Nil(w)
	})
}
