package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/htrandev/gophermart/internal/domain"
)

func (s *Service) GetBalance(ctx context.Context, userId uuid.UUID) (domain.Balance, error) {
	balance, err := s.opts.Repository.GetBalance(ctx, userId)
	if err != nil {
		return domain.Balance{}, fmt.Errorf("service: get balance: %w", err)
	}

	return balance, nil
}

func (s *Service) Withdraw(ctx context.Context, withdraw domain.WithdrawRequest) error {
	err := s.opts.Repository.Withdraw(ctx, withdraw)
	if err != nil {
		return fmt.Errorf("service: withdraw: %w", err)
	}
	return nil
}

func (s *Service) GetWithdrawals(ctx context.Context, userId uuid.UUID) ([]domain.Withdraw, error) {
	withdrawals, err := s.opts.Repository.GetWithdrawals(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("service: get balance: %w", err)
	}

	return withdrawals, nil
}
