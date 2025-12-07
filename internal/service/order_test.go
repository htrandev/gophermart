package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/htrandev/gophermart/internal/domain"
)

type OrderSuite struct {
	ServiceSuite
}

func (s *OrderSuite) SetupSuite() {
	s.ServiceSuite.SetupSuite()
}

func (s *OrderSuite) TearDownSuite() {
	s.service.Close()
}

func TestOrder(t *testing.T) {
	suite.Run(t, new(OrderSuite))
}

func (s *OrderSuite) TestAddOrder() {
	ctx := context.Background()

	id := uuid.New()
	order := domain.Order{
		Number: "test",
		UserID: id,
	}

	s.Run("valid", func() {
		s.repository.EXPECT().CreateOrder(gomock.Any(), order).Return(nil).Times(1)

		err := s.service.AddOrder(ctx, order)
		s.Require().NoError(err)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().CreateOrder(gomock.Any(), order).Return(errTest).Times(1)

		err := s.service.AddOrder(ctx, order)
		s.Require().ErrorIs(err, errTest)
	})
}

func (s *OrderSuite) TestGetOrders() {
	ctx := context.Background()

	id := uuid.New()

	orderID1 := uuid.New()
	orderID2 := uuid.New()
	orderID3 := uuid.New()
	dbOrders := []domain.Order{
		{ID: orderID1, Number: "order_1", Status: domain.OrderStatusNew, Accrual: 0, UserID: id},
		{ID: orderID2, Number: "order_2", Status: domain.OrderStatusProcessing, Accrual: 0, UserID: id},
		{ID: orderID3, Number: "order_3", Status: domain.OrderStatusProcessed, Accrual: 100, UserID: id},
	}

	s.Run("valid", func() {
		s.repository.EXPECT().GetOrders(gomock.Any(), id).Return(dbOrders, nil).Times(1)

		orders, err := s.service.GetOrders(ctx, id)
		s.Require().NoError(err)
		s.Require().Equal(dbOrders, orders)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().GetOrders(gomock.Any(), id).Return(nil, errTest).Times(1)

		_, err := s.service.GetOrders(ctx, id)
		s.Require().ErrorIs(err, errTest)
	})
}

func (s *OrderSuite) TestUpdateOrder() {
	ctx := context.Background()

	id := uuid.New()
	order := domain.Order{
		Number: "test",
		UserID: id,
		Status: domain.OrderStatusProcessing,
	}

	s.Run("valid", func() {
		s.repository.EXPECT().UpdateOrder(gomock.Any(), order).Return(nil).Times(1)

		o, err := s.service.updateOrder(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(order, o)
	})

	s.Run("order processed", func() {
		processedOrder := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessed,
		}
		s.repository.EXPECT().UpdateOrder(gomock.Any(), processedOrder).Return(nil).Times(1)

		order, err := s.service.updateOrder(ctx, processedOrder)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{}, order)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().UpdateOrder(gomock.Any(), order).Return(errTest).Times(1)

		o, err := s.service.updateOrder(ctx, order)
		s.Require().ErrorIs(err, errTest)
		s.Require().Equal(order, o)
	})
}

func (s *OrderSuite) TestProcessOrder() {
	ctx := context.Background()

	id := uuid.New()

	s.Run("order status new", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusNew,
		}

		accrual := domain.Accrual{
			Order:   order.Number,
			Status:  domain.AccrualStatusRegistered,
			Accrual: 10,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(accrual, nil).Times(1)
		s.repository.EXPECT().UpdateOrder(gomock.Any(), domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}).Return(nil).Times(1)

		o, err := s.service.processOrderAccrual(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}, o)
	})

	s.Run("order status new", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}

		accrual := domain.Accrual{
			Order:   order.Number,
			Status:  domain.AccrualStatusRegistered,
			Accrual: 10,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(accrual, nil).Times(1)

		o, err := s.service.processOrderAccrual(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}, o)
	})

	s.Run("accural status unknown", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}

		accrual := domain.Accrual{
			Order:  order.Number,
			Status: domain.AccrualStatusUnknown,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(accrual, nil).Times(1)

		o, err := s.service.processOrderAccrual(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}, o)
	})

	s.Run("accural status invalid", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}

		accrual := domain.Accrual{
			Order:  order.Number,
			Status: domain.AccrualStatusInvalid,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(accrual, nil).Times(1)
		s.repository.EXPECT().UpdateOrder(gomock.Any(), domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusInvalid,
		}).Return(nil).Times(1)

		o, err := s.service.processOrderAccrual(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusInvalid,
		}, o)
	})

	s.Run("accural status processed", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}

		accrual := domain.Accrual{
			Order:  order.Number,
			Status: domain.AccrualStatusProcessed,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(accrual, nil).Times(1)
		s.repository.EXPECT().UpdateOrder(gomock.Any(), domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessed,
		}).Return(nil).Times(1)

		o, err := s.service.processOrderAccrual(ctx, order)
		s.Require().NoError(err)
		s.Require().Equal(domain.Order{}, o)
	})

	s.Run("client error", func() {
		order := domain.Order{
			Number: "test",
			UserID: id,
			Status: domain.OrderStatusProcessing,
		}

		s.client.EXPECT().GetAccrual(gomock.Any(), order.Number).Return(domain.Accrual{}, errTest).Times(1)

		_, err := s.service.processOrderAccrual(ctx, order)
		s.Require().ErrorIs(err, errTest)
	})
}
