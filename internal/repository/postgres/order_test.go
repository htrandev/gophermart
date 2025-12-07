package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/stretchr/testify/suite"
)

type OrderSuite struct {
	RepositorySuite
}

func (s *OrderSuite) SetupSuite() {
	s.RepositorySuite.SetupSuite()
}

// Чистим базу после каждого теста, чтобы не было ошибки уникальности логина.
func (s *OrderSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.RepositorySuite.repository.Truncate(ctx)
}

func TestOrder(t *testing.T) {
	suite.Run(t, new(OrderSuite))
}

func (s *OrderSuite) TearDownSuite() {
	s.RepositorySuite.TearDownSuite()
}

func (s *OrderSuite) TestCreateOrder() {
	ctx := context.Background()

	user := domain.User{
		Login:    "test",
		Password: "password",
	}
	number := "test"

	id, err := s.repository.Register(ctx, user)
	s.Require().NoError(err)
	userId, err := uuid.Parse(id)
	s.Require().NoError(err)

	s.Run("valid", func() {
		err = s.repository.CreateOrder(ctx, domain.Order{Number: number, UserId: userId})
		s.Require().NoError(err)
	})

	s.Run("order already exists", func() {
		err = s.repository.CreateOrder(ctx, domain.Order{Number: number, UserId: userId})
		s.Require().ErrorIs(err, domain.ErrOrderAlreadyAddToUser)
	})

	s.Run("order add to another user", func() {
		err = s.repository.CreateOrder(ctx, domain.Order{Number: number, UserId: uuid.New()})
		s.Require().ErrorIs(err, domain.ErrOrderCreatedByAnotherUser)
	})
}

func (s *OrderSuite) TestGetOrders() {
	ctx := context.Background()

	user := domain.User{
		Login:    "test",
		Password: "password",
	}

	id, err := s.repository.Register(ctx, user)
	s.Require().NoError(err)
	userId, err := uuid.Parse(id)
	s.Require().NoError(err)

	err = s.repository.CreateOrder(ctx, domain.Order{Number: "number_1", UserId: userId})
	s.Require().NoError(err)
	err = s.repository.CreateOrder(ctx, domain.Order{Number: "number_2", UserId: userId})
	s.Require().NoError(err)
	err = s.repository.CreateOrder(ctx, domain.Order{Number: "number_3", UserId: userId})
	s.Require().NoError(err)

	s.Run("valid", func() {
		o, err := s.repository.GetOrders(ctx, userId)
		s.Require().NoError(err)
		s.Require().Len(o, 3)

		s.Require().Equal("number_3", o[0].Number)
		s.Require().Equal(userId, o[0].UserId)
		s.Require().Equal("number_2", o[1].Number)
		s.Require().Equal(userId, o[1].UserId)
		s.Require().Equal("number_1", o[2].Number)
		s.Require().Equal(userId, o[2].UserId)
	})

	s.Run("no orders", func() {
		o, err := s.repository.GetOrders(ctx, uuid.New())
		s.Require().ErrorIs(err, domain.ErrNotFound)
		s.Require().Nil(o)
	})
}

func (s *OrderSuite) TestUpdateOrder() {
	ctx := context.Background()

	user := domain.User{
		Login:    "test",
		Password: "password",
	}

	id, err := s.repository.Register(ctx, user)
	s.Require().NoError(err)
	userId, err := uuid.Parse(id)
	s.Require().NoError(err)

	s.Run("order processed", func() {

		order := domain.Order{Number: "number_1", UserId: userId}

		err = s.repository.CreateOrder(ctx, order)
		s.Require().NoError(err)

		order.Status = domain.OrderStatusProcessed
		err = s.repository.UpdateOrder(ctx, order)
		s.Require().NoError(err)
	})

	s.Run("order processing", func() {
		order := domain.Order{Number: "number_2", UserId: userId}

		err = s.repository.CreateOrder(ctx, order)
		s.Require().NoError(err)

		order.Status = domain.OrderStatusProcessing
		err = s.repository.UpdateOrder(ctx, order)
		s.Require().NoError(err)
	})
}
