package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/htrandev/gophermart/internal/domain"
)

type AuthorizationRepositorySuite struct {
	RepositorySuite
}

func TestAuthorization(t *testing.T) {
	suite.Run(t, new(AuthorizationRepositorySuite))
}

func (s *AuthorizationRepositorySuite) SetupSuite() {
	s.RepositorySuite.SetupSuite()
}

// Чистим базу после каждого теста, чтобы не было ошибки уникальности логина.
func (s *AuthorizationRepositorySuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.RepositorySuite.repository.Truncate(ctx)
}

func (s *AuthorizationRepositorySuite) TearDownSuite() {
	s.RepositorySuite.TearDownSuite()
}

func (s *AuthorizationRepositorySuite) TestRegister() {
	ctx := context.Background()

	user := domain.User{
		Login:        "test",
		HashPassword: "password",
	}

	s.Run("valid", func() {
		id, err := s.repository.Register(ctx, user)

		s.Require().NoError(err)
		s.Require().NotEmpty(id)
		s.Require().NotPanics(func() { uuid.MustParse(id) })
	})

	s.Run("duplicate", func() {
		_, err := s.repository.Register(ctx, user)
		s.Require().ErrorIs(err, domain.ErrNotUniqueLogin)
	})

	s.Run("context canceled", func() {
		newCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.repository.Register(newCtx, user)
		s.Require().ErrorIs(err, context.Canceled)
	})
}

func (s *AuthorizationRepositorySuite) TestLogin() {
	ctx := context.Background()

	user := domain.User{
		Login:        "test",
		HashPassword: "password",
	}

	s.Run("vallid", func() {
		_, err := s.repository.Register(ctx, user)
		s.Require().NoError(err)

		u, err := s.repository.Login(ctx, user.Login)
		s.Require().NoError(err)
		s.Require().Equal(user.Login, u.Login)
		s.Require().Equal(user.HashPassword, u.HashPassword)
	})

	s.Run("unknown user", func() {
		_, err := s.repository.Login(ctx, "unknown login")
		s.Require().Error(err)
	})
}
