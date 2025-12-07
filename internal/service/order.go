package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/htrandev/gophermart/internal/domain"
)

func (s *Service) AddOrder(ctx context.Context, order domain.Order) error {
	if err := s.opts.Repository.CreateOrder(ctx, order); err != nil {
		return fmt.Errorf("service: create order: %w", err)
	}

	s.enqueue(order)
	return nil
}

func (s *Service) GetOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	orders, err := s.opts.Repository.GetOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: create order: %w", err)
	}
	return orders, nil
}

func (s *Service) updateOrder(ctx context.Context, order domain.Order) (domain.Order, error) {
	if err := s.opts.Repository.UpdateOrder(ctx, order); err != nil {
		return order, fmt.Errorf("update order: %w", err)
	}
	if order.Status == domain.OrderStatusProcessed {
		return domain.Order{}, nil
	}
	return order, nil
}

func (s *Service) processOrderAccrual(ctx context.Context, order domain.Order) (domain.Order, error) {
	accrual, err := s.opts.Client.GetAccrual(ctx, order.Number)
	if err != nil {
		return order, fmt.Errorf("get accrual from client: %w", err)
	}

	s.opts.Logger.Debug("get accrual", zap.Any("", accrual))

	switch accrual.Status {
	case domain.AccrualStatusUnknown:
		return order, nil
	case domain.AccrualStatusInvalid:
		order.Status = domain.OrderStatusInvalid
		return s.updateOrder(ctx, order)
	case domain.AccrualStatusRegistered, domain.AccrualStatusProcessing:
		if order.Status == domain.OrderStatusNew {
			order.Status = domain.OrderStatusProcessing
			return s.updateOrder(ctx, order)
		}
		return order, nil
	case domain.AccrualStatusProcessed:
		order.Status = domain.OrderStatusProcessed
		order.Accrual = float64(accrual.Accrual)
		return s.updateOrder(ctx, order)
	}

	return domain.Order{}, nil
}
