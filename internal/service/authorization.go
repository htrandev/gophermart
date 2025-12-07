package service

import (
	"context"
	"fmt"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
)

func (s *Service) Register(ctx context.Context, user application.AuthorizationRequest) (string, error) {
	hash, err := s.opts.Authorizer.HashPassword(user.Password)
	if err != nil {
		return "", fmt.Errorf("service: hash password: %w", err)
	}

	id, err := s.opts.Repository.Register(ctx, domain.User{
		Login:        user.Login,
		HashPassword: hash,
	})
	if err != nil {
		return "", fmt.Errorf("register new user: %w", err)
	}

	token, err := s.opts.Authorizer.Token(id)
	if err != nil {
		return "", fmt.Errorf("create new token: %w", err)
	}
	return token, nil
}

func (s *Service) Login(ctx context.Context, user application.AuthorizationRequest) (string, error) {
	u, err := s.opts.Repository.Login(ctx, user.Login)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if valid := s.opts.Authorizer.ValidatePassword(u.HashPassword, user.Password); !valid {
		return "", domain.ErrIncorrectPassword
	}

	token, err := s.opts.Authorizer.Token(u.ID.String())
	if err != nil {
		return "", fmt.Errorf("create new token: %w", err)
	}
	return token, nil
}
